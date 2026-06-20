package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/model"
	"time"
)

type DingTalkService struct{}

func NewDingTalkService() *DingTalkService {
	return &DingTalkService{}
}

type DingTextMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		AtAll     bool     `json:"isAtAll"`
	} `json:"at"`
}

type DingMarkdownMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	} `json:"markdown"`
	At struct {
		AtMobiles []string `json:"atMobiles"`
		AtAll     bool     `json:"isAtAll"`
	} `json:"at"`
}

type DingResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (s *DingTalkService) generateSign(secret string, timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return url.QueryEscape(sign), nil
}

func (s *DingTalkService) SendText(content string, atAll bool, atMobiles []string) error {
	cfg := config.AppConfig.DingTalk
	if cfg.Webhook == "" {
		return errors.New("dingtalk webhook not configured")
	}

	msg := DingTextMsg{
		MsgType: "text",
	}
	msg.Text.Content = content
	msg.At.AtAll = atAll
	msg.At.AtMobiles = atMobiles

	return s.send(msg)
}

func (s *DingTalkService) SendMarkdown(title, text string, atAll bool, atMobiles []string) error {
	cfg := config.AppConfig.DingTalk
	if cfg.Webhook == "" {
		return errors.New("dingtalk webhook not configured")
	}

	msg := DingMarkdownMsg{
		MsgType: "markdown",
	}
	msg.Markdown.Title = title
	msg.Markdown.Text = text
	msg.At.AtAll = atAll
	msg.At.AtMobiles = atMobiles

	return s.send(msg)
}

func (s *DingTalkService) send(msg interface{}) error {
	cfg := config.AppConfig.DingTalk
	webhook := cfg.Webhook

	if cfg.Secret != "" {
		timestamp := time.Now().UnixMilli()
		sign, err := s.generateSign(cfg.Secret, timestamp)
		if err != nil {
			return err
		}
		webhook = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhook, timestamp, sign)
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result DingResponse
	json.Unmarshal(respBody, &result)

	if result.ErrCode != 0 {
		return fmt.Errorf("dingtalk error: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

func (s *DingTalkService) SendStockWarning(warnings []model.StockWarning, storeName string) error {
	if len(warnings) == 0 {
		return nil
	}

	title := fmt.Sprintf("【库存预警】%s - 共%d个SKU低于安全库存", storeName, len(warnings))

	text := fmt.Sprintf("## 🚨 库存预警通知\n\n> 门店：%s\n> 预警时间：%s\n> 预警数量：%d个SKU\n\n",
		storeName,
		time.Now().Format("2006-01-02 15:04:05"),
		len(warnings),
	)

	text += "| 商品 | SKU | 当前库存 | 安全阈值 | 差额 |\n"
	text += "|------|-----|---------|---------|------|\n"

	for i, w := range warnings {
		if i >= 10 {
			text += fmt.Sprintf("| ... 还有 %d 个 ... | | | | |\n", len(warnings)-10)
			break
		}
		diff := w.CurrentStock - w.Threshold
		diffStr := strconv.Itoa(diff)
		if diff < 0 {
			diffStr = fmt.Sprintf("-%d", w.Threshold-w.CurrentStock)
		}
		productName := w.Product.Name
		specName := w.SKU.SpecName
		if specName != "" {
			productName += " / " + specName
		}
		text += fmt.Sprintf("| %s | %s | %d | %d | %s |\n",
			productName, w.SKU.SKUCode, w.CurrentStock, w.Threshold, diffStr)
	}

	text += "\n⚠️ 请及时补货，避免影响销售！"

	return s.SendMarkdown(title, text, false, nil)
}

func (s *DingTalkService) SendStockCheckComplete(checkNo, title string, totalSKU, diffCount int, diffAmount float64) error {
	titleMsg := fmt.Sprintf("【盘点完成】%s", title)
	text := fmt.Sprintf("## 📋 盘点单完成通知\n\n"+
		"- **盘点单号**：%s\n"+
		"- **盘点标题**：%s\n"+
		"- **盘点时间**：%s\n"+
		"- **SKU总数**：%d个\n"+
		"- **差异数量**：%d个SKU\n"+
		"- **差异金额**：%.2f元\n\n",
		checkNo,
		title,
		time.Now().Format("2006-01-02 15:04:05"),
		totalSKU,
		diffCount,
		diffAmount,
	)

	if diffCount > 0 {
		text += "⚠️ 存在库存差异，请及时处理！\n"
	} else {
		text += "✅ 本次盘点无差异，库存准确。\n"
	}

	return s.SendMarkdown(titleMsg, text, false, nil)
}
