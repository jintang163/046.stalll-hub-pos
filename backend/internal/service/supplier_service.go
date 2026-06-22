package service

import (
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
)

var paymentTermMap = map[int]string{
	0:  "货到付款",
	7:  "7天账期",
	15: "15天账期",
	30: "月结30天",
	45: "月结45天",
	60: "月结60天",
	90: "月结90天",
}

var settlementMethodMap = map[string]string{
	"bank_transfer": "银行转账",
	"cash":          "现金",
	"check":         "支票",
	"alipay":        "支付宝",
	"wechat":        "微信",
	"other":         "其他",
}

var payableStatusMap = map[string]string{
	"unpaid":   "未付款",
	"partial":  "部分付款",
	"paid":     "已付清",
	"overdue":  "已逾期",
	"disputed": "有争议",
}

var reconcileStatusMap = map[string]string{
	"draft":     "草稿",
	"pending":   "待确认",
	"confirmed": "已确认",
	"rejected":  "已驳回",
}

var businessTypeMap = map[string]string{
	"purchase": "采购订单",
	"return":   "采购退货",
	"adjust":   "金额调整",
	"other":    "其他",
}

type SupplierService struct {
	supplierRepo *repository.SupplierRepository
	smsService   *SmsService
}

func NewSupplierService() *SupplierService {
	return &SupplierService{
		supplierRepo: repository.NewSupplierRepository(nil),
		smsService:   NewSmsService(),
	}
}

func (s *SupplierService) generateSupplierNo(storeID uint) string {
	return fmt.Sprintf("SUP%s%04d", time.Now().Format("20060102"), storeID%10000)
}

func (s *SupplierService) CreateSupplier(req *dto.SupplierCreateDTO) (*dto.SupplierResponse, error) {
	supplierNo := s.generateSupplierNo(req.StoreID)

	supplier := &model.Supplier{
		StoreID:          req.StoreID,
		SupplierNo:       supplierNo,
		Name:             req.Name,
		ShortName:        req.ShortName,
		Category:         req.Category,
		ContactPerson:    req.ContactPerson,
		Phone:            req.Phone,
		Mobile:           req.Mobile,
		Email:            req.Email,
		Fax:              req.Fax,
		Address:          req.Address,
		Province:         req.Province,
		City:             req.City,
		District:         req.District,
		BankName:         req.BankName,
		BankAccount:      req.BankAccount,
		BankAccountName:  req.BankAccountName,
		TaxNo:            req.TaxNo,
		PaymentTerm:      req.PaymentTerm,
		PaymentTermDesc:  req.PaymentTermDesc,
		SettlementMethod: req.SettlementMethod,
		CreditLimit:      req.CreditLimit,
		Status:           req.Status,
		Remark:           req.Remark,
	}

	if supplier.Status == 0 {
		supplier.Status = 1
	}
	if supplier.SettlementMethod == "" {
		supplier.SettlementMethod = "bank_transfer"
	}

	if err := s.supplierRepo.Create(supplier); err != nil {
		return nil, fmt.Errorf("create supplier failed: %w", err)
	}

	return s.GetSupplier(supplier.ID)
}

func (s *SupplierService) UpdateSupplier(id uint, req *dto.SupplierUpdateDTO) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	if req.Name != "" {
		supplier.Name = req.Name
	}
	supplier.ShortName = req.ShortName
	supplier.Category = req.Category
	supplier.ContactPerson = req.ContactPerson
	supplier.Phone = req.Phone
	supplier.Mobile = req.Mobile
	supplier.Email = req.Email
	supplier.Fax = req.Fax
	supplier.Address = req.Address
	supplier.Province = req.Province
	supplier.City = req.City
	supplier.District = req.District
	supplier.BankName = req.BankName
	supplier.BankAccount = req.BankAccount
	supplier.BankAccountName = req.BankAccountName
	supplier.TaxNo = req.TaxNo
	supplier.PaymentTerm = req.PaymentTerm
	supplier.PaymentTermDesc = req.PaymentTermDesc
	supplier.SettlementMethod = req.SettlementMethod
	supplier.CreditLimit = req.CreditLimit
	supplier.Status = req.Status
	supplier.Remark = req.Remark

	if err := s.supplierRepo.Update(supplier); err != nil {
		return nil, fmt.Errorf("update supplier failed: %w", err)
	}

	return s.GetSupplier(id)
}

