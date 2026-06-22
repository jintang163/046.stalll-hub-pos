package service

import (
	"errors"
	"fmt"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransferService struct {
	dingTalk   *DingTalkService
	kdnService *KuaiDiNiaoService
}

func NewTransferService() *TransferService {
	return &TransferService{
		dingTalk:   NewDingTalkService(),
		kdnService: NewKuaiDiNiaoService(),
	}
}

func (s *TransferService) CreateTransfer(req *dto.CreateTransferOrderDTO) (*model.TransferOrder, error) {
	if req.FromStoreID == req.ToStoreID {
		return nil, errors.New("调出门店和调入门店不能相同")
	}

	transferNo := generateTransferNo()

	var totalQty decimal.Decimal
	var totalAmount decimal.Decimal

	for _, item := range req.Items {
		if item.OutQty.LessThanOrEqual(decimal.Zero) {
			return nil, errors.New("调拨数量必须大于0")
		}
		totalQty = totalQty.Add(item.OutQty)
		totalAmount = totalAmount.Add(item.OutQty.Mul(item.UnitPrice))
	}

	transfer := &model.TransferOrder{
		TransferNo:      transferNo,
		FromStoreID:     req.FromStoreID,
		ToStoreID:       req.ToStoreID,
		Status:          model.TransferStatusPendingAccept,
		TotalQty:        totalQty,
		TotalAmount:     totalAmount,
		TransferType:    req.TransferType,
		Priority:        req.Priority,
		SenderName:      req.SenderName,
		SenderPhone:     req.SenderPhone,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		Remark:          req.Remark,
	}

	tx := database.DB.Begin()
	if err := tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range req.Items {
		var ingredient model.Ingredient
		if err := tx.Where("id = ? AND store_id = ?", item.IngredientID, req.FromStoreID).First(&ingredient).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("食材不存在: %d", item.IngredientID)
		}

		if ingredient.CurrentStock.LessThan(item.OutQty) {
			tx.Rollback()
			return nil, fmt.Errorf("食材[%s]库存不足，当前库存: %s", ingredient.Name, ingredient.CurrentStock.String())
		}

		transferItem := model.TransferOrderItem{
			TransferID:     transfer.ID,
			IngredientID:   item.IngredientID,
			IngredientNo:   ingredient.IngredientNo,
			IngredientName: ingredient.Name,
			Unit:           ingredient.Unit,
			OutQty:         item.OutQty,
			UnitPrice:      ingredient.CurrentPrice,
			Amount:         item.OutQty.Mul(ingredient.CurrentPrice),
			Remark:         item.Remark,
		}

		if err := tx.Create(&transferItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	transfer.Items, _ = s.getItems(transfer.ID)

	go s.dingTalk.SendText(fmt.Sprintf("收到新的调拨单[%s]，请及时确认接单", transfer.TransferNo), false, nil)

	return transfer, nil
}

func (s *TransferService) GetTransferByID(id uint) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	err := database.DB.
		Preload("FromStore").
		Preload("ToStore").
		Preload("Items").
		Preload("LogisticsTracks", func(db *gorm.DB) *gorm.DB {
			return db.Order("track_time DESC, id DESC")
		}).
		First(&transfer, id).Error
	return &transfer, err
}

func (s *TransferService) GetTransferByNo(transferNo string) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	err := database.DB.
		Preload("FromStore").
		Preload("ToStore").
		Preload("Items").
		Where("transfer_no = ?", transferNo).
		First(&transfer).Error
	return &transfer, err
}

