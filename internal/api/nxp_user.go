package api

import (
	"github.com/gin-gonic/gin"

	"sop/internal/middleware"
	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// UserApi 用户管理接口处理器
type UserApi struct{}

// GetUserList 分页获取用户列表
func (u UserApi) GetUserList(c *gin.Context) {
	var req UserListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	req.Normalize()

	cond := service.UserQuery{
		Name:     req.Name,
		Tel:      req.Tel,
		Email:    req.Email,
		SFC:      req.SFC,
		FSL:      req.FSL,
		Dept:     req.Dept,
		RoleID:   req.RoleID,
		Position: req.Position,
		Status:   req.Status,
	}

	list, total, err := service.GetUserList(req.Page, req.PageSize, req.Keyword, req.OrderKey, req.Desc, cond)
	if err != nil {
		response.FailError(c, err)
		return
	}

	if list == nil {
		list = []service.UserNxp{}
	}
	response.OKWithPage(c, toNxpUserDTOs(list), total, req.Page, req.PageSize)
}

// UserByIDsRequest 按 ID 列表查询用户请求
type UserByIDsRequest struct {
	IDs []uint `json:"ids"`
}

// GetUsersByIDs 按 ID 列表批量获取用户（用于提问等记录反查提出工程师）
func (u UserApi) GetUsersByIDs(c *gin.Context) {
	var req UserByIDsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	list, err := service.GetUsersByIDs(req.IDs)
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithData(c, toNxpUserDTOs(list))
}

// NxpUserDTO 用户信息响应结构（nxp_user + nxp_user_info 组合）
type NxpUserDTO struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	CreateTime       *string `json:"createTime"`
	RoleID           *int    `json:"roleId"`
	Status           *int8   `json:"status"`
	Tel              string  `json:"tel"`
	ChineseName      string  `json:"chineseName"`
	EnglishName      string  `json:"englishName"`
	Sex              int8    `json:"sex"`
	Email            string  `json:"email"`
	Position         *int    `json:"position"`
	Shift            *int    `json:"shift"`
	SFCNumber        string  `json:"sfcNumber"`
	FSLNumber        string  `json:"fslNumber"`
	DepartmentNumber string  `json:"departmentNumber"`
	PID              string  `json:"pid"`
	Section          string  `json:"section"`
	WBICode          string  `json:"wbiCode"`
}

// toNxpUserDTOs 将 []service.UserNxp 映射为 []NxpUserDTO
func toNxpUserDTOs(users []service.UserNxp) []NxpUserDTO {
	dtos := make([]NxpUserDTO, 0, len(users))
	for _, u := range users {
		if u.User == nil {
			continue
		}
		dto := NxpUserDTO{ID: int(u.User.ID), Name: u.User.Name}
		if !u.User.CreateTime.IsZero() {
			s := u.User.CreateTime.Format("2006-01-02 15:04:05")
			dto.CreateTime = &s
		}
		if u.Info != nil {
			dto.RoleID = u.Info.RoleID
			dto.Status = u.Info.Status
			dto.Tel = u.Info.Tel
			dto.ChineseName = u.Info.ChineseName
			dto.EnglishName = u.Info.EnglishName
			dto.Sex = u.Info.Sex
			dto.Email = u.Info.Email
			dto.Position = u.Info.Position
			dto.Shift = u.Info.Shift
			dto.SFCNumber = u.Info.SFCNumber
			dto.FSLNumber = u.Info.FSLNumber
			dto.DepartmentNumber = u.Info.DepartmentNumber
			dto.PID = u.Info.PID
			dto.Section = u.Info.Section
			dto.WBICode = u.Info.WBICode
		} else {
			// info 缺失（幽灵账号）时兜底为禁用，避免前端拿到 null 误判
			dto.Status = ptrInt8(model.UserStatusDisabled)
		}
		dtos = append(dtos, dto)
	}
	return dtos
}

// ptrInt8 返回 int8 的指针（辅助兜底默认值）
func ptrInt8(v int8) *int8 { return &v }

// GetUserInfo 获取自身信息。
// 返回扁平结构：将 nxp_user 与 nxp_user_info 的全部字段铺到 userInfo 顶层，
// 同时保留 User 子结构（前端依赖 u.User.id / u.User.name）与 role 字符串角色。
func (u UserApi) GetUserInfo(c *gin.Context) {
	claims, ok := middleware.GetJWTClaims(c)
	if !ok {
		response.Fail(c, "未获取到用户信息")
		return
	}

	// GetNxpUserByID 一次查出 nxp_user + nxp_user_info，避免二次查询
	user, err := service.GetNxpUserByID(claims.UserID)
	if err != nil {
		response.FailError(c, err)
		return
	}

	// 前端通过用户名（admin/sopadmin，大小写不敏感）判定是否系统管理员；
	// role 顶层字段按此口径输出字符串角色，非管理员统一给 JUNIOR_ENGINEER。
	roleName := "JUNIOR_ENGINEER"
	name := user.User.Name
	if service.IsAdminUserName(name) {
		roleName = "ADMIN"
	}

	// createTime 统一格式化（与列表接口一致）
	createTime := ""
	if !user.User.CreateTime.IsZero() {
		createTime = user.User.CreateTime.Format("2006-01-02 15:04:05")
	}

	// nxp_user 全字段 + role，均铺到顶层
	userInfo := gin.H{
		"id":         user.User.ID,
		"userId":     user.User.UserID,
		"name":       name,
		"username":   name,
		"employeeId": user.User.UserID,
		"role":       roleName,
		"createTime": createTime,
	}

	// nxp_user_info 全部字段铺到顶层（info 为 nil 时跳过）
	if user.Info != nil {
		userInfo["roleId"] = user.Info.RoleID
		userInfo["status"] = user.Info.Status
		userInfo["leaderUid"] = user.Info.LeaderUID
		userInfo["pid"] = user.Info.PID
		userInfo["phone"] = user.Info.Tel
		userInfo["chineseName"] = user.Info.ChineseName
		userInfo["englishName"] = user.Info.EnglishName
		userInfo["sex"] = user.Info.Sex
		userInfo["email"] = user.Info.Email
		userInfo["position"] = user.Info.Position
		userInfo["shift"] = user.Info.Shift
		userInfo["area"] = user.Info.Area
		userInfo["stationId"] = user.Info.StationID
		userInfo["sfcNumber"] = user.Info.SFCNumber
		userInfo["fslNumber"] = user.Info.FSLNumber
		userInfo["mealCardCode"] = user.Info.MealCardCode
		userInfo["wbiCode"] = user.Info.WBICode
		userInfo["badgeCode"] = user.Info.BadgeCode
		userInfo["skill"] = user.Info.Skill
		userInfo["department"] = user.Info.DepartmentNumber
		userInfo["section"] = user.Info.Section
		userInfo["gembaRole"] = user.Info.GembaRole
		userInfo["team"] = user.Info.Team
		userInfo["shoeType"] = user.Info.ShoeType
		userInfo["newArea"] = user.Info.NewArea
		userInfo["sectionIr"] = user.Info.SectionIR
	}

	// 保留 User 子结构：nxp_user 主键 id / userId / name / createTime，
	// 兼容前端 u.User.id / u.User.name 访问（与 JWT/后端 admin 判定一致）。
	userInfo["User"] = gin.H{
		"id":         user.User.ID,
		"userId":     user.User.UserID,
		"name":       name,
		"createTime": createTime,
	}

	response.Result(c, gin.H{"userInfo": userInfo}, "获取成功")
}
