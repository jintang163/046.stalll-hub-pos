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

type AccountsPayableService struct {
	payableRepo  *repository.AccountsPayableRepository
	paymentRepo  *repository.PayablePaymentRepository
	reconcileRepo *repository.ReconciliationRepository
	supplierRepo *repository.SupplierRepository
}

func NewAccountsPayableService() *AccountsPayableService {
	return &AccountsPayableService{
		payableRepo:   repository.NewAccountsPayableRepository(nil),
		paymentRepo:   repository.NewPayablePaymentRepository(nil),
		reconcileRepo: repository.NewReconciliationRepository(nil),
		supplierRepo:  repository.NewSupplierRepository(nil),
	}
}

func (s *AccountsPayableService) generatePayableNo(storeID uint) string {
	return fmt.Sprintf("AP%s%04d", time.Now().Format("20060102"), storeID%10000)
}

func (s *AccountsPayableService) generatePaymentNo(storeID uint) string {
	return fmt.Sprintf("PAY%s%04d", time.Now().Format("20060102150405"), storeID%10000)
}

func (s *AccountsPayableService) generateReconcileNo(storeID uint) string {
	return fmt.Sprintf("REC%s%04d", time.Now().Format("20060102"), storeID%10000)
}

func (s *AccountsPayableService) CreatePayableFromPurchase(purchase *model.PurchaseOrder) error {
	existing, _ := s.payableRepo.GetByBusinessID("purchase", purchase.ID)
	if existing != nil {
		return nil
	}

	var paymentTermDays int
	if purchase.SupplierID > 0 {
		supplier, err := s.supplierRepo.GetByID(purchase.SupplierID)
		if err == nil && supplier.PaymentTerm > 0 {
			paymentTermDays = supplier.PaymentTerm
		}
	}
	if purchase.PaymentTerm > 0 {
		paymentTermDays = purchase.PaymentTerm
	}

	var dueDate string
	if paymentTermDays > 0 {
		dueDate = time.Now().AddDate(0, 0, paymentTermDays).Format("2006-01-02")
	} else {
		dueDate = purchase.CreatedAt.Format("2006-01-02")
	}

	payable := &model.AccountsPayable{
		StoreID:      purchase.StoreID,
		SupplierID:   purchase.SupplierID,
		SupplierName: purchase.SupplierName,
		PayableNo:    s.generatePayableNo(purchase.StoreID),
		BusinessType: "purchase",
		BusinessID:   purchase.ID,
		BusinessNo:   purchase.PurchaseNo,
		Amount:       purchase.TotalAmount,
		PaidAmount:   decimal.Zero,
		Balance:      purchase.TotalAmount,
		DueDate:      dueDate,
		Status:       "unpaid",
		IsOverdue:    0,
		Remark:       "采购订单自动生成应付账款",
	}

	if err := s.payableRepo.Create(payable); err != nil {
		return fmt.Errorf("create payable failed: %w", err)
	}

	return nil
}

func (s *AccountsPayableService) GetPayable(id uint) (*dto.PayableResponse, error) {
	payable, err := s.payableRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToPayableResponse(payable), nil
}

