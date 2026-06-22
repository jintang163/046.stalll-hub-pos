package service

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type PurchaseOrderV2Service struct {
	purchaseSvc   *PurchaseService
	supplierSvc   *SupplierService
	payableSvc    *AccountsPayableService
	smsService    *SmsService
	dingTalkSvc   *DingTalkService
	emailService  *EmailService
}

func NewPurchaseOrderV2Service() *PurchaseOrderV2Service {
	return &PurchaseOrderV2Service{
		purchaseSvc:  NewPurchaseService(),
		supplierSvc:  NewSupplierService(),
		payableSvc:   NewAccountsPayableService(),
		smsService:   NewSmsService(),
		dingTalkSvc:  NewDingTalkService(),
		emailService: NewEmailService(),
	}
}

func (s *PurchaseOrderV2Service) CreatePurchaseOrder(req *dto.PurchaseOrderCreateV2DTO) (*dto.PurchaseOrderV2Response, error) {
	var supplierName, supplierPhone, supplierEmail string
	var paymentTerm int

	if req.SupplierID > 0 {
		supplier, err := s.supplierSvc.supplierRepo.GetByID(req.SupplierID)
		if err != nil {
			return nil, fmt.Errorf("supplier not found: %w", err)
		}
		supplierName = supplier.Name
		supplierPhone = supplier.Phone
		supplierEmail = supplier.Email
		if supplier.Mobile != "" {
			supplierPhone = supplier.Mobile
		}
		paymentTerm = supplier.PaymentTerm
	} else {
		supplierName = req.SupplierName
		supplierPhone = req.SupplierPhone
		supplierEmail = req.SupplierEmail
		paymentTerm = req.PaymentTerm
	}

	createReq := &dto.PurchaseOrderCreateRequest{
		StoreID:       req.StoreID,
		ForecastDate:  req.ForecastDate,
		ForecastDays:  req.ForecastDays,
		SupplierName:  supplierName,
		SupplierPhone: supplierPhone,
		SupplierEmail: supplierEmail,
		Items:         req.Items,
		Remark:        req.Remark,
	}

	purchase, err := s.purchaseSvc.GeneratePurchaseOrder(createReq)
	if err != nil {
		return nil, err
	}

	database.DB.Model(&model.PurchaseOrder{}).Where("id = ?", purchase.ID).Updates(map[string]interface{}{
		"supplier_id":   req.SupplierID,
		"payment_term":  paymentTerm,
		"expected_date": req.ExpectedDate,
	})

	purchase.SupplierID = req.SupplierID
	purchase.PaymentTerm = paymentTerm
	purchase.ExpectedDate = req.ExpectedDate
	purchase.SupplierPhone = supplierPhone
	purchase.SupplierEmail = supplierEmail

	log.Printf("[PurchaseOrderV2Service] Created purchase order %s for supplier %s",
		purchase.PurchaseNo, supplierName)

	sendPurchaseNotifications(purchase, s.smsService, s.emailService, s.dingTalkSvc)

	_ = s.purchaseSvc.UpdateStatus(purchase.ID, 1)

	return s.GetPurchaseOrder(purchase.ID)
}

