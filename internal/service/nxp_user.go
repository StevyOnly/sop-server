package service

import (
	"crypto/subtle"
	"errors"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"sop/global"
	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
	"sop/pkg/jwt"
)

// UserNxp 用户行（nxp_user JOIN nxp_user_info）
type UserNxp struct {
	User *model.NxpUser
	Info *model.NxpUserInfo
}

// IsAdminUserName 判断是否为系统管理员（admin / sopadmin，大小写不敏感）。
// 与 api.GetUserInfo 的 role 判定口径保持一致，供菜单等按用户释放全部权限的逻辑复用。
func IsAdminUserName(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return n == "admin" || n == "sopadmin"
}

// LoginResult 登录返回结果
type LoginResult struct {
	User      *model.NxpUser
	Token     string
	ExpiresAt int64
}

// checkPassword 校验密码（bcrypt）
func checkPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// subtleConstantTimeEqual 常量时间字符串比较（避免 == 的时序侧信道泄露）。
// 任一为空或长度不等直接返回 false，不执行底层更耗时路径。
func subtleConstantTimeEqual(a, b string) bool {
	if a == "" && b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// Login 用户登录，按 name 查询用户，校验 bcrypt 密码与启用状态
func Login(name, password string) (*LoginResult, error) {
	user, err := GetNxpUserByName(name)
	if err != nil {
		// 仅「用户不存在」伪装成密码错误（防止用户名枚举）；
		// Onebe 基础设施故障（连接/超时/驱动错误）不伪装，原样上抛
		// 并由上层 FailError 回 500 + 服务端日志，避免掩盖老库故障。
		if errors.Is(err, bizerr.ErrUserNotFound) {
			return nil, bizerr.ErrPasswordOrUser
		}
		global.Logger.Error("login failed: query user from onebe", zap.Error(err))
		return nil, err
	}

	// 配置开启时跳过密码校验（测试用）
	if !global.GetConfig().Server.SkipPasswordCheck {
		if !checkPassword(user.User.Password, password) {
			return nil, bizerr.ErrPasswordOrUser
		}
	}

	if user.Info == nil {
		return nil, bizerr.ErrUserDisabled
	}
	if user.Info.Status != nil && *user.Info.Status != model.UserStatusEnabled {
		return nil, bizerr.ErrUserDisabled
	}

	token, expiresAt, err := jwt.GenerateToken(uint(user.User.ID))
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:      user.User,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// AppLogin APP 端免密码登录：校验 appKey 与用户存在且启用后，签发永久 AppToken。
// appKey 是"是否 APP 端"的固定标志（配置 server.appTokenKey），与用户密码无关。
// 返回的 ExpiresAt 为 0（永久有效，不设过期时间），仅作占位。
func AppLogin(userID uint, appKey string) (*LoginResult, error) {
	// 校验固定密钥：不匹配直接 401，且不暴露期望值，防止探测
	expected := global.GetConfig().Server.AppTokenKey
	if expected == "" || !subtleConstantTimeEqual(appKey, expected) {
		return nil, bizerr.NewUnauthorized("无效的 App 凭证")
	}

	// 反查 Onebe 确认用户存在且启用（与 Login/EnsureActiveUser 口径一致）：
	// 禁用/缺失（幽灵账号）拒绝签发；Onebe 基础设施故障原样上抛由上层 FailError 回 500，避免伪装
	user, err := EnsureActiveUser(userID)
	if err != nil {
		global.Logger.Warn("app login rejected",
			zap.Uint("userId", userID),
			zap.Error(err))
		return nil, err
	}
	if user == nil {
		// 仅发生在 Onebe 抖动被 EnsureActiveUser 放行（返回 nil,nil）时：拒绝签发并提示重试
		return nil, bizerr.NewParam("用户信息暂时无法确认，请稍后重试")
	}

	token, err := jwt.GenerateAppToken(uint(user.User.ID))
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:      user.User,
		Token:     token,
		ExpiresAt: 0,
	}, nil
}

// GetNxpUserByName 根据 name 查询用户（nxp_user + nxp_user_info）
func GetNxpUserByName(name string) (*UserNxp, error) {
	var user model.NxpUser
	err := database.Onebe().Where("name = ?", name).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, bizerr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	var info model.NxpUserInfo
	if err := database.Onebe().Where("id = ?", user.UserID).First(&info).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// info 记录缺失时返回 Info 为 nil，便于调用方区分“数据不完整”
		return &UserNxp{User: &user, Info: nil}, nil
	}
	return &UserNxp{User: &user, Info: &info}, nil
}

