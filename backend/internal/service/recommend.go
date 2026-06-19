package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	redisPkg "stalll-hub-pos/backend/pkg/redis"

	"github.com/shopspring/decimal"
)

const (
	redisRecommendKeyPrefix  = "pos:recommend:"
	redisConfigKey           = "pos:recommend:config:"
	redisRefreshLockPrefix   = "pos:recommend:lock:"
	redisRecommendTTL        = 24 * time.Hour
	redisRefreshLockTTL      = 30 * time.Minute
	redisHotProductsKey      = "pos:recommend:hot:"
)

type RecommendService struct {
	recRepo        *repository.RecommendRepository
	productRepo    *repository.ProductRepository
	runningStores  map[uint]bool
	runningMutex   sync.Mutex
}

func NewRecommendService() *RecommendService {
	return &RecommendService{
		recRepo:     repository.NewRecommendRepository(),
		productRepo: repository.NewProductRepository(),
		runningStores: make(map[uint]bool),
	}
}

func (s *RecommendService) GetOrCreateConfig(storeID uint) (*model.RecommendConfig, error) {
	cfg, err := s.recRepo.GetConfigByStoreID(storeID)
	if err == nil && cfg != nil {
		return cfg, nil
	}

	cfg = &model.RecommendConfig{
		StoreID:                 storeID,
		CFWeight:                0.6,
		HotWeight:               0.3,
		CategoryDiversityWeight: 0.1,
		RecommendCount:          8,
		MinOrderPairs:           3,
		MinSimilarity:           0.05,
		HotDays:                 30,
		CFDays:                  90,
		Enabled:                 true,
		AutoRefresh:             true,
		RefreshIntervalHours:    6,
	}
	err = s.recRepo.CreateConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *RecommendService) UpdateConfig(storeID uint, req *dto.UpdateRecommendConfigRequest) (*model.RecommendConfig, error) {
	cfg, err := s.GetOrCreateConfig(storeID)
	if err != nil {
		return nil, err
	}

	if req.CFWeight != nil {
		cfg.CFWeight = *req.CFWeight
	}
	if req.HotWeight != nil {
		cfg.HotWeight = *req.HotWeight
	}
	if req.CategoryDiversityWeight != nil {
		cfg.CategoryDiversityWeight = *req.CategoryDiversityWeight
	}
	if req.RecommendCount != nil {
		cfg.RecommendCount = *req.RecommendCount
	}
	if req.MinOrderPairs != nil {
		cfg.MinOrderPairs = *req.MinOrderPairs
	}
	if req.MinSimilarity != nil {
		cfg.MinSimilarity = *req.MinSimilarity
	}
	if req.HotDays != nil {
		cfg.HotDays = *req.HotDays
	}
	if req.CFDays != nil {
		cfg.CFDays = *req.CFDays
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.AutoRefresh != nil {
		cfg.AutoRefresh = *req.AutoRefresh
	}
	if req.RefreshIntervalHours != nil {
		cfg.RefreshIntervalHours = *req.RefreshIntervalHours
	}

	err = s.recRepo.UpdateConfig(cfg)
	if err != nil {
		return nil, err
	}
	s.cacheConfig(cfg)
	s.clearRecommendCache(storeID)
	return cfg, nil
}

func (s *RecommendService) TriggerRefresh(storeID uint) error {
	s.runningMutex.Lock()
	if s.runningStores[storeID] {
		s.runningMutex.Unlock()
		return fmt.Errorf("门店 %d 推荐刷新正在运行中，请稍候", storeID)
	}
	s.runningStores[storeID] = true
	s.runningMutex.Unlock()

	defer func() {
		s.runningMutex.Lock()
		delete(s.runningStores, storeID)
		s.runningMutex.Unlock()
	}()

	lockKey := fmt.Sprintf("%s%d", redisRefreshLockPrefix, storeID)
	ok, err := redisPkg.Client.SetNX(redisPkg.Ctx, lockKey, "1", redisRefreshLockTTL).Result()
	if err != nil {
		log.Printf("[Recommend] 获取Redis锁失败: %v", err)
	}
	if !ok {
		return fmt.Errorf("推荐刷新任务已在运行，请稍候")
	}
	defer redisPkg.Client.Del(redisPkg.Ctx, lockKey)

	return s.computeAndSaveRecommendations(storeID)
}

func (s *RecommendService) computeAndSaveRecommendations(storeID uint) error {
	cfg, err := s.GetOrCreateConfig(storeID)
	if err != nil {
		return err
	}

	start := time.Now()
	log.Printf("[Recommend] 开始计算门店 %d 推荐结果...", storeID)

	cfMatrix, err := s.computeItemCF(storeID, cfg.CFDays, cfg.MinOrderPairs, cfg.MinSimilarity)
	if err != nil {
		log.Printf("[Recommend] 协同过滤计算出错: %v", err)
		cfMatrix = make(map[uint]map[uint]float64)
	}

	hotList, err := s.recRepo.GetHotProducts(storeID, cfg.HotDays, 100)
	if err != nil {
		log.Printf("[Recommend] 热门商品获取出错: %v", err)
		hotList = []model.HotProduct{}
	}
	hotMap := make(map[uint]float64)
	for _, h := range hotList {
		hotMap[h.ProductID] = h.HotScore
	}
	s.cacheHotProducts(storeID, hotList)

	products, err := s.recRepo.GetValidProducts(storeID)
	if err != nil {
		return err
	}
	productCategoryMap := make(map[uint]uint)
	for _, p := range products {
		productCategoryMap[p.ID] = p.CategoryID
	}

	totalWeight := cfg.CFWeight + cfg.HotWeight + cfg.CategoryDiversityWeight
	if totalWeight <= 0 {
		totalWeight = 1.0
	}

	var allResults []model.RecommendResult
	maxResultsPerProduct := cfg.RecommendCount * 2

	for _, product := range products {
		pid := product.ID
		scoreMap := make(map[uint]*model.RecommendResult)

		if simMap, ok := cfMatrix[pid]; ok {
			for otherID, sim := range simMap {
				if otherID == pid {
					continue
				}
				weighted := sim * cfg.CFWeight / totalWeight
				scoreMap[otherID] = &model.RecommendResult{
					StoreID:            storeID,
					ProductID:          pid,
					RecommendProductID: otherID,
					CFScore:            sim,
					Score:              weighted,
					Reason:             "常一起购买",
				}
			}
		}

		for hotID, hotScore := range hotMap {
			if hotID == pid {
				continue
			}
			weighted := hotScore * cfg.HotWeight / totalWeight
			if r, exists := scoreMap[hotID]; exists {
				r.HotScore = hotScore
				r.Score += weighted
			} else {
				scoreMap[hotID] = &model.RecommendResult{
					StoreID:            storeID,
					ProductID:          pid,
					RecommendProductID: hotID,
					HotScore:           hotScore,
					Score:              weighted,
					Reason:             "热门推荐",
				}
			}
		}

		ranked := make([]*model.RecommendResult, 0, len(scoreMap))
		for _, r := range scoreMap {
			if r.Score < cfg.MinSimilarity*cfg.CFWeight/totalWeight && r.HotScore <= 0 {
				continue
			}
			ranked = append(ranked, r)
		}
		sort.Slice(ranked, func(i, j int) bool {
			return ranked[i].Score > ranked[j].Score
		})

		if cfg.CategoryDiversityWeight > 0 && len(ranked) > cfg.RecommendCount {
			ranked = s.applyCategoryDiversity(ranked, productCategoryMap, cfg.RecommendCount)
		}

		for i, r := range ranked {
			if i >= maxResultsPerProduct {
				break
			}
			allResults = append(allResults, *r)
		}
	}

	if err := s.recRepo.ClearResultsByStore(storeID); err != nil {
		log.Printf("[Recommend] 清理旧推荐结果失败: %v", err)
	}

	if len(allResults) > 0 {
		if err := s.recRepo.BatchCreateResults(allResults); err != nil {
			log.Printf("[Recommend] 批量写入推荐结果失败: %v", err)
			return err
		}
	}

	now := time.Now()
	cfg.LastRefreshedAt = &now
	if err := s.recRepo.UpdateConfig(cfg); err != nil {
		log.Printf("[Recommend] 更新配置时间戳失败: %v", err)
	}
	s.cacheConfig(cfg)
	s.clearRecommendCache(storeID)

	log.Printf("[Recommend] 门店 %d 推荐计算完成，共 %d 条结果，耗时 %v",
		storeID, len(allResults), time.Since(start))
	return nil
}

func (s *RecommendService) computeItemCF(storeID uint, days int, minPairs int, minSim float64) (map[uint]map[uint]float64, error) {
	items, err := s.recRepo.GetOrderItemsForCF(storeID, days)
	if err != nil {
		return nil, err
	}

	orderProducts := make(map[uint]map[uint]bool)
	productOrderCount := make(map[uint]int)

	for _, it := range items {
		if _, ok := orderProducts[it.OrderID]; !ok {
			orderProducts[it.OrderID] = make(map[uint]bool)
		}
		orderProducts[it.OrderID][it.ProductID] = true
	}

	for _, products := range orderProducts {
		for pid := range products {
			productOrderCount[pid]++
		}
	}

	pairCount := make(map[[2]uint]int)
	for _, products := range orderProducts {
		pids := make([]uint, 0, len(products))
		for pid := range products {
			pids = append(pids, pid)
		}
		for i := 0; i < len(pids); i++ {
			for j := i + 1; j < len(pids); j++ {
				a, b := pids[i], pids[j]
				if a > b {
					a, b = b, a
				}
				pairCount[[2]uint{a, b}]++
			}
		}
	}

	simMatrix := make(map[uint]map[uint]float64)
	for pair, cnt := range pairCount {
		if cnt < minPairs {
			continue
		}
		a, b := pair[0], pair[1]
		countA := productOrderCount[a]
		countB := productOrderCount[b]
		if countA == 0 || countB == 0 {
			continue
		}
		sim := float64(cnt) / math.Sqrt(float64(countA)*float64(countB))
		if sim < minSim {
			continue
		}
		if _, ok := simMatrix[a]; !ok {
			simMatrix[a] = make(map[uint]float64)
		}
		if _, ok := simMatrix[b]; !ok {
			simMatrix[b] = make(map[uint]float64)
		}
		simMatrix[a][b] = sim
		simMatrix[b][a] = sim
	}

	log.Printf("[Recommend] ItemCF: 订单数=%d, 商品数=%d, 共现对=%d, 有效相似对=%d",
		len(orderProducts), len(productOrderCount), len(pairCount), len(simMatrix))
	return simMatrix, nil
}

func (s *RecommendService) applyCategoryDiversity(ranked []*model.RecommendResult, categoryMap map[uint]uint, limit int) []*model.RecommendResult {
	selected := make([]*model.RecommendResult, 0, limit)
	categoryUsed := make(map[uint]int)

	for _, r := range ranked {
		catID := categoryMap[r.RecommendProductID]
		catCount := categoryUsed[catID]
		boost := 1.0 / math.Pow(float64(catCount+1), 0.5)
		r.Score *= boost
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Score > ranked[j].Score
	})

	for _, r := range ranked {
		if len(selected) >= limit {
			break
		}
		selected = append(selected, r)
		catID := categoryMap[r.RecommendProductID]
		categoryUsed[catID]++
	}
	return selected
}