func (s *AccountsPayableService) ListPayables(query *dto.PayableQueryDTO) (*dto.PageResponse, error) {
	s.UpdateOverdueStatus()

	payables, total, err := s.payableRepo.List(query)
	if err != nil {
		return nil, err
	}

	var list []dto.PayableResponse
	for _, p := range payables {
		list = append(list, *s.convertToPayableResponse(&p))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *AccountsPayableService) GetPayableStats(query *dto.PayableQueryDTO) (*dto.PayableStatsResponse, error) {
	s.UpdateOverdueStatus()
	return s.payableRepo.GetStats(query)
}

func (s *AccountsPayableService) UpdateOverdueStatus() {
	count, err := s.payableRepo.UpdateOverdueStatus()
	if err != nil {
		log.Printf("[AccountsPayableService] Update overdue status failed: %v", err)
	} else if count > 0 {
		log.Printf("[AccountsPayableService] Updated %d overdue payables", count)
	}
}

func (s *AccountsPayableService) CreatePayment(req *dto.PayablePaymentCreateDTO) (*dto.PayablePaymentResponse, error) {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	paymentNo := s.generatePaymentNo(req.StoreID)
	paymentDate := req.PaymentDate
	if paymentDate == "" {
		paymentDate = time.Now().Format("2006-01-02")
	}

	if req.PayableID == 0 {
		unpaidPayables, err := s.payableRepo.GetUnpaidBySupplier(req.StoreID, req.SupplierID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("get unpaid payables failed: %w", err)
		}

		remaining := req.Amount
		for _, payable := range unpaidPayables {
			if remaining.LessThanOrEqual(decimal.Zero) {
				break
			}

			applyAmount := decimal.Min(remaining, payable.Balance)

			payment := &model.PayablePayment{
				StoreID:       req.StoreID,
				SupplierID:    req.SupplierID,
				SupplierName:  payable.SupplierName,
				PayableID:     payable.ID,
				PaymentNo:     paymentNo,
				Amount:        applyAmount,
				PaymentMethod: req.PaymentMethod,
				PaymentDate:   paymentDate,
				OperatorID:    req.OperatorID,
				OperatorName:  req.OperatorName,
				VoucherNo:     req.VoucherNo,
				VoucherURL:    req.VoucherURL,
				Remark:        req.Remark,
			}

			if err := tx.Create(payment).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("create payment failed: %w", err)
			}

			newPaidAmount := payable.PaidAmount.Add(applyAmount)
			newBalance := payable.Balance.Sub(applyAmount)
			newStatus := "partial"
			now := time.Now()
			var paidAt *time.Time
			if newBalance.LessThanOrEqual(decimal.Zero) {
				newStatus = "paid"
				paidAt = &now
			}

			updates := map[string]interface{}{
				"paid_amount": newPaidAmount,
				"balance":     decimal.Max(decimal.Zero, newBalance),
				"status":      newStatus,
			}
			if paidAt != nil {
				updates["paid_at"] = paidAt
			}

			if err := tx.Model(&model.AccountsPayable{}).Where("id = ?", payable.ID).Updates(updates).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("update payable failed: %w", err)
			}

			remaining = remaining.Sub(applyAmount)
		}
	} else {
		payable, err := s.payableRepo.GetByID(req.PayableID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("payable not found: %w", err)
		}

		if payable.Balance.LessThan(req.Amount) {
			tx.Rollback()
			return nil, fmt.Errorf("payment amount exceeds balance")
		}

		payment := &model.PayablePayment{
			StoreID:       req.StoreID,
			SupplierID:    req.SupplierID,
			SupplierName:  payable.SupplierName,
			PayableID:     req.PayableID,
			PaymentNo:     paymentNo,
			Amount:        req.Amount,
			PaymentMethod: req.PaymentMethod,
			PaymentDate:   paymentDate,
			OperatorID:    req.OperatorID,
			OperatorName:  req.OperatorName,
			VoucherNo:     req.VoucherNo,
			VoucherURL:    req.VoucherURL,
			Remark:        req.Remark,
		}

		if err := tx.Create(payment).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("create payment failed: %w", err)
		}

		newPaidAmount := payable.PaidAmount.Add(req.Amount)
		newBalance := payable.Balance.Sub(req.Amount)
		newStatus := "partial"
		now := time.Now()
		var paidAt *time.Time
		if newBalance.LessThanOrEqual(decimal.Zero) {
			newStatus = "paid"
			paidAt = &now
		}

		updates := map[string]interface{}{
			"paid_amount": newPaidAmount,
			"balance":     decimal.Max(decimal.Zero, newBalance),
			"status":      newStatus,
		}
		if paidAt != nil {
			updates["paid_at"] = paidAt
		}

		if err := tx.Model(&model.AccountsPayable{}).Where("id = ?", payable.ID).Updates(updates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("update payable failed: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if req.SupplierID > 0 {
		go s.supplierRepo.UpdateStats(req.SupplierID)
	}

	payment, err := s.paymentRepo.GetByID(0)
	if err != nil {
		payments, _, _ := s.paymentRepo.List(&dto.PayablePaymentQueryDTO{
			StoreID:   req.StoreID,
			Page:      1,
			PageSize:  1,
		})
		if len(payments) > 0 {
			payment = &payments[0]
		}
	}
	if payment != nil {
		return s.convertToPaymentResponse(payment), nil
	}

	return &dto.PayablePaymentResponse{
		PaymentNo:         paymentNo,
		Amount:            req.Amount,
		PaymentMethod:     req.PaymentMethod,
		PaymentMethodText: settlementMethodMap[req.PaymentMethod],
		PaymentDate:       paymentDate,
	}, nil
}