func (s *SupplierService) DeleteSupplier(id uint) error {
	_, err := s.supplierRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("supplier not found: %w", err)
	}
	return s.supplierRepo.Delete(id)
}

func (s *SupplierService) GetSupplier(id uint) (*dto.SupplierResponse, error) {
	supplier, err := s.supplierRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToSupplierResponse(supplier), nil
}

func (s *SupplierService) ListSuppliers(query *dto.SupplierQueryDTO) (*dto.PageResponse, error) {
	suppliers, total, err := s.supplierRepo.List(query)
	if err != nil {
		return nil, err
	}

	var list []dto.SupplierResponse
	for _, sup := range suppliers {
		list = append(list, *s.convertToSupplierResponse(&sup))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *SupplierService) GetSupplierCategories(storeID uint) ([]string, error) {
	return s.supplierRepo.GetCategories(storeID)
}

func (s *SupplierService) GetSupplierStats(storeID uint) (*dto.SupplierStatsResponse, error) {
	return s.supplierRepo.GetStats(storeID)
}

func (s *SupplierService) NotifySupplier(id uint, notifyTypes []string, content string) error {
	supplier, err := s.supplierRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("supplier not found: %w", err)
	}

	for _, notifyType := range notifyTypes {
		switch notifyType {
		case "sms":
			if supplier.Mobile != "" {
				go func() {
					err := s.smsService.SendSms(supplier.StoreID, 0, supplier.Mobile,
						"供应商通知", "SUPPLIER_NOTIFY", content, 0)
					if err != nil {
						log.Printf("[SupplierService] Send SMS to supplier %s failed: %v", supplier.Name, err)
					}
				}()
			}
		case "email":
			if supplier.Email != "" {
				log.Printf("[SupplierService] Would send email to %s (%s): %s", supplier.Name, supplier.Email, content)
			}
		}
	}

	return nil
}

func (s *SupplierService) convertToSupplierResponse(sup *model.Supplier) *dto.SupplierResponse {
	storeName := ""
	if sup.Store.ID > 0 {
		storeName = sup.Store.Name
	}

	statusText := "禁用"
	if sup.Status == 1 {
		statusText = "启用"
	}

	return &dto.SupplierResponse{
		ID:               sup.ID,
		StoreID:          sup.StoreID,
		StoreName:        storeName,
		SupplierNo:       sup.SupplierNo,
		Name:             sup.Name,
		ShortName:        sup.ShortName,
		Category:         sup.Category,
		ContactPerson:    sup.ContactPerson,
		Phone:            sup.Phone,
		Mobile:           sup.Mobile,
		Email:            sup.Email,
		Fax:              sup.Fax,
		Address:          sup.Address,
		Province:         sup.Province,
		City:             sup.City,
		District:         sup.District,
		BankName:         sup.BankName,
		BankAccount:      sup.BankAccount,
		BankAccountName:  sup.BankAccountName,
		TaxNo:            sup.TaxNo,
		PaymentTerm:      sup.PaymentTerm,
		PaymentTermDesc:  sup.PaymentTermDesc,
		SettlementMethod: sup.SettlementMethod,
		CreditLimit:      sup.CreditLimit,
		CurrentPayable:   sup.CurrentPayable,
		TotalPurchase:    sup.TotalPurchase,
		TotalPaid:        sup.TotalPaid,
		Status:           sup.Status,
		StatusText:       statusText,
		Remark:           sup.Remark,
		CreatedAt:        sup.CreatedAt,
		UpdatedAt:        sup.UpdatedAt,
	}
}

type PurchaseReceiveService struct {
	receiveRepo  *repository.PurchaseReceiveRepository
	supplierRepo *repository.SupplierRepository
	purchaseSvc  *PurchaseService
}

