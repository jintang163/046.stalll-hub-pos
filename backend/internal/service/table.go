package service

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/config"
	"stalll-hub-pos/backend/pkg/nsq"
	"stalll-hub-pos/backend/pkg/redis"
)

type TableService struct {
	tableRepo       *repository.TableRepository
	areaRepo        *repository.TableAreaRepository
	reservationRepo *repository.ReservationRepository
	queueRepo       *repository.QueueRepository
	orderRepo       *repository.OrderRepository
	productRepo     *repository.ProductRepository
}

func NewTableService() *TableService {
	return &TableService{
		tableRepo:       repository.NewTableRepository(),
		areaRepo:        repository.NewTableAreaRepository(),
		reservationRepo: repository.NewReservationRepository(),
		queueRepo:       repository.NewQueueRepository(),
		orderRepo:       repository.NewOrderRepository(),
		productRepo:     repository.NewProductRepository(),
	}
}

func (s *TableService) CreateTable(dto *dto.TableCreateDTO) (*model.Table, error) {
	table := &model.Table{
		StoreID:  dto.StoreID,
		TableNo:  dto.TableNo,
		Name:     dto.Name,
		Type:     dto.Type,
		Capacity: dto.Capacity,
		Floor:    dto.Floor,
		Area:     dto.Area,
		QRCode:   dto.QRCode,
		QRCodeUrl: dto.QRCodeUrl,
		Status:   dto.Status,
	}
	if table.Type == "" {
		table.Type = "normal"
	}
	if table.Capacity == 0 {
		table.Capacity = 4
	}
	if table.Floor == 0 {
		table.Floor = 1
	}
	if table.Status == 0 {
		table.Status = 1
	}
	err := s.tableRepo.Create(table)
	if err != nil {
		return nil, err
	}
	scene := fmt.Sprintf("table_%d_%d", table.StoreID, table.ID)
	qrCode, qrCodeUrl, err := s.generateQRCode(table.ID, scene, dto.StoreID, table.TableNo)
	if err == nil {
		s.tableRepo.UpdateQRCode(table.ID, qrCode, qrCodeUrl)
		table.QRCode = qrCode
		table.QRCodeUrl = qrCodeUrl
	}
	return table, nil
}

func (s *TableService) UpdateTable(id uint, dto *dto.TableUpdateDTO) error {
	table := &model.Table{}
	if dto.TableNo != "" {
		table.TableNo = dto.TableNo
	}
	if dto.Name != "" {
		table.Name = dto.Name
	}
	if dto.Type != "" {
		table.Type = dto.Type
	}
	if dto.Capacity > 0 {
		table.Capacity = dto.Capacity
	}
	if dto.Floor > 0 {
		table.Floor = dto.Floor
	}
	if dto.Area != "" {
		table.Area = dto.Area
	}
	if dto.Status >= 0 {
		table.Status = dto.Status
	}
	return s.tableRepo.Update(id, table)
}

func (s *TableService) DeleteTable(id uint) error {
	return s.tableRepo.Delete(id)
}

func (s *TableService) GetTable(id uint) (*model.Table, error) {
	return s.tableRepo.GetByID(id)
}

func (s *TableService) ListTables(query *dto.TableQueryDTO) ([]model.Table, int64, error) {
	return s.tableRepo.List(query)
}

func (s *TableService) GetOccupiedTables(storeID uint) ([]dto.TableOccupiedInfo, error) {
	tables, err := s.tableRepo.GetOccupiedTables(storeID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.TableOccupiedInfo, 0, len(tables))
	for _, t := range tables {
		info := dto.TableOccupiedInfo{
			ID:                  t.ID,
			TableNo:             t.TableNo,
			Name:                t.Name,
			Capacity:            t.Capacity,
			CurrentCustomerCount: t.CurrentCustomerCount,
			CurrentOrderID:      t.CurrentOrderID,
			Status:              t.Status,
		}
		if t.CheckinTime != nil {
			info.CheckinTime = t.CheckinTime.Format("2006-01-02 15:04:05")
			info.OccupiedDuration = int(time.Since(*t.CheckinTime).Minutes())
		}
		if t.CurrentOrderID > 0 {
			order, _ := s.orderRepo.GetByID(t.CurrentOrderID)
			if order != nil {
				info.CurrentOrderAmount = order.TotalAmount
			}
		}
		result = append(result, info)
	}
	return result, nil
}

