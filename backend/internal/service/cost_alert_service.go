package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type CostAlertService struct {
	httpClient *http.Client
}

func NewCostAlertService() *CostAlertService {
	return &CostAlertService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type DingTalkMessage struct {
	MsgType string                 `json:"msgtype"`
	Text    *DingTalkTextContent   `json:"text,omitempty"`
	Markdown *DingTalkMarkdownContent `json:"markdown,omitempty"`
}

type DingTalkTextContent struct {
	Content string `json:"content"`
}

type DingTalkMarkdownContent struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

func (s *CostAlertService) DetectAndAlert(storeID uint, ingredients []model.Ingredient, batchNo string) {
	threshold := decimal.NewFromFloat(config.AppConfig.CostAlert.PriceChangeThreshold)
	if threshold.LessThanOrEqual(decimal.Zero) {
		threshold = decimal.NewFromInt(20)
	}

	var alertIngredients []model.Ingredient
	for _, ing := range ingredients {
		var lastPrice model.IngredientPrice
		if err := database.DB.Where("store_id = ? AND ingredient_id = ?", storeID, ing.ID).
			Order("id DESC").First(&lastPrice).Error; err == nil {
			changeRate := lastPrice.PriceChange.Abs()
			if changeRate.GreaterThanOrEqual(threshold) {
				alertIngredients = append(alertIngredients, ing)
			}
		}
	}

	if len(alertIngredients) == 0 {
		log.Printf("[CostAlert] No price changes exceed threshold (%.2f%%)", threshold.InexactFloat64())
		return
	}

	cooldownHours := config.AppConfig.CostAlert.CooldownHours
	if cooldownHours <= 0 {
		cooldownHours = 24
	}

	var alerts []model.CostAlert
	var store model.Store
	database.DB.First(&store, storeID)

	for _, ing := range alertIngredients {
		var lastAlert model.CostAlert
		recentAlert := database.DB.Where("store_id = ? AND ingredient_id = ? AND status = 0", storeID, ing.ID).
			Order("id DESC").First(&lastAlert)

		skipAlert := false
		if recentAlert.Error == nil && lastAlert.NotifiedAt != nil {
			hoursSince := time.Since(*lastAlert.NotifiedAt).Hours()
			if hoursSince < float64(cooldownHours) {
				skipAlert = true
			}
		}

		if skipAlert {
			continue
		}

		var lastPrice model.IngredientPrice
		database.DB.Where("store_id = ? AND ingredient_id = ?", storeID, ing.ID).
			Order("id DESC").First(&lastPrice)

		alert := model.CostAlert{
			StoreID:        storeID,
			IngredientID:   ing.ID,
			IngredientName: ing.Name,
			AlertType:      "price_rise",
			PreviousPrice:  lastPrice.PreviousPrice,
			CurrentPrice:   lastPrice.Price,
			ChangeRate:     lastPrice.PriceChange.Abs(),
			Threshold:      threshold,
			Status:         0,
		}
		database.DB.Create(&alert)
		alerts = append(alerts, alert)
	}

	if len(alerts) > 0 {
		if err := s.sendDingTalkAlert(store, alerts); err != nil {
			log.Printf("[CostAlert] Failed to send DingTalk alert: %v", err)
		} else {
			now := time.Now()
			for i := range alerts {
				alerts[i].NotifiedAt = &now
				database.DB.Save(&alerts[i])
			}
			log.Printf("[CostAlert] Sent %d cost alerts via DingTalk", len(alerts))
		}
	}
}

func (s *CostAlertService) sendDingTalkAlert(store model.Store, alerts []model.CostAlert) error {
	cfg := config.AppConfig.DingTalk
	if cfg.Webhook == "" {
		return fmt.Errorf("dingtalk webhook not configured")
	}

	webhookURL := cfg.Webhook
	if cfg.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := s.generateDingTalkSign(timestamp, cfg.Secret)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, url.QueryEscape(sign))
	}

	title := fmt.Sprintf("【成本预警】%s - %d种食材价格异常波动", store.Name, len(alerts))

	var markdownText string
	markdownText += fmt.Sprintf("### %s\n\n", title)
	markdownText += fmt.Sprintf("**门店**: %s\n\n", store.Name)
	markdownText += fmt.Sprintf("**告警数量**: %d 种食材\n\n", len(alerts))
	markdownText += "---\n\n"

	for i, alert := range alerts {
		direction := "上涨"
		if alert.CurrentPrice.LessThan(alert.PreviousPrice) {
			direction = "下跌"
		}
		markdownText += fmt.Sprintf("**%d. %s**\n\n", i+1, alert.IngredientName)
		markdownText += fmt.Sprintf("- 原价: ¥%s / %s\n\n", alert.PreviousPrice.String(), "单位")
		markdownText += fmt.Sprintf("- 现价: ¥%s / %s\n\n", alert.CurrentPrice.String(), "单位")
		markdownText += fmt.Sprintf("- 变动: **%s%%** %s\n\n", alert.ChangeRate.StringFixed(2), direction)
		markdownText += "\n"
	}

	markdownText += "---\n"
	markdownText += fmt.Sprintf("*告警时间: %s*", time.Now().Format("2006-01-02 15:04:05"))

	msg := DingTalkMessage{
		MsgType: "markdown",
		Markdown: &DingTalkMarkdownContent{
			Title: title,
			Text:  markdownText,
		},
	}

	msgBody, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	resp, err := s.httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(msgBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	json.Unmarshal(respBody, &result)

	if result.ErrCode != 0 {
		return fmt.Errorf("dingtalk api error: %s", result.ErrMsg)
	}

	return nil
}

func (s *CostAlertService) generateDingTalkSign(timestamp int64, secret string) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (s *CostAlertService) GetAlerts(storeID uint, status int, page, pageSize int) ([]model.CostAlert, int64, error) {
	var alerts []model.CostAlert
	var total int64

	db := database.DB.Model(&model.CostAlert{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	db.Count(&total)

	if err := db.Preload("Ingredient").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

func (s *CostAlertService) HandleAlert(alertID uint, handler, remark string) error {
	var alert model.CostAlert
	if err := database.DB.First(&alert, alertID).Error; err != nil {
		return fmt.Errorf("alert not found: %v", err)
	}

	now := time.Now()
	alert.Status = 1
	alert.HandledAt = &now
	alert.Handler = handler
	alert.Remark = remark

	return database.DB.Save(&alert).Error
}

func (s *CostAlertService) GetPriceHistory(ingredientID uint, startDate, endDate string, limit int) ([]model.IngredientPrice, error) {
	var prices []model.IngredientPrice

	db := database.DB.Where("ingredient_id = ?", ingredientID)
	if startDate != "" {
		db = db.Where("effective_date >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("effective_date <= ?", endDate)
	}

	if limit > 0 {
		db = db.Limit(limit)
	}

	if err := db.Order("effective_date DESC, id DESC").Find(&prices).Error; err != nil {
		return nil, err
	}

	return prices, nil
}

func ParseInt(s string, defaultVal int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultVal
}
