package service

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"stalll-hub-pos/backend/pkg/config"
	"time"
)

type KuaiDiNiaoService struct {
	EBusinessID string
	AppKey      string
	APIURL      string
}

func NewKuaiDiNiaoService() *KuaiDiNiaoService {
	cfg := config.AppConfig.KuaiDiNiao
	return &KuaiDiNiaoService{
		EBusinessID: cfg.EBusinessID,
		AppKey:      cfg.AppKey,
		APIURL:      cfg.APIURL,
	}
}

type KDNQueryRequest struct {
	OrderCode    string `json:"OrderCode"`
	ShipperCode  string `json:"ShipperCode"`
	LogisticCode string `json:"LogisticCode"`
}

type KDNQueryResponse struct {
	EBusinessID   string        `json:"EBusinessID"`
	OrderCode     string        `json:"OrderCode"`
	ShipperCode   string        `json:"ShipperCode"`
	LogisticCode  string        `json:"LogisticCode"`
	Success       bool          `json:"Success"`
	Reason        string        `json:"Reason"`
	State         string        `json:"State"`
	StateEx       string        `json:"StateEx"`
	Location      string        `json:"Location"`
	Traces        []KDNTrace     `json:"Traces"`
}

type KDNTrace struct {
	AcceptTime    string `json:"AcceptTime"`
	AcceptStation string `json:"AcceptStation"`
	Remark        string `json:"Remark"`
	Location      string `json:"Location"`
	Action        string `json:"Action"`
}

func (s *KuaiDiNiaoService) GetLogisticsTrack(trackingNo string, shipperCode string) (*KDNQueryResponse, error) {
	if s.EBusinessID == "" || s.AppKey == "" {
		return nil, errors.New("快递鸟API未配置")
	}

	reqData := KDNQueryRequest{
		LogisticCode: trackingNo,
		ShipperCode:  shipperCode,
	}

	reqDataJSON, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	dataSign := s.encrypt(string(reqDataJSON), s.AppKey)

	form := url.Values{}
	form.Set("RequestData", string(reqDataJSON))
	form.Set("EBusinessID", s.EBusinessID)
	form.Set("RequestType", "1002")
	form.Set("DataSign", dataSign)
	form.Set("DataType", "2")

	apiURL := s.APIURL
	if apiURL == "" {
		apiURL = "https://api.kdniao.com/Ebusiness/EbusinessOrderHandle.aspx"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result KDNQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, errors.New(result.Reason)
	}

	return &result, nil
}

func (s *KuaiDiNiaoService) encrypt(content string, keyValue string) string {
	str := content + keyValue
	h := md5.New()
	h.Write([]byte(str))
	md5Str := fmt.Sprintf("%x", h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(md5Str))
}

func GetLogisticsStatusText(state string) string {
	switch state {
	case "0":
		return "暂无轨迹信息"
	case "1":
		return "已揽收"
	case "2":
		return "在途中"
	case "3":
		return "签收"
	case "4":
		return "问题件"
	default:
		return "未知状态"
	}
}

func GetLogisticsCompanyCode(companyName string) string {
	companyMap := map[string]string{
		"顺丰":      "SF",
		"顺丰速运":   "SF",
		"圆通":      "YTO",
		"圆通速递":   "YTO",
		"中通":      "ZTO",
		"中通快递":   "ZTO",
		"申通":      "STO",
		"申通快递":   "STO",
		"韵达":      "YD",
		"韵达快递":   "YD",
		"百世":      "HTKY",
		"百世快递":   "HTKY",
		"EMS":      "EMS",
		"邮政":      "YZPY",
		"邮政快递":   "YZPY",
		"京东":      "JD",
		"京东物流":   "JD",
		"德邦":      "DBL",
		"德邦物流":   "DBL",
	}
	return companyMap[companyName]
}