func (s *RecommendService) GetCartRecommendations(storeID uint, productIDs []uint, count int) ([]dto.RecommendItemDTO, error) {
	cfg, err := s.GetOrCreateConfig(storeID)
	if err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		return s.getFallbackHotRecommendations(storeID, count)
	}

	if count <= 0 {
		count = cfg.RecommendCount
	}

	cacheKey := s.buildCacheKey(storeID, productIDs, count)
	if cached, err := s.getRecommendFromCache(cacheKey); err == nil && len(cached) > 0 {
		return cached, nil
	}

	var inputIDs []uint
	for _, pid := range productIDs {
		if pid > 0 {
			inputIDs = append(inputIDs, pid)
		}
	}

	results := []model.RecommendResult{}
	if len(inputIDs) > 0 {
		dbResults, err := s.recRepo.GetResultsByProducts(storeID, inputIDs, count)
		if err == nil {
			results = dbResults
		}
	}

	merged := make(map[uint]*dto.RecommendItemDTO)
	inputSet := make(map[uint]bool)
	for _, pid := range inputIDs {
		inputSet[pid] = true
	}

	for _, r := range results {
		if inputSet[r.RecommendProductID] {
			continue
		}
		if r.RecommendProduct.ID == 0 || r.RecommendProduct.Status != 1 {
			continue
		}
		if existing, ok := merged[r.RecommendProductID]; ok {
			if r.Score > existing.Score {
				existing.Score = r.Score
				existing.Reason = r.Reason
			}
			continue
		}
		price := decimal.Zero
		if len(r.RecommendProduct.SKUs) > 0 {
			price = r.RecommendProduct.SKUs[0].Price
		}
		merged[r.RecommendProductID] = &dto.RecommendItemDTO{
			ProductID:   r.RecommendProductID,
			ProductName: r.RecommendProduct.Name,
			CategoryID:  r.RecommendProduct.CategoryID,
			MainImage:   r.RecommendProduct.MainImage,
			Price:       price.String(),
			Score:       r.Score,
			Reason:      r.Reason,
		}
	}

	list := make([]dto.RecommendItemDTO, 0, len(merged))
	for _, item := range merged {
		list = append(list, *item)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	if len(list) < count {
		hotItems, _ := s.getFallbackHotRecommendations(storeID, count)
		for _, hi := range hotItems {
			if _, exists := merged[hi.ProductID]; exists {
				continue
			}
			if inputSet[hi.ProductID] {
				continue
			}
			list = append(list, hi)
			merged[hi.ProductID] = &hi
			if len(list) >= count {
				break
			}
		}
	}

	if len(list) > count {
		list = list[:count]
	}

	s.setRecommendCache(cacheKey, list)
	return list, nil
}