func (s *TransferService) ListTransfers(query *dto.TransferOrderQueryDTO) ([]model.TransferOrder, int64, error) {
	var list []model.TransferOrder
	var total int64

	db := database.DB.Model(&model.TransferOrder{})

	if query.FromStoreID > 0 {
		db = db.Where("from_store_id = ?", query.FromStoreID)
	}
	if query.ToStoreID > 0 {
		db = db.Where("to_store_id = ?", query.ToStoreID)
	}
	if query.Status >= 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.TransferNo != "" {
		db = db.Where("transfer_no LIKE ?", "%"+query.TransferNo+"%")
	}
	if query.Keyword != "" {
		db = db.Where("transfer_no LIKE ? OR remark LIKE ?",
			"%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}

	db.Count(&total)

	err := db.Preload("FromStore").
		Preload("ToStore").
		Order("id DESC").
		Offset(query.GetOffset()).
		Limit(query.GetLimit()).
		Find(&list).Error

	return list, total, err
}

func (s *TransferService) AcceptTransfer(id uint, operatorID uint, operatorName string, storeID uint) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.ToStoreID != storeID {
		return nil, errors.New("只有调入门店可以确认接单")
	}

	if transfer.Status != model.TransferStatusPendingAccept {
		return nil, errors.New("只有待接单状态的调拨单可以确认接单")
	}

	now := time.Now()
	transfer.Status = model.TransferStatusPendingOut
	transfer.AcceptOperatorID = operatorID
	transfer.AcceptOperatorName = operatorName
	transfer.AcceptedAt = &now

	if err := database.DB.Save(&transfer).Error; err != nil {
		return nil, err
	}

	transfer.Items, _ = s.getItems(transfer.ID)

	go s.dingTalk.SendText(fmt.Sprintf("调拨单[%s]已被确认接单，请及时安排出库", transfer.TransferNo), false, nil)

	return &transfer, nil
}

func (s *TransferService) ConfirmOutbound(id uint, operatorID uint, operatorName string, storeID uint, remark string) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.FromStoreID != storeID {
		return nil, errors.New("只有调出门店可以执行出库操作")
	}

	if transfer.Status != model.TransferStatusPendingOut {
		return nil, errors.New("只有待出库状态的调拨单可以确认出库")
	}

	now := time.Now()

	tx := database.DB.Begin()

	items, err := s.getItems(transfer.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range items {
		result := tx.Model(&model.Ingredient{}).
			Where("id = ? AND store_id = ? AND current_stock >= ?", item.IngredientID, transfer.FromStoreID, item.OutQty).
			Update("current_stock", gorm.Expr("current_stock - ?", item.OutQty))

		if result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("食材[%s]库存不足", item.IngredientName)
		}
	}

	transfer.Status = model.TransferStatusOutConfirmed
	transfer.OutOperatorID = operatorID
	transfer.OutOperatorName = operatorName
	transfer.OutConfirmedAt = &now
	if remark != "" {
		transfer.Remark = transfer.Remark + "\n出库备注: " + remark
	}

	if err := tx.Save(&transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	transfer.Items = items

	go s.dingTalk.SendText(fmt.Sprintf("调拨单[%s]已确认出库，请门店及时收货", transfer.TransferNo), false, nil)

	return &transfer, nil
}

func (s *TransferService) StartShipping(id uint, logisticsCompany, trackingNo, logisticsCode string, storeID uint) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.FromStoreID != storeID {
		return nil, errors.New("只有调出门店可以执行发货操作")
	}

	if transfer.Status != model.TransferStatusOutConfirmed {
		return nil, errors.New("只有已出库状态的调拨单可以开始发货")
	}

	transfer.Status = model.TransferStatusInTransit
	transfer.LogisticsCompany = logisticsCompany
	transfer.TrackingNo = trackingNo

	tx := database.DB.Begin()

	if err := tx.Save(&transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	initTrack := model.TransferLogistics{
		TransferID:    transfer.ID,
		TrackingNo:    trackingNo,
		LogisticsCode: logisticsCode,
		LogisticsName: logisticsCompany,
		Status:        "已发货",
		Description:   "货物已发出，开始运输",
		TrackTime:     &time.Time{},
	}
	now := time.Now()
	initTrack.TrackTime = &now

	if err := tx.Create(&initTrack).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &transfer, nil
}

func (s *TransferService) ReceiveTransfer(id uint, req *dto.ReceiveTransferDTO, storeID uint) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.ToStoreID != storeID {
		return nil, errors.New("只有调入门店可以执行收货操作")
	}

	if transfer.Status != model.TransferStatusInTransit && transfer.Status != model.TransferStatusOutConfirmed {
		return nil, errors.New("当前状态不能收货")
	}

	now := time.Now()

	tx := database.DB.Begin()

	hasDiff := false

	for _, item := range req.Items {
		var transferItem model.TransferOrderItem
		if err := tx.First(&transferItem, item.ItemID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("调拨明细不存在: %d", item.ItemID)
		}
		if transferItem.TransferID != transfer.ID {
			tx.Rollback()
			return nil, errors.New("调拨明细与调拨单不匹配")
		}

		diffQty := item.InQty.Sub(transferItem.OutQty)
		diffAmount := diffQty.Mul(transferItem.UnitPrice)

		if !diffQty.Equal(decimal.Zero) {
			hasDiff = true
		}

		transferItem.InQty = item.InQty
		transferItem.DiffQty = diffQty
		transferItem.DiffAmount = diffAmount
		if item.Remark != "" {
			transferItem.Remark = item.Remark
		}

		var existingIngredient model.Ingredient
		err := tx.Where("ingredient_no = ? AND store_id = ?", transferItem.IngredientNo, transfer.ToStoreID).
			First(&existingIngredient).Error

		if err != nil {
			newIngredient := model.Ingredient{
				StoreID:       transfer.ToStoreID,
				IngredientNo:  transferItem.IngredientNo,
				Name:          transferItem.IngredientName,
				Unit:          transferItem.Unit,
				CurrentPrice:  transferItem.UnitPrice,
				CurrentStock:  item.InQty,
				StockUnit:     transferItem.Unit,
				Status:        1,
			}
			if err := tx.Create(&newIngredient).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			tx.Model(&model.Ingredient{}).
				Where("id = ?", existingIngredient.ID).
				Update("current_stock", gorm.Expr("current_stock + ?", item.InQty))
		}

		if err := tx.Save(&transferItem).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	transfer.Status = model.TransferStatusReceived
	transfer.InOperatorID = req.OperatorID
	transfer.InOperatorName = req.OperatorName
	transfer.ReceivedAt = &now
	transfer.HasDiff = hasDiff

	if err := tx.Save(&transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	transfer.Items, _ = s.getItems(transfer.ID)

	if hasDiff {
		go s.dingTalk.SendText(fmt.Sprintf("调拨单[%s]收货完成，存在差异，请及时处理", transfer.TransferNo), false, nil)
	}

	return &transfer, nil
}

func (s *TransferService) CompleteTransfer(id uint, diffRemark string, storeID uint) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.FromStoreID != storeID && transfer.ToStoreID != storeID {
		return nil, errors.New("只有相关门店可以完成调拨单")
	}

	if transfer.Status != model.TransferStatusReceived {
		return nil, errors.New("只有已收货状态的调拨单可以完成")
	}

	tx := database.DB.Begin()

	items, err := s.getItems(transfer.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if transfer.HasDiff {
		for _, item := range items {
			if !item.DiffQty.Equal(decimal.Zero) {
				var fromIngredient model.Ingredient
				err := tx.Where("id = ? AND store_id = ?", item.IngredientID, transfer.FromStoreID).
					First(&fromIngredient).Error
				if err == nil {
					if item.DiffQty.GreaterThan(decimal.Zero) {
						tx.Model(&model.Ingredient{}).
							Where("id = ?", fromIngredient.ID).
							Update("current_stock", gorm.Expr("current_stock - ?", item.DiffQty))
					} else {
						tx.Model(&model.Ingredient{}).
							Where("id = ?", fromIngredient.ID).
							Update("current_stock", gorm.Expr("current_stock + ?", item.DiffQty.Abs()))
					}
				}
			}
		}
	}

	now := time.Now()
	transfer.Status = model.TransferStatusCompleted
	transfer.CompletedAt = &now
	transfer.DiffRemark = diffRemark

	if err := tx.Save(&transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	transfer.Items = items

	return &transfer, nil
}

func (s *TransferService) CancelTransfer(id uint, remark string, storeID uint) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.FromStoreID != storeID {
		return nil, errors.New("只有调出门店可以取消调拨单")
	}

	if transfer.Status >= model.TransferStatusOutConfirmed {
		return nil, errors.New("已出库的调拨单不能取消")
	}

	transfer.Status = model.TransferStatusCancelled
	if remark != "" {
		transfer.Remark = transfer.Remark + "\n取消原因: " + remark
	}

	if err := database.DB.Save(&transfer).Error; err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (s *TransferService) RefreshLogisticsTrack(id uint) ([]model.TransferLogistics, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.TrackingNo == "" {
		return nil, errors.New("暂无物流单号")
	}

	logisticsCode := transfer.LogisticsCompany
	if code := GetLogisticsCompanyCode(transfer.LogisticsCompany); code != "" {
		logisticsCode = code
	}

	result, err := s.kdnService.GetLogisticsTrack(transfer.TrackingNo, logisticsCode)
	if err != nil {
		return nil, err
	}

	tx := database.DB.Begin()

	if err := tx.Where("transfer_id = ?", id).Delete(&model.TransferLogistics{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tracks := make([]model.TransferLogistics, 0, len(result.Traces))
	for _, trace := range result.Traces {
		trackTime, _ := time.ParseInLocation("2006-01-02 15:04:05", trace.AcceptTime, time.Local)
		track := model.TransferLogistics{
			TransferID:    id,
			TrackingNo:    transfer.TrackingNo,
			LogisticsCode: logisticsCode,
			LogisticsName: transfer.LogisticsCompany,
			Status:        GetLogisticsStatusText(result.State),
			Location:      trace.Location,
			Description:   trace.AcceptStation,
			Operator:      "",
			OperatorPhone: "",
			TrackTime:     &trackTime,
		}
		tracks = append(tracks, track)
	}

	if len(tracks) > 0 {
		if err := tx.Create(&tracks).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return s.GetLogisticsTracks(id)
}

func (s *TransferService) UpdateLogisticsTrack(id uint, tracks []model.TransferLogistics) error {
	tx := database.DB.Begin()

	if err := tx.Where("transfer_id = ?", id).Delete(&model.TransferLogistics{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range tracks {
		tracks[i].TransferID = id
	}

	if len(tracks) > 0 {
		if err := tx.Create(&tracks).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *TransferService) GetLogisticsTracks(id uint) ([]model.TransferLogistics, error) {
	var tracks []model.TransferLogistics
	err := database.DB.Where("transfer_id = ?", id).
		Order("track_time DESC, id DESC").
		Find(&tracks).Error
	return tracks, err
}

func (s *TransferService) getItems(transferID uint) ([]model.TransferOrderItem, error) {
	var items []model.TransferOrderItem
	err := database.DB.Where("transfer_id = ?", transferID).
		Preload("Ingredient").
		Order("id ASC").
		Find(&items).Error
	return items, err
}

func (s *TransferService) GetItems(transferID uint) ([]model.TransferOrderItem, error) {
	return s.getItems(transferID)
}

func (s *TransferService) RejectTransfer(id uint, operatorID uint, operatorName string, storeID uint, reason string) (*model.TransferOrder, error) {
	var transfer model.TransferOrder
	if err := database.DB.First(&transfer, id).Error; err != nil {
		return nil, errors.New("调拨单不存在")
	}

	if transfer.ToStoreID != storeID {
		return nil, errors.New("只有调入门店可以拒单")
	}

	if transfer.Status != model.TransferStatusPendingAccept {
		return nil, errors.New("只有待接单状态可以拒单")
	}

	transfer.Status = model.TransferStatusCancelled
	if reason != "" {
		transfer.Remark = transfer.Remark + "\n拒单原因: " + reason
	}

	if err := database.DB.Save(&transfer).Error; err != nil {
		return nil, err
	}

	go s.dingTalk.SendText(fmt.Sprintf("调拨单[%s]已被拒绝，请及时处理", transfer.TransferNo), false, nil)

	return &transfer, nil
}

func generateTransferNo() string {
	now := time.Now()
	return fmt.Sprintf("DB%s%06d", now.Format("20060102150405"), time.Now().UnixNano()%1000000)
}