func (s *TableService) Checkin(dto *dto.TableCheckinDTO) error {
	return s.tableRepo.Checkin(dto.TableID, dto.PeopleCount, 0)
}

func (s *TableService) Checkout(dto *dto.TableCheckoutDTO) error {
	order, err := s.orderRepo.GetByID(dto.OrderID)
	if err != nil {
		return err
	}
	err = s.tableRepo.Checkout(dto.TableID, order.TotalAmount)
	if err != nil {
		return err
	}
	return nil
}

func (s *TableService) generateQRCode(tableID uint, scene string, storeID uint, tableNo string) (string, string, error) {
	sceneStr := fmt.Sprintf("table_%d_%d_%d", storeID, tableID, time.Now().Unix())
	qrCode := fmt.Sprintf("%x", md5.Sum([]byte(sceneStr)))
	baseURL := config.AppConfig.Wechat.AppURL
	qrCodeUrl := fmt.Sprintf("%s?scene=%s&table=%d&store=%d", baseURL, qrCode, tableID, storeID)
	cacheKey := fmt.Sprintf("table:qrcode:%d", tableID)
	redis.Client.Set(cacheKey, sceneStr, 7*24*time.Hour)
	return qrCode, qrCodeUrl, nil
}

func (s *TableService) GenerateTableQRCode(tableID uint) (string, string, error) {
	table, err := s.tableRepo.GetByID(tableID)
	if err != nil {
		return "", "", err
	}
	scene := fmt.Sprintf("table_%d_%d", table.StoreID, table.ID)
	qrCode, qrCodeUrl, err := s.generateQRCode(tableID, scene, table.StoreID, table.TableNo)
	if err != nil {
		return "", "", err
	}
	err = s.tableRepo.UpdateQRCode(tableID, qrCode, qrCodeUrl)
	return qrCode, qrCodeUrl, err
}

func (s *TableService) ScanQRCode(scene string) (*dto.ScanResultDTO, error) {
	parts := splitScene(scene)
	if len(parts) >= 3 && parts[0] == "table" {
		storeID, _ := strconv.ParseUint(parts[1], 10, 32)
		tableID, _ := strconv.ParseUint(parts[2], 10, 32)
		table, err := s.tableRepo.GetByID(uint(tableID))
		if err == nil && table.StoreID == uint(storeID) {
			msg := ""
			if table.Status != 1 {
				msg = "该桌位已暂停使用"
			} else if table.CurrentOrderID > 0 {
				msg = "该桌位已有人"
			}
			store, _ := repository.NewStoreRepository().GetByID(uint(storeID))
			storeName := ""
			if store != nil {
				storeName = store.Name
			}
			return &dto.ScanResultDTO{
				StoreID:   table.StoreID,
				StoreName: storeName,
				TableID:   table.ID,
				TableNo:   table.TableNo,
				TableType: table.Type,
				Capacity:  table.Capacity,
				Area:      table.Area,
				Floor:     table.Floor,
				Status:    table.Status,
				Message:   msg,
			}, nil
		}
	}
	return nil, errors.New("二维码无效或已过期")
}

func (s *TableService) GetAvailableTables(storeID uint, peopleCount int) ([]model.Table, error) {
	return s.tableRepo.GetAvailableTables(storeID, peopleCount)
}

func (s *TableService) GetStoreMap() ([]dto.StoreMapDTO, error) {
	stores, err := s.tableRepo.GetStoreMapInfo()
	if err != nil {
		return nil, err
	}
	result := make([]dto.StoreMapDTO, 0, len(stores))
	for _, store := range stores {
		lat, _ := strconv.ParseFloat(store.Latitude, 64)
		lng, _ := strconv.ParseFloat(store.Longitude, 64)
		result = append(result, dto.StoreMapDTO{
			ID:          store.ID,
			Name:        store.Name,
			Address:     store.Address,
			Latitude:    lat,
			Longitude:   lng,
			Status:      store.Status,
			TablesCount: store.TablesCount,
			OpenTime:    store.OpenTime,
			CloseTime:   store.CloseTime,
			Phone:       store.Phone,
		})
	}
	return result, nil
}

