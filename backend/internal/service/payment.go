package service

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
)

const (
	wechatUnifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	wechatRefundURL       = "https://api.mch.weixin.qq.com/secapi/pay/refund"
	wechatNotifyURL       = "/api/payment/wechat/notify"
	wechatMicropayURL     = "https://api.mch.weixin.qq.com/pay/micropay"
	wechatFacepayURL      = "https://payapp.weixin.qq.com/face/pay"
	alipayTradePayURL     = "https://openapi.alipay.com/gateway.do"
	alipayTradePayURLSandbox = "https://openapi.alipaydev.com/gateway.do"
)

type WechatUnifiedOrderRequest struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int      `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	NotifyURL      string   `xml:"notify_url"`
	TradeType      string   `xml:"trade_type"`
	OpenID         string   `xml:"openid,omitempty"`
}

type WechatUnifiedOrderResponse struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	AppID      string   `xml:"appid"`
	MchID      string   `xml:"mch_id"`
	NonceStr   string   `xml:"nonce_str"`
	Sign       string   `xml:"sign"`
	ResultCode string   `xml:"result_code"`
	PrepayID   string   `xml:"prepay_id"`
	TradeType  string   `xml:"trade_type"`
	CodeURL    string   `xml:"code_url"`
	ErrCode    string   `xml:"err_code"`
	ErrCodeDes string   `xml:"err_code_des"`
}

type WechatPaymentNotify struct {
	XMLName       xml.Name `xml:"xml"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	ResultCode    string   `xml:"result_code"`
	OpenID        string   `xml:"openid"`
	TradeType     string   `xml:"trade_type"`
	BankType      string   `xml:"bank_type"`
	TotalFee      int      `xml:"total_fee"`
	FeeType       string   `xml:"fee_type"`
	CashFee       int      `xml:"cash_fee"`
	TransactionID string   `xml:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no"`
	TimeEnd       string   `xml:"time_end"`
}

type WechatRefundRequest struct {
	XMLName         xml.Name `xml:"xml"`
	AppID           string   `xml:"appid"`
	MchID           string   `xml:"mch_id"`
	NonceStr        string   `xml:"nonce_str"`
	Sign            string   `xml:"sign"`
	TransactionID   string   `xml:"transaction_id,omitempty"`
	OutTradeNo      string   `xml:"out_trade_no"`
	OutRefundNo     string   `xml:"out_refund_no"`
	TotalFee        int      `xml:"total_fee"`
	RefundFee       int      `xml:"refund_fee"`
	RefundFeeType   string   `xml:"refund_fee_type"`
	OpUserID        string   `xml:"op_user_id"`
	RefundAccount   string   `xml:"refund_account"`
}

type WechatRefundResponse struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	ResultCode    string   `xml:"result_code"`
	ErrCode       string   `xml:"err_code"`
	ErrCodeDes    string   `xml:"err_code_des"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	TransactionID string   `xml:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no"`
	OutRefundNo   string   `xml:"out_refund_no"`
	RefundID      string   `xml:"refund_id"`
	RefundFee     int      `xml:"refund_fee"`
}

type PaymentService struct {
	orderRepo *repository.OrderRepository
	cfg       *config.Config
}

func NewPaymentService(cfg *config.Config) *PaymentService {
	return &PaymentService{
		orderRepo: repository.NewOrderRepository(nil),
		cfg:       cfg,
	}
}

func (s *PaymentService) generateMD5Sign(params map[string]string, apiKey string) string {
	var keys []string
	for k := range params {
		if k != "sign" && params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var signStr string
	for _, k := range keys {
		signStr += k + "=" + params[k] + "&"
	}
	signStr += "key=" + apiKey

	h := md5.New()
	h.Write([]byte(signStr))
	return strings.ToUpper(fmt.Sprintf("%x", h.Sum(nil)))
}

func (s *PaymentService) generateNonceStr() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
	}
	return fmt.Sprintf("%x", b)
}

func (s *PaymentService) generateRefundNo() string {
	now := time.Now()
	return fmt.Sprintf("RF%s%06d", now.Format("20060102150405"), time.Now().UnixNano()%1000000)
}

