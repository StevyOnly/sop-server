package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// MenuTreeNode 菜单树节点（预留层级扩展位，当前菜单为平铺结构）
type MenuTreeNode struct {
	model.NxpSopMenu
	Children []*MenuTreeNode `json:"children"`
}

// GetMenuList 分页查询菜单列表，支持按名称模糊搜索
func GetMenuList(page, pageSize int, keyword string) ([]model.NxpSopMenu, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopMenu{})
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("menu_name LIKE ?", like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopMenu
	if err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpSopMenu{}
	}
	return list, total, nil
}

// GetAllMenus 返回全部菜单列表（供职务-菜单配置整表展示）
func GetAllMenus() ([]model.NxpSopMenu, error) {
	var list []model.NxpSopMenu
	if err := database.SOPDB().Model(&model.NxpSopMenu{}).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.NxpSopMenu{}
	}
	return list, nil
}

// SaveMenu 新增或更新菜单
// 新增时校验名称唯一；更新时校验记录存在，并做名称唯一性（排除自身）
func SaveMenu(menu *model.NxpSopMenu) error {
	if menu.MenuName == "" {
		return bizerr.NewParam("菜单名称不能为空")
	}

	if menu.ID == 0 {
		// 主键自增：直接插入；预置名称重复时返回冲突
		var count int64
		if err := database.SOPDB().Model(&model.NxpSopMenu{}).Where("menu_name = ?", menu.MenuName).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return bizerr.NewConflict("菜单名称已存在")
		}
		return database.SOPDB().Create(menu).Error
	}

	var existing model.NxpSopMenu
	if err := database.SOPDB().First(&existing, menu.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrMenuNotFound
		}
		return err
	}

	var count int64
	if err := database.SOPDB().Model(&model.NxpSopMenu{}).
		Where("menu_name = ? AND id <> ?", menu.MenuName, menu.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewConflict("菜单名称已存在")
	}

	return database.SOPDB().Model(&model.NxpSopMenu{}).Where("id = ?", menu.ID).Update("menu_name", menu.MenuName).Error
}

// DeleteMenu 删除菜单，同时清理职务-菜单关联表中的关联，避免残留脏数据
func DeleteMenu(id uint) error {
	return database.SOPDB().Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&model.NxpSopMenu{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return bizerr.ErrMenuNotFound
		}
		// join 表 nxp_sop_position_menu 由 many2many 自动管理（无 id 列），直接按 menu_id 原生删除
		return tx.Exec("DELETE FROM nxp_sop_position_menu WHERE menu_id = ?", id).Error
	})
}

// GetPositionList 返回全部职务（预置数据，按 id 排序）
func GetPositionList() ([]model.NxpSopPosition, error) {
	var list []model.NxpSopPosition
	if err := database.SOPDB().Model(&model.NxpSopPosition{}).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.NxpSopPosition{}
	}
	return list, nil
}

// GetPositionMenus 按职务返回该职务可见的菜单列表
func GetPositionMenus(positionID int) ([]model.NxpSopMenu, error) {
	var position model.NxpSopPosition
	if err := database.SOPDB().First(&position, positionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ErrPositionNotFound
		}
		return nil, err
	}

	// join 表 nxp_sop_position_menu 由 many2many 自动管理（无 id 列），直接查 menu_id
	menuIDs := database.SOPDB().Table("nxp_sop_position_menu").
		Where("position_id = ?", positionID).
		Select("menu_id")
	var list []model.NxpSopMenu
	if err := database.SOPDB().Where("id IN (?)", menuIDs).Order("id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.NxpSopMenu{}
	}
	return list, nil
}

// GetPositionsWithMenus 返回全部职务及其可见菜单（供前端一次性加载配置）
// 通过 NxpSopPosition.Menus 的 many2many 关联 Preload 一次性取出，避免 N+1 查询。
func GetPositionsWithMenus() ([]model.NxpSopPosition, error) {
	var list []model.NxpSopPosition
	if err := database.SOPDB().Model(&model.NxpSopPosition{}).
		Preload("Menus").
		Order("id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.NxpSopPosition{}
	}
	return list, nil
}

