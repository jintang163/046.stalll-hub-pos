package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/nsq"
	pkgredis "stalll-hub-pos/backend/pkg/redis"
)

const (
	StallDeviceOnlineKey = "stall:device:online:%s"
	StallDeviceOnlineTTL = 5 * time.Minute
	StallDeviceOfflineThreshold = 30 * time.Minute
)

type StallService struct {
	stallRepo          *repository.StallRepository
	deviceRepo         *repository.StallDeviceRepository
	userRepo           *repository.StallUserRepository
	settlementRepo     *repository.StallSettlementRepository
	dailyReportRepo    *repository.StallDailyReportRepository
	orderRepo          *repository.OrderRepository
	productRepo        *repository.ProductRepository
}

func NewStallService() *StallService {
	return &StallService{
		stallRepo:       repository.NewStallRepository(nil),
		deviceRepo:      repository.NewStallDeviceRepository(nil),
		userRepo:        repository.NewStallUserRepository(nil),
		settlementRepo:  repository.NewStallSettlementRepository(nil),
		dailyReportRepo: repository.NewStallDailyReportRepository(nil),
		orderRepo:       repository.NewOrderRepository(nil),
		productRepo:     repository.NewProductRepository(),
	}
}

func (s *StallService) CreateStall(req *dto.StallCreateDTO) (*model.Stall, error) {
	_, err := s.stallRepo.GetByStallNo(req.StoreID, req.StallNo)
	if err == nil {
		return nil, errors.New("摊位编号已存在")
	}

	if req.RevenueRatio.IsZero() && req.PlatformRatio.IsZero() {
		req.RevenueRatio = decimal.NewFromFloat(0.7)
		req.PlatformRatio = decimal.NewFromFloat(0.3)
	}

	if req.RevenueRatio.Add(req.PlatformRatio).Sub(decimal.NewFromInt(1)).Abs().GreaterThan(decimal.NewFromFloat(0.0001)) {
		return nil, errors.New("分账比例之和必须等于1")
	}

	stall := &model.Stall{
		StoreID:       req.StoreID,
		StallNo:       req.StallNo,
		Name:          req.Name,
		Type:          req.Type,
		Description:   req.Description,
		Logo:          req.Logo,
		RevenueRatio:  req.RevenueRatio,
		PlatformRatio: req.PlatformRatio,
		ContactName:   req.ContactName,
		ContactPhone:  req.ContactPhone,
		PrinterName:   req.PrinterName,
		SortOrder:     req.SortOrder,
		Status:        req.Status,
	}
	if stall.Status == 0 {
		stall.Status = 1
	}
	if stall.Type == "" {
		stall.Type = "normal"
	}

	err = s.stallRepo.Create(stall)
	if err != nil {
		return nil, err
	}

	nsq.PublishStallChange("create", req.StoreID, stall.ID, stall.StallNo, stall.Name, stall.Status, stall)

	return s.stallRepo.GetByID(stall.ID)
}

func (s *StallService) GetStall(id uint) (*model.Stall, error) {
	return s.stallRepo.GetByID(id)
}

func (s *StallService) UpdateStall(id uint, req *dto.StallUpdateDTO) (*model.Stall, error) {
	stall, err := s.stallRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("摊位不存在")
	}

	if req.Name != "" {
		stall.Name = req.Name
	}
	if req.Type != "" {
		stall.Type = req.Type
	}
	stall.Description = req.Description
	stall.Logo = req.Logo
	if req.RevenueRatio != nil && req.PlatformRatio != nil {
		if req.RevenueRatio.Add(*req.PlatformRatio).Sub(decimal.NewFromInt(1)).Abs().GreaterThan(decimal.NewFromFloat(0.0001)) {
			return nil, errors.New("分账比例之和必须等于1")
		}
		stall.RevenueRatio = *req.RevenueRatio
		stall.PlatformRatio = *req.PlatformRatio
	}
	stall.ContactName = req.ContactName
	stall.ContactPhone = req.ContactPhone
	stall.PrinterName = req.PrinterName
	if req.SortOrder != nil {
		stall.SortOrder = *req.SortOrder
	}
	if req.Status != 0 {
		stall.Status = req.Status
	}

	err = s.stallRepo.Update(stall)
	if err != nil {
		return nil, err
	}

	nsq.PublishStallChange("update", stall.StoreID, stall.ID, stall.StallNo, stall.Name, stall.Status, stall)

	return s.stallRepo.GetByID(id)
}

