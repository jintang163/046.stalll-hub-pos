package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/redis"
	"time"
)

const (
	QueueTypeSmall  = "small"
	QueueTypeMedium = "medium"
	QueueTypeLarge  = "large"

	QueueStatusWaiting   = 1
	QueueStatusCalled    = 2
	QueueStatusArrived   = 3
	QueueStatusCancelled = 4
	QueueStatusExpired   = 5

	queueKeyPrefix       = "queue:"
	queueNumberKeyPrefix = "queue:number:"
	queueSeqKeyPrefix    = "queue:seq:"
	queueConfigKey       = "queue:config:"
	preOrderKeyPrefix    = "queue:preorder:"
)

type QueueInfo struct {
	StoreID      uint              `json:"store_id"`
	QueueType    string            `json:"queue_type"`
	QueuePrefix  string            `json:"queue_prefix"`
	WaitCount    int               `json:"wait_count"`
	CurrentNum   string            `json:"current_num"`
	LatestNumbers []string         `json:"latest_numbers"`
	MyNumber     *QueueNumberInfo  `json:"my_number,omitempty"`
}

type QueueNumberInfo struct {
	QueueNumber string `json:"queue_number"`
	Sequence    int    `json:"sequence"`
	AheadCount  int    `json:"ahead_count"`
	PeopleCount int    `json:"people_count"`
	Status      int    `json:"status"`
	CallCount   int    `json:"call_count"`
	CreatedAt   string `json:"created_at"`
	TableNo     string `json:"table_no,omitempty"`
}

