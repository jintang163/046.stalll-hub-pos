package service

import (
	"errors"
	"fmt"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/nsq"

	"github.com/shopspring/decimal"
)

type ProductService struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
}

func NewProductService() *ProductService {
	return &ProductService{
		productRepo:  repository.NewProductRepository(),
		categoryRepo: repository.NewCategoryRepository(),
	}
}

func (s *ProductService) CreateProduct(dto *dto.ProductCreateDTO) (*model.Product, error) {
	product := &model.Product{
		StoreID:               dto.StoreID,
		CategoryID:            dto.CategoryID,
		Name:                  dto.Name,
		Description:           dto.Description,
		MainImage:             dto.MainImage,
		Images:                dto.Images,
		SortOrder:             dto.SortOrder,
		Status:                dto.Status,
		IsHot:                 dto.IsHot,
		IsRecommend:           dto.IsRecommend,
		StockWarningThreshold: dto.StockWarningThreshold,
	}

	if product.StockWarningThreshold == 0 {
		product.StockWarningThreshold = 10
	}

	err := s.productRepo.Create(product)
	if err != nil {
		return nil, err
	}

	for _, skuDTO := range dto.SKUs {
		sku := &model.ProductSKU{
			ProductID:     product.ID,
			StoreID:       dto.StoreID,
			SKUCode:       skuDTO.SKUCode,
			SpecName:      skuDTO.SpecName,
			Price:         skuDTO.Price,
			OriginalPrice: skuDTO.OriginalPrice,
			Stock:         skuDTO.Stock,
			Image:         skuDTO.Image,
			Status:        skuDTO.Status,
		}
		if sku.Status == 0 {
			sku.Status = 1
		}

		err = s.productRepo.CreateSKU(sku)
		if err != nil {
			return nil, err
		}

		for _, attrVal := range skuDTO.AttributeValues {
			sav := &model.SKUAttributeValue{
				SKUID:       sku.ID,
				AttributeID: attrVal.AttributeID,
				ValueID:     attrVal.ValueID,
			}
			s.productRepo.CreateSKUAttributeValue(sav)
		}

		s.productRepo.CheckStockWarning(dto.StoreID, sku.ID, product.ID, sku.Stock, product.StockWarningThreshold)
	}

	for _, attrDTO := range dto.Attributes {
		attr := &model.ProductAttribute{
			ProductID: product.ID,
			Name:      attrDTO.Name,
			SortOrder: attrDTO.SortOrder,
			Status:    attrDTO.Status,
		}
		if attr.Status == 0 {
			attr.Status = 1
		}

		err = s.productRepo.CreateAttribute(attr)
		if err != nil {
			return nil, err
		}

		for _, valDTO := range attrDTO.Values {
			val := &model.AttributeValue{
				AttributeID: attr.ID,
				Value:       valDTO.Value,
				SortOrder:   valDTO.SortOrder,
				Status:      valDTO.Status,
				ExtraPrice:  valDTO.ExtraPrice,
				Stock:       valDTO.Stock,
			}
			if val.Status == 0 {
				val.Status = 1
			}
			s.productRepo.CreateAttributeValue(val)
		}
	}

	nsq.PublishProductChange("create", product.StoreID, product.ID, product)

	return s.productRepo.GetByID(product.ID)
}

