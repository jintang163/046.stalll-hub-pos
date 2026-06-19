package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/config"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/internal/repository"
	"stalll-hub-pos/backend/pkg/database"
)

type MemberService struct {
	memberRepo         *repository.MemberRepository
	memberLevelRepo    *repository.MemberLevelRepository
	pointsEngine       *PointsEngineService
	cfg                *config.Config
}

func NewMemberService(cfg *config.Config) *MemberService {
	return &MemberService{
		memberRepo:         repository.NewMemberRepository(nil),
		memberLevelRepo:    repository.NewMemberLevelRepository(nil),
		pointsEngine:       NewPointsEngineService(),
		cfg:                cfg,
	}
}

func (s *MemberService) generateMemberNo() string {
	now := time.Now()
	n, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("M%s%04d", now.Format("20060102150405"), n.Int64())
}

func (s *MemberService) generateToken(member *model.Member) string {
	claims := jwt.MapClaims{
		"member_id": member.ID,
		"store_id":  member.StoreID,
		"phone":     member.Phone,
		"exp":       time.Now().Add(time.Hour * 24 * 30).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(s.cfg.JWT.Secret))
	return tokenStr
}

func (s *MemberService) Register(req *dto.MemberCreateDTO) (*dto.MemberResponse, error) {
	existing, _ := s.memberRepo.GetByPhone(req.StoreID, req.Phone)
	if existing != nil {
		return nil, errors.New("phone number already registered")
	}

	levelID := req.LevelID
	if levelID == 0 {
		levelID = 1
	}

	member := &model.Member{
		StoreID:  req.StoreID,
		MemberNo: s.generateMemberNo(),
		Name:     req.Name,
		Phone:    req.Phone,
		Gender:   req.Gender,
		Avatar:   req.Avatar,
		LevelID:  levelID,
		Points:   0,
		Status:   req.Status,
	}
	if member.Status == 0 {
		member.Status = 1
	}

	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			member.Birthday = &t
		}
	}

	err := s.memberRepo.Create(member)
	if err != nil {
		return nil, fmt.Errorf("register failed: %w", err)
	}

	registerBonus := s.pointsEngine.CalculateRegisterBonus(req.StoreID)
	if registerBonus > 0 {
		member.Points = registerBonus
		member.TotalPoints = registerBonus
		if err := s.memberRepo.Update(member); err == nil {
			record := &model.MemberPointsRecord{
				MemberID: member.ID,
				StoreID:  req.StoreID,
				Type:     "earn",
				Points:   registerBonus,
				Balance:  registerBonus,
				Remark:   "新会员注册赠送积分",
			}
			_ = database.DB.Create(record)
		}
	}

	member, err = s.memberRepo.GetByID(member.ID)
	if err != nil {
		return nil, err
	}

	return s.convertToMemberResponse(member), nil
}

func (s *MemberService) Login(req *dto.MemberLoginDTO) (*dto.MemberLoginResponse, error) {
	member, err := s.memberRepo.GetByPhone(req.StoreID, req.Phone)
	if err != nil {
		return nil, errors.New("invalid phone or password")
	}

	if member.Status != 1 {
		return nil, errors.New("account is disabled")
	}

	if req.Code != "" {
		if !s.verifySMSCode(req.Phone, req.Code) {
			return nil, errors.New("invalid verification code")
		}
	} else if req.Password != "" {
		if !s.verifyPassword(member, req.Password) {
			return nil, errors.New("invalid phone or password")
		}
	}

	token := s.generateToken(member)

	_ = s.memberRepo.UpdateLastActive(member.ID)

	return &dto.MemberLoginResponse{
		Token:  token,
		Member: *s.convertToMemberResponse(member),
	}, nil
}

func (s *MemberService) verifySMSCode(phone, code string) bool {
	return code == "123456"
}

func (s *MemberService) verifyPassword(member *model.Member, password string) bool {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]) == member.Password
}

func (s *MemberService) GetMember(id uint) (*dto.MemberResponse, error) {
	member, err := s.memberRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToMemberResponse(member), nil
}

func (s *MemberService) UpdateMember(id uint, req *dto.MemberUpdateDTO) (*dto.MemberResponse, error) {
	member, err := s.memberRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("member not found")
	}

	if req.Name != "" {
		member.Name = req.Name
	}
	if req.Phone != "" {
		member.Phone = req.Phone
	}
	if req.Gender != 0 {
		member.Gender = req.Gender
	}
	member.Avatar = req.Avatar
	if req.LevelID > 0 {
		member.LevelID = req.LevelID
	}
	if req.Status != 0 {
		member.Status = req.Status
	}
	if req.Birthday != "" {
		if t, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			member.Birthday = &t
		}
	}

	err = s.memberRepo.Update(member)
	if err != nil {
		return nil, err
	}

	member, err = s.memberRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return s.convertToMemberResponse(member), nil
}

func (s *MemberService) DeleteMember(id uint) error {
	_, err := s.memberRepo.GetByID(id)
	if err != nil {
		return errors.New("member not found")
	}
	return s.memberRepo.Update(&model.Member{BaseModel: model.BaseModel{ID: id}, Status: -1})
}