func sendPurchaseNotifications(purchase *model.PurchaseOrder, smsSvc *SmsService,
	emailSvc *EmailService, dingTalkSvc *DingTalkService) {

	smsContent := fmt.Sprintf("您有新的采购订单：单号%s，金额¥%s，请及时备货发货。",
		purchase.PurchaseNo, purchase.TotalAmount.String())

	if purchase.SupplierPhone != "" {
		go func(phone, content string) {
			err := smsSvc.SendSms(purchase.StoreID, 0, phone,
				"采购订单通知", "PURCHASE_NEW", content, 0)
			if err != nil {
				log.Printf("[PurchaseOrderV2Service] Auto send SMS to supplier %s (%s) failed: %v",
					purchase.SupplierName, phone, err)
			} else {
				log.Printf("[PurchaseOrderV2Service] Auto SMS sent to %s (%s): %s",
					purchase.SupplierName, phone, content)
			}
		}(purchase.SupplierPhone, smsContent)
	}

	if purchase.SupplierEmail != "" {
		go func(email, supplierName string, po *model.PurchaseOrder) {
			subject := fmt.Sprintf("【大排档POS】新采购订单 %s 请及时备货", po.PurchaseNo)
			htmlBody := buildPurchaseOrderEmail(po)
			err := emailSvc.SendEmail(&EmailMessage{
				To:      []string{email},
				Subject: subject,
				Body:    htmlBody,
				IsHTML:  true,
			})
			if err != nil {
				log.Printf("[PurchaseOrderV2Service] Auto send email to supplier %s (%s) failed: %v",
					supplierName, email, err)
			} else {
				log.Printf("[PurchaseOrderV2Service] Auto email queued for %s (%s), subject: %s",
					supplierName, email, subject)
			}
		}(purchase.SupplierEmail, purchase.SupplierName, purchase)
	}

	go func(po *model.PurchaseOrder) {
		dingMsg := fmt.Sprintf("📋 新采购订单已发送\n单号：%s\n供应商：%s\n总金额：¥%s\n明细数量：%d种食材",
			po.PurchaseNo, po.SupplierName, po.TotalAmount.String(), len(po.Items))
		_ = dingTalkSvc.SendDingTalkNotify(dingMsg, "采购通知")
	}(purchase)
}

func (s *PurchaseOrderV2Service) GetPurchaseOrder(id uint) (*dto.PurchaseOrderV2Response, error) {
	var purchase model.PurchaseOrder
	if err := database.DB.Preload("Items").Preload("Store").Preload("Supplier").
		First(&purchase, id).Error; err != nil {
		return nil, fmt.Errorf("purchase order not found: %w", err)
	}
	return s.convertToPurchaseOrderV2Response(&purchase), nil
}

func (s *PurchaseOrderV2Service) ListPurchaseOrders(query *dto.PurchaseOrderQuery) (*dto.PageResponse, error) {
	orders, total, err := s.purchaseSvc.ListPurchaseOrders(query)
	if err != nil {
		return nil, err
	}

	var list []dto.PurchaseOrderV2Response
	for i := range orders {
		po := &orders[i]
		database.DB.Preload("Supplier").First(po, po.ID)
		list = append(list, *s.convertToPurchaseOrderV2Response(po))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.Size,
	}, nil
}

func (s *PurchaseOrderV2Service) SendToSupplier(id uint, notifyTypes []string, content string) error {
	purchase, err := s.purchaseSvc.GetPurchaseOrder(id)
	if err != nil {
		return err
	}

	if err := s.purchaseSvc.SendToSupplier(id); err != nil {
		return err
	}

	notifyContent := content
	if notifyContent == "" {
		notifyContent = fmt.Sprintf("您有新的采购订单：单号%s，金额¥%s，请及时处理。",
			purchase.PurchaseNo, purchase.TotalAmount.String())
	}

	if purchase.SupplierID > 0 {
		_ = s.supplierSvc.NotifySupplier(purchase.SupplierID, notifyTypes, notifyContent)
	} else {
		for _, notifyType := range notifyTypes {
			switch notifyType {
			case "sms":
				if purchase.SupplierPhone != "" {
					go func() {
						err := s.smsService.SendSms(purchase.StoreID, 0, purchase.SupplierPhone,
							"采购通知", "PURCHASE_NOTIFY", notifyContent, 0)
						if err != nil {
							log.Printf("[PurchaseOrderV2Service] Send SMS failed: %v", err)
						}
					}()
				}
			case "email":
				if purchase.SupplierEmail != "" {
					log.Printf("[PurchaseOrderV2Service] Sending email to %s: %s",
						purchase.SupplierEmail, notifyContent)
					subject := fmt.Sprintf("【大排档POS】采购订单通知 %s", purchase.PurchaseNo)
					htmlBody := buildPurchaseOrderEmail(purchase)
					if content != "" {
						htmlBody = buildSupplierNotifyEmail(purchase.SupplierName, content)
					}
					_ = s.emailService.SendEmail(&EmailMessage{
						To:      []string{purchase.SupplierEmail},
						Subject: subject,
						Body:    htmlBody,
						IsHTML:  true,
					})
				}
			}
		}
	}

	return nil
}