func (s *TableService) BatchCreateTables(storeID uint, count int, prefix string, startNo int, capacity int, floor int, area string) ([]model.Table, error) {
	tables := make([]model.Table, 0, count)
	for i := 0; i < count; i++ {
		tableNo := fmt.Sprintf("%s%d", prefix, startNo+i)
		table := model.Table{
			StoreID:  storeID,
			TableNo:  tableNo,
			Name:     tableNo,
			Type:     "normal",
			Capacity: capacity,
			Floor:    floor,
			Area:     area,
			Status:   1,
		}
		tables = append(tables, table)
	}
	err := s.tableRepo.BatchCreate(tables)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (s *TableService) CreateArea(dto *dto.TableAreaCreateDTO) (*model.TableArea, error) {
	area := &model.TableArea{
		StoreID:   dto.StoreID,
		Name:      dto.Name,
		SortOrder: dto.SortOrder,
		Status:    dto.Status,
	}
	if area.Status == 0 {
		area.Status = 1
	}
	err := s.areaRepo.Create(area)
	return area, err
}

func (s *TableService) UpdateArea(id uint, dto *dto.TableAreaUpdateDTO) error {
	area := &model.TableArea{}
	if dto.Name != "" {
		area.Name = dto.Name
	}
	if dto.SortOrder >= 0 {
		area.SortOrder = dto.SortOrder
	}
	if dto.Status >= 0 {
		area.Status = dto.Status
	}
	return s.areaRepo.Update(id, area)
}

func (s *TableService) DeleteArea(id uint) error {
	return s.areaRepo.Delete(id)
}

func (s *TableService) ListAreas(storeID uint) ([]model.TableArea, error) {
	return s.areaRepo.List(storeID)
}

func (s *TableService) CreateReservation(dto *dto.ReservationCreateDTO) (*model.TableReservation, error) {
	store, err := repository.NewStoreRepository().GetByID(dto.StoreID)
	if err != nil {
		return nil, err
	}
	if !store.ReserveEnabled {
		return nil, errors.New("该门店暂不支持预约")
	}
	if dto.TableID > 0 {
		count, err := s.reservationRepo.GetCountByTableAndTime(dto.TableID, dto.ReserveDate, dto.ReserveTime)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("该时段该桌位已被预约")
		}
	}
	cacheKey := fmt.Sprintf("reserve:slot:%d:%s:%s", dto.StoreID, dto.ReserveDate, dto.ReserveTime)
	booked, _ := redis.Client.Get(cacheKey).Int()
	maxTables := 5
	if booked >= maxTables {
		return nil, errors.New("该时段预约已满")
	}
	reservation := &model.TableReservation{
		StoreID:     dto.StoreID,
		TableID:     dto.TableID,
		MemberID:    dto.MemberID,
		MemberName:  dto.MemberName,
		MemberPhone: dto.MemberPhone,
		TableNo:     dto.TableNo,
		ReserveDate: dto.ReserveDate,
		ReserveTime: dto.ReserveTime,
		PeopleCount: dto.PeopleCount,
		Status:      1,
		Remark:      dto.Remark,
		Source:      dto.Source,
	}
	if reservation.Source == "" {
		reservation.Source = "wechat"
	}
	err = s.reservationRepo.Create(reservation)
	if err != nil {
		return nil, err
	}
	redis.Client.Incr(cacheKey)
	redis.Client.Expire(cacheKey, 24*time.Hour)
	return reservation, nil
}

func (s *TableService) UpdateReservation(id uint, dto *dto.ReservationUpdateDTO) error {
	reservation := &model.TableReservation{}
	if dto.Status > 0 {
		reservation.Status = dto.Status
	}
	if dto.TableID > 0 {
		reservation.TableID = dto.TableID
	}
	if dto.TableNo != "" {
		reservation.TableNo = dto.TableNo
	}
	if dto.ReserveDate != "" {
		reservation.ReserveDate = dto.ReserveDate
	}
	if dto.ReserveTime != "" {
		reservation.ReserveTime = dto.ReserveTime
	}
	if dto.PeopleCount > 0 {
		reservation.PeopleCount = dto.PeopleCount
	}
	if dto.Remark != "" {
		reservation.Remark = dto.Remark
	}
	return s.reservationRepo.Update(id, reservation)
}

func (s *TableService) CancelReservation(id uint) error {
	return s.reservationRepo.Cancel(id)
}

func (s *TableService) CheckinReservation(id uint) error {
	return s.reservationRepo.Checkin(id)
}