func (s *RecommendService) getFallbackHotRecommendations(storeID uint, count int) ([]dto.RecommendItemDTO, error) {
	if count <= 0 {
		count = 8
	}
	hotList, err := s.getCachedHotProducts(storeID)
	if err != nil || len(hotList) == 0 {
		hotList, err = s.recRepo.GetHotProducts(storeID, 30, count)
		if err != nil {
			return []dto.RecommendItemDTO{}, nil
		}
		s.cacheHotProducts(storeID, hotList)
	}

	productIDs := make([]uint, 0, len(hotList))
	for _, h := range hotList {
		productIDs = append(productIDs, h.ProductID)
	}

	result := make([]dto.RecommendItemDTO, 0, count)
	for i, h := range hotList {
		if i >= count {
			break
		}
		product, err := s.productRepo.GetByID(h.ProductID)
		price := decimal.Zero
		if err == nil && product != nil && len(product.SKUs) > 0 {
			price = product.SKUs[0].Price
		}
		result = append(result, dto.RecommendItemDTO{
			ProductID:   h.ProductID,
			ProductName: h.ProductName,
			CategoryID:  h.CategoryID,
			Score:       h.HotScore,
			Price:       price.String(),
			Reason:      "热门推荐",
		})
	}
	return result, nil
}

