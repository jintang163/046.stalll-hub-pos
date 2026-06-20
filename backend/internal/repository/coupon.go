package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"stalll-hub-pos/backend/internal/model"
	"stalll-hub-pos/backend/pkg/database"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

func (r *CouponRepository) Create(coupon *model.Coupon) error {
	return database.DB.Create(coupon).Error
}

func (r *CouponRepository) GetByID(id uint) (*model.Coupon, error) {
	var coupon model.Coupon
	err := database.DB.First(&coupon, id).Error
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *CouponRepository) List(name, couponType string, status int, page, pageSize int) ([]model.Coupon, int64, error) {
	var coupons []model.Coupon
	var total int64

	db := database.DB.Model(&model.Coupon{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if couponType != "" {
		db = db.Where("type = ?", couponType)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&coupons).Error
	return coupons, total, err
}

func (r *CouponRepository) Update(coupon *model.Coupon) error {
	return database.DB.Save(coupon).Error
}

func (r *CouponRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Coupon{}, id).Error
}

func (r *CouponRepository) IncrementUsedCount(id uint) error {
	return database.DB.Model(&model.Coupon{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}

func (r *CouponRepository) IssueToMembers(couponID uint, memberIDs []uint) ([]model.MemberCoupon, error) {
	coupon, err := r.GetByID(couponID)
	if err != nil {
		return nil, err
	}

	var memberCoupons []model.MemberCoupon
	now := time.Now()

	for _, memberID := range memberIDs {
		var existingCount int64
		database.DB.Model(&model.MemberCoupon{}).
			Where("coupon_id = ? AND member_id = ?", couponID, memberID).
			Count(&existingCount)
		if int(existingCount) >= coupon.PerUserLimit {
			continue
		}

		var expireAt *time.Time
		if coupon.ValidityType == "fixed" {
			expireAt = coupon.EndTime
		} else {
			t := now.AddDate(0, 0, coupon.ValidityDays)
			expireAt = &t
		}

		mc := model.MemberCoupon{
			MemberID: memberID,
			CouponID: couponID,
			Code:     r.generateCouponCode(),
			Status:   1,
			ExpireAt: expireAt,
		}
		if err := database.DB.Create(&mc).Error; err != nil {
			continue
		}
		mc.Coupon = *coupon
		memberCoupons = append(memberCoupons, mc)
	}

	return memberCoupons, nil
}

func (r *CouponRepository) ClaimCoupon(memberID, couponID uint) (*model.MemberCoupon, error) {
	coupon, err := r.GetByID(couponID)
	if err != nil {
		return nil, err
	}

	if coupon.Status != 1 {
		return nil, errors.New("coupon is not active")
	}

	now := time.Now()
	if coupon.StartTime != nil && coupon.StartTime.After(now) {
		return nil, errors.New("coupon is not available yet")
	}
	if coupon.EndTime != nil && coupon.EndTime.Before(now) {
		return nil, errors.New("coupon has expired")
	}

	if coupon.TotalCount > 0 && coupon.UsedCount >= coupon.TotalCount {
		return nil, errors.New("coupon is out of stock")
	}

	var existingCount int64
	database.DB.Model(&model.MemberCoupon{}).
		Where("coupon_id = ? AND member_id = ?", couponID, memberID).
		Count(&existingCount)
	if int(existingCount) >= coupon.PerUserLimit {
		return nil, errors.New("coupon limit per user reached")
	}

	var expireAt *time.Time
	if coupon.ValidityType == "fixed" {
		expireAt = coupon.EndTime
	} else {
		t := now.AddDate(0, 0, coupon.ValidityDays)
		expireAt = &t
	}

	mc := &model.MemberCoupon{
		StoreID:  coupon.StoreID,
		MemberID: memberID,
		CouponID: couponID,
		Code:     r.generateCouponCode(),
		Status:   1,
		ExpireAt: expireAt,
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(mc).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Coupon{}).Where("id = ?", couponID).
			UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	mc.Coupon = *coupon
	return mc, nil
}

func (r *CouponRepository) GetMemberCoupons(memberID uint, status int, page, pageSize int) ([]model.MemberCoupon, int64, error) {
	var memberCoupons []model.MemberCoupon
	var total int64

	db := database.DB.Model(&model.MemberCoupon{}).Where("member_id = ?", memberID)
	if status > 0 {
		if status == 1 {
			db = db.Where("status = 1 AND (expire_at IS NULL OR expire_at > ?)", time.Now())
		} else {
			db = db.Where("status = ? OR (status = 1 AND expire_at <= ?)", status, time.Now())
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Coupon").
		Order("id DESC").Offset(offset).Limit(pageSize).Find(&memberCoupons).Error
	return memberCoupons, total, err
}

func (r *CouponRepository) GetMemberCouponByID(id uint) (*model.MemberCoupon, error) {
	var mc model.MemberCoupon
	err := database.DB.Where("id = ?", id).
		Preload("Coupon").First(&mc).Error
	if err != nil {
		return nil, err
	}
	return &mc, nil
}

func (r *CouponRepository) GetMemberCoupon(id, memberID uint) (*model.MemberCoupon, error) {
	var mc model.MemberCoupon
	err := database.DB.Where("id = ? AND member_id = ?", id, memberID).
		Preload("Coupon").First(&mc).Error
	if err != nil {
		return nil, err
	}
	return &mc, nil
}

func (r *CouponRepository) UseCoupon(id, orderID uint) error {
	now := time.Now()
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var mc model.MemberCoupon
		if err := tx.First(&mc, id).Error; err != nil {
			return err
		}

		if err := tx.Model(&mc).Updates(map[string]interface{}{
			"status":   2,
			"used_at":  now,
			"order_id": orderID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Coupon{}).Where("id = ?", mc.CouponID).
			UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *CouponRepository) ReturnCoupon(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var mc model.MemberCoupon
		if err := tx.First(&mc, id).Error; err != nil {
			return err
		}

		if err := tx.Model(&mc).Updates(map[string]interface{}{
			"status":   1,
			"used_at":  nil,
			"order_id": nil,
		}).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Coupon{}).Where("id = ?", mc.CouponID).
			UpdateColumn("used_count", gorm.Expr("used_count - 1")).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *CouponRepository) GetAvailableCoupons(memberID, storeID uint, amount float64, productIDs []uint) ([]model.MemberCoupon, error) {
	var memberCoupons []model.MemberCoupon

	query := database.DB.Where("member_id = ? AND status = 1 AND (expire_at IS NULL OR expire_at > ?)",
		memberID, time.Now()).Preload("Coupon")

	if err := query.Find(&memberCoupons).Error; err != nil {
		return nil, err
	}

	var validCoupons []model.MemberCoupon
	for _, mc := range memberCoupons {
		if mc.Coupon.Status != 1 {
			continue
		}
		if mc.Coupon.MinAmount.GreaterThan(decimal.NewFromFloat(amount)) {
			continue
		}

		if mc.Coupon.ApplicableType != "all" && len(productIDs) > 0 && mc.Coupon.ApplicableIDs != "" {
			parts := strings.Split(mc.Coupon.ApplicableIDs, ",")
			found := false
			for _, pid := range productIDs {
				for _, aidStr := range parts {
					var aid uint
					if _, err := fmt.Sscanf(aidStr, "%d", &aid); err == nil && pid == aid {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				continue
			}
		}

		validCoupons = append(validCoupons, mc)
	}

	return validCoupons, nil
}

func (r *CouponRepository) GetClaimableCoupons(storeID, memberID uint) ([]model.Coupon, map[uint]int, error) {
	var coupons []model.Coupon
	now := time.Now()

	db := database.DB.Model(&model.Coupon{}).
		Where("status = 1").
		Where("(start_time IS NULL OR start_time <= ?)", now).
		Where("(end_time IS NULL OR end_time >= ?)", now)
	if storeID > 0 {
		db = db.Where("(store_id = ? OR store_id = 0)", storeID)
	}
	db = db.Order("priority DESC, id DESC")
	if err := db.Find(&coupons).Error; err != nil {
		return nil, nil, err
	}

	claimedCounts := make(map[uint]int)
	if memberID > 0 {
		couponIDs := make([]uint, 0, len(coupons))
		for _, c := range coupons {
			couponIDs = append(couponIDs, c.ID)
		}
		if len(couponIDs) > 0 {
			type CountRow struct {
				CouponID uint
				Cnt      int64
			}
			var rows []CountRow
			database.DB.Model(&model.MemberCoupon{}).
				Select("coupon_id, COUNT(*) as cnt").
				Where("member_id = ? AND coupon_id IN ?", memberID, couponIDs).
				Group("coupon_id").
				Scan(&rows)
			for _, row := range rows {
				claimedCounts[row.CouponID] = int(row.Cnt)
			}
		}
	}

	return coupons, claimedCounts, nil
}

func (r *CouponRepository) generateCouponCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