func (s *TableService) GetReservation(id uint) (*model.TableReservation, error) {
	return s.reservationRepo.GetByID(id)
}

func (s *TableService) ListReservations(query *dto.ReservationQueryDTO) ([]model.TableReservation, int64, error) {
	return s.reservationRepo.List(query)
}

func (s *TableService) GetTimeSlots(query *dto.ReservationTimeSlotDTO) ([]dto.TimeSlotInfo, error) {
	availability, err := s.reservationRepo.GetTimeSlotAvailability(query.StoreID, query.ReserveDate, query.PeopleCount)
	if err != nil {
		return nil, err
	}
	store, err := repository.NewStoreRepository().GetByID(query.StoreID)
	if err != nil {
		return nil, err
	}
	timeSlots := generateTimeSlots(store.OpenTime, store.CloseTime)
	result := make([]dto.TimeSlotInfo, 0, len(timeSlots))
	for _, ts := range timeSlots {
		booked := availability[ts]
		maxTables := 5
		available := maxTables - booked
		status := 1
		if available <= 0 {
			status = 2
		}
		result = append(result, dto.TimeSlotInfo{
			Time:      ts,
			Available: available,
			Total:     maxTables,
			Status:    status,
		})
	}
	return result, nil
}

func (s *TableService) CreateQueue(dto *dto.QueueCreateDTO) (*model.QueueNumber, error) {
	store, err := repository.NewStoreRepository().GetByID(dto.StoreID)
	if err != nil {
		return nil, err
	}
	if !store.QueueEnabled {
		return nil, errors.New("该门店暂不支持排队取号")
	}
	config, err := s.queueRepo.GetConfig(dto.StoreID)
	if err != nil {
		return nil, err
	}
	queueType := dto.QueueType
	if queueType == "" {
		if dto.PeopleCount <= config.SmallCapacity {
			queueType = "small"
		} else if dto.PeopleCount <= config.MediumCapacity {
			queueType = "medium"
		} else {
			queueType = "large"
		}
	}
	if dto.MemberID > 0 {
		myQueues, _ := s.queueRepo.GetMyQueue(dto.MemberID, dto.StoreID)
		if len(myQueues) > 0 {
			return nil, errors.New("您已在排队中")
		}
	}
	queue := &model.QueueNumber{
		StoreID:     dto.StoreID,
		QueueType:   queueType,
		MemberID:    dto.MemberID,
		MemberName:  dto.MemberName,
		MemberPhone: dto.MemberPhone,
		PeopleCount: dto.PeopleCount,
		Status:      1,
		Remark:      dto.Remark,
	}
	err = s.queueRepo.Create(queue, config)
	if err != nil {
		return nil, err
	}
	queueListKey := fmt.Sprintf("queue:list:%d:%s", dto.StoreID, queueType)
	queueData, _ := json.Marshal(queue)
	redis.Client.LPush(queueListKey, queueData)
	redis.Client.Expire(queueListKey, 24*time.Hour)
	return queue, nil
}

