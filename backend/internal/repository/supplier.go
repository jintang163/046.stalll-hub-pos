package repository

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) Create(supplier *model.Supplier) error {
	return database.DB.Create(supplier).Error
}

func (r *SupplierRepository) Update(supplier *model.Supplier) error {
	return database.DB.Save(supplier).Error
}

func (r *SupplierRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Supplier{}, id).Error
}

func (r *SupplierRepository) GetByID(id uint) (*model.Supplier, error) {
	var supplier model.Supplier
	err := database.DB.Preload("Store").First(&supplier, id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *SupplierRepository) GetByNo(storeID uint, supplierNo string) (*model.Supplier, error) {
	var supplier model.Supplier
	err := database.DB.Where("store_id = ? AND supplier_no = ?", storeID, supplierNo).First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

func (r *SupplierRepository) List(query *dto.SupplierQueryDTO) ([]model.Supplier, int64, error) {
	var suppliers []model.Supplier
	var total int64

	db := database.DB.Model(&model.Supplier{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR supplier_no LIKE ? OR contact_person LIKE ? OR phone LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Store").
		Order("id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&suppliers).Error

	return suppliers, total, err
}

func (r *SupplierRepository) GetCategories(storeID uint) ([]string, error) {
	var categories []string
	db := database.DB.Model(&model.Supplier{})
	if storeID > 0 {
		db = db.Where("store_id = ?", storeID)
	}
	err := db.Distinct("category").Where("category <> ''").Pluck("category", &categories).Error
	return categories, err
}

func (r *SupplierRepository) UpdateStats(supplierID uint) error {
	type Stats struct {
		TotalPurchase decimal.Decimal
		TotalPaid     decimal.Decimal
	}
	var stats Stats

	database.DB.Model(&model.PurchaseOrder{}).
		Select("COALESCE(SUM(total_amount), 0) as total_purchase").
		Where("supplier_id = ? AND status IN (1,2,3,4)", supplierID).
		Scan(&stats)

	database.DB.Model(&model.PayablePayment{}).
		Select("COALESCE(SUM(amount), 0) as total_paid").
		Where("supplier_id = ?", supplierID).
		Scan(&stats)

	var currentPayable decimal.Decimal
	database.DB.Model(&model.AccountsPayable{}).
		Select("COALESCE(SUM(balance), 0)").
		Where("supplier_id = ? AND status <> 'paid'", supplierID).
		Scan(&currentPayable)

	return database.DB.Model(&model.Supplier{}).Where("id = ?", supplierID).Updates(map[string]interface{}{
		"total_purchase":  stats.TotalPurchase,
		"total_paid":      stats.TotalPaid,
		"current_payable": currentPayable,
	}).Error
}

func (r *SupplierRepository) GetStats(storeID uint) (*dto.SupplierStatsResponse, error) {
	var stats dto.SupplierStatsResponse

	database.DB.Model(&model.Supplier{}).
		Where("store_id = ?", storeID).
		Select("COUNT(*)").Scan(&stats.TotalSupplier)

	database.DB.Model(&model.Supplier{}).
		Where("store_id = ? AND status = 1", storeID).
		Select("COUNT(*)").Scan(&stats.ActiveSupplier)

	database.DB.Model(&model.AccountsPayable{}).
		Where("store_id = ?", storeID).
		Select("COALESCE(SUM(balance), 0)").Scan(&stats.TotalPayable)

	database.DB.Model(&model.AccountsPayable{}).
		Where("store_id = ? AND is_overdue = 1", storeID).
		Select("COALESCE(SUM(balance), 0)").Scan(&stats.OverduePayable)

	database.DB.Model(&model.PurchaseOrder{}).
		Where("store_id = ? AND status IN (1,2,3,4)", storeID).
		Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TotalPurchase)

	database.DB.Model(&model.PayablePayment{}).
		Where("store_id = ?", storeID).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalPaid)

	return &stats, nil
}

type PurchaseReceiveRepository struct {
	db *gorm.DB
}

func NewPurchaseReceiveRepository(db *gorm.DB) *PurchaseReceiveRepository {
	return &PurchaseReceiveRepository{db: db}
}

func (r *PurchaseReceiveRepository) Create(receive *model.PurchaseReceive) error {
	return database.DB.Create(receive).Error
}

func (r *PurchaseReceiveRepository) GetByID(id uint) (*model.PurchaseReceive, error) {
	var receive model.PurchaseReceive
	err := database.DB.Preload("Items").Preload("Store").Preload("Purchase").First(&receive, id).Error
	if err != nil {
		return nil, err
	}
	return &receive, nil
}

func (r *PurchaseReceiveRepository) List(query *dto.PurchaseReceiveQueryDTO) ([]model.PurchaseReceive, int64, error) {
	var receives []model.PurchaseReceive
	var total int64

	db := database.DB.Model(&model.PurchaseReceive{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.PurchaseID > 0 {
		db = db.Where("purchase_id = ?", query.PurchaseID)
	}
	if query.SupplierID > 0 {
		db = db.Where("supplier_id = ?", query.SupplierID)
	}
	if query.Keyword != "" {
		db = db.Where("receive_no LIKE ? OR purchase_no LIKE ? OR supplier_name LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.StartDate != "" {
		db = db.Where("DATE(created_at) >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("DATE(created_at) <= ?", query.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Store").
		Order("id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&receives).Error

	return receives, total, err
}

type AccountsPayableRepository struct {
	db *gorm.DB
}

func NewAccountsPayableRepository(db *gorm.DB) *AccountsPayableRepository {
	return &AccountsPayableRepository{db: db}
}

func (r *AccountsPayableRepository) Create(payable *model.AccountsPayable) error {
	return database.DB.Create(payable).Error
}

func (r *AccountsPayableRepository) Update(payable *model.AccountsPayable) error {
	return database.DB.Save(payable).Error
}

func (r *AccountsPayableRepository) GetByID(id uint) (*model.AccountsPayable, error) {
	var payable model.AccountsPayable
	err := database.DB.Preload("Supplier").Preload("Store").Preload("Payments").First(&payable, id).Error
	if err != nil {
		return nil, err
	}
	return &payable, nil
}

func (r *AccountsPayableRepository) GetByBusinessID(businessType string, businessID uint) (*model.AccountsPayable, error) {
	var payable model.AccountsPayable
	err := database.DB.Where("business_type = ? AND business_id = ?", businessType, businessID).First(&payable).Error
	if err != nil {
		return nil, err
	}
	return &payable, nil
}

func (r *AccountsPayableRepository) List(query *dto.PayableQueryDTO) ([]model.AccountsPayable, int64, error) {
	var payables []model.AccountsPayable
	var total int64

	db := database.DB.Model(&model.AccountsPayable{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.SupplierID > 0 {
		db = db.Where("supplier_id = ?", query.SupplierID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.IsOverdue > 0 {
		db = db.Where("is_overdue = ?", query.IsOverdue)
	}
	if query.Keyword != "" {
		db = db.Where("payable_no LIKE ? OR business_no LIKE ? OR supplier_name LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.StartDate != "" {
		db = db.Where("DATE(created_at) >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("DATE(created_at) <= ?", query.EndDate)
	}
	if query.DueBefore != "" {
		db = db.Where("due_date <= ? AND status <> 'paid'", query.DueBefore)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Supplier").Preload("Store").
		Order("is_overdue DESC, due_date ASC, id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&payables).Error

	return payables, total, err
}

func (r *AccountsPayableRepository) UpdateOverdueStatus() (int64, error) {
	today := time.Now().Format("2006-01-02")
	result := database.DB.Model(&model.AccountsPayable{}).
		Where("due_date < ? AND status = 'unpaid'", today).
		Update("is_overdue", 1)
	return result.RowsAffected, result.Error
}

func (r *AccountsPayableRepository) GetStats(query *dto.PayableQueryDTO) (*dto.PayableStatsResponse, error) {
	var stats dto.PayableStatsResponse

	db := database.DB.Model(&model.AccountsPayable{})
	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.SupplierID > 0 {
		db = db.Where("supplier_id = ?", query.SupplierID)
	}

	db.Count(&stats.TotalCount)
	db.Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalAmount)
	db.Select("COALESCE(SUM(paid_amount), 0)").Scan(&stats.TotalPaid)
	db.Select("COALESCE(SUM(balance), 0)").Scan(&stats.TotalBalance)

	overdueDB := database.DB.Model(&model.AccountsPayable{}).Where("is_overdue = 1")
	if query.StoreID > 0 {
		overdueDB = overdueDB.Where("store_id = ?", query.StoreID)
	}
	if query.SupplierID > 0 {
		overdueDB = overdueDB.Where("supplier_id = ?", query.SupplierID)
	}
	overdueDB.Count(&stats.OverdueCount)
	overdueDB.Select("COALESCE(SUM(balance), 0)").Scan(&stats.OverdueAmount)

	return &stats, nil
}

func (r *AccountsPayableRepository) GetUnpaidBySupplier(storeID, supplierID uint) ([]model.AccountsPayable, error) {
	var payables []model.AccountsPayable
	err := database.DB.Where("store_id = ? AND supplier_id = ? AND status <> 'paid'", storeID, supplierID).
		Order("due_date ASC").Find(&payables).Error
	return payables, err
}

type PayablePaymentRepository struct {
	db *gorm.DB
}

func NewPayablePaymentRepository(db *gorm.DB) *PayablePaymentRepository {
	return &PayablePaymentRepository{db: db}
}

func (r *PayablePaymentRepository) Create(payment *model.PayablePayment) error {
	return database.DB.Create(payment).Error
}

func (r *PayablePaymentRepository) GetByID(id uint) (*model.PayablePayment, error) {
	var payment model.PayablePayment
	err := database.DB.Preload("Payable").First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PayablePaymentRepository) List(query *dto.PayablePaymentQueryDTO) ([]model.PayablePayment, int64, error) {
	var payments []model.PayablePayment
	var total int64

	db := database.DB.Model(&model.PayablePayment{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.SupplierID > 0 {
		db = db.Where("supplier_id = ?", query.SupplierID)
	}
	if query.PayableID > 0 {
		db = db.Where("payable_id = ?", query.PayableID)
	}
	if query.Keyword != "" {
		db = db.Where("payment_no LIKE ? OR voucher_no LIKE ? OR supplier_name LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.StartDate != "" {
		db = db.Where("payment_date >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("payment_date <= ?", query.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Payable").
		Order("payment_date DESC, id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&payments).Error

	return payments, total, err
}

type ReconciliationRepository struct {
	db *gorm.DB
}

func NewReconciliationRepository(db *gorm.DB) *ReconciliationRepository {
	return &ReconciliationRepository{db: db}
}

func (r *ReconciliationRepository) Create(reconcile *model.Reconciliation) error {
	return database.DB.Create(reconcile).Error
}

func (r *ReconciliationRepository) Update(reconcile *model.Reconciliation) error {
	return database.DB.Save(reconcile).Error
}

func (r *ReconciliationRepository) GetByID(id uint) (*model.Reconciliation, error) {
	var reconcile model.Reconciliation
	err := database.DB.Preload("Items").Preload("Supplier").Preload("Store").First(&reconcile, id).Error
	if err != nil {
		return nil, err
	}
	return &reconcile, nil
}

func (r *ReconciliationRepository) List(query *dto.ReconciliationQueryDTO) ([]model.Reconciliation, int64, error) {
	var reconciles []model.Reconciliation
	var total int64

	db := database.DB.Model(&model.Reconciliation{})

	if query.StoreID > 0 {
		db = db.Where("store_id = ?", query.StoreID)
	}
	if query.SupplierID > 0 {
		db = db.Where("supplier_id = ?", query.SupplierID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		db = db.Where("reconcile_no LIKE ? OR supplier_name LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.StartDate != "" {
		db = db.Where("period_end >= ?", query.StartDate)
	}
	if query.EndDate != "" {
		db = db.Where("period_start <= ?", query.EndDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Supplier").Preload("Store").
		Order("id DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&reconciles).Error

	return reconciles, total, err
}

func (r *ReconciliationRepository) GetSupplierPeriodPurchases(storeID, supplierID uint, periodStart, periodEnd string) ([]model.PurchaseOrder, error) {
	var purchases []model.PurchaseOrder
	err := database.DB.Preload("Items").
		Where("store_id = ? AND supplier_id = ? AND status IN (1,2,3,4) AND DATE(created_at) >= ? AND DATE(created_at) <= ?",
			storeID, supplierID, periodStart, periodEnd).
		Order("created_at ASC").
		Find(&purchases).Error
	return purchases, err
}

func (r *ReconciliationRepository) GetSupplierPeriodPayments(storeID, supplierID uint, periodStart, periodEnd string) ([]model.PayablePayment, error) {
	var payments []model.PayablePayment
	err := database.DB.
		Where("store_id = ? AND supplier_id = ? AND payment_date >= ? AND payment_date <= ?",
			storeID, supplierID, periodStart, periodEnd).
		Order("payment_date ASC").
		Find(&payments).Error
	return payments, err
}
