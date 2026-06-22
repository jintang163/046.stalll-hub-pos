package service

import (
	"fmt"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"time"
)

var adTypeMap = map[string]string{
	"image":  "图片",
	"qrcode": "二维码",
	"text":   "文字",
}

var adPositionMap = map[string]string{
	"header": "顶部",
	"footer": "底部",
}

var adStatusMap = map[int]string{
	0: "禁用",
	1: "启用",
}

type ReceiptAdService struct {
	adRepo    *repository.ReceiptAdRepository
	clickRepo *repository.ReceiptAdClickRepository
}

func NewReceiptAdService() *ReceiptAdService {
	return &ReceiptAdService{
		adRepo:    repository.NewReceiptAdRepository(),
		clickRepo: repository.NewReceiptAdClickRepository(),
	}
}

func (s *ReceiptAdService) CreateReceiptAd(storeID uint, req *dto.ReceiptAdDTO) (*model.ReceiptAd, error) {
	ad := &model.ReceiptAd{
		StoreID:       storeID,
		Title:         req.Title,
		AdType:        req.AdType,
		ImageURL:      req.ImageURL,
		QRCodeContent: req.QRCodeContent,
		LinkURL:       req.LinkURL,
		Content:       req.Content,
		Subtitle:      req.Subtitle,
		Position:      req.Position,
		SortOrder:     req.SortOrder,
		Status:        req.Status,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Remark:        req.Remark,
	}

	if ad.Position == "" {
		ad.Position = "footer"
	}
	if ad.Status == 0 {
		ad.Status = 1
	}

	err := s.adRepo.Create(ad)
	if err != nil {
		return nil, fmt.Errorf("创建广告失败: %w", err)
	}
	return ad, nil
}

func (s *ReceiptAdService) UpdateReceiptAd(id uint, req *dto.ReceiptAdDTO) (*model.ReceiptAd, error) {
	ad, err := s.adRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("广告不存在: %w", err)
	}

	ad.Title = req.Title
	ad.AdType = req.AdType
	ad.ImageURL = req.ImageURL
	ad.QRCodeContent = req.QRCodeContent
	ad.LinkURL = req.LinkURL
	ad.Content = req.Content
	ad.Subtitle = req.Subtitle
	ad.Position = req.Position
	ad.SortOrder = req.SortOrder
	ad.Status = req.Status
	ad.StartDate = req.StartDate
	ad.EndDate = req.EndDate
	ad.StartTime = req.StartTime
	ad.EndTime = req.EndTime
	ad.Remark = req.Remark

	err = s.adRepo.Update(ad)
	if err != nil {
		return nil, fmt.Errorf("更新广告失败: %w", err)
	}
	return ad, nil
}

func (s *ReceiptAdService) DeleteReceiptAd(id uint) error {
	_, err := s.adRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("广告不存在: %w", err)
	}
	return s.adRepo.Delete(id)
}

func (s *ReceiptAdService) GetReceiptAd(id uint) (*dto.ReceiptAdResponse, error) {
	ad, err := s.adRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("广告不存在: %w", err)
	}
	return s.convertToResponse(ad), nil
}

func (s *ReceiptAdService) ListReceiptAds(storeID uint, req *dto.ReceiptAdListRequest) ([]dto.ReceiptAdResponse, int64, error) {
	list, total, err := s.adRepo.List(storeID, req.Page, req.PageSize, req.Status, req.Position, req.AdType, req.Keyword)
	if err != nil {
		return nil, 0, fmt.Errorf("获取广告列表失败: %w", err)
	}

	var result []dto.ReceiptAdResponse
	for _, ad := range list {
		result = append(result, *s.convertToResponse(&ad))
	}
	return result, total, nil
}

func (s *ReceiptAdService) GetActiveAds(storeID uint, position string) ([]model.ReceiptAd, error) {
	return s.adRepo.GetActiveAds(storeID, position)
}

func (s *ReceiptAdService) GetAdStoreID(adID uint) (uint, *model.ReceiptAd, error) {
	ad, err := s.adRepo.GetByID(adID)
	if err != nil {
		return 0, nil, err
	}
	return ad.StoreID, ad, nil
}