func NewPurchaseReceiveService() *PurchaseReceiveService {
	return &PurchaseReceiveService{
		receiveRepo:  repository.NewPurchaseReceiveRepository(nil),
		supplierRepo: repository.NewSupplierRepository(nil),
		purchaseSvc:  NewPurchaseService(),
	}
}

func (s *PurchaseReceiveService) generateReceiveNo(storeID uint) string {
	return fmt.Sprintf("REC%s%04d", time.Now().Format("20060102150405"), storeID%10000)
}

func (s *PurchaseReceiveService) CreateReceive(req *dto.PurchaseReceiveCreateDTO) (*dto.PurchaseReceiveResponse, error) {
	purchase, err := s.purchaseSvc.GetPurchaseOrder(req.PurchaseID)
	if err != nil {
		return nil, fmt.Errorf("purchase order not found: %w", err)
	}

	if purchase.Status == 4 || purchase.Status == 5 {
		return nil, fmt.Errorf("purchase order is completed or cancelled")
	}

	receiveNo := s.generateReceiveNo(req.StoreID)
	now := time.Now()

	var totalQty decimal.Decimal
	var totalAmount decimal.Decimal
	var items []model.PurchaseReceiveItem

	for idx, item := range req.Items {
		subtotal := item.ReceivedQty.Mul(item.UnitPrice)
		totalQty = totalQty.Add(item.ReceivedQty)
		totalAmount = totalAmount.Add(subtotal)

		qualifiedQty := item.QualifiedQty
		if qualifiedQty.IsZero() {
			qualifiedQty = item.ReceivedQty
		}

		items = append(items, model.PurchaseReceiveItem{
			PurchaseItemID: item.PurchaseItemID,
			IngredientID:   item.IngredientID,
			IngredientName: item.IngredientName,
			Category:       item.Category,
			Unit:           item.Unit,
			PurchaseQty:    item.PurchaseQty,
			ReceivedQty:    item.ReceivedQty,
			QualifiedQty:   qualifiedQty,
			RejectedQty:    item.RejectedQty,
			UnitPrice:      item.UnitPrice,
			Subtotal:       subtotal,
			BatchNo:        item.BatchNo,
			ExpiryDate:     item.ExpiryDate,
			RejectReason:   item.RejectReason,
			SortOrder:      idx,
		})
	}

	receive := &model.PurchaseReceive{
		StoreID:      req.StoreID,
		PurchaseID:   req.PurchaseID,
		PurchaseNo:   purchase.PurchaseNo,
		SupplierID:   purchase.SupplierID,
		SupplierName: purchase.SupplierName,
		ReceiveNo:    receiveNo,
		ReceiveType:  req.ReceiveType,
		TotalQty:     totalQty,
		TotalAmount:  totalAmount,
		Remark:       req.Remark,
		OperatorID:   req.OperatorID,
		OperatorName: req.OperatorName,
		ReceivedAt:   &now,
		Items:        items,
	}

	if receive.ReceiveType == "" {
		receive.ReceiveType = "full"
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(receive).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("create receive failed: %w", err)
	}

	if err := s.updatePurchaseAfterReceive(tx, purchase, receive); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.updateIngredientStock(tx, receive); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if purchase.SupplierID > 0 {
		go s.supplierRepo.UpdateStats(purchase.SupplierID)
	}

	return s.GetReceive(receive.ID)
}