func (s *MemberService) ListMembers(query *dto.MemberQueryDTO) (*dto.PageResponse, error) {
	members, total, err := s.memberRepo.List(
		query.StoreID,
		query.Name,
		query.Phone,
		int(query.LevelID),
		query.Status,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.MemberResponse
	for _, m := range members {
		list = append(list, *s.convertToMemberResponse(&m))
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *MemberService) AdjustPoints(req *dto.AdjustPointsDTO) error {
	member, err := s.memberRepo.GetByID(req.MemberID)
	if err != nil {
		return errors.New("member not found")
	}

	req.Points = int(req.Points)
	if req.Points < 0 && member.Points+req.Points < 0 {
		return errors.New("insufficient points")
	}

	err = s.memberRepo.AdjustPoints(req.MemberID, req.Points, req.Remark, 0)
	if err != nil {
		return err
	}

	member, _ = s.memberRepo.GetByID(req.MemberID)
	s.checkAndUpdateLevel(member)

	return nil
}

func (s *MemberService) checkAndUpdateLevel(member *model.Member) {
	level, err := s.memberLevelRepo.GetByPoints(member.TotalPoints)
	if err != nil {
		return
	}
	if level.ID != member.LevelID {
		member.LevelID = level.ID
		_ = s.memberRepo.Update(member)
	}
}

func (s *MemberService) GetPointsRecords(query *dto.PointsRecordQueryDTO) (*dto.PageResponse, error) {
	records, total, err := s.memberRepo.GetPointsRecords(
		query.MemberID,
		query.StoreID,
		query.Type,
		query.StartDate,
		query.EndDate,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		return nil, err
	}

	var list []dto.PointsRecordResponse
	for _, r := range records {
		memberName := ""
		if r.Member.Name != "" {
			memberName = r.Member.Name
		}
		storeName := ""
		if r.Store.Name != "" {
			storeName = r.Store.Name
		}
		orderNo := ""
		if r.Order != nil {
			orderNo = r.Order.OrderNo
		}

		list = append(list, dto.PointsRecordResponse{
			ID:         r.ID,
			MemberID:   r.MemberID,
			MemberName: memberName,
			StoreID:    r.StoreID,
			StoreName:  storeName,
			Type:       r.Type,
			Points:     r.Points,
			Balance:    r.Balance,
			OrderID:    r.OrderID,
			OrderNo:    orderNo,
			Remark:     r.Remark,
			CreatedAt:  r.CreatedAt,
		})
	}

	return &dto.PageResponse{
		List:  list,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}, nil
}

func (s *MemberService) CreateLevel(req *dto.MemberLevelCreateDTO) (*model.MemberLevel, error) {
	level := &model.MemberLevel{
		Name:         req.Name,
		MinPoints:    req.PointsRequired,
		DiscountRate: req.Discount,
		Description:  req.Description,
		Color:        req.Color,
		Status:       req.Status,
	}
	if level.Status == 0 {
		level.Status = 1
	}

	err := s.memberLevelRepo.Create(level)
	if err != nil {
		return nil, err
	}
	return s.memberLevelRepo.GetByID(level.ID)
}

func (s *MemberService) GetLevel(id uint) (*model.MemberLevel, error) {
	return s.memberLevelRepo.GetByID(id)
}

func (s *MemberService) UpdateLevel(id uint, req *dto.MemberLevelUpdateDTO) (*model.MemberLevel, error) {
	level, err := s.memberLevelRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("level not found")
	}

	if req.Name != "" {
		level.Name = req.Name
	}
	if req.PointsRequired >= 0 {
		level.MinPoints = req.PointsRequired
	}
	if !req.Discount.IsZero() {
		level.DiscountRate = req.Discount
	}
	level.Description = req.Description
	level.Color = req.Color
	if req.Status != 0 {
		level.Status = req.Status
	}

	err = s.memberLevelRepo.Update(level)
	if err != nil {
		return nil, err
	}
	return s.memberLevelRepo.GetByID(id)
}

func (s *MemberService) DeleteLevel(id uint) error {
	_, err := s.memberLevelRepo.GetByID(id)
	if err != nil {
		return errors.New("level not found")
	}
	return s.memberLevelRepo.Delete(id)
}

func (s *MemberService) ListLevels() ([]dto.MemberLevelResponse, error) {
	levels, err := s.memberLevelRepo.List()
	if err != nil {
		return nil, err
	}

	var response []dto.MemberLevelResponse
	for _, level := range levels {
		count, _ := s.memberLevelRepo.CountMembers(level.ID)
		response = append(response, dto.MemberLevelResponse{
			ID:              level.ID,
			Name:            level.Name,
			PointsRequired:  level.MinPoints,
			Discount:        level.DiscountRate,
			Description:     level.Description,
			Color:           level.Color,
			MemberCount:     int(count),
			Status:          level.Status,
			CreatedAt:       level.CreatedAt,
		})
	}

	return response, nil
}

func (s *MemberService) convertToMemberResponse(m *model.Member) *dto.MemberResponse {
	levelName := ""
	levelDiscount := decimal.NewFromInt(100)
	if m.Level.ID > 0 {
		levelName = m.Level.Name
		levelDiscount = m.Level.DiscountRate
	}

	storeName := ""
	if m.Store.Name != "" {
		storeName = m.Store.Name
	}

	birthday := ""
	if m.Birthday != nil {
		birthday = m.Birthday.Format("2006-01-02")
	}

	return &dto.MemberResponse{
		ID:            m.ID,
		StoreID:       m.StoreID,
		StoreName:     storeName,
		Name:          m.Name,
		Phone:         m.Phone,
		Gender:        m.Gender,
		Birthday:      birthday,
		Avatar:        m.Avatar,
		LevelID:       m.LevelID,
		LevelName:     levelName,
		LevelDiscount: levelDiscount,
		Points:        m.Points,
		TotalSpent:    m.TotalConsume,
		TotalOrders:   m.OrderCount,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
	}
}
