package initialize

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"go.uber.org/zap"

	"sop/global"
	"sop/internal/database"
	"sop/internal/model"
)

// InitDatabase 初始化三个数据库连接；SOPHub 表不存在时自动建表，已存在的表不做任何改动。
// Onebe / SPAREPART 为只读老库：只建连接，绝不执行任何建表/种子等 DDL/DML。
func InitDatabase() error {
	if err := database.InitSOPHub(); err != nil {
		return fmt.Errorf("init SOPHub database failed: %w", err)
	}
	if err := database.InitOnebe(); err != nil {
		return fmt.Errorf("init Onebe database failed: %w", err)
	}
	if err := database.InitSparepart(); err != nil {
		return fmt.Errorf("init SPAREPART database failed: %w", err)
	}
	// 以下建表与种子数据全部只作用于 SOPHub（SOPDB），Onebe 表结构由老系统维护
	models := []interface{}{
		&model.NxpSopGuide{},
		&model.NxpSopGuideStep{},
		&model.NxpSopInquiry{},
		&model.NxpSopMedia{},
		&model.NxpSopRepairRecord{},
		&model.NxpSopMenu{},
		&model.NxpSopPosition{},
		&model.NxpSopExecutionOption{},
		&model.NxpSopFaultCategory{},
		&model.NxpSopSparepartAttr{},
	}
	if err := ensureTables(models); err != nil {
		return err
	}
	if err := seedPresetData(); err != nil {
		return err
	}
	if err := migrateRepairRecordColumns(); err != nil {
		return err
	}
	if err := migrateGuideStepColumns(); err != nil {
		return err
	}
	// 废弃列清理：自动删除已不再使用的历史列（幂等，列不存在时跳过）
	if err := dropDeprecatedColumns(); err != nil {
		return err
	}
	return nil
}

// migrateGuideStepColumns 为 nxp_sop_guide_step 表幂等补充新增列。
// ensureTables 对已存在的表不会执行 AutoMigrate，因此这里显式补列（每列可重复执行，列已存在则跳过）。
func migrateGuideStepColumns() error {
	db := database.SOPDB()
	table := (&model.NxpSopGuideStep{}).TableName()
	cols := map[string]string{
		// number 列值不可为空：存量表加 NOT NULL 列需 DEFAULT 兜底（SQL Server 回填 0）
		"number": "INT NOT NULL DEFAULT 0",
	}
	for col, typ := range cols {
		if err := ensureColumnIfMissing(db, table, col,
			`ALTER TABLE dbo.`+table+` ADD `+col+` `+typ); err != nil {
			return err
		}
	}
	global.Logger.Info("已为步骤表补充新增列", zap.String("table", table))
	return nil
}

// dropDeprecatedColumns 清理历史遗留的废弃列，列不存在时幂等跳过。
// 说明：删除的列对应的模型字段已随版本移除，这里只是物理回收表结构。
func dropDeprecatedColumns() error {
	db := database.SOPDB()
	tables := []struct {
		table string
		col   string
	}{
		{table: (&model.NxpSopGuideStep{}).TableName(), col: "branches"},
	}
	for _, tc := range tables {
		if err := ensureColumnRemoved(db, tc.table, tc.col); err != nil {
			return err
		}
	}
	global.Logger.Info("已清理废弃数据库列")
	return nil
}

// ensureColumnRemoved 目标列存在时执行 dropSQL（幂等，已不存在则直接跳过）
func ensureColumnRemoved(db *gorm.DB, table, column string) error {
	var n int64
	if err := db.Raw(`SELECT COUNT(*) FROM sys.columns
WHERE object_id = OBJECT_ID(?) AND name = ?`, table, column).Scan(&n).Error; err != nil {
		return fmt.Errorf("检查 %s.%s 列失败: %w", table, column, err)
	}
	if n == 0 {
		return nil
	}
	if err := db.Exec(`ALTER TABLE dbo.` + table + ` DROP COLUMN ` + column).Error; err != nil {
		return fmt.Errorf("删除 %s.%s 列失败: %w", table, column, err)
	}
	return nil
}

// migrateRepairRecordColumns 为 nxp_sop_repair_record 表幂等补充缺失列（fault_code 及各新增字段）。
// ensureTables 对已存在的表不会执行 AutoMigrate，因此这里显式补列（每列可重复执行）。
func migrateRepairRecordColumns() error {
	db := database.SOPDB()
	table := (&model.NxpSopRepairRecord{}).TableName()
	cols := map[string]string{
		"fault_code":            "NVARCHAR(255) NULL",
		"machine_model_id":      "BIGINT NULL",
		"step_id":               "BIGINT NULL",
		"repair_position_id":    "BIGINT NULL",
		"repair_content_id":     "BIGINT NULL",
		"situation_description": "NVARCHAR(1000) NULL",
		"request_number":        "NVARCHAR(100) NULL",
		"repair_time":           "DATETIME2 NULL",
		"repair_number":         "NVARCHAR(100) NULL",
		"sparepart_s_ids":       "NVARCHAR(MAX) NULL",
	}
	for col, typ := range cols {
		if err := ensureColumnIfMissing(db, table, col,
			`ALTER TABLE dbo.`+table+` ADD `+col+` `+typ); err != nil {
			return err
		}
	}
	global.Logger.Info("已为维修记录表补充新增列", zap.String("table", table))
	return nil
}

