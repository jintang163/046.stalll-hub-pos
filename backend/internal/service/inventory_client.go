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
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
)

type InventoryClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

type InventoryIngredient struct {
	IngredientNo string          `json:"ingredient_no"`
	Name         string          `json:"name"`
	Category     string          `json:"category"`
	Unit         string          `json:"unit"`
	Price        decimal.Decimal `json:"price"`
	Supplier     string          `json:"supplier"`
	Status       int             `json:"status"`
}

type InventoryResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    []InventoryIngredient  `json:"data"`
	Total   int                    `json:"total"`
}

func NewInventoryClient() *InventoryClient {
	cfg := config.AppConfig.Inventory
	timeout := 30
	if cfg.TimeoutSeconds > 0 {
		timeout = cfg.TimeoutSeconds
	}
	return &InventoryClient{
		baseURL:   cfg.BaseURL,
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

func (c *InventoryClient) generateSignature(timestamp int64) string {
	message := fmt.Sprintf("%s%d", c.apiKey, timestamp)
	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *InventoryClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	timestamp := time.Now().Unix()
	signature := c.generateSignature(timestamp)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-Signature", signature)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *InventoryClient) GetIngredients(storeID uint, page, pageSize int) (*InventoryResponse, error) {
	path := fmt.Sprintf("/api/ingredients?page=%d&page_size=%d", page, pageSize)
	if storeID > 0 {
		path += fmt.Sprintf("&store_id=%d", storeID)
	}

	respBody, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp InventoryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("inventory api error: %s", resp.Message)
	}

	return &resp, nil
}

func (c *InventoryClient) GetAllIngredients(storeID uint) ([]InventoryIngredient, error) {
	var allIngredients []InventoryIngredient
	page := 1
	pageSize := 100

	for {
		resp, err := c.GetIngredients(storeID, page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("failed to get ingredients page %d: %v", page, err)
		}

		allIngredients = append(allIngredients, resp.Data...)

		if len(allIngredients) >= resp.Total || len(resp.Data) == 0 {
			break
		}

		page++
	}

	log.Printf("[Inventory] Fetched %d ingredients from inventory system (store_id=%d)", len(allIngredients), storeID)
	return allIngredients, nil
}