func (s *StallService) DeleteStall(id uint) error {
	stall, err := s.stallRepo.GetByID(id)
	if err != nil {
		return errors.New("摊位不存在")
	}
	err = s.stallRepo.Delete(id)
	if err != nil {
		return err
	}
	nsq.PublishStallChange("delete", stall.StoreID, id, stall.StallNo, stall.Name, 0, nil)
	return nil
}

func (s *StallService) ListStalls(query *dto.StallQueryDTO) ([]dto.StallResponse, int64, error) {
	stalls, total, err := s.stallRepo.List(query.StoreID, query.Name, query.Status, query.Page, query.PageSize)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.StallResponse
	for _, stall := range stalls {
		storeName := ""
		if stall.Store.Name != "" {
			storeName = stall.Store.Name
		}
		response = append(response, dto.StallResponse{
			ID:            stall.ID,
			StoreID:       stall.StoreID,
			StoreName:     storeName,
			StallNo:       stall.StallNo,
			Name:          stall.Name,
			Type:          stall.Type,
			Description:   stall.Description,
			Logo:          stall.Logo,
			RevenueRatio:  stall.RevenueRatio,
			PlatformRatio: stall.PlatformRatio,
			ContactName:   stall.ContactName,
			ContactPhone:  stall.ContactPhone,
			PrinterName:   stall.PrinterName,
			SortOrder:     stall.SortOrder,
			Status:        stall.Status,
			CreatedAt:     stall.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, total, nil
}

func (s *StallService) GetAllStalls(storeID uint) ([]dto.StallResponse, error) {
	stalls, err := s.stallRepo.GetAll(storeID)
	if err != nil {
		return nil, err
	}

	var response []dto.StallResponse
	for _, stall := range stalls {
		response = append(response, dto.StallResponse{
			ID:            stall.ID,
			StoreID:       stall.StoreID,
			StallNo:       stall.StallNo,
			Name:          stall.Name,
			Type:          stall.Type,
			Description:   stall.Description,
			Logo:          stall.Logo,
			RevenueRatio:  stall.RevenueRatio,
			PlatformRatio: stall.PlatformRatio,
			ContactName:   stall.ContactName,
			ContactPhone:  stall.ContactPhone,
			PrinterName:   stall.PrinterName,
			SortOrder:     stall.SortOrder,
			Status:        stall.Status,
			CreatedAt:     stall.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, nil
}

func (s *StallService) RegisterDevice(req *dto.StallDeviceRegisterDTO) (*model.StallDevice, error) {
	_, err := s.stallRepo.GetByID(req.StallID)
	if err != nil {
		return nil, errors.New("摊位不存在")
	}

	existing, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err == nil {
		existing.StallID = req.StallID
		existing.DeviceName = req.DeviceName
		existing.DeviceType = req.DeviceType
		existing.OSVersion = req.OSVersion
		existing.AppVersion = req.AppVersion
		now := time.Now()
		existing.LastOnlineAt = &now
		existing.LastHeartbeatAt = &now
		err = s.deviceRepo.Update(existing)
		if err != nil {
			return nil, err
		}
		return s.deviceRepo.GetByID(existing.ID)
	}

	now := time.Now()
	device := &model.StallDevice{
		StoreID:         req.StoreID,
		StallID:         req.StallID,
		DeviceID:        req.DeviceID,
		DeviceName:      req.DeviceName,
		DeviceType:      req.DeviceType,
		OSVersion:       req.OSVersion,
		AppVersion:      req.AppVersion,
		LastOnlineAt:    &now,
		LastHeartbeatAt: &now,
		Status:          1,
	}
	if device.DeviceType == "" {
		device.DeviceType = "mobile"
	}

	err = s.deviceRepo.Create(device)
	if err != nil {
		return nil, err
	}

	s.setDeviceOnline(req.DeviceID, device.StallID)

	return s.deviceRepo.GetByID(device.ID)
}

func (s *StallService) Heartbeat(deviceID string, appVersion string) error {
	err := s.deviceRepo.UpdateHeartbeat(deviceID)
	if err != nil {
		return err
	}

	device, err := s.deviceRepo.GetByDeviceID(deviceID)
	if err != nil {
		return err
	}

	s.setDeviceOnline(deviceID, device.StallID)

	if appVersion != "" && device.AppVersion != appVersion {
		device.AppVersion = appVersion
		s.deviceRepo.Update(device)
	}

	return nil
}

func (s *StallService) setDeviceOnline(deviceID string, stallID uint) {
	key := fmt.Sprintf(StallDeviceOnlineKey, deviceID)
	pkgredis.HSet(key, "stall_id", stallID, "last_heartbeat", time.Now().Unix())
	pkgredis.Client.Expire(pkgredis.Ctx, key, StallDeviceOnlineTTL)
}

func (s *StallService) IsDeviceOnline(deviceID string) bool {
	key := fmt.Sprintf(StallDeviceOnlineKey, deviceID)
	exists, _ := pkgredis.Exists(key)
	return exists > 0
}

func (s *StallService) GetDevice(id uint) (*dto.StallDeviceResponse, error) {
	device, err := s.deviceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	isOnline := s.IsDeviceOnline(device.DeviceID)
	stallName := ""
	if device.Stall.Name != "" {
		stallName = device.Stall.Name
	}

	lastOnlineAt := ""
	if device.LastOnlineAt != nil {
		lastOnlineAt = device.LastOnlineAt.Format("2006-01-02 15:04:05")
	}
	lastHeartbeatAt := ""
	if device.LastHeartbeatAt != nil {
		lastHeartbeatAt = device.LastHeartbeatAt.Format("2006-01-02 15:04:05")
	}

	return &dto.StallDeviceResponse{
		ID:              device.ID,
		StoreID:         device.StoreID,
		StallID:         device.StallID,
		StallName:       stallName,
		DeviceID:        device.DeviceID,
		DeviceName:      device.DeviceName,
		DeviceType:      device.DeviceType,
		OSVersion:       device.OSVersion,
		AppVersion:      device.AppVersion,
		IsOnline:        isOnline,
		LastOnlineAt:    lastOnlineAt,
		LastHeartbeatAt: lastHeartbeatAt,
		Status:          device.Status,
	}, nil
}

func (s *StallService) ListDevices(storeID, stallID uint, page, pageSize int) ([]dto.StallDeviceResponse, int64, error) {
	devices, total, err := s.deviceRepo.List(storeID, stallID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.StallDeviceResponse
	for _, device := range devices {
		isOnline := s.IsDeviceOnline(device.DeviceID)
		stallName := ""
		if device.Stall.Name != "" {
			stallName = device.Stall.Name
		}

		lastOnlineAt := ""
		if device.LastOnlineAt != nil {
			lastOnlineAt = device.LastOnlineAt.Format("2006-01-02 15:04:05")
		}
		lastHeartbeatAt := ""
		if device.LastHeartbeatAt != nil {
			lastHeartbeatAt = device.LastHeartbeatAt.Format("2006-01-02 15:04:05")
		}

		response = append(response, dto.StallDeviceResponse{
			ID:              device.ID,
			StoreID:         device.StoreID,
			StallID:         device.StallID,
			StallName:       stallName,
			DeviceID:        device.DeviceID,
			DeviceName:      device.DeviceName,
			DeviceType:      device.DeviceType,
			OSVersion:       device.OSVersion,
			AppVersion:      device.AppVersion,
			IsOnline:        isOnline,
			LastOnlineAt:    lastOnlineAt,
			LastHeartbeatAt: lastHeartbeatAt,
			Status:          device.Status,
		})
	}

	return response, total, nil
}

func (s *StallService) DeleteDevice(id uint) error {
	_, err := s.deviceRepo.GetByID(id)
	if err != nil {
		return errors.New("设备不存在")
	}
	return s.deviceRepo.Delete(id)
}

func (s *StallService) CreateStallUser(req *dto.StallUserCreateDTO) (*dto.StallUserResponse, error) {
	_, err := s.stallRepo.GetByID(req.StallID)
	if err != nil {
		return nil, errors.New("摊位不存在")
	}

	_, err = s.userRepo.GetByUsername(req.Username)
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.StallUser{
		StoreID:  req.StoreID,
		StallID:  req.StallID,
		Username: req.Username,
		Password: string(hashedPassword),
		RealName: req.RealName,
		Phone:    req.Phone,
		Role:     req.Role,
		Status:   req.Status,
	}
	if user.Status == 0 {
		user.Status = 1
	}
	if user.Role == "" {
		user.Role = "stall_staff"
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	created, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	stallName := ""
	if created.Stall.Name != "" {
		stallName = created.Stall.Name
	}

	return &dto.StallUserResponse{
		ID:        created.ID,
		StoreID:   created.StoreID,
		StallID:   created.StallID,
		StallName: stallName,
		Username:  created.Username,
		RealName:  created.RealName,
		Phone:     created.Phone,
		Role:      created.Role,
		Status:    created.Status,
		CreatedAt: created.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *StallService) UpdateStallUser(id uint, req *dto.StallUserUpdateDTO) (*dto.StallUserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	user.RealName = req.RealName
	user.Phone = req.Phone
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != 0 {
		user.Status = req.Status
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	updated, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	stallName := ""
	if updated.Stall.Name != "" {
		stallName = updated.Stall.Name
	}

	return &dto.StallUserResponse{
		ID:        updated.ID,
		StoreID:   updated.StoreID,
		StallID:   updated.StallID,
		StallName: stallName,
		Username:  updated.Username,
		RealName:  updated.RealName,
		Phone:     updated.Phone,
		Role:      updated.Role,
		Status:    updated.Status,
		CreatedAt: updated.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *StallService) ListStallUsers(query *dto.StallUserQueryDTO) ([]dto.StallUserResponse, int64, error) {
	users, total, err := s.userRepo.List(query.StoreID, query.StallID, query.Username, query.Status, query.Page, query.PageSize)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.StallUserResponse
	for _, user := range users {
		stallName := ""
		if user.Stall.Name != "" {
			stallName = user.Stall.Name
		}
		response = append(response, dto.StallUserResponse{
			ID:        user.ID,
			StoreID:   user.StoreID,
			StallID:   user.StallID,
			StallName: stallName,
			Username:  user.Username,
			RealName:  user.RealName,
			Phone:     user.Phone,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, total, nil
}

func (s *StallService) DeleteStallUser(id uint) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return errors.New("用户不存在")
	}
	return s.userRepo.Delete(id)
}

func (s *StallService) StallLogin(req *dto.StallLoginDTO) (*dto.StallLoginResponse, error) {
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账号已禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	stall, err := s.stallRepo.GetByID(user.StallID)
	if err != nil {
		return nil, errors.New("摊位不存在")
	}

	token, err := s.generateStallToken(user.ID)
	if err != nil {
		return nil, err
	}

	userResp := dto.StallUserResponse{
		ID:        user.ID,
		StoreID:   user.StoreID,
		StallID:   user.StallID,
		StallName: stall.Name,
		Username:  user.Username,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	stallResp := dto.StallResponse{
		ID:            stall.ID,
		StoreID:       stall.StoreID,
		StallNo:       stall.StallNo,
		Name:          stall.Name,
		Type:          stall.Type,
		Description:   stall.Description,
		Logo:          stall.Logo,
		RevenueRatio:  stall.RevenueRatio,
		PlatformRatio: stall.PlatformRatio,
		ContactName:   stall.ContactName,
		ContactPhone:  stall.ContactPhone,
		PrinterName:   stall.PrinterName,
		SortOrder:     stall.SortOrder,
		Status:        stall.Status,
		CreatedAt:     stall.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return &dto.StallLoginResponse{
		Token: token,
		User:  userResp,
		Stall: stallResp,
	}, nil
}

func (s *StallService) generateStallToken(userID uint) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("stall_%d_%x", userID, b), nil
}

func (s *StallService) CalculateStallSettlement(stallID uint, settlementDate string) (*model.StallSettlement, error) {
	stall, err := s.stallRepo.GetByID(stallID)
	if err != nil {
		return nil, errors.New("摊位不存在")
	}

	orders, err := s.orderRepo.GetPaidOrdersByStallAndDate(stallID, settlementDate)
	if err != nil {
		return nil, err
	}

	totalAmount := decimal.Zero
	refundAmount := decimal.Zero
	stallAmount := decimal.Zero
	platformAmount := decimal.Zero
	orderCount := 0

	for _, order := range orders {
		if order.OrderStatus == -1 {
			refundAmount = refundAmount.Add(order.PayAmount)
			continue
		}

		orderCount++
		totalAmount = totalAmount.Add(order.PayAmount)

		for _, item := range order.Items {
			if item.StallID == stallID {
				stallAmount = stallAmount.Add(item.StallAmount)
				platformAmount = platformAmount.Add(item.PlatformAmount)
			}
		}
	}

	netAmount := totalAmount.Sub(refundAmount)

	settlementNo := s.generateSettlementNo()

	settlement := &model.StallSettlement{
		StoreID:          stall.StoreID,
		StallID:          stallID,
		SettlementNo:     settlementNo,
		SettlementDate:   settlementDate,
		OrderCount:       orderCount,
		TotalAmount:      totalAmount,
		RefundAmount:     refundAmount,
		NetAmount:        netAmount,
		StallAmount:      stallAmount,
		PlatformAmount:   platformAmount,
		SettlementStatus: 0,
	}

	return settlement, nil
}

func (s *StallService) generateSettlementNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("ST%s%06d", now.Format("20060102150405"), n.Int64())
}

func (s *StallService) CreateSettlement(req *dto.StallSettlementCreateDTO) (*dto.StallSettlementResponse, error) {
	settlement, err := s.CalculateStallSettlement(req.StallID, req.SettlementDate)
	if err != nil {
		return nil, err
	}

	settlement.Remark = req.Remark

	err = s.settlementRepo.Create(settlement)
	if err != nil {
		return nil, err
	}

	created, err := s.settlementRepo.GetByID(settlement.ID)
	if err != nil {
		return nil, err
	}

	stallName := ""
	if created.Stall.Name != "" {
		stallName = created.Stall.Name
	}

	settledAt := ""
	if created.SettledAt != nil {
		settledAt = created.SettledAt.Format("2006-01-02 15:04:05")
	}

	return &dto.StallSettlementResponse{
		ID:               created.ID,
		StoreID:          created.StoreID,
		StallID:          created.StallID,
		StallName:        stallName,
		SettlementNo:     created.SettlementNo,
		SettlementDate:   created.SettlementDate,
		OrderCount:       created.OrderCount,
		TotalAmount:      created.TotalAmount,
		RefundAmount:     created.RefundAmount,
		NetAmount:        created.NetAmount,
		StallAmount:      created.StallAmount,
		PlatformAmount:   created.PlatformAmount,
		SettlementStatus: created.SettlementStatus,
		SettledAt:        settledAt,
		Remark:           created.Remark,
	}, nil
}

func (s *StallService) ListSettlements(query *dto.StallSettlementQueryDTO) ([]dto.StallSettlementResponse, int64, error) {
	settlements, total, err := s.settlementRepo.List(
		query.StoreID, query.StallID, query.SettlementDate, query.SettlementStatus,
		query.Page, query.PageSize,
	)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.StallSettlementResponse
	for _, s := range settlements {
		stallName := ""
		if s.Stall.Name != "" {
			stallName = s.Stall.Name
		}
		settledAt := ""
		if s.SettledAt != nil {
			settledAt = s.SettledAt.Format("2006-01-02 15:04:05")
		}
		response = append(response, dto.StallSettlementResponse{
			ID:               s.ID,
			StoreID:          s.StoreID,
			StallID:          s.StallID,
			StallName:        stallName,
			SettlementNo:     s.SettlementNo,
			SettlementDate:   s.SettlementDate,
			OrderCount:       s.OrderCount,
			TotalAmount:      s.TotalAmount,
			RefundAmount:     s.RefundAmount,
			NetAmount:        s.NetAmount,
			StallAmount:      s.StallAmount,
			PlatformAmount:   s.PlatformAmount,
			SettlementStatus: s.SettlementStatus,
			SettledAt:        settledAt,
			Remark:           s.Remark,
		})
	}

	return response, total, nil
}

func (s *StallService) GenerateDailyReport(storeID, stallID uint, reportDate string) (*model.StallDailyReport, error) {
	stall, err := s.stallRepo.GetByID(stallID)
	if err != nil {
		return nil, errors.New("摊位不存在")
	}

	settlement, err := s.CalculateStallSettlement(stallID, reportDate)
	if err != nil {
		return nil, err
	}

	discountAmount := decimal.Zero
	couponAmount := decimal.Zero

	report := &model.StallDailyReport{
		StoreID:        storeID,
		StallID:        stallID,
		ReportDate:     reportDate,
		OrderCount:     settlement.OrderCount,
		TotalAmount:    settlement.TotalAmount,
		DiscountAmount: discountAmount,
		CouponAmount:   couponAmount,
		RefundAmount:   settlement.RefundAmount,
		NetAmount:      settlement.NetAmount,
		StallAmount:    settlement.StallAmount,
		PlatformAmount: settlement.PlatformAmount,
	}

	existing, err := s.dailyReportRepo.GetByDate(storeID, stallID, reportDate)
	if err == nil {
		report.ID = existing.ID
		report.CreatedAt = existing.CreatedAt
	}

	err = s.dailyReportRepo.Upsert(report)
	if err != nil {
		return nil, err
	}

	return s.dailyReportRepo.GetByID(report.ID)
}

func (s *StallService) GetDailyReport(query *dto.StallDailyReportQueryDTO) ([]dto.StallDailyReportResponse, error) {
	reports, err := s.dailyReportRepo.List(query.StoreID, query.StallID, query.StartDate, query.EndDate)
	if err != nil {
		return nil, err
	}

	var response []dto.StallDailyReportResponse
	for _, r := range reports {
		stallName := ""
		if r.Stall.Name != "" {
			stallName = r.Stall.Name
		}
		response = append(response, dto.StallDailyReportResponse{
			ID:             r.ID,
			StoreID:        r.StoreID,
			StallID:        r.StallID,
			StallName:      stallName,
			ReportDate:     r.ReportDate,
			OrderCount:     r.OrderCount,
			TotalAmount:    r.TotalAmount,
			DiscountAmount: r.DiscountAmount,
			CouponAmount:   r.CouponAmount,
			RefundAmount:   r.RefundAmount,
			NetAmount:      r.NetAmount,
			StallAmount:    r.StallAmount,
			PlatformAmount: r.PlatformAmount,
		})
	}

	return response, nil
}

func (s *StallService) CheckOfflineDevices() (int, error) {
	devices, _, err := s.deviceRepo.List(0, 0, 1, 1000)
	if err != nil {
		log.Printf("检查离线设备失败: %v", err)
		return 0, err
	}

	alertCount := 0
	threshold := time.Now().Add(-StallDeviceOfflineThreshold)

	for _, device := range devices {
		if device.LastHeartbeatAt == nil || device.LastHeartbeatAt.Before(threshold) {
			key := fmt.Sprintf(StallDeviceOnlineKey, device.DeviceID)
			pkgredis.Del(key)

			offlineMinutes := 0
			if device.LastHeartbeatAt != nil {
				offlineMinutes = int(time.Since(*device.LastHeartbeatAt).Minutes())
			} else {
				offlineMinutes = int(StallDeviceOfflineThreshold.Minutes()) + 1
			}

			stallName := ""
			if device.Stall.Name != "" {
				stallName = device.Stall.Name
			}

			alertErr := nsq.PublishStallDeviceAlert(
				device.ID,
				device.DeviceName,
				device.DeviceID,
				stallName,
				device.StoreID,
				device.StallID,
				"offline",
				offlineMinutes,
				0,
			)
			if alertErr != nil {
				log.Printf("发送设备告警失败: device=%d, err=%v", device.ID, alertErr)
			}

			alertCount++
		}
	}

	return alertCount, nil
}