func (s *AccountsPayableService) ListPayments(query *dto.PayablePaymentQueryDTO) (*dto.PageResponse, error) {
	payments, total, err := s.paymentRepo.List(query)
	if err != nil {
		return nil, err
	}

	var list []dto.PayablePaymentResponse
	for _, p := range payments {
		list = append(list, *s.convertToPaymentResponse(&p))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *AccountsPayableService) CreateReconciliation(req *dto.ReconciliationCreateDTO) (*dto.ReconciliationResponse, error) {
	supplier, err := s.supplierRepo.GetByID(req.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	reconcileNo := s.generateReconcileNo(req.StoreID)

	purchases, err := s.reconcileRepo.GetSupplierPeriodPurchases(req.StoreID, req.SupplierID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("get period purchases failed: %w", err)
	}

	payments, err := s.reconcileRepo.GetSupplierPeriodPayments(req.StoreID, req.SupplierID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("get period payments failed: %w", err)
	}

	var systemAmount decimal.Decimal
	var items []model.ReconciliationItem

	for _, p := range purchases {
		systemAmount = systemAmount.Add(p.TotalAmount)
		items = append(items, model.ReconciliationItem{
			BusinessType:  "purchase",
			BusinessID:    p.ID,
			BusinessNo:    p.PurchaseNo,
			BusinessDate:  p.CreatedAt.Format("2006-01-02"),
			SystemAmount:  p.TotalAmount,
			SupplierAmount: p.TotalAmount,
			DiffAmount:    decimal.Zero,
		})
	}

	for _, p := range payments {
		systemAmount = systemAmount.Sub(p.Amount)
		items = append(items, model.ReconciliationItem{
			BusinessType:  "payment",
			BusinessID:    p.ID,
			BusinessNo:    p.PaymentNo,
			BusinessDate:  p.PaymentDate,
			SystemAmount:  p.Amount.Neg(),
			SupplierAmount: p.Amount.Neg(),
			DiffAmount:    decimal.Zero,
		})
	}

	reconcile := &model.Reconciliation{
		StoreID:        req.StoreID,
		SupplierID:     req.SupplierID,
		SupplierName:   supplier.Name,
		ReconcileNo:    reconcileNo,
		PeriodStart:    req.PeriodStart,
		PeriodEnd:      req.PeriodEnd,
		SystemAmount:   systemAmount,
		SupplierAmount: decimal.Zero,
		DiffAmount:     decimal.Zero,
		Status:         "draft",
		Remark:         req.Remark,
		Items:          items,
	}

	if err := s.reconcileRepo.Create(reconcile); err != nil {
		return nil, fmt.Errorf("create reconciliation failed: %w", err)
	}

	return s.GetReconciliation(reconcile.ID)
}

func (s *AccountsPayableService) GetReconciliation(id uint) (*dto.ReconciliationResponse, error) {
	reconcile, err := s.reconcileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToReconcileResponse(reconcile), nil
}

func (s *AccountsPayableService) ListReconciliations(query *dto.ReconciliationQueryDTO) (*dto.PageResponse, error) {
	reconciles, total, err := s.reconcileRepo.List(query)
	if err != nil {
		return nil, err
	}

	var list []dto.ReconciliationResponse
	for _, r := range reconciles {
		list = append(list, *s.convertToReconcileResponse(&r))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *AccountsPayableService) ConfirmReconciliation(id uint, req *dto.ReconciliationConfirmDTO) (*dto.ReconciliationResponse, error) {
	reconcile, err := s.reconcileRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("reconciliation not found: %w", err)
	}

	if reconcile.Status == "confirmed" {
		return nil, fmt.Errorf("reconciliation already confirmed")
	}

	now := time.Now()
	reconcile.SupplierAmount = req.SupplierAmount
	reconcile.DiffAmount = reconcile.SystemAmount.Sub(req.SupplierAmount)
	reconcile.Status = "confirmed"
	reconcile.ConfirmedAt = &now
	reconcile.ConfirmedBy = req.ConfirmedBy
	reconcile.Remark = req.Remark

	if err := s.reconcileRepo.Update(reconcile); err != nil {
		return nil, fmt.Errorf("confirm reconciliation failed: %w", err)
	}

	return s.GetReconciliation(id)
}

func (s *AccountsPayableService) InputSupplierAmount(id uint, req *dto.ReconciliationSupplierAmountDTO) (*dto.ReconciliationResponse, error) {
	reconcile, err := s.reconcileRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("reconciliation not found: %w", err)
	}

	reconcile.SupplierAmount = req.SupplierAmount
	reconcile.DiffAmount = reconcile.SystemAmount.Sub(req.SupplierAmount)
	reconcile.DifferenceRemark = req.DifferenceRemark

	if reconcile.DiffAmount.IsZero() {
		reconcile.Status = "matched"
	} else if reconcile.Status != "confirmed" {
		reconcile.Status = "pending"
	}

	if err := s.reconcileRepo.Update(reconcile); err != nil {
		return nil, fmt.Errorf("input supplier amount failed: %w", err)
	}

	return s.GetReconciliation(id)
}

func (s *AccountsPayableService) convertToPayableResponse(p *model.AccountsPayable) *dto.PayableResponse {
	storeName := ""
	if p.Store.ID > 0 {
		storeName = p.Store.Name
	}

	statusText := payableStatusMap[p.Status]
	if statusText == "" {
		statusText = p.Status
	}

	var payments []dto.PayablePaymentResponse
	for _, pm := range p.Payments {
		payments = append(payments, *s.convertToPaymentResponse(&pm))
	}

	return &dto.PayableResponse{
		ID:           p.ID,
		StoreID:      p.StoreID,
		StoreName:    storeName,
		SupplierID:   p.SupplierID,
		SupplierName: p.SupplierName,
		PayableNo:    p.PayableNo,
		BusinessType: p.BusinessType,
		BusinessID:   p.BusinessID,
		BusinessNo:   p.BusinessNo,
		Amount:       p.Amount,
		PaidAmount:   p.PaidAmount,
		Balance:      p.Balance,
		DueDate:      p.DueDate,
		Status:       p.Status,
		StatusText:   statusText,
		IsOverdue:    p.IsOverdue,
		Remark:       p.Remark,
		PaidAt:       p.PaidAt,
		CreatedAt:    p.CreatedAt,
		Payments:     payments,
	}
}

func (s *AccountsPayableService) convertToPaymentResponse(p *model.PayablePayment) *dto.PayablePaymentResponse {
	methodText := settlementMethodMap[p.PaymentMethod]
	if methodText == "" {
		methodText = p.PaymentMethod
	}

	return &dto.PayablePaymentResponse{
		ID:                p.ID,
		StoreID:           p.StoreID,
		SupplierID:        p.SupplierID,
		SupplierName:      p.SupplierName,
		PayableID:         p.PayableID,
		PaymentNo:         p.PaymentNo,
		Amount:            p.Amount,
		PaymentMethod:     p.PaymentMethod,
		PaymentMethodText: methodText,
		PaymentDate:       p.PaymentDate,
		OperatorID:        p.OperatorID,
		OperatorName:      p.OperatorName,
		VoucherNo:         p.VoucherNo,
		VoucherURL:        p.VoucherURL,
		Remark:            p.Remark,
		CreatedAt:         p.CreatedAt,
	}
}

func (s *AccountsPayableService) convertToReconcileResponse(r *model.Reconciliation) *dto.ReconciliationResponse {
	storeName := ""
	if r.Store.ID > 0 {
		storeName = r.Store.Name
	}

	statusText := reconcileStatusMap[r.Status]
	if statusText == "" {
		statusText = r.Status
	}

	var items []dto.ReconciliationItemResponse
	for _, item := range r.Items {
		bTypeText := businessTypeMap[item.BusinessType]
		if bTypeText == "" {
			bTypeText = item.BusinessType
		}
		items = append(items, dto.ReconciliationItemResponse{
			ID:               item.ID,
			ReconcileID:      item.ReconcileID,
			BusinessType:     item.BusinessType,
			BusinessTypeText: bTypeText,
			BusinessID:       item.BusinessID,
			BusinessNo:       item.BusinessNo,
			BusinessDate:     item.BusinessDate,
			SystemAmount:     item.SystemAmount,
			SupplierAmount:   item.SupplierAmount,
			DiffAmount:       item.DiffAmount,
			Remark:           item.Remark,
		})
	}

	return &dto.ReconciliationResponse{
		ID:             r.ID,
		StoreID:        r.StoreID,
		StoreName:      storeName,
		SupplierID:     r.SupplierID,
		SupplierName:   r.SupplierName,
		ReconcileNo:    r.ReconcileNo,
		PeriodStart:    r.PeriodStart,
		PeriodEnd:      r.PeriodEnd,
		SystemAmount:   r.SystemAmount,
		SupplierAmount: r.SupplierAmount,
		DiffAmount:     r.DiffAmount,
		Status:         r.Status,
		StatusText:     statusText,
		ConfirmedAt:    r.ConfirmedAt,
		ConfirmedBy:    r.ConfirmedBy,
		Remark:         r.Remark,
		CreatedAt:      r.CreatedAt,
		Items:          items,
	}
}