func (s *PaymentService) WechatUnifiedOrder(req *dto.WechatUnifiedOrderDTO) (*dto.WechatUnifiedOrderResponse, error) {
	order, err := s.orderRepo.GetByOrderNo(req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.PayStatus == 1 {
		return nil, errors.New("order already paid")
	}

	if order.OrderStatus == -1 {
		return nil, errors.New("order already cancelled")
	}

	totalFee := int(req.Amount.Mul(decimal.NewFromInt(100)).IntPart())
	if totalFee <= 0 {
		totalFee = 1
	}

	params := map[string]string{
		"appid":            s.cfg.Wechat.AppID,
		"mch_id":           s.cfg.Wechat.MchID,
		"nonce_str":        s.generateNonceStr(),
		"body":             req.Description,
		"out_trade_no":     req.OrderNo,
		"total_fee":        fmt.Sprintf("%d", totalFee),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       req.NotifyURL,
		"trade_type":       req.TradeType,
	}

	if req.OpenID != "" {
		params["openid"] = req.OpenID
	}

	params["sign"] = s.generateMD5Sign(params, s.cfg.Wechat.APIKey)

	wxReq := &WechatUnifiedOrderRequest{
		AppID:          params["appid"],
		MchID:          params["mch_id"],
		NonceStr:       params["nonce_str"],
		Sign:           params["sign"],
		Body:           params["body"],
		OutTradeNo:     params["out_trade_no"],
		TotalFee:       totalFee,
		SpbillCreateIP: params["spbill_create_ip"],
		NotifyURL:      params["notify_url"],
		TradeType:      params["trade_type"],
		OpenID:         params["openid"],
	}

	xmlData, err := xml.Marshal(wxReq)
	if err != nil {
		return nil, fmt.Errorf("marshal xml failed: %w", err)
	}

	resp, err := http.Post(wechatUnifiedOrderURL, "application/xml", bytes.NewReader(xmlData))
	if err != nil {
		return nil, fmt.Errorf("request wechat api failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var wxResp WechatUnifiedOrderResponse
	if err := xml.Unmarshal(body, &wxResp); err != nil {
		return nil, fmt.Errorf("unmarshal xml failed: %w", err)
	}

	if wxResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat return error: %s", wxResp.ReturnMsg)
	}

	if wxResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat result error: %s - %s", wxResp.ErrCode, wxResp.ErrCodeDes)
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonceStr := s.generateNonceStr()
	packageStr := "prepay_id=" + wxResp.PrepayID

	signParams := map[string]string{
		"appId":     s.cfg.Wechat.AppID,
		"timeStamp": timestamp,
		"nonceStr":  nonceStr,
		"package":   packageStr,
		"signType":  "MD5",
	}
	paySign := s.generateMD5Sign(signParams, s.cfg.Wechat.APIKey)

	return &dto.WechatUnifiedOrderResponse{
		PrepayID:  wxResp.PrepayID,
		CodeURL:   wxResp.CodeURL,
		AppID:     s.cfg.Wechat.AppID,
		TimeStamp: timestamp,
		NonceStr:  nonceStr,
		Package:   packageStr,
		SignType:  "MD5",
		PaySign:   paySign,
	}, nil
}

func (s *PaymentService) WechatPaymentNotify(xmlData []byte) (bool, string, error) {
	var notify WechatPaymentNotify
	if err := xml.Unmarshal(xmlData, &notify); err != nil {
		return false, "", fmt.Errorf("unmarshal notify failed: %w", err)
	}

	params := map[string]string{
		"appid":          notify.AppID,
		"mch_id":         notify.MchID,
		"nonce_str":      notify.NonceStr,
		"result_code":    notify.ResultCode,
		"openid":         notify.OpenID,
		"trade_type":     notify.TradeType,
		"bank_type":      notify.BankType,
		"total_fee":      fmt.Sprintf("%d", notify.TotalFee),
		"fee_type":       notify.FeeType,
		"cash_fee":       fmt.Sprintf("%d", notify.CashFee),
		"transaction_id": notify.TransactionID,
		"out_trade_no":   notify.OutTradeNo,
		"time_end":       notify.TimeEnd,
	}

	calculatedSign := s.generateMD5Sign(params, s.cfg.Wechat.APIKey)
	if calculatedSign != notify.Sign {
		return false, notify.OutTradeNo, errors.New("invalid sign")
	}

	if notify.ResultCode != "SUCCESS" {
		return false, notify.OutTradeNo, errors.New("payment failed")
	}

	amount := decimal.NewFromInt(int64(notify.TotalFee)).Div(decimal.NewFromInt(100))
	order, err := s.orderRepo.GetByOrderNo(notify.OutTradeNo)
	if err != nil {
		return false, notify.OutTradeNo, fmt.Errorf("order not found: %w", err)
	}

	payTime := time.Now()
	err = s.orderRepo.UpdatePayStatus(order.ID, 1, "wechat", &payTime)
	if err != nil {
		return false, notify.OutTradeNo, fmt.Errorf("update order status failed: %w", err)
	}

	payment := &model.OrderPayment{
		OrderID:       order.ID,
		PayMethod:     "wechat",
		Amount:        amount,
		TransactionID: notify.TransactionID,
		PayStatus:     1,
		PayTime:       &payTime,
	}
	_ = s.orderRepo.CreatePayment(payment)
	_ = s.orderRepo.UpdateStatus(order.ID, 2)

	return true, notify.OutTradeNo, nil
}

func (s *PaymentService) WechatRefund(req *dto.WechatRefundDTO) (*dto.WechatRefundResponse, error) {
	order, err := s.orderRepo.GetByOrderNo(req.OrderNo)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if order.PayStatus != 1 {
		return nil, errors.New("order not paid")
	}

	totalFee := int(req.TotalFee.Mul(decimal.NewFromInt(100)).IntPart())
	refundFee := int(req.RefundFee.Mul(decimal.NewFromInt(100)).IntPart())

	if refundFee > totalFee {
		return nil, errors.New("refund amount exceeds paid amount")
	}

	outRefundNo := req.RefundNo
	if outRefundNo == "" {
		outRefundNo = s.generateRefundNo()
	}

	params := map[string]string{
		"appid":           s.cfg.Wechat.AppID,
		"mch_id":          s.cfg.Wechat.MchID,
		"nonce_str":       s.generateNonceStr(),
		"out_trade_no":    req.OrderNo,
		"out_refund_no":   outRefundNo,
		"total_fee":       fmt.Sprintf("%d", totalFee),
		"refund_fee":      fmt.Sprintf("%d", refundFee),
		"refund_desc":     req.RefundDesc,
		"notify_url":      req.NotifyURL,
		"op_user_id":      s.cfg.Wechat.MchID,
	}

	params["sign"] = s.generateMD5Sign(params, s.cfg.Wechat.APIKey)

	wxReq := &WechatRefundRequest{
		AppID:        params["appid"],
		MchID:        params["mch_id"],
		NonceStr:     params["nonce_str"],
		Sign:         params["sign"],
		OutTradeNo:   params["out_trade_no"],
		OutRefundNo:  params["out_refund_no"],
		TotalFee:     totalFee,
		RefundFee:    refundFee,
		OpUserID:     params["op_user_id"],
	}

	xmlData, err := xml.Marshal(wxReq)
	if err != nil {
		return nil, fmt.Errorf("marshal xml failed: %w", err)
	}

	resp, err := http.Post(wechatRefundURL, "application/xml", bytes.NewReader(xmlData))
	if err != nil {
		return nil, fmt.Errorf("request wechat refund api failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var wxResp WechatRefundResponse
	if err := xml.Unmarshal(body, &wxResp); err != nil {
		return nil, fmt.Errorf("unmarshal xml failed: %w", err)
	}

	if wxResp.ReturnCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat return error: %s", wxResp.ReturnMsg)
	}

	if wxResp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat result error: %s - %s", wxResp.ErrCode, wxResp.ErrCodeDes)
	}

	refund := &model.OrderRefund{
		OrderID:      order.ID,
		RefundNo:     outRefundNo,
		RefundAmount: req.RefundFee,
		RefundReason: req.RefundDesc,
		RefundType:   "wechat",
		RefundStatus: 1,
	}
	_ = s.orderRepo.CreateRefund(refund)

	successTime := time.Now().Format("2006-01-02 15:04:05")

	return &dto.WechatRefundResponse{
		RefundID:          wxResp.RefundID,
		OutRefundNo:       wxResp.OutRefundNo,
		RefundStatus:      "SUCCESS",
		RefundRecvAccout:  "",
		SuccessTime:       successTime,
	}, nil
}

func (s *PaymentService) VerifyWechatSign(params map[string]string, sign string) bool {
	calculatedSign := s.generateMD5Sign(params, s.cfg.Wechat.APIKey)
	return calculatedSign == strings.ToUpper(sign)
}

func (s *PaymentService) GetPaymentParams(req *dto.PaymentParamsRequest) (*dto.PaymentParamsResponse, error) {
	order, err := s.orderRepo.GetByID(req.OrderID)
	if err != nil {
		return nil, err
	}

	if order.PayStatus == 1 {
		return nil, errors.New("order already paid")
	}

	if order.OrderStatus == -1 {
		return nil, errors.New("order already cancelled")
	}

	params := make(map[string]string)

	switch req.PayType {
	case "wechat":
		unifiedReq := &dto.WechatUnifiedOrderDTO{
			StoreID:     order.StoreID,
			OrderNo:     order.OrderNo,
			Amount:      order.PayAmount,
			Description: fmt.Sprintf("大排档订单-%s", order.OrderNo),
			TradeType:   "NATIVE",
			NotifyURL:   fmt.Sprintf("%s/api/payment/wechat/notify", "http://localhost:8080"),
		}
		unifiedResp, err := s.WechatUnifiedOrder(unifiedReq)
		if err != nil {
			return nil, err
		}

		params["appId"] = unifiedResp.AppID
		params["prepayId"] = unifiedResp.PrepayID
		params["package"] = unifiedResp.Package
		params["nonceStr"] = unifiedResp.NonceStr
		params["timeStamp"] = unifiedResp.TimeStamp
		params["signType"] = unifiedResp.SignType
		params["paySign"] = unifiedResp.PaySign
		if unifiedResp.CodeURL != "" {
			params["codeUrl"] = unifiedResp.CodeURL
		}

	case "alipay":
		params["appId"] = s.cfg.Alipay.AppID
		params["outTradeNo"] = order.OrderNo
		params["subject"] = fmt.Sprintf("大排档订单-%s", order.OrderNo)
		params["totalAmount"] = order.PayAmount.String()
		params["notifyUrl"] = s.cfg.Alipay.NotifyURL

	case "cash":
		params["cashier"] = "system"
		params["amount"] = order.PayAmount.String()
	}

	return &dto.PaymentParamsResponse{
		PayType: req.PayType,
		OrderID: order.ID,
		OrderNo: order.OrderNo,
		Amount:  order.PayAmount,
		Params:  params,
	}, nil
}

type WechatMicropayRequest struct {
	XMLName        xml.Name `xml:"xml"`
	AppID          string   `xml:"appid"`
	MchID          string   `xml:"mch_id"`
	NonceStr       string   `xml:"nonce_str"`
	Sign           string   `xml:"sign"`
	Body           string   `xml:"body"`
	OutTradeNo     string   `xml:"out_trade_no"`
	TotalFee       int      `xml:"total_fee"`
	SpbillCreateIP string   `xml:"spbill_create_ip"`
	AuthCode       string   `xml:"auth_code"`
	FaceCode       string   `xml:"face_code,omitempty"`
}

type WechatMicropayResponse struct {
	XMLName       xml.Name `xml:"xml"`
	ReturnCode    string   `xml:"return_code"`
	ReturnMsg     string   `xml:"return_msg"`
	AppID         string   `xml:"appid"`
	MchID         string   `xml:"mch_id"`
	DeviceInfo    string   `xml:"device_info"`
	NonceStr      string   `xml:"nonce_str"`
	Sign          string   `xml:"sign"`
	ResultCode    string   `xml:"result_code"`
	OpenID        string   `xml:"openid"`
	IsSubscribe   string   `xml:"is_subscribe"`
	TradeType     string   `xml:"trade_type"`
	BankType      string   `xml:"bank_type"`
	TotalFee      int      `xml:"total_fee"`
	FeeType       string   `xml:"fee_type"`
	CashFee       int      `xml:"cash_fee"`
	TransactionID string   `xml:"transaction_id"`
	OutTradeNo    string   `xml:"out_trade_no"`
	Attach        string   `xml:"attach"`
	TimeEnd       string   `xml:"time_end"`
	ErrCode       string   `xml:"err_code"`
	ErrCodeDes    string   `xml:"err_code_des"`
}

type AlipayTradePayResponse struct {
	AlipayTradePayResponse struct {
		Code              string `json:"code"`
		Msg               string `json:"msg"`
		SubCode           string `json:"sub_code"`
		SubMsg            string `json:"sub_msg"`
		TradeNo           string `json:"trade_no"`
		OutTradeNo        string `json:"out_trade_no"`
		BuyerLogonID      string `json:"buyer_logon_id"`
		TotalAmount       string `json:"total_amount"`
		ReceiptAmount     string `json:"receipt_amount"`
		InvoiceAmount     string `json:"invoice_amount"`
		BuyerPayAmount    string `json:"buyer_pay_amount"`
		PointAmount       string `json:"point_amount"`
		DiscountGoodsDetail string `json:"discount_goods_detail"`
		GmtPayment        string `json:"gmt_payment"`
		FundBillList      string `json:"fund_bill_list"`
		CardBalance       string `json:"card_balance"`
	} `json:"alipay_trade_pay_response"`
	Sign string `json:"sign"`
}

type WechatFacePayResult struct {
	Success       bool
	TransactionID string
	PayTime       time.Time
	ErrMsg        string
}

type AlipayFacePayResult struct {
	Success       bool
	TransactionID string
	PayTime       time.Time
	ErrMsg        string
}

func (s *PaymentService) WechatFacePay(orderNo string, amount decimal.Decimal, authCode string, faceCode string) (*WechatFacePayResult, error) {
	totalFee := int(amount.Mul(decimal.NewFromInt(100)).IntPart())
	if totalFee <= 0 {
		totalFee = 1
	}

	params := map[string]string{
		"appid":            s.cfg.Wechat.AppID,
		"mch_id":           s.cfg.Wechat.MchID,
		"nonce_str":        s.generateNonceStr(),
		"body":             "刷脸支付-" + orderNo,
		"out_trade_no":     orderNo,
		"total_fee":        fmt.Sprintf("%d", totalFee),
		"spbill_create_ip": "127.0.0.1",
		"auth_code":        authCode,
	}
	if faceCode != "" {
		params["face_code"] = faceCode
	}

	params["sign"] = s.generateMD5Sign(params, s.cfg.Wechat.APIKey)

	wxReq := &WechatMicropayRequest{
		AppID:          params["appid"],
		MchID:          params["mch_id"],
		NonceStr:       params["nonce_str"],
		Sign:           params["sign"],
		Body:           params["body"],
		OutTradeNo:     params["out_trade_no"],
		TotalFee:       totalFee,
		SpbillCreateIP: params["spbill_create_ip"],
		AuthCode:       params["auth_code"],
		FaceCode:       params["face_code"],
	}

	xmlData, err := xml.Marshal(wxReq)
	if err != nil {
		return nil, fmt.Errorf("marshal xml failed: %w", err)
	}

	resp, err := http.Post(wechatMicropayURL, "application/xml", bytes.NewReader(xmlData))
	if err != nil {
		return nil, fmt.Errorf("request wechat micropay api failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var wxResp WechatMicropayResponse
	if err := xml.Unmarshal(body, &wxResp); err != nil {
		return nil, fmt.Errorf("unmarshal xml failed: %w", err)
	}

	if wxResp.ReturnCode != "SUCCESS" {
		return &WechatFacePayResult{
			Success: false,
			ErrMsg:  wxResp.ReturnMsg,
		}, fmt.Errorf("wechat return error: %s", wxResp.ReturnMsg)
	}

	if wxResp.ResultCode != "SUCCESS" {
		return &WechatFacePayResult{
			Success: false,
			ErrMsg:  fmt.Sprintf("%s - %s", wxResp.ErrCode, wxResp.ErrCodeDes),
		}, fmt.Errorf("wechat result error: %s - %s", wxResp.ErrCode, wxResp.ErrCodeDes)
	}

	if !s.VerifyWechatSign(map[string]string{
		"return_code":     wxResp.ReturnCode,
		"appid":           wxResp.AppID,
		"mch_id":          wxResp.MchID,
		"nonce_str":       wxResp.NonceStr,
		"result_code":     wxResp.ResultCode,
		"openid":          wxResp.OpenID,
		"trade_type":      wxResp.TradeType,
		"bank_type":       wxResp.BankType,
		"total_fee":       fmt.Sprintf("%d", wxResp.TotalFee),
		"cash_fee":        fmt.Sprintf("%d", wxResp.CashFee),
		"transaction_id":  wxResp.TransactionID,
		"out_trade_no":    wxResp.OutTradeNo,
		"time_end":        wxResp.TimeEnd,
	}, wxResp.Sign) {
		return &WechatFacePayResult{
			Success: false,
			ErrMsg:  "签名验证失败",
		}, errors.New("invalid wechat response sign")
	}

	payTime, _ := time.Parse("20060102150405", wxResp.TimeEnd)
	if payTime.IsZero() {
		payTime = time.Now()
	}

	return &WechatFacePayResult{
		Success:       true,
		TransactionID: wxResp.TransactionID,
		PayTime:       payTime,
	}, nil
}

func (s *PaymentService) AlipayFacePay(orderNo string, amount decimal.Decimal, authCode string) (*AlipayFacePayResult, error) {
	gatewayURL := alipayTradePayURL
	if s.cfg.Alipay.Sandbox {
		gatewayURL = alipayTradePayURLSandbox
	}

	bizContent := map[string]interface{}{
		"out_trade_no": orderNo,
		"scene":        "security_code",
		"auth_code":    authCode,
		"subject":      "刷脸支付-" + orderNo,
		"total_amount": amount.String(),
	}
	bizContentStr, _ := json.Marshal(bizContent)

	publicParams := map[string]string{
		"app_id":      s.cfg.Alipay.AppID,
		"method":      "alipay.trade.pay",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"notify_url":  s.cfg.Alipay.NotifyURL,
		"biz_content": string(bizContentStr),
	}

	signStr := s.buildAlipaySignString(publicParams)
	signature, err := s.signAlipayRSA2(signStr, s.cfg.Alipay.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("sign alipay request failed: %w", err)
	}
	publicParams["sign"] = signature

	postData := s.buildAlipayPostData(publicParams)

	resp, err := http.Post(gatewayURL, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		return nil, fmt.Errorf("request alipay trade pay api failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read alipay response failed: %w", err)
	}

	var alipayResp AlipayTradePayResponse
	if err := json.Unmarshal(body, &alipayResp); err != nil {
		return nil, fmt.Errorf("unmarshal alipay response failed: %w", err)
	}

	if alipayResp.AlipayTradePayResponse.Code != "10000" {
		return &AlipayFacePayResult{
			Success: false,
			ErrMsg:  fmt.Sprintf("%s - %s", alipayResp.AlipayTradePayResponse.SubCode, alipayResp.AlipayTradePayResponse.SubMsg),
		}, fmt.Errorf("alipay error: code=%s msg=%s sub_code=%s sub_msg=%s",
			alipayResp.AlipayTradePayResponse.Code,
			alipayResp.AlipayTradePayResponse.Msg,
			alipayResp.AlipayTradePayResponse.SubCode,
			alipayResp.AlipayTradePayResponse.SubMsg)
	}

	if err := s.verifyAlipaySign(string(body), alipayResp.Sign, s.cfg.Alipay.PublicKey); err != nil {
		return &AlipayFacePayResult{
			Success: false,
			ErrMsg:  "支付宝签名验证失败",
		}, fmt.Errorf("alipay sign verify failed: %w", err)
	}

	payTime, _ := time.Parse("2006-01-02 15:04:05", alipayResp.AlipayTradePayResponse.GmtPayment)
	if payTime.IsZero() {
		payTime = time.Now()
	}

	return &AlipayFacePayResult{
		Success:       true,
		TransactionID: alipayResp.AlipayTradePayResponse.TradeNo,
		PayTime:       payTime,
	}, nil
}

func (s *PaymentService) buildAlipaySignString(params map[string]string) string {
	var keys []string
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		pairs = append(pairs, k+"="+params[k])
	}
	return strings.Join(pairs, "&")
}

func (s *PaymentService) buildAlipayPostData(params map[string]string) string {
	var pairs []string
	for k, v := range params {
		pairs = append(pairs, k+"="+v)
	}
	return strings.Join(pairs, "&")
}

func (s *PaymentService) signAlipayRSA2(data string, privateKey string) (string, error) {
	_ = data
	_ = privateKey
	return "", errors.New("alipay RSA2 signing requires a proper crypto implementation, please configure private key pem content in config.alipay.private_key")
}

func (s *PaymentService) verifyAlipaySign(body string, sign string, publicKey string) error {
	_ = body
	_ = sign
	_ = publicKey
	return errors.New("alipay sign verification requires public key pem content")
}
