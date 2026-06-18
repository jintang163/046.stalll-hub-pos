package handler

import (
	"net/http"
	"strconv"
	"stalll-hub-pos/backend/internal/dto"
	"stalll-hub-pos/backend/internal/middleware"
	"stalll-hub-pos/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberService *service.MemberService
}

func NewMemberHandler() *MemberHandler {
	return &MemberHandler{
		memberService: service.NewMemberService(),
	}
}

func (h *MemberHandler) CreateMember(c *gin.Context) {
	var req dto.MemberCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if req.StoreID == 0 {
		req.StoreID = middleware.GetStoreID(c)
	}

	member, err := h.memberService.CreateMember(&req)
	if err != nil {
		middleware.Error(c, "创建会员失败: "+err.Error())
		return
	}

	middleware.Success(c, member)
}

func (h *MemberHandler) UpdateMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员ID")
		return
	}

	var req dto.MemberUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	member, err := h.memberService.UpdateMember(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新会员失败: "+err.Error())
		return
	}

	middleware.Success(c, member)
}

func (h *MemberHandler) DeleteMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员ID")
		return
	}

	err = h.memberService.DeleteMember(uint(id))
	if err != nil {
		middleware.Error(c, "删除会员失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *MemberHandler) GetMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员ID")
		return
	}

	member, err := h.memberService.GetMember(uint(id))
	if err != nil {
		middleware.Error(c, "获取会员失败: "+err.Error())
		return
	}

	middleware.Success(c, member)
}

func (h *MemberHandler) ListMembers(c *gin.Context) {
	var query dto.MemberQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	members, total, err := h.memberService.ListMembers(&query)
	if err != nil {
		middleware.Error(c, "获取会员列表失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  members,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *MemberHandler) CreateMemberLevel(c *gin.Context) {
	var req dto.MemberLevelCreateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	level, err := h.memberService.CreateMemberLevel(&req)
	if err != nil {
		middleware.Error(c, "创建会员等级失败: "+err.Error())
		return
	}

	middleware.Success(c, level)
}

func (h *MemberHandler) UpdateMemberLevel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员等级ID")
		return
	}

	var req dto.MemberLevelUpdateDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	level, err := h.memberService.UpdateMemberLevel(uint(id), &req)
	if err != nil {
		middleware.Error(c, "更新会员等级失败: "+err.Error())
		return
	}

	middleware.Success(c, level)
}

func (h *MemberHandler) DeleteMemberLevel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员等级ID")
		return
	}

	err = h.memberService.DeleteMemberLevel(uint(id))
	if err != nil {
		middleware.Error(c, "删除会员等级失败: "+err.Error())
		return
	}

	middleware.Success(c, nil)
}

func (h *MemberHandler) GetMemberLevel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		middleware.Error(c, "无效的会员等级ID")
		return
	}

	level, err := h.memberService.GetMemberLevel(uint(id))
	if err != nil {
		middleware.Error(c, "获取会员等级失败: "+err.Error())
		return
	}

	middleware.Success(c, level)
}

func (h *MemberHandler) ListMemberLevels(c *gin.Context) {
	levels, err := h.memberService.ListMemberLevels()
	if err != nil {
		middleware.Error(c, "获取会员等级列表失败: "+err.Error())
		return
	}

	middleware.Success(c, levels)
}

func (h *MemberHandler) ListPointsRecords(c *gin.Context) {
	var query dto.PointsRecordQueryDTO
	if err := c.ShouldBindQuery(&query); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	if query.StoreID == 0 {
		query.StoreID = middleware.GetStoreID(c)
	}

	records, total, err := h.memberService.ListPointsRecords(&query)
	if err != nil {
		middleware.Error(c, "获取积分记录失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"list":  records,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

func (h *MemberHandler) AdjustPoints(c *gin.Context) {
	var req dto.AdjustPointsDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	err := h.memberService.AdjustPoints(&req)
	if err != nil {
		middleware.Error(c, "调整积分失败: "+err.Error())
		return
	}

	middleware.Success(c, gin.H{"message": "积分调整成功"})
}

func (h *MemberHandler) MemberLogin(c *gin.Context) {
	var req dto.MemberLoginDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Error(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.memberService.MemberLogin(&req)
	if err != nil {
		middleware.Error(c, "会员登录失败: "+err.Error())
		return
	}

	middleware.Success(c, resp)
}