func (s *ReceiptAdService) UpdateStatus(id uint, status int) error {
	_, err := s.adRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("广告不存在: %w", err)
	}
	return s.adRepo.UpdateStatus(id, status)
}

func (s *ReceiptAdService) RecordClick(storeID uint, req *dto.ReceiptAdClickRequest) error {
	ad, err := s.adRepo.GetByID(req.AdID)
	if err != nil {
		return fmt.Errorf("广告不存在: %w", err)
	}

	click := &model.ReceiptAdClick{
		StoreID:   storeID,
		AdID:      req.AdID,
		OrderID:   req.OrderID,
		OrderNo:   req.OrderNo,
		ClickType: req.ClickType,
		IP:        req.IP,
		UserAgent: req.UserAgent,
	}
	if click.ClickType == "" {
		click.ClickType = "scan"
	}

	err = s.clickRepo.Create(click)
	if err != nil {
		return err
	}

	return s.adRepo.IncrementClickCount(ad.ID)
}

func (s *ReceiptAdService) IncrementViewCount(adID uint) {
	_ = s.adRepo.IncrementViewCount(adID)
}

func (s *ReceiptAdService) GetStats(storeID uint, req *dto.ReceiptAdStatsRequest) ([]model.ReceiptAdStats, error) {
	return s.clickRepo.GetStats(storeID, req.AdID, req.StartDate, req.EndDate)
}

func (s *ReceiptAdService) convertToResponse(ad *model.ReceiptAd) *dto.ReceiptAdResponse {
	adTypeText := adTypeMap[ad.AdType]
	if adTypeText == "" {
		adTypeText = ad.AdType
	}
	positionText := adPositionMap[ad.Position]
	if positionText == "" {
		positionText = ad.Position
	}
	statusText := adStatusMap[ad.Status]
	if statusText == "" {
		statusText = fmt.Sprintf("%d", ad.Status)
	}

	return &dto.ReceiptAdResponse{
		ID:            ad.ID,
		StoreID:       ad.StoreID,
		Title:         ad.Title,
		AdType:        ad.AdType,
		AdTypeText:    adTypeText,
		ImageURL:      ad.ImageURL,
		QRCodeContent: ad.QRCodeContent,
		LinkURL:       ad.LinkURL,
		Content:       ad.Content,
		Subtitle:      ad.Subtitle,
		Position:      ad.Position,
		PositionText:  positionText,
		SortOrder:     ad.SortOrder,
		Status:        ad.Status,
		StatusText:    statusText,
		ViewCount:     ad.ViewCount,
		ClickCount:    ad.ClickCount,
		StartDate:     ad.StartDate,
		EndDate:       ad.EndDate,
		StartTime:     ad.StartTime,
		EndTime:       ad.EndTime,
		Remark:        ad.Remark,
		CreatedAt:     ad.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *ReceiptAdService) GetAdContentForPrint(storeID uint) []string {
	ads, err := s.adRepo.GetActiveAds(storeID, "footer")
	if err != nil || len(ads) == 0 {
		return nil
	}

	var lines []string
	for _, ad := range ads {
		if ad.AdType == "text" {
			if ad.Title != "" {
				lines = append(lines, ad.Title)
			}
			if ad.Content != "" {
				lines = append(lines, ad.Content)
			}
			if ad.Subtitle != "" {
				lines = append(lines, ad.Subtitle)
			}
		} else if ad.AdType == "qrcode" {
			if ad.Title != "" {
				lines = append(lines, ad.Title)
			}
			if ad.Content != "" {
				lines = append(lines, ad.Content)
			}
			if ad.QRCodeContent != "" {
				lines = append(lines, "扫码查看: "+ad.QRCodeContent)
			}
		} else if ad.AdType == "image" {
			if ad.Title != "" {
				lines = append(lines, ad.Title)
			}
			if ad.Content != "" {
				lines = append(lines, ad.Content)
			}
		}

		go s.adRepo.IncrementViewCount(ad.ID)
	}

	return lines
}