func (s *RecommendService) GetRefreshStatus(storeID uint) (*dto.RefreshStatusDTO, error) {
	cfg, err := s.GetOrCreateConfig(storeID)
	if err != nil {
		return nil, err
	}
	s.runningMutex.Lock()
	isRunning := s.runningStores[storeID]
	s.runningMutex.Unlock()

	productCount, pairCount := s.recRepo.GetResultStats(storeID)
	lastStr := ""
	if cfg.LastRefreshedAt != nil {
		lastStr = cfg.LastRefreshedAt.Format("2006-01-02 15:04:05")
	}
	return &dto.RefreshStatusDTO{
		StoreID:         storeID,
		Enabled:         cfg.Enabled,
		LastRefreshedAt: lastStr,
		IsRunning:       isRunning,
		TotalProducts:   productCount,
		TotalPairs:      pairCount,
	}, nil
}

func (s *RecommendService) StartAutoRefreshScheduler() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		log.Println("[Recommend] 自动刷新调度器已启动")
		s.scheduleAllStores()
		for range ticker.C {
			s.scheduleAllStores()
		}
	}()
}

func (s *RecommendService) scheduleAllStores() {
	storeIDs, err := s.recRepo.GetAllStoreIDs()
	if err != nil {
		log.Printf("[Recommend] 获取门店列表失败: %v", err)
		return
	}
	for _, sid := range storeIDs {
		cfg, err := s.recRepo.GetConfigByStoreID(sid)
		if err != nil || cfg == nil || !cfg.AutoRefresh {
			continue
		}
		if cfg.LastRefreshedAt != nil {
			nextRun := cfg.LastRefreshedAt.Add(time.Duration(cfg.RefreshIntervalHours) * time.Hour)
			if time.Now().Before(nextRun) {
				continue
			}
		}
		go func(storeID uint) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Recommend] 门店 %d 刷新异常: %v", storeID, r)
				}
			}()
			if err := s.TriggerRefresh(storeID); err != nil {
				log.Printf("[Recommend] 门店 %d 自动刷新跳过: %v", storeID, err)
			}
		}(sid)
	}
}