// GetUsersByIDs 按 ID 列表批量查询用户（nxp_user + nxp_user_info）
// SQL Server 单条语句参数上限约 2100，按 1000 个 ID 一批分片查询，避免超限报错
func GetUsersByIDs(ids []uint) ([]UserNxp, error) {
	if len(ids) == 0 {
		return []UserNxp{}, nil
	}
	const chunkSize = 1000
	result := make([]UserNxp, 0, len(ids))
	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}
		chunk, err := getUsersByIDsChunk(ids[start:end])
		if err != nil {
			return nil, err
		}
		result = append(result, chunk...)
	}
	return result, nil
}

// getUsersByIDsChunk 单批次查询（调用方保证 len(ids) <= 1000）
func getUsersByIDsChunk(ids []uint) ([]UserNxp, error) {
	var users []model.NxpUser
	if err := database.Onebe().Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []UserNxp{}, nil
	}

	var infos []model.NxpUserInfo
	userIDs := make([]int, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, int(u.UserID))
	}
	if err := database.Onebe().Model(&model.NxpUserInfo{}).Where("id IN ?", userIDs).Find(&infos).Error; err != nil {
		return nil, err
	}
	infoMap := make(map[uint]*model.NxpUserInfo, len(infos))
	for i := range infos {
		infoMap[infos[i].ID] = &infos[i]
	}

	result := make([]UserNxp, 0, len(users))
	for i := range users {
		item := UserNxp{User: &users[i]}
		if info, ok := infoMap[uint(users[i].UserID)]; ok {
			item.Info = info
		}
		result = append(result, item)
	}
	return result, nil
}

// GetNxpUserByID 根据 nxp_user 主键 id 查询用户（nxp_user + nxp_user_info）
func GetNxpUserByID(id uint) (*UserNxp, error) {
	var user model.NxpUser
	err := database.Onebe().First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, bizerr.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	var info model.NxpUserInfo
	if err := database.Onebe().Where("id = ?", user.UserID).First(&info).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// info 记录缺失时返回 Info 为 nil，便于调用方区分“数据不完整”
		return &UserNxp{User: &user, Info: nil}, nil
	}
	return &UserNxp{User: &user, Info: &info}, nil
}

// EnsureActiveUser 关键写操作前置校验：反查 Onebe 确认用户存在且处于启用状态，
// 并返回用户数据供调用方复用（避免为取用户名等字段二次反查）。
// 背景：JWT 只验签不查库，Onebe 侧禁用/删除用户后，已签发 token 在有效期内仍可用。
// 兜底约定（与 Step 6 一致）：仅在明确判定「用户不存在」或「已禁用」时拒绝；
// 若 Onebe 查询发生基础设施错误（不可用/超时），则放行，避免因老库抖动误伤正常写请求
// （此时返回的 *UserNxp 为 nil，调用方不得依赖其非空）。
func EnsureActiveUser(userID uint) (*UserNxp, error) {
	user, err := GetNxpUserByID(userID)
	if err != nil {
		if errors.Is(err, bizerr.ErrUserNotFound) {
			return nil, err
		}
		// 非「未找到」类错误视为 Onebe 抖动，直接放行，不阻断业务写入
		return nil, nil
	}
	if user.Info == nil {
		return nil, bizerr.ErrUserDisabled
	}
	if user.Info.Status != nil && *user.Info.Status != model.UserStatusEnabled {
		return nil, bizerr.ErrUserDisabled
	}
	return user, nil
}

// UserQuery 用户列表查询条件
type UserQuery struct {
	Name     string
	Tel      string
	Email    string
	RoleID   *int
	Status   *int8
	SFC      string
	FSL      string
	Dept     string
	Position *int
}