func (s *PurchaseReceiveService) updatePurchaseAfterReceive(tx *gorm.DB, purchase *model.PurchaseOrder, receive *model.PurchaseReceive) error {
	var newReceivedQty decimal.Decimal
	var newReceivedAmount decimal.Decimal
	var newReceivedQuantity int
	allReceived := true

	for idx, purchaseItem := range purchase.Items {
		itemReceived := purchaseItem.ReceivedQty
		for _, recvItem := range receive.Items {
			if recvItem.IngredientID == purchaseItem.IngredientID {
				itemReceived = itemReceived.Add(recvItem.QualifiedQty)
			}
		}
		purchase.Items[idx].ReceivedQty = itemReceived

		newReceivedQty = newReceivedQty.Add(itemReceived)
		newReceivedAmount = newReceivedAmount.Add(itemReceived.Mul(purchaseItem.UnitPrice))

		if itemReceived.LessThan(purchaseItem.PurchaseQty) {
			allReceived = false
		}
	}

	newReceivedQuantity = int(newReceivedQty.IntPart())

	newStatus := purchase.Status
	if allReceived {
		newStatus = 4
	} else if newReceivedQty.GreaterThan(decimal.Zero) {
		newStatus = 3
	}

	updates := map[string]interface{}{
		"received_quantity": newReceivedQuantity,
		"received_amount":   newReceivedAmount,
		"status":            newStatus,
	}

	if err := tx.Model(&model.PurchaseOrder{}).Where("id = ?", purchase.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("update purchase order failed: %w", err)
	}

	for _, item := range purchase.Items {
		if err := tx.Model(&model.PurchaseOrderItem{}).Where("id = ?", item.ID).
			Update("received_qty", item.ReceivedQty).Error; err != nil {
			return fmt.Errorf("update purchase item failed: %w", err)
		}
	}

	return nil
}

func (s *PurchaseReceiveService) updateIngredientStock(tx *gorm.DB, receive *model.PurchaseReceive) error {
	for _, item := range receive.Items {
		if item.QualifiedQty.IsZero() {
			continue
		}

		result := tx.Model(&model.Ingredient{}).
			Where("id = ?", item.IngredientID).
			UpdateColumn("current_stock", gorm.Expr("current_stock + ?", item.QualifiedQty))

		if result.Error != nil {
			return fmt.Errorf("update ingredient stock failed: %w", result.Error)
		}

		log.Printf("[PurchaseReceiveService] Ingredient %s stock increased by %s",
			item.IngredientName, item.QualifiedQty.String())
	}

	return nil
}

func (s *PurchaseReceiveService) GetReceive(id uint) (*dto.PurchaseReceiveResponse, error) {
	receive, err := s.receiveRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToReceiveResponse(receive), nil
}

func (s *PurchaseReceiveService) ListReceives(query *dto.PurchaseReceiveQueryDTO) (*dto.PageResponse, error) {
	receives, total, err := s.receiveRepo.List(query)
	if err != nil {
		return nil, err
	}

	var list []dto.PurchaseReceiveResponse
	for _, r := range receives {
		list = append(list, *s.convertToReceiveResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *PurchaseReceiveService) convertToReceiveResponse(r *model.PurchaseReceive) *dto.PurchaseReceiveResponse {
	storeName := ""
	if r.Store.ID > 0 {
		storeName = r.Store.Name
	}

	var items []dto.PurchaseReceiveItemResponse
	for _, item := range r.Items {
		items = append(items, dto.PurchaseReceiveItemResponse{
			ID:             item.ID,
			ReceiveID:      item.ReceiveID,
			PurchaseItemID: item.PurchaseItemID,
			IngredientID:   item.IngredientID,
			IngredientName: item.IngredientName,
			Category:       item.Category,
			Unit:           item.Unit,
			PurchaseQty:    item.PurchaseQty,
			ReceivedQty:    item.ReceivedQty,
			QualifiedQty:   item.QualifiedQty,
			RejectedQty:    item.RejectedQty,
			UnitPrice:      item.UnitPrice,
			Subtotal:       item.Subtotal,
			BatchNo:        item.BatchNo,
			ExpiryDate:     item.ExpiryDate,
			RejectReason:   item.RejectReason,
		})
	}

	return &dto.PurchaseReceiveResponse{
		ID:           r.ID,
		StoreID:      r.StoreID,
		StoreName:    storeName,
		PurchaseID:   r.PurchaseID,
		PurchaseNo:   r.PurchaseNo,
		SupplierID:   r.SupplierID,
		SupplierName: r.SupplierName,
		ReceiveNo:    r.ReceiveNo,
		ReceiveType:  r.ReceiveType,
		TotalQty:     r.TotalQty,
		TotalAmount:  r.TotalAmount,
		Remark:       r.Remark,
		OperatorID:   r.OperatorID,
		OperatorName: r.OperatorName,
		ReceivedAt:   r.ReceivedAt,
		CreatedAt:    r.CreatedAt,
		Items:        items,
	}
}
