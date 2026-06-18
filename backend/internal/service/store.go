package service

import (
	"errors"
	"fmt"
	"net"
	"time"

	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
)

type StoreService struct {
	storeRepo   *repository.StoreRepository
	printerRepo *repository.PrinterRepository
}

func NewStoreService() *StoreService {
	return &StoreService{
		storeRepo:   repository.NewStoreRepository(nil),
		printerRepo: repository.NewPrinterRepository(nil),
	}
}

func (s *StoreService) CreateStore(dto *dto.StoreCreateDTO) (*model.Store, error) {
	store := &model.Store{
		Name:          dto.Name,
		Address:       dto.Address,
		Phone:         dto.Phone,
		BusinessHours: dto.BusinessHours,
		Description:   dto.Description,
		Logo:          dto.Logo,
		Status:        dto.Status,
	}
	if store.Status == 0 {
		store.Status = 1
	}
	err := s.storeRepo.Create(store)
	if err != nil {
		return nil, err
	}
	return s.storeRepo.GetByID(store.ID)
}

func (s *StoreService) GetStore(id uint) (*model.Store, error) {
	return s.storeRepo.GetByID(id)
}

func (s *StoreService) UpdateStore(id uint, dto *dto.StoreUpdateDTO) (*model.Store, error) {
	store, err := s.storeRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("store not found")
	}
	if dto.Name != "" {
		store.Name = dto.Name
	}
	if dto.Address != "" {
		store.Address = dto.Address
	}
	if dto.Phone != "" {
		store.Phone = dto.Phone
	}
	store.BusinessHours = dto.BusinessHours
	store.Description = dto.Description
	store.Logo = dto.Logo
	if dto.Status != 0 {
		store.Status = dto.Status
	}
	err = s.storeRepo.Update(store)
	if err != nil {
		return nil, err
	}
	return s.storeRepo.GetByID(id)
}

func (s *StoreService) DeleteStore(id uint) error {
	_, err := s.storeRepo.GetByID(id)
	if err != nil {
		return errors.New("store not found")
	}
	return s.storeRepo.Delete(id)
}

func (s *StoreService) ListStores(query *dto.StoreQueryDTO) ([]model.Store, int64, error) {
	return s.storeRepo.List(query.Name, query.Status, query.Page, query.PageSize)
}

func (s *StoreService) GetAllStores() ([]model.Store, error) {
	return s.storeRepo.GetAll()
}

func (s *StoreService) CreatePrinter(dto *dto.PrinterCreateDTO) (*model.Printer, error) {
	printer := &model.Printer{
		StoreID:     dto.StoreID,
		Name:        dto.Name,
		Type:        dto.Type,
		IPAddress:   dto.IPAddress,
		PrintType:   dto.Type,
		Status:      dto.Status,
	}
	if printer.Status == 0 {
		printer.Status = 1
	}
	err := s.printerRepo.Create(printer)
	if err != nil {
		return nil, err
	}
	return s.printerRepo.GetByID(printer.ID)
}

func (s *StoreService) GetPrinter(id uint) (*model.Printer, error) {
	return s.printerRepo.GetByID(id)
}

func (s *StoreService) UpdatePrinter(id uint, dto *dto.PrinterUpdateDTO) (*model.Printer, error) {
	printer, err := s.printerRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("printer not found")
	}
	if dto.Name != "" {
		printer.Name = dto.Name
	}
	if dto.Type != "" {
		printer.Type = dto.Type
		printer.PrintType = dto.Type
	}
	if dto.IPAddress != "" {
		printer.IPAddress = dto.IPAddress
	}
	if dto.Status != 0 {
		printer.Status = dto.Status
	}
	err = s.printerRepo.Update(printer)
	if err != nil {
		return nil, err
	}
	return s.printerRepo.GetByID(id)
}

func (s *StoreService) DeletePrinter(id uint) error {
	_, err := s.printerRepo.GetByID(id)
	if err != nil {
		return errors.New("printer not found")
	}
	return s.printerRepo.Delete(id)
}

func (s *StoreService) ListPrinters(query *dto.PrinterQueryDTO) ([]model.Printer, int64, error) {
	return s.printerRepo.List(query.StoreID, query.Type, query.Status, query.Page, query.PageSize)
}

func (s *StoreService) GetPrintersByStore(storeID uint) ([]model.Printer, error) {
	return s.printerRepo.GetByStore(storeID)
}

func (s *StoreService) GetPrintersByStoreAndType(storeID uint, printerType string) ([]model.Printer, error) {
	return s.printerRepo.GetByStoreAndType(storeID, printerType)
}

func (s *StoreService) TestPrinter(printerID uint) error {
	printer, err := s.printerRepo.GetByID(printerID)
	if err != nil {
		return fmt.Errorf("printer not found: %w", err)
	}
	if printer.Status != 1 {
		return errors.New("printer is offline")
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", printer.IPAddress, 9100), 3*time.Second)
	if err != nil {
		return fmt.Errorf("cannot connect to printer: %w", err)
	}
	defer conn.Close()

	testContent := s.generateTestPrintContent(printer)
	_, err = conn.Write(testContent)
	if err != nil {
		return fmt.Errorf("print failed: %w", err)
	}

	return nil
}

func (s *StoreService) generateTestPrintContent(printer *model.Printer) []byte {
	var content []byte

	content = append(content, 27, 64)
	content = append(content, 27, 97, 1)
	content = append(content, 27, 33, 16)
	content = append(content, []byte("测试打印\n")...)
	content = append(content, 27, 33, 0)
	content = append(content, []byte("================\n")...)
	content = append(content, 27, 97, 0)
	content = append(content, []byte(fmt.Sprintf("打印机ID: %d\n", printer.ID))...)
	content = append(content, []byte(fmt.Sprintf("打印机名称: %s\n", printer.Name))...)
	content = append(content, []byte(fmt.Sprintf("打印机类型: %s\n", printer.Type))...)
	content = append(content, []byte(fmt.Sprintf("门店: %s\n", printer.Store.Name))...)
	content = append(content, []byte(fmt.Sprintf("打印时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))...)
	content = append(content, []byte("================\n")...)
	content = append(content, []byte("打印成功！\n")...)
	content = append(content, 10, 10, 10)
	content = append(content, 29, 86, 48, 0)

	return content
}