// GetUserList 联表分页查询用户列表（nxp_user + nxp_user_info）
func GetUserList(page, pageSize int, keyword, orderKey string, desc bool, cond UserQuery) ([]UserNxp, int64, error) {
	// 过滤条件可能涉及 info 字段，故用 EXISTS / 内联过滤
	db := database.Onebe().Model(&model.NxpUser{})
	hasInfoCond := keyword != "" ||
		cond.Name != "" || cond.Tel != "" || cond.Email != "" || cond.SFC != "" ||
		cond.FSL != "" || cond.Dept != "" || cond.RoleID != nil || cond.Position != nil || cond.Status != nil
	// 排序字段落在 nxp_user_info 上时同样必须 JOIN，否则 ORDER BY 引用不到该表
	_, orderByInfo := infoOrderFields[orderKey]

	if hasInfoCond || orderByInfo {
		// 用户信息通过 nxp_user.user_id 关联到 nxp_user_info.id（两表主键独立取号，不能用 id=id 拼接）
		db = db.Joins("LEFT JOIN nxp_user_info ON nxp_user_info.id = nxp_user.user_id")
	}

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("(nxp_user.name LIKE ? OR nxp_user_info.chinese_name LIKE ? OR nxp_user_info.tel LIKE ? OR nxp_user_info.email LIKE ? OR nxp_user_info.sfc_number LIKE ? OR nxp_user_info.english_name LIKE ?)",
			like, like, like, like, like, like)
	}
	if cond.Name != "" {
		db = db.Where("nxp_user.name = ?", cond.Name)
	}
	if cond.Tel != "" {
		db = db.Where("nxp_user_info.tel = ?", cond.Tel)
	}
	if cond.Email != "" {
		db = db.Where("nxp_user_info.email = ?", cond.Email)
	}
	if cond.SFC != "" {
		db = db.Where("nxp_user_info.sfc_number = ?", cond.SFC)
	}
	if cond.FSL != "" {
		db = db.Where("nxp_user_info.fsl_number = ?", cond.FSL)
	}
	if cond.Dept != "" {
		db = db.Where("nxp_user_info.department_number = ?", cond.Dept)
	}
	if cond.RoleID != nil {
		db = db.Where("nxp_user_info.role_id = ?", *cond.RoleID)
	}
	if cond.Position != nil {
		db = db.Where("nxp_user_info.position = ?", *cond.Position)
	}
	if cond.Status != nil {
		db = db.Where("nxp_user_info.status = ?", *cond.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页取 user 主键列表。
	// JOIN 条件是 nxp_user_info.id（主键）= nxp_user.user_id，一对一关联不会产生重复行，
	// 因此不能用 DISTINCT：SQL Server 要求 SELECT DISTINCT 时 ORDER BY 列必须出现在选择列表中
	var userIDs []int
	orderCls := orderClause(orderKey, desc)
	if err := db.Order(orderCls).Offset((page-1)*pageSize).Limit(pageSize).Pluck("nxp_user.id", &userIDs).Error; err != nil {
		return nil, 0, err
	}
	if len(userIDs) == 0 {
		return []UserNxp{}, total, nil
	}

	// 批量查询用户信息
	var users []model.NxpUser
	if err := database.Onebe().Model(&model.NxpUser{}).Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 批量查询用户信息：按 nxp_user.user_id（即 nxp_user_info.id）关联
	infoIDs := make([]int, 0, len(users))
	for _, u := range users {
		if u.UserID > 0 {
			infoIDs = append(infoIDs, u.UserID)
		}
	}
	if len(infoIDs) == 0 {
		return []UserNxp{}, total, nil
	}
	var infos []model.NxpUserInfo
	if err := database.Onebe().Model(&model.NxpUserInfo{}).Where("id IN ?", infoIDs).Find(&infos).Error; err != nil {
		return nil, 0, err
	}
	infoMap := make(map[uint]*model.NxpUserInfo, len(infos))
	for i := range infos {
		infoMap[infos[i].ID] = &infos[i]
	}

	// 按分页顺序重排
	orderIdx := make(map[int]int, len(userIDs))
	for i, id := range userIDs {
		orderIdx[id] = i
	}
	// 按 len(userIDs) 而非 len(users) 开槽：users 来自 `id IN userIDs`，
	// Pluck 与 Find 之间存在并发删除窗口（len(users) < len(userIDs)）时，
	// 幸存用户的位置 idx 可达 len(userIDs)-1，按较短长度开槽会越界 panic。
	// 常规路径两表行数相等、键必命中；返回前剔除空槽，保证行的 User 恒非 nil。
	sorted := make([]UserNxp, len(userIDs))
	for i := range users {
		idx, ok := orderIdx[int(users[i].ID)]
		if !ok {
			continue
		}
		sorted[idx] = UserNxp{User: &users[i], Info: infoMap[uint(users[i].UserID)]}
	}
	out := make([]UserNxp, 0, len(users))
	for _, r := range sorted {
		if r.User != nil {
			out = append(out, r)
		}
	}
	return out, total, nil
}

// userOrderFields 允许的外部排序字段白名单（防止 order 注入）
var userOrderFields = map[string]string{
	"id":           "nxp_user.id",
	"create_time":  "nxp_user.create_time",
	"name":         "nxp_user.name",
	"chinese_name": "nxp_user_info.chinese_name",
	"tel":          "nxp_user_info.tel",
}

// infoOrderFields 排序列位于 nxp_user_info 表的 orderKey 集合，
// 命中时 GetUserList 必须附加 LEFT JOIN
var infoOrderFields = map[string]struct{}{
	"chinese_name": {},
	"tel":          {},
}

// orderClause 将外部 orderKey 映射为安全的排序子句；非法字段回退为默认排序
func orderClause(orderKey string, desc bool) string {
	col, ok := userOrderFields[orderKey]
	if !ok {
		return "nxp_user.id DESC"
	}
	if desc {
		return col + " DESC"
	}
	return col + " ASC"
}