type PreOrderItem struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	SkuID       uint    `json:"sku_id"`
	SkuName     string  `json:"sku_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type PreOrder struct {
	QueueID     string        `json:"queue_id"`
	QueueNumber string        `json:"queue_number"`
	StoreID     uint          `json:"store_id"`
	MemberID    uint          `json:"member_id"`
	Items       []PreOrderItem `json:"items"`
	TotalAmount float64       `json:"total_amount"`
	Remark      string        `json:"remark"`
	UpdatedAt   string        `json:"updated_at"`
}

type QueueService struct{}

func NewQueueService() *QueueService {
	return &QueueService{}
}

func (s *QueueService) getQueueKey(storeID uint, queueType string) string {
	return fmt.Sprintf("%s%d:%s", queueKeyPrefix, storeID, queueType)
}

func (s *QueueService) getSeqKey(storeID uint, queueType string) string {
	return fmt.Sprintf("%s%d:%s", queueSeqKeyPrefix, storeID, queueType)
}

func (s *QueueService) getNumberKey(queueID string) string {
	return queueNumberKeyPrefix + queueID
}

func (s *QueueService) getConfigKey(storeID uint) string {
	return queueConfigKey + strconv.Itoa(int(storeID))
}

func (s *QueueService) getPreOrderKey(queueID string) string {
	return preOrderKeyPrefix + queueID
}

func (s *QueueService) GetQueueConfig(storeID uint) (*model.QueueConfig, error) {
	configKey := s.getConfigKey(storeID)
	configMap, err := redis.HGetAll(configKey)
	if err == nil && len(configMap) > 0 {
		cfg := &model.QueueConfig{StoreID: storeID}
		if v, ok := configMap["small_prefix"]; ok { cfg.SmallPrefix = v }
		if v, ok := configMap["small_capacity"]; ok { 
			if n, e := strconv.Atoi(v); e == nil { cfg.SmallCapacity = n }
		}
		if v, ok := configMap["medium_prefix"]; ok { cfg.MediumPrefix = v }
		if v, ok := configMap["medium_capacity"]; ok { 
			if n, e := strconv.Atoi(v); e == nil { cfg.MediumCapacity = n }
		}
		if v, ok := configMap["large_prefix"]; ok { cfg.LargePrefix = v }
		if v, ok := configMap["large_capacity"]; ok { 
			if n, e := strconv.Atoi(v); e == nil { cfg.LargeCapacity = n }
		}
		if v, ok := configMap["expire_minutes"]; ok { 
			if n, e := strconv.Atoi(v); e == nil { cfg.ExpireMinutes = n }
		}
		if v, ok := configMap["max_call_count"]; ok { 
			if n, e := strconv.Atoi(v); e == nil { cfg.MaxCallCount = n }
		}
		if v, ok := configMap["voice_notify"]; ok { 
			cfg.VoiceNotify = v == "1" || v == "true"
		}
		return cfg, nil
	}

	cfg := &model.QueueConfig{
		StoreID:        storeID,
		SmallPrefix:    "A",
		SmallCapacity:  4,
		MediumPrefix:   "B",
		MediumCapacity: 6,
		LargePrefix:    "C",
		LargeCapacity:  10,
		AutoCall:       true,
		CallInterval:   300,
		MaxCallCount:   3,
		AutoExpire:     true,
		ExpireMinutes:  15,
		VoiceNotify:    true,
		SMSNotify:      false,
	}

	redis.HSet(configKey,
		"small_prefix", cfg.SmallPrefix,
		"small_capacity", strconv.Itoa(cfg.SmallCapacity),
		"medium_prefix", cfg.MediumPrefix,
		"medium_capacity", strconv.Itoa(cfg.MediumCapacity),
		"large_prefix", cfg.LargePrefix,
		"large_capacity", strconv.Itoa(cfg.LargeCapacity),
		"expire_minutes", strconv.Itoa(cfg.ExpireMinutes),
		"max_call_count", strconv.Itoa(cfg.MaxCallCount),
		"voice_notify", "1",
	)
	redis.Expire(configKey, 24*time.Hour)
	return cfg, nil
}

func (s *QueueService) determineQueueType(peopleCount int) string {
	if peopleCount <= 4 {
		return QueueTypeSmall
	} else if peopleCount <= 6 {
		return QueueTypeMedium
	}
	return QueueTypeLarge
}

func (s *QueueService) getQueuePrefix(cfg *model.QueueConfig, queueType string) string {
	switch queueType {
	case QueueTypeSmall:
		return cfg.SmallPrefix
	case QueueTypeMedium:
		return cfg.MediumPrefix
	case QueueTypeLarge:
		return cfg.LargePrefix
	}
	return "A"
}

func (s *QueueService) TakeNumber(storeID uint, memberID uint, memberName, memberPhone string, peopleCount int, remark string) (*QueueNumberInfo, error) {
	if peopleCount <= 0 {
		return nil, errors.New("invalid people count")
	}

	queueType := s.determineQueueType(peopleCount)
	cfg, err := s.GetQueueConfig(storeID)
	if err != nil {
		return nil, err
	}
	prefix := s.getQueuePrefix(cfg, queueType)

	seqKey := s.getSeqKey(storeID, queueType)
	seq, err := redis.Incr(seqKey)
	if err != nil {
		return nil, err
	}
	redis.Expire(seqKey, 24*time.Hour)

	queueNumber := fmt.Sprintf("%s%03d", prefix, seq)
	queueID := fmt.Sprintf("%d:%s", storeID, queueNumber)

	now := time.Now()
	numInfo := map[string]interface{}{
		"queue_id":     queueID,
		"queue_number": queueNumber,
		"sequence":     strconv.FormatInt(seq, 10),
		"store_id":     strconv.FormatUint(uint64(storeID), 10),
		"member_id":    strconv.FormatUint(uint64(memberID), 10),
		"member_name":  memberName,
		"member_phone": memberPhone,
		"people_count": strconv.Itoa(peopleCount),
		"queue_type":   queueType,
		"status":       strconv.Itoa(QueueStatusWaiting),
		"call_count":   "0",
		"remark":       remark,
		"created_at":   now.Format(time.RFC3339),
	}

	numKey := s.getNumberKey(queueID)
	for k, v := range numInfo {
		redis.HSet(numKey, k, v)
	}
	redis.Expire(numKey, 24*time.Hour)

	queueKey := s.getQueueKey(storeID, queueType)
	redis.RPush(queueKey, queueID)
	redis.Expire(queueKey, 24*time.Hour)

	waitCount, _ := redis.LLen(queueKey)
	aheadCount := int(waitCount) - 1
	if aheadCount < 0 {
		aheadCount = 0
	}

	return &QueueNumberInfo{
		QueueNumber: queueNumber,
		Sequence:    int(seq),
		AheadCount:  aheadCount,
		PeopleCount: peopleCount,
		Status:      QueueStatusWaiting,
		CallCount:   0,
		CreatedAt:   now.Format(time.RFC3339),
	}, nil
}

func (s *QueueService) GetQueueInfo(storeID uint, queueType string, queueID string) (*QueueInfo, error) {
	cfg, err := s.GetQueueConfig(storeID)
	if err != nil {
		return nil, err
	}

	queueKey := s.getQueueKey(storeID, queueType)
	waitCount, _ := redis.LLen(queueKey)

	prefix := s.getQueuePrefix(cfg, queueType)
	latest, _ := redis.LRange(queueKey, 0, 4)
	latestNumbers := make([]string, 0, len(latest))
	for _, qid := range latest {
		parts := strings.SplitN(qid, ":", 2)
		if len(parts) == 2 {
			latestNumbers = append(latestNumbers, parts[1])
		}
	}

	info := &QueueInfo{
		StoreID:       storeID,
		QueueType:     queueType,
		QueuePrefix:   prefix,
		WaitCount:     int(waitCount),
		CurrentNum:    "",
		LatestNumbers: latestNumbers,
	}

	if len(latestNumbers) > 0 {
		info.CurrentNum = latestNumbers[0]
	}

	if queueID != "" {
		numKey := s.getNumberKey(queueID)
		numMap, err := redis.HGetAll(numKey)
		if err == nil && len(numMap) > 0 {
			myInfo := &QueueNumberInfo{
				QueueNumber: numMap["queue_number"],
				PeopleCount: 0,
				Status:      0,
				CallCount:   0,
				CreatedAt:   numMap["created_at"],
				TableNo:     numMap["table_no"],
			}
			if v, ok := numMap["sequence"]; ok {
				if n, e := strconv.Atoi(v); e == nil { myInfo.Sequence = n }
			}
			if v, ok := numMap["people_count"]; ok {
				if n, e := strconv.Atoi(v); e == nil { myInfo.PeopleCount = n }
			}
			if v, ok := numMap["status"]; ok {
				if n, e := strconv.Atoi(v); e == nil { myInfo.Status = n }
			}
			if v, ok := numMap["call_count"]; ok {
				if n, e := strconv.Atoi(v); e == nil { myInfo.CallCount = n }
			}

			rank, err := redis.LIndex(queueKey, 0)
			if err == nil && rank == queueID {
				myInfo.AheadCount = 0
			} else {
				ahead, _ := s.getAheadCount(queueKey, queueID)
				myInfo.AheadCount = ahead
			}
			info.MyNumber = myInfo
		}
	}

	return info, nil
}

func (s *QueueService) getAheadCount(queueKey, queueID string) (int, error) {
	all, err := redis.LRange(queueKey, 0, -1)
	if err != nil {
		return 0, err
	}
	for i, id := range all {
		if id == queueID {
			return i, nil
		}
	}
	return 0, nil
}

func (s *QueueService) CallNumber(storeID uint, queueType string) (*QueueNumberInfo, error) {
	queueKey := s.getQueueKey(storeID, queueType)

	queueID, err := redis.LIndex(queueKey, 0)
	if err != nil || queueID == "" {
		return nil, errors.New("no waiting numbers")
	}

	numKey := s.getNumberKey(queueID)
	numMap, err := redis.HGetAll(numKey)
	if err != nil || len(numMap) == 0 {
		redis.LPop(queueKey)
		return nil, errors.New("queue number not found")
	}

	callCount := 0
	if v, ok := numMap["call_count"]; ok {
		if n, e := strconv.Atoi(v); e == nil { callCount = n }
	}
	callCount++
	redis.HSet(numKey, "call_count", strconv.Itoa(callCount))
	redis.HSet(numKey, "status", strconv.Itoa(QueueStatusCalled))
	redis.HSet(numKey, "last_call_time", time.Now().Format(time.RFC3339))

	info := &QueueNumberInfo{
		QueueNumber: numMap["queue_number"],
		CallCount:   callCount,
		Status:      QueueStatusCalled,
	}
	if v, ok := numMap["people_count"]; ok {
		if n, e := strconv.Atoi(v); e == nil { info.PeopleCount = n }
	}
	if v, ok := numMap["sequence"]; ok {
		if n, e := strconv.Atoi(v); e == nil { info.Sequence = n }
	}

	callMsg := map[string]interface{}{
		"type":         "call",
		"store_id":     storeID,
		"queue_type":   queueType,
		"queue_id":     queueID,
		"queue_number": numMap["queue_number"],
		"call_count":   callCount,
		"timestamp":    time.Now().Unix(),
	}
	callJSON, _ := json.Marshal(callMsg)
	redis.Publish(fmt.Sprintf("queue:call:%d", storeID), string(callJSON))

	return info, nil
}

func (s *QueueService) Arrive(storeID uint, queueID string, tableNo string) error {
	queueKey := s.getQueueKey(storeID, "")
	queueIDFound := false

	types := []string{QueueTypeSmall, QueueTypeMedium, QueueTypeLarge}
	for _, t := range types {
		qk := s.getQueueKey(storeID, t)
		all, err := redis.LRange(qk, 0, -1)
		if err == nil {
			for _, id := range all {
				if id == queueID {
					redis.LRem(qk, 1, queueID)
					queueKey = qk
					queueIDFound = true
					break
				}
			}
		}
		if queueIDFound {
			break
		}
	}

	numKey := s.getNumberKey(queueID)
	redis.HSet(numKey,
		"status", strconv.Itoa(QueueStatusArrived),
		"table_no", tableNo,
		"arrive_time", time.Now().Format(time.RFC3339),
	)

	arriveMsg := map[string]interface{}{
		"type":       "arrive",
		"store_id":   storeID,
		"queue_id":   queueID,
		"table_no":   tableNo,
		"timestamp":  time.Now().Unix(),
	}
	arriveJSON, _ := json.Marshal(arriveMsg)
	redis.Publish(fmt.Sprintf("queue:call:%d", storeID), string(arriveJSON))

	return nil
}

func (s *QueueService) Cancel(storeID uint, queueID string) error {
	types := []string{QueueTypeSmall, QueueTypeMedium, QueueTypeLarge}
	found := false
	for _, t := range types {
		qk := s.getQueueKey(storeID, t)
		redis.LRem(qk, 1, queueID)
		found = true
	}
	if !found {
		return errors.New("queue number not found")
	}

	numKey := s.getNumberKey(queueID)
	redis.HSet(numKey,
		"status", strconv.Itoa(QueueStatusCancelled),
		"cancel_time", time.Now().Format(time.RFC3339),
	)

	cancelMsg := map[string]interface{}{
		"type":      "cancel",
		"store_id":  storeID,
		"queue_id":  queueID,
		"timestamp": time.Now().Unix(),
	}
	cancelJSON, _ := json.Marshal(cancelMsg)
	redis.Publish(fmt.Sprintf("queue:call:%d", storeID), string(cancelJSON))

	return nil
}

func (s *QueueService) SavePreOrder(queueID string, storeID uint, memberID uint, items []PreOrderItem, totalAmount float64, remark string) error {
	if queueID == "" {
		return errors.New("queue id required")
	}

	preOrder := PreOrder{
		QueueID:     queueID,
		StoreID:     storeID,
		MemberID:    memberID,
		Items:       items,
		TotalAmount: totalAmount,
		Remark:      remark,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	key := s.getPreOrderKey(queueID)
	data, err := json.Marshal(preOrder)
	if err != nil {
		return err
	}

	numMap, _ := redis.HGetAll(s.getNumberKey(queueID))
	if qn, ok := numMap["queue_number"]; ok {
		preOrder.QueueNumber = qn
		data, _ = json.Marshal(preOrder)
	}

	redis.Set(key, string(data), 24*time.Hour)
	return nil
}

func (s *QueueService) GetPreOrder(queueID string) (*PreOrder, error) {
	key := s.getPreOrderKey(queueID)
	data, err := redis.Get(key)
	if err != nil || data == "" {
		return nil, errors.New("pre-order not found")
	}

	var preOrder PreOrder
	if err := json.Unmarshal([]byte(data), &preOrder); err != nil {
		return nil, err
	}

	return &preOrder, nil
}

func (s *QueueService) GetAllWaiting(storeID uint) (map[string][]QueueNumberInfo, error) {
	result := make(map[string][]QueueNumberInfo)
	types := []string{QueueTypeSmall, QueueTypeMedium, QueueTypeLarge}

	for _, t := range types {
		qk := s.getQueueKey(storeID, t)
		ids, err := redis.LRange(qk, 0, -1)
		if err != nil {
			continue
		}

		list := make([]QueueNumberInfo, 0, len(ids))
		for _, qid := range ids {
			numKey := s.getNumberKey(qid)
			numMap, err := redis.HGetAll(numKey)
			if err != nil || len(numMap) == 0 {
				continue
			}
			info := QueueNumberInfo{
				QueueNumber: numMap["queue_number"],
				Status:      0,
				CallCount:   0,
				CreatedAt:   numMap["created_at"],
			}
			if v, ok := numMap["sequence"]; ok {
				if n, e := strconv.Atoi(v); e == nil { info.Sequence = n }
			}
			if v, ok := numMap["people_count"]; ok {
				if n, e := strconv.Atoi(v); e == nil { info.PeopleCount = n }
			}
			if v, ok := numMap["status"]; ok {
				if n, e := strconv.Atoi(v); e == nil { info.Status = n }
			}
			if v, ok := numMap["call_count"]; ok {
				if n, e := strconv.Atoi(v); e == nil { info.CallCount = n }
			}
			if v, ok := numMap["member_name"]; ok {
				info.TableNo = v
			}
			list = append(list, info)
		}
		result[t] = list
	}

	return result, nil
}
