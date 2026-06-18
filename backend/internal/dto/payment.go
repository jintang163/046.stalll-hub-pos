package dto

import (
	"github.com/shopspring/decimal"
)

type WechatUnifiedOrderDTO struct {
	StoreID     uint            `json:"store_id"`
	OrderNo     string          `json:"order_no" binding:"required"`
	Amount      decimal.Decimal `json:"amount" binding:"required,min=0.01"`
	Description string          `json:"description" binding:"max=128"`
	OpenID      string          `json:"openid"`
	NotifyURL   string          `json:"notify_url"`
	TradeType   string          `json:"trade_type" binding:"oneof=JSAPI NATIVE APP MWEB"`
}

type WechatUnifiedOrderResponse struct {
	PrepayID string `json:"prepay_id"`
	CodeURL  string `json:"code_url"`
	AppID     string `json:"appid"`
	TimeStamp string `json:"timestamp"`
	NonceStr  string `json:"noncestr"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

type WechatQueryOrderResponse struct {
	TradeState     string          `json:"trade_state"`
	TradeStateDesc string          `json:"trade_state_desc"`
	TransactionID  string          `json:"transaction_id"`
	OutTradeNo   string          `json:"out_trade_no"`
	TotalFee     decimal.Decimal `json:"total_fee"`
	CashFee      decimal.Decimal `json:"cash_fee"`
	TimeEnd       string          `json:"time_end"`
}

type WechatRefundDTO struct {
	OrderNo       string          `json:"order_no" binding:"required"`
	RefundNo      string          `json:"refund_no" binding:"required"`
	TotalFee      decimal.Decimal `json:"total_fee" binding:"required,min=0.01"`
	RefundFee     decimal.Decimal `json:"refund_fee" binding:"required,min=0.01"`
	RefundDesc    string          `json:"refund_desc" binding:"max=80"`
	NotifyURL     string          `json:"notify_url"`
}

type WechatRefundResponse struct {
	RefundID     string `json:"refund_id"`
	OutRefundNo  string `json:"out_refund_no"`
	RefundStatus string `json:"refund_status"`
	RefundRecvAccout string `json:"refund_recv_accout"`
	SuccessTime  string `json:"success_time"`
}