func (s *ProductService) UpdateProduct(id uint, dto *dto.ProductUpdateDTO) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if dto.CategoryID > 0 {
		product.CategoryID = dto.CategoryID
	}
	if dto.Name != "" {
		product.Name = dto.Name
	}
	product.Description = dto.Description
	product.MainImage = dto.MainImage
	product.Images = dto.Images
	product.SortOrder = dto.SortOrder
	product.Status = dto.Status
	product.IsHot = dto.IsHot
	product.IsRecommend = dto.IsRecommend
	if dto.StockWarningThreshold > 0 {
		product.StockWarningThreshold = dto.StockWarningThreshold
	}

	err = s.productRepo.Update(product)
	if err != nil {
		return nil, err
	}

	for _, skuDTO := range dto.SKUs {
		if skuDTO.ID > 0 {
			sku, err := s.productRepo.GetSKUByID(skuDTO.ID)
			if err == nil {
				if skuDTO.SKUCode != "" {
					sku.SKUCode = skuDTO.SKUCode
				}
				if skuDTO.SpecName != "" {
					sku.SpecName = skuDTO.SpecName
				}
				if !skuDTO.Price.IsZero() {
					sku.Price = skuDTO.Price
				}
				sku.OriginalPrice = skuDTO.OriginalPrice
				sku.Image = skuDTO.Image
				sku.Status = skuDTO.Status

				s.productRepo.UpdateSKU(sku)

				s.productRepo.DeleteSKUAttributeValues(sku.ID)
				for _, attrVal := range skuDTO.AttributeValues {
					sav := &model.SKUAttributeValue{
						SKUID:       sku.ID,
						AttributeID: attrVal.AttributeID,
						ValueID:     attrVal.ValueID,
					}
					s.productRepo.CreateSKUAttributeValue(sav)
				}
			}
		} else {
			sku := &model.ProductSKU{
				ProductID:     product.ID,
				StoreID:       product.StoreID,
				SKUCode:       skuDTO.SKUCode,
				SpecName:      skuDTO.SpecName,
				Price:         skuDTO.Price,
				OriginalPrice: skuDTO.OriginalPrice,
				Stock:         skuDTO.Stock,
				Image:         skuDTO.Image,
				Status:        skuDTO.Status,
			}
			if sku.Status == 0 {
				sku.Status = 1
			}
			s.productRepo.CreateSKU(sku)
		}
	}

	for _, attrDTO := range dto.Attributes {
		if attrDTO.ID > 0 {
			// 更新属性
		} else {
			attr := &model.ProductAttribute{
				ProductID: product.ID,
				Name:      attrDTO.Name,
				SortOrder: attrDTO.SortOrder,
				Status:    attrDTO.Status,
			}
			s.productRepo.CreateAttribute(attr)

			for _, valDTO := range attrDTO.Values {
				val := &model.AttributeValue{
					AttributeID: attr.ID,
					Value:       valDTO.Value,
					SortOrder:   valDTO.SortOrder,
					Status:      valDTO.Status,
					ExtraPrice:  valDTO.ExtraPrice,
					Stock:       valDTO.Stock,
				}
				s.productRepo.CreateAttributeValue(val)
			}
		}
	}

	nsq.PublishProductChange("update", product.StoreID, product.ID, product)

	return s.productRepo.GetByID(product.ID)
}

func (s *ProductService) DeleteProduct(id uint) error {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return errors.New("product not found")
	}

	err = s.productRepo.Delete(id)
	if err != nil {
		return err
	}

	nsq.PublishProductChange("delete", product.StoreID, product.ID, nil)

	return nil
}

func (s *ProductService) GetProduct(id uint) (*dto.ProductDetailResponse, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(product), nil
}

func (s *ProductService) ListProducts(query *dto.ProductQueryDTO) ([]dto.ProductListResponse, int64, error) {
	products, total, err := s.productRepo.List(
		query.StoreID,
		query.CategoryID,
		query.Name,
		query.Status,
		query.IsHot,
		query.IsRecommend,
		query.GetOffset(),
		query.GetLimit(),
	)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.ProductListResponse
	for _, p := range products {
		response = append(response, s.convertToListResponse(p))
	}

	return response, total, nil
}

func (s *ProductService) UpdateSKUStock(storeID uint, items []dto.SKUStockItem) error {
	for _, item := range items {
		sku, err := s.productRepo.GetSKUByID(item.SKUID)
		if err != nil {
			continue
		}
		oldStock, err := s.productRepo.UpdateSKUStock(item.SKUID, item.Stock)
		if err != nil {
			continue
		}

		product, _ := s.productRepo.GetByID(sku.ProductID)
		threshold := 10
		if product != nil {
			threshold = product.StockWarningThreshold
		}

		s.productRepo.CheckStockWarning(storeID, sku.ID, sku.ProductID, item.Stock, threshold)
		nsq.PublishStockChange(storeID, sku.ID, sku.ProductID, oldStock, item.Stock, "manual")
	}
	return nil
}