func (s *PurchaseOrderV2Service) CompletePurchase(id uint) error {
	purchase, err := s.purchaseSvc.GetPurchaseOrder(id)
	if err != nil {
		return err
	}

	if purchase.Status != 3 && purchase.Status != 2 && purchase.Status != 1 {
		return fmt.Errorf("purchase order cannot be completed in current status")
	}

	if err := s.purchaseSvc.UpdateStatus(id, 4); err != nil {
		return err
	}

	if err := s.payableSvc.CreatePayableFromPurchase(purchase); err != nil {
		log.Printf("[PurchaseOrderV2Service] Create payable failed: %v", err)
	}

	if purchase.SupplierID > 0 {
		go s.supplierSvc.supplierRepo.UpdateStats(purchase.SupplierID)
	}

	return nil
}

func (s *PurchaseOrderV2Service) CancelPurchase(id uint, remark string) error {
	purchase, err := s.purchaseSvc.GetPurchaseOrder(id)
	if err != nil {
		return err
	}

	if purchase.Status >= 3 {
		return fmt.Errorf("purchase order already received, cannot cancel")
	}

	if err := s.purchaseSvc.UpdateStatus(id, 5); err != nil {
		return err
	}

	if remark != "" {
		database.DB.Model(&model.PurchaseOrder{}).Where("id = ?", id).
			Update("remark", purchase.Remark+" | 取消原因:"+remark)
	}

	return nil
}

func (s *PurchaseOrderV2Service) convertToPurchaseOrderV2Response(p *model.PurchaseOrder) *dto.PurchaseOrderV2Response {
	storeName := ""
	if p.Store.ID > 0 {
		storeName = p.Store.Name
	}

	statusText := purchaseStatusMap[p.Status]
	paymentTermText := paymentTermMap[p.PaymentTerm]

	var items []dto.PurchaseItemV2Response
	for _, item := range p.Items {
		items = append(items, dto.PurchaseItemV2Response{
			ID:             item.ID,
			IngredientID:   item.IngredientID,
			IngredientName: item.IngredientName,
			Category:       item.Category,
			Unit:           item.Unit,
			ForecastQty:    item.ForecastQty,
			SafetyStockQty: item.SafetyStockQty,
			CurrentStock:   item.CurrentStock,
			PurchaseQty:    item.PurchaseQty,
			ReceivedQty:    item.ReceivedQty,
			UnitPrice:      item.UnitPrice,
			Subtotal:       item.Subtotal,
		})
	}

	return &dto.PurchaseOrderV2Response{
		ID:               p.ID,
		StoreID:          p.StoreID,
		StoreName:        storeName,
		SupplierID:       p.SupplierID,
		PurchaseNo:       p.PurchaseNo,
		SupplierName:     p.SupplierName,
		SupplierPhone:    p.SupplierPhone,
		SupplierEmail:    p.SupplierEmail,
		TotalAmount:      p.TotalAmount,
		ReceivedAmount:   p.ReceivedAmount,
		TotalQuantity:    p.TotalQuantity,
		ReceivedQuantity: p.ReceivedQuantity,
		ItemCount:        p.ItemCount,
		Status:           p.Status,
		StatusText:       statusText,
		ForecastDate:     p.ForecastDate,
		ForecastDays:     p.ForecastDays,
		PaymentTerm:      p.PaymentTerm,
		PaymentTermText:  paymentTermText,
		ExpectedDate:     p.ExpectedDate,
		Remark:           p.Remark,
		SentAt:           p.SentAt,
		CreatedAt:        p.CreatedAt,
		Items:            items,
	}
}

func GetPaymentTermMap() map[int]string {
	return paymentTermMap
}

func GetSettlementMethodMap() map[string]string {
	return settlementMethodMap
}

func GetPayableStatusMap() map[string]string {
	return payableStatusMap
}

func GetReconcileStatusMap() map[string]string {
	return reconcileStatusMap
}

func ParseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