// ensureColumnIfMissing 目标列不存在时执行 addSQL（幂等，已存在则直接跳过）
func ensureColumnIfMissing(db *gorm.DB, table, column, addSQL string) error {
	var n int64
	if err := db.Raw(`SELECT COUNT(*) FROM sys.columns
WHERE object_id = OBJECT_ID(?) AND name = ?`, table, column).Scan(&n).Error; err != nil {
		return fmt.Errorf("检查 %s.%s 列失败: %w", table, column, err)
	}
	if n > 0 {
		return nil
	}
	if err := db.Exec(addSQL).Error; err != nil {
		return fmt.Errorf("为 %s 补充 %s 列失败: %w", table, column, err)
	}
	return nil
}

// seedPresetData 种入预置的菜单与职务数据（仅当表为空时插入，防止重复），
// 职务表为预置只读表：写入后即校验其恒等于初始化数据，非法时启动失败。
func seedPresetData() error {
	if err := seedIfEmpty(&model.NxpSopMenu{}, seedMenus(model.NxpSopMenuNames)); err != nil {
		return err
	}
	if err := seedIfEmpty(&model.NxpSopPosition{}, seedPositions(model.NxpSopPositionNames)); err != nil {
		return err
	}
	return verifyPositions()
}

// verifyPositions 校验 nxp_sop_position 与预置职务完全一致（数量、id 顺序、名称），
// 一旦被篡改即返回错误，该错误会向上传播导致服务启动失败（fail-fast）。
// 错误信息与日志中会给出正确的期望顺序与 id，便于人工修复。
func verifyPositions() error {
	var rows []model.NxpSopPosition
	if err := database.SOPDB().Model(&model.NxpSopPosition{}).Order("id ASC").Find(&rows).Error; err != nil {
		return fmt.Errorf("查询 nxp_sop_position 失败: %w", err)
	}

	names := model.NxpSopPositionNames
	expected := formatPositionExpected(names)

	if len(rows) != len(names) {
		msg := fmt.Sprintf("nxp_sop_position 记录数不正确：当前 %d 条，期望 %d 条。正确顺序与 id：%s",
			len(rows), len(names), expected)
		abortPositionInconsistent(msg)
		return fmt.Errorf("nxp_sop_position 一致性校验失败: %s", msg)
	}

	for i := range rows {
		if rows[i].PositionName != names[i] {
			msg := fmt.Sprintf("nxp_sop_position 第 %d 行不符：当前 [id=%d] %q，期望 [id=%d] %q。正确顺序与 id：%s",
				i+1, rows[i].ID, rows[i].PositionName, i+1, names[i], expected)
			abortPositionInconsistent(msg)
			return fmt.Errorf("nxp_sop_position 一致性校验失败: %s", msg)
		}
	}
	return nil
}

// abortPositionInconsistent 记录一致性校验失败日志，作为人工修复时的对照依据
func abortPositionInconsistent(msg string) {
	global.Logger.Error("nxp_sop_position 数据与初始化不一致，服务拒绝启动",
		zap.String("detail", msg))
	fmt.Printf("%s\n", msg)
}

// formatPositionExpected 生成 "id:职务" 的顺序清单，用于人工修复对照
func formatPositionExpected(names []string) string {
	parts := make([]string, 0, len(names))
	for i, name := range names {
		parts = append(parts, fmt.Sprintf("%d:%s", i+1, name))
	}
	return strings.Join(parts, "、")
}

// seedIfEmpty 目标表为空时执行 seedFn 插入预置数据
func seedIfEmpty(m interface{}, seedFn func() error) error {
	var n int64
	if err := database.SOPDB().Model(m).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return seedFn()
}

// seedMenus 预置菜单名称转结构体列表
func seedMenus(names []string) func() error {
	return func() error {
		rows := make([]model.NxpSopMenu, 0, len(names))
		for _, name := range names {
			rows = append(rows, model.NxpSopMenu{MenuName: name})
		}
		return database.SOPDB().CreateInBatches(rows, 100).Error
	}
}

// seedPositions 预置职务名称转结构体列表
// 表已定义只读钩子（见 model.NxpSopPosition），这里必须跳过钩子才能写入初始化数据
func seedPositions(names []string) func() error {
	return func() error {
		rows := make([]model.NxpSopPosition, 0, len(names))
		for _, name := range names {
			rows = append(rows, model.NxpSopPosition{PositionName: name})
		}
		return database.SOPDB().Session(&gorm.Session{SkipHooks: true}).CreateInBatches(rows, 100).Error
	}
}

// ensureTables 逐表检查：已存在则跳过，不存在才 AutoMigrate 建表，避免对老库触发列类型变更
func ensureTables(models []interface{}) error {
	for _, m := range models {
		stmt := &gorm.Statement{DB: database.SOPDB()}
		if err := stmt.Parse(m); err != nil {
			return err
		}
		var n int64
		// sys.tables 对比表名（默认 dbo schema）
		if err := database.SOPDB().Raw("SELECT COUNT(*) FROM sys.tables WHERE name = ?", stmt.Table).Scan(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		if err := database.SOPDB().AutoMigrate(m); err != nil {
			return fmt.Errorf("auto migrate %s failed: %w", stmt.Table, err)
		}
	}
	return nil
}