func (s *ProductService) BatchUpdatePrice(dto *dto.BatchPriceUpdateDTO) error {
	price := dto.Price.InexactFloat64()
	err := s.productRepo.BatchUpdatePrice(dto.ProductIDs, dto.PriceType, price)
	if err != nil {
		return err
	}

	products, _ := s.productRepo.GetByIDs(dto.ProductIDs)
	for _, p := range products {
		nsq.PublishProductChange("update", p.StoreID, p.ID, p)
	}

	return nil
}

func (s *ProductService) CopyProduct(dto *dto.ProductCopyDTO) (*model.Product, error) {
	sourceProduct, err := s.productRepo.GetByID(dto.ProductID)
	if err != nil {
		return nil, errors.New("source product not found")
	}

	newName := dto.NewName
	if newName == "" {
		newName = sourceProduct.Name + " (副本)"
	}

	categoryID := dto.CategoryID
	if categoryID == 0 {
		categoryID = sourceProduct.CategoryID
	}

	newProduct := &model.Product{
		StoreID:               dto.StoreID,
		CategoryID:            categoryID,
		Name:                  newName,
		Description:           sourceProduct.Description,
		MainImage:             sourceProduct.MainImage,
		Images:                sourceProduct.Images,
		SortOrder:             sourceProduct.SortOrder,
		Status:                0,
		IsHot:                 sourceProduct.IsHot,
		IsRecommend:           sourceProduct.IsRecommend,
		StockWarningThreshold: sourceProduct.StockWarningThreshold,
	}

	err = s.productRepo.Create(newProduct)
	if err != nil {
		return nil, err
	}

	for _, sku := range sourceProduct.SKUs {
		newSKU := &model.ProductSKU{
			ProductID:     newProduct.ID,
			StoreID:       dto.StoreID,
			SKUCode:       fmt.Sprintf("%s_copy_%d", sku.SKUCode, newProduct.ID),
			SpecName:      sku.SpecName,
			Price:         sku.Price,
			OriginalPrice: sku.OriginalPrice,
			Stock:         0,
			Image:         sku.Image,
			Status:        sku.Status,
		}
		s.productRepo.CreateSKU(newSKU)

		for _, av := range sku.AttributeValues {
			newAV := &model.SKUAttributeValue{
				SKUID:       newSKU.ID,
				AttributeID: av.AttributeID,
				ValueID:     av.ValueID,
			}
			s.productRepo.CreateSKUAttributeValue(newAV)
		}
	}

	for _, attr := range sourceProduct.Attributes {
		newAttr := &model.ProductAttribute{
			ProductID: newProduct.ID,
			Name:      attr.Name,
			SortOrder: attr.SortOrder,
			Status:    attr.Status,
		}
		s.productRepo.CreateAttribute(newAttr)

		for _, val := range attr.Values {
			newVal := &model.AttributeValue{
				AttributeID: newAttr.ID,
				Value:       val.Value,
				SortOrder:   val.SortOrder,
				Status:      val.Status,
				ExtraPrice:  val.ExtraPrice,
				Stock:       val.Stock,
			}
			s.productRepo.CreateAttributeValue(newVal)
		}
	}

	return s.productRepo.GetByID(newProduct.ID)
}

func (s *ProductService) SyncProducts(storeID uint, lastSyncID uint, limit int) (*dto.SyncProductResponse, error) {
	products, lastID, total, err := s.productRepo.SyncProducts(storeID, lastSyncID, limit)
	if err != nil {
		return nil, err
	}

	var productResponses []dto.ProductDetailResponse
	for _, p := range products {
		productResponses = append(productResponses, *s.convertToDetailResponse(&p))
	}

	return &dto.SyncProductResponse{
		LastSyncID: lastID,
		Total:      total,
		Products:   productResponses,
	}, nil
}