// SavePositionMenus 整体覆盖指定职务的可见菜单（事务内先删后插，原子替换）
// 校验职务存在；menuIDs 为空表示清空该职务的菜单。
// 实现直接删除并插入连接表 nxp_sop_position_menu（而不是 Association("Menus").Replace，
// 后者无事务，many2many 分支实际是先逐条插入新行、再 DELETE 不在新集合中的旧行，
// 中途失败会留下新旧混合状态）。也不触碰只读的 nxp_sop_position 本身（不触发其 BeforeUpdate 钩子）。
func SavePositionMenus(positionID int, menuIDs []int) error {
	// 过滤非法菜单 ID（<=0），并按原顺序去重，避免重复行
	seen := make(map[int]struct{}, len(menuIDs))
	cleanIDs := make([]int, 0, len(menuIDs))
	for _, id := range menuIDs {
		if id <= 0 {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		cleanIDs = append(cleanIDs, id)
	}

	return database.SOPDB().Transaction(func(tx *gorm.DB) error {
		// 校验职务存在
		var position model.NxpSopPosition
		if err := tx.First(&position, positionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bizerr.ErrPositionNotFound
			}
			return err
		}

		// 先删旧关联，整体清空该职务的连接表行
		// join 表 nxp_sop_position_menu 由 many2many 自动管理（无 id 列），直接原生删插
		if err := tx.Exec("DELETE FROM nxp_sop_position_menu WHERE position_id = ?", positionID).Error; err != nil {
			return err
		}

		// 再插新关联
		if len(cleanIDs) == 0 {
			return nil
		}
		values := make([]string, 0, len(cleanIDs))
		args := make([]interface{}, 0, len(cleanIDs)*2)
		for _, id := range cleanIDs {
			values = append(values, "(?, ?)")
			args = append(args, positionID, id)
		}
		insertSQL := "INSERT INTO nxp_sop_position_menu (position_id, menu_id) VALUES " + strings.Join(values, ", ")
		return tx.Exec(insertSQL, args...).Error
	})
}

// GetUserMenus 按用户返回其可见的菜单列表
//   - 系统管理员（admin / sopadmin）或用户 id 为 1（nxp_user.id=1）时，返回全部菜单；
//     admin/sopadmin 按用户名判定（大小写不敏感），不受职位限制
//   - 其他用户按 nxp_user_info.position 关联 nxp_sop_position_menu 与 nxp_sop_menu；
//     position 为空（未设置职务）时返回空列表
//
// 双库改造（Step 6 内存组装）：nxp_user/nxp_user_info 在 Onebe、菜单表在 SOPHub，
// 不能再用子查询联查。经 GetNxpUserByID（内部两次 Onebe 查询）取 position，
// 再用 SOPDB 查职务菜单。兜底约定：用户/信息不存在→空菜单；Onebe 连接异常→抛错。
func GetUserMenus(userID uint) ([]model.NxpSopMenu, error) {
	// 兜底：id 为 1 的用户即为 admin，拥有全部菜单权限（防御性保留，兼容查不到用户名的边界）
	if userID == 1 {
		return GetAllMenus()
	}

	user, err := GetNxpUserByID(userID)
	if err != nil {
		// 用户不存在按空列表处理；连接级错误向上抛出
		if errors.Is(err, bizerr.ErrUserNotFound) {
			return []model.NxpSopMenu{}, nil
		}
		return nil, err
	}
	// 系统管理员（admin/sopadmin）无论职位如何，均返回全部菜单
	if IsAdminUserName(user.User.Name) {
		return GetAllMenus()
	}
	// info 缺失或未设置职务时返回空列表
	if user.Info == nil || user.Info.Position == nil || *user.Info.Position <= 0 {
		return []model.NxpSopMenu{}, nil
	}

	return GetPositionMenus(*user.Info.Position)
}