func (s *TableService) CallQueue(dto *dto.QueueCallDTO) (*model.QueueNumber, error) {
	err := s.queueRepo.Call(dto.QueueID)
	if err != nil {
		return nil, err
	}
	queue, err := s.queueRepo.GetByID(dto.QueueID)
	if err != nil {
		return nil, err
	}
	nsq.Producer.Publish("queue_call", fmt.Sprintf("%d", dto.QueueID)
	return queue, nil
}

func (s *TableService) CallNextQueue(storeID uint, queueType string) (*model.QueueNumber, error) {
	queue, err := s.queueRepo.GetNextToCall(storeID, queueType)
	if err != nil {
		return nil, err
	}
	return s.CallQueue(&dto.QueueCallDTO{QueueID: queue.ID})
}

func (s *TableService) CancelQueue(dto *dto.QueueCancelDTO) error {
	err := s.queueRepo.Cancel(dto.QueueID, dto.Reason)
	if err != nil {
		return err
	}
	queue, _ := s.queueRepo.GetByID(dto.QueueID)
	if queue != nil {
		queueListKey := fmt.Sprintf("queue:list:%d:%s", queue.StoreID, queue.QueueType)
		redis.Client.LRem(queueListKey, 0, fmt.Sprintf("%d", dto.QueueID))
	}
	return nil
}

func (s *TableService) ArriveQueue(queueID uint, tableID uint, tableNo string) error {
	return s.queueRepo.Arrive(queueID, tableID, tableNo)
}

func (s *TableService) GetQueue(id uint) (*model.QueueNumber, error) {
	return s.queueRepo.GetByID(id)
}

func (s *TableService) ListQueues(query *dto.QueueQueryDTO) ([]model.QueueNumber, int64, error) {
	return s.queueRepo.List(query)
}

func (s *TableService) GetQueueStatus(query *dto.QueueStatusDTO) (*dto.QueueInfoDTO, error) {
	var queue *model.QueueNumber
	var err error
	if query.QueueNumber != "" {
		queues, _, _ := s.queueRepo.List(&dto.QueueQueryDTO{
			StoreID:     query.StoreID,
			QueueNumber: query.QueueNumber,
			PageSize:    1,
		})
		if len(queues) > 0 {
			queue = &queues[0]
		}
	} else if query.MemberID > 0 {
		myQueues, _ := s.queueRepo.GetMyQueue(query.MemberID, query.StoreID)
		if len(myQueues) > 0 {
			queue = &myQueues[0]
		}
	}
	if queue == nil {
		return nil, errors.New("未找到排队信息")
	}
	aheadCount, err := s.queueRepo.GetAheadCount(query.StoreID, queue.QueueType, queue.Sequence)
	if err != nil {
		return nil, err
	}
	waitTime := 0
	if queue.LastCallTime != nil {
		waitTime = int(time.Since(*queue.LastCallTime).Minutes())
	}
	return &dto.QueueInfoDTO{
		QueueNumber: queue.QueueNumber,
		QueueType:   queue.QueueType,
		Status:      queue.Status,
		Sequence:    queue.Sequence,
		AheadCount:  aheadCount,
		WaitTime:    waitTime,
		PeopleCount: queue.PeopleCount,
		CreatedAt:   queue.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *TableService) GetQueueConfig(storeID uint) (*model.QueueConfig, error) {
	return s.queueRepo.GetConfig(storeID)
}

func (s *TableService) SaveQueueConfig(dto *dto.QueueConfigDTO) error {
	config := &model.QueueConfig{
		StoreID:        dto.StoreID,
		SmallPrefix:    dto.SmallPrefix,
		SmallCapacity:  dto.SmallCapacity,
		MediumPrefix:   dto.MediumPrefix,
		MediumCapacity: dto.MediumCapacity,
		LargePrefix:    dto.LargePrefix,
		LargeCapacity:  dto.LargeCapacity,
		AutoCall:       dto.AutoCall,
		CallInterval:   dto.CallInterval,
		MaxCallCount:   dto.MaxCallCount,
		AutoExpire:     dto.AutoExpire,
		ExpireMinutes:  dto.ExpireMinutes,
		VoiceNotify:    dto.VoiceNotify,
		SMSNotify:      dto.SMSNotify,
	}
	return s.queueRepo.SaveConfig(config)
}

func (s *TableService) GetMyQueue(memberID uint, storeID uint) ([]model.QueueNumber, error) {
	return s.queueRepo.GetMyQueue(memberID, storeID)
}

func (s *TableService) GetWaitingCount(storeID uint) (map[string]int, error) {
	result := make(map[string]int)
	for _, t := range []string{"small", "medium", "large"} {
		count, err := s.queueRepo.GetWaitingCount(storeID, t)
		if err == nil {
			result[t] = count
		}
	}
	return result, nil
}

func splitScene(scene string) []string {
	return strings.Split(scene, "_")
}

func generateTimeSlots(openTime, closeTime string) []string {
	slots := make([]string, 0)
	if openTime == "" || closeTime == "" {
		openTime = "10:00"
		closeTime = "22:00"
	}
	startHour, _ := strconv.Atoi(openTime[:2])
	startMin, _ := strconv.Atoi(openTime[3:])
	endHour, _ := strconv.Atoi(closeTime[:2])
	endMin, _ := strconv.Atoi(closeTime[3:])
	current := startHour*60 + startMin
	end := endHour*60 + endMin
	interval := 30
	for current < end {
		h := current / 60
		m := current % 60
		slot := fmt.Sprintf("%02d:%02d", h, m)
		slots = append(slots, slot)
		current += interval
	}
	return slots
}