func (s *ProductService) GetStockWarnings(storeID uint, page, pageSize int) ([]model.StockWarning, int64, error) {
	return s.productRepo.GetStockWarnings(storeID)
}

func (s *ProductService) convertToListResponse(p model.Product) dto.ProductListResponse {
	minPrice := decimal.NewFromInt(0)
	maxPrice := decimal.NewFromInt(0)
	totalStock := 0
	skuCount := len(p.SKUs)

	for i, sku := range p.SKUs {
		if i == 0 {
			minPrice = sku.Price
			maxPrice = sku.Price
		} else {
			if sku.Price.LessThan(minPrice) {
				minPrice = sku.Price
			}
			if sku.Price.GreaterThan(maxPrice) {
				maxPrice = sku.Price
			}
		}
		totalStock += sku.Stock
	}

	return dto.ProductListResponse{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		MainImage:   p.MainImage,
		Status:      p.Status,
		IsHot:       p.IsHot,
		IsRecommend: p.IsRecommend,
		SortOrder:   p.SortOrder,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		TotalStock:  totalStock,
		SKUCount:    skuCount,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *ProductService) convertToDetailResponse(p *model.Product) *dto.ProductDetailResponse {
	if p == nil {
		return nil
	}

	var category *dto.CategorySimpleDTO
	if p.Category.ID > 0 {
		category = &dto.CategorySimpleDTO{
			ID:   p.Category.ID,
			Name: p.Category.Name,
		}
	}

	var skus []dto.ProductSKUResponse
	for _, sku := range p.SKUs {
		var attrValues []dto.SKUAttributeResponse
		for _, av := range sku.AttributeValues {
			attrValues = append(attrValues, dto.SKUAttributeResponse{
				AttributeID:   av.AttributeID,
				AttributeName: av.Attribute.Name,
				ValueID:       av.ValueID,
				ValueName:     av.Value.Value,
			})
		}

		skus = append(skus, dto.ProductSKUResponse{
			ID:              sku.ID,
			ProductID:       sku.ProductID,
			SKUCode:         sku.SKUCode,
			SpecName:        sku.SpecName,
			Price:           sku.Price,
			OriginalPrice:   sku.OriginalPrice,
			Stock:           sku.Stock,
			SoldCount:       sku.SoldCount,
			Image:           sku.Image,
			Status:          sku.Status,
			IsSoldOut:       sku.IsSoldOut,
			AttributeValues: attrValues,
		})
	}

	var attributes []dto.AttributeResponse
	for _, attr := range p.Attributes {
		var values []dto.AttributeValueResponse
		for _, val := range attr.Values {
			values = append(values, dto.AttributeValueResponse{
				ID:         val.ID,
				Value:      val.Value,
				SortOrder:  val.SortOrder,
				Status:     val.Status,
				ExtraPrice: val.ExtraPrice,
				Stock:      val.Stock,
			})
		}

		attributes = append(attributes, dto.AttributeResponse{
			ID:        attr.ID,
			ProductID: attr.ProductID,
			Name:      attr.Name,
			SortOrder: attr.SortOrder,
			Status:    attr.Status,
			Values:    values,
		})
	}

	return &dto.ProductDetailResponse{
		ID:                    p.ID,
		StoreID:               p.StoreID,
		CategoryID:            p.CategoryID,
		Name:                  p.Name,
		Description:           p.Description,
		MainImage:             p.MainImage,
		Images:                p.Images,
		SortOrder:             p.SortOrder,
		Status:                p.Status,
		IsHot:                 p.IsHot,
		IsRecommend:           p.IsRecommend,
		StockWarningThreshold: p.StockWarningThreshold,
		Category:              category,
		SKUs:                  skus,
		Attributes:            attributes,
		CreatedAt:             p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:             p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *ProductService) BatchSoldOut(dtoReq *dto.SoldOutBatchDTO) error {
	skus, err := s.productRepo.GetSKUsByIDs(dtoReq.SKUIds)
	if err != nil {
		return err
	}
	if len(skus) == 0 {
		return nil
	}

	err = s.productRepo.BatchUpdateSoldOut(dtoReq.SKUIds, true)
	if err != nil {
		return err
	}

	storeID := dtoReq.StoreID
	if storeID == 0 && len(skus) > 0 {
		storeID = skus[0].StoreID
	}

	for _, sku := range skus {
		if sku.IsSoldOut {
			continue
		}
		record := &model.SoldOutRecord{
			StoreID:       storeID,
			ProductID:     sku.ProductID,
			SKUID:         sku.ID,
			SKUCode:       sku.SKUCode,
			ProductName:   sku.Product.Name,
			SpecName:      sku.SpecName,
			CategoryID:    sku.Product.CategoryID,
			ActionType:    "sold_out",
			OperatorID:    dtoReq.OperatorID,
			OperatorName:  dtoReq.OperatorName,
			Source:        dtoReq.Source,
			Remark:        dtoReq.Remark,
			StockAtAction: sku.Stock,
		}
		s.productRepo.CreateSoldOutRecord(record)

		product, _ := s.productRepo.GetByID(sku.ProductID)
		if product != nil {
			nsq.PublishProductChange("update", storeID, sku.ProductID, product)
		}
	}

	return nil
}

func (s *ProductService) BatchRestoreSoldOut(dtoReq *dto.SoldOutBatchDTO) error {
	skus, err := s.productRepo.GetSKUsByIDs(dtoReq.SKUIds)
	if err != nil {
		return err
	}
	if len(skus) == 0 {
		return nil
	}

	err = s.productRepo.BatchUpdateSoldOut(dtoReq.SKUIds, false)
	if err != nil {
		return err
	}

	storeID := dtoReq.StoreID
	if storeID == 0 && len(skus) > 0 {
		storeID = skus[0].StoreID
	}

	for _, sku := range skus {
		if !sku.IsSoldOut {
			continue
		}
		record := &model.SoldOutRecord{
			StoreID:       storeID,
			ProductID:     sku.ProductID,
			SKUID:         sku.ID,
			SKUCode:       sku.SKUCode,
			ProductName:   sku.Product.Name,
			SpecName:      sku.SpecName,
			CategoryID:    sku.Product.CategoryID,
			ActionType:    "restore",
			OperatorID:    dtoReq.OperatorID,
			OperatorName:  dtoReq.OperatorName,
			Source:        dtoReq.Source,
			Remark:        dtoReq.Remark,
			StockAtAction: sku.Stock,
		}
		s.productRepo.CreateSoldOutRecord(record)

		product, _ := s.productRepo.GetByID(sku.ProductID)
		if product != nil {
			nsq.PublishProductChange("update", storeID, sku.ProductID, product)
		}
	}

	return nil
}

func (s *ProductService) ListSoldOutRecords(query *dto.SoldOutRecordQueryDTO) ([]dto.SoldOutRecordResponse, int64, error) {
	records, total, err := s.productRepo.ListSoldOutRecords(
		query.StoreID,
		query.ProductID,
		query.SKUID,
		query.ActionType,
		query.StartDate,
		query.EndDate,
		query.GetOffset(),
		query.GetLimit(),
	)
	if err != nil {
		return nil, 0, err
	}

	var response []dto.SoldOutRecordResponse
	for _, r := range records {
		response = append(response, dto.SoldOutRecordResponse{
			ID:            r.ID,
			StoreID:       r.StoreID,
			ProductID:     r.ProductID,
			SKUID:         r.SKUID,
			SKUCode:       r.SKUCode,
			ProductName:   r.ProductName,
			SpecName:      r.SpecName,
			CategoryID:    r.CategoryID,
			CategoryName:  r.CategoryName,
			ActionType:    r.ActionType,
			OperatorID:    r.OperatorID,
			OperatorName:  r.OperatorName,
			Source:        r.Source,
			Remark:        r.Remark,
			StockAtAction: r.StockAtAction,
			CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return response, total, nil
}