func (s *RecommendService) buildCacheKey(storeID uint, productIDs []uint, count int) string {
	sortedIDs := make([]uint, len(productIDs))
	copy(sortedIDs, productIDs)
	sort.Slice(sortedIDs, func(i, j int) bool { return sortedIDs[i] < sortedIDs[j] })
	return fmt.Sprintf("%s%d:%v:%d", redisRecommendKeyPrefix, storeID, sortedIDs, count)
}

func (s *RecommendService) setRecommendCache(key string, data []dto.RecommendItemDTO) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	redisPkg.Client.Set(redisPkg.Ctx, key, string(b), redisRecommendTTL)
}

func (s *RecommendService) getRecommendFromCache(key string) ([]dto.RecommendItemDTO, error) {
	val, err := redisPkg.Client.Get(redisPkg.Ctx, key).Result()
	if err != nil || val == "" {
		return nil, err
	}
	var result []dto.RecommendItemDTO
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RecommendService) cacheConfig(cfg *model.RecommendConfig) {
	key := fmt.Sprintf("%s%d", redisConfigKey, cfg.StoreID)
	b, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	redisPkg.Client.Set(redisPkg.Ctx, key, string(b), 2*redisRecommendTTL)
}

func (s *RecommendService) cacheHotProducts(storeID uint, hots []model.HotProduct) {
	key := fmt.Sprintf("%s%d", redisHotProductsKey, storeID)
	b, err := json.Marshal(hots)
	if err != nil {
		return
	}
	redisPkg.Client.Set(redisPkg.Ctx, key, string(b), redisRecommendTTL)
}

func (s *RecommendService) getCachedHotProducts(storeID uint) ([]model.HotProduct, error) {
	key := fmt.Sprintf("%s%d", redisHotProductsKey, storeID)
	val, err := redisPkg.Client.Get(redisPkg.Ctx, key).Result()
	if err != nil || val == "" {
		return nil, err
	}
	var result []model.HotProduct
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RecommendService) clearRecommendCache(storeID uint) {
	pattern := fmt.Sprintf("%s%d:*", redisRecommendKeyPrefix, storeID)
	var cursor uint64
	for {
		keys, next, err := redisPkg.Client.Scan(redisPkg.Ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			redisPkg.Client.Del(redisPkg.Ctx, keys...)
		}
		if next == 0 {
			break
		}
		cursor = next
	}
}
