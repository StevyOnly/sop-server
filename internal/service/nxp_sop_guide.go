package service

import (
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// parseModelId 将字符串形式的机型 ID（即 nxp_facility.model）解析为 uint；
// 非数字或非法时返回 0。
func parseModelId(s string) uint {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(v)
}

// GuideWithLabel SOP 指南（列表展示），扩展故障分类文案
// 不直接在 model.NxpSopGuide 上添加展示字段，避免污染 SaveGuide 的持久化写入。
type GuideWithLabel struct {
	model.NxpSopGuide
	FaultCategoryLabel string `json:"faultCategoryLabel"`
}

// GetGuideList 分页查询指南列表（含步骤），支持按故障码/故障分类ID/关键词筛选
func GetGuideList(page, pageSize int, faultCode string, faultCategoryID uint, keyword string) ([]GuideWithLabel, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopGuide{})

	if faultCode != "" {
		db = db.Where("fault_code LIKE ?", "%"+faultCode+"%")
	}
	if faultCategoryID > 0 {
		db = db.Where("fault_category_id = ?", faultCategoryID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("fault_code LIKE ?", like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var guides []model.NxpSopGuide
	if err := db.Preload("Steps").Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&guides).Error; err != nil {
		return nil, 0, err
	}
	if guides == nil {
		guides = []model.NxpSopGuide{}
	}

	// 批量查询本页涉及的非零故障分类 id -> label 映射，填充 faultCategoryLabel
	labelMap, err := faultCategoryLabelMap(guides)
	if err != nil {
		return nil, 0, err
	}
	withLabel := make([]GuideWithLabel, 0, len(guides))
	for _, g := range guides {
		w := GuideWithLabel{NxpSopGuide: g}
		if g.FaultCategoryID > 0 {
			w.FaultCategoryLabel = labelMap[g.FaultCategoryID]
		}
		withLabel = append(withLabel, w)
	}
	return withLabel, total, nil
}

// faultCategoryLabelMap 批量查询 guides 涉及的非零故障分类 id->label 映射
func faultCategoryLabelMap(guides []model.NxpSopGuide) (map[uint]string, error) {
	ids := make([]uint, 0, len(guides))
	seen := make(map[uint]bool, len(guides))
	for _, g := range guides {
		if g.FaultCategoryID > 0 && !seen[g.FaultCategoryID] {
			seen[g.FaultCategoryID] = true
			ids = append(ids, g.FaultCategoryID)
		}
	}
	if len(ids) == 0 {
		return map[uint]string{}, nil
	}
	var cats []model.NxpSopFaultCategory
	if err := database.SOPDB().Where("id IN ?", ids).Find(&cats).Error; err != nil {
		return nil, err
	}
	labelMap := make(map[uint]string, len(cats))
	for _, c := range cats {
		labelMap[c.ID] = c.Label
	}
	return labelMap, nil
}

// GuideSample 指南精简视图（不含步骤），用于下拉/概览等无需步骤明细的场景。
// 显式列字段而非嵌入 NxpSopGuide，确保响应中不出现 steps（嵌入会带出 json:"steps"）。
type GuideSample struct {
	model.BaseModel
	MachineModelId       uint   `json:"machineModelId"`
	FaultCode            string `json:"faultCode"`
	FaultCategoryID      uint   `json:"faultCategoryId"`
	FaultCategoryLabel   string `json:"faultCategoryLabel"`
	QuestionTypeID       uint   `json:"questionTypeId"`
	RepairPositionID     uint   `json:"repairPositionId"`
	RepairContentID      uint   `json:"repairContentId"`
	FaultCodeDescription string `json:"faultCodeDescription"`
	Enabled              bool   `json:"enabled"`
	TotalOccurrenceCount int64  `json:"totalOccurrenceCount"`
}

// GetGuideSamplesByMachineModelIds 按设备（Facility ID）列表分页查询指南（不含步骤），按 id DESC 排序。
// 入参 ids 为 nxp_facility.id，通过 nxp_facility.model（即指南的 machine_model_id）中转后查询。
// 仅返回启用状态（Enabled=true）的指南。
// id 上限 1000（API binding 保证）：GetFacilitiesByIDs 内部已分片，后续 machine_model_id IN 参数 < 2100，无需再分片。
func GetGuideSamplesByMachineModelIds(page, pageSize int, ids []uint) ([]GuideSample, int64, error) {
	if len(ids) == 0 {
		return []GuideSample{}, 0, nil
	}
	// 设备 ID -> 机型 ID：nxp_facility.model 为字符串形态的 machine_model_id
	facilities, err := GetFacilitiesByIDs(ids)
	if err != nil {
		return nil, 0, err
	}
	modelIds := make([]uint, 0, len(facilities))
	seen := make(map[uint]struct{}, len(facilities))
	for _, f := range facilities {
		mid := parseModelId(f.Model)
		if mid == 0 {
			continue
		}
		if _, ok := seen[mid]; ok {
			continue
		}
		seen[mid] = struct{}{}
		modelIds = append(modelIds, mid)
	}
	if len(modelIds) == 0 {
		return []GuideSample{}, 0, nil
	}

	db := database.SOPDB().Model(&model.NxpSopGuide{}).Where("machine_model_id IN ?", modelIds).
		Where("enabled = ?", true) // 仅返回启用状态（Enabled=true）的指南，count 与 find 均受此条件约束

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var guides []model.NxpSopGuide
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&guides).Error; err != nil {
		return nil, 0, err
	}

	// 批量填充 faultCategoryLabel（复用 faultCategoryLabelMap，一次查出本页分类文案）
	labelMap, err := faultCategoryLabelMap(guides)
	if err != nil {
		return nil, 0, err
	}
	list := make([]GuideSample, 0, len(guides))
	for _, g := range guides {
		s := GuideSample{
			BaseModel:            g.BaseModel,
			MachineModelId:       g.MachineModelId,
			FaultCode:            g.FaultCode,
			FaultCategoryID:      g.FaultCategoryID,
			QuestionTypeID:       g.QuestionTypeID,
			RepairPositionID:     g.RepairPositionID,
			RepairContentID:      g.RepairContentID,
			FaultCodeDescription: g.FaultCodeDescription,
			Enabled:              g.Enabled,
			TotalOccurrenceCount: g.TotalOccurrenceCount,
		}
		if g.FaultCategoryID > 0 {
			s.FaultCategoryLabel = labelMap[g.FaultCategoryID]
		}
		list = append(list, s)
	}
	return list, total, nil
}

// GetGuidesByIds 按指南 ID 列表分页查询指南（含步骤），按 id ASC 排序。
// ids 上限 1000（API binding 保证），单条 SQL IN 参数 < 2100，无需分片。
func GetGuidesByIds(page, pageSize int, ids []uint) ([]GuideWithLabel, int64, error) {
	if len(ids) == 0 {
		return []GuideWithLabel{}, 0, nil
	}
	db := database.SOPDB().Model(&model.NxpSopGuide{}).Where("id IN ?", ids)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var guides []model.NxpSopGuide
	if err := db.Preload("Steps").Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&guides).Error; err != nil {
		return nil, 0, err
	}

	labelMap, err := faultCategoryLabelMap(guides)
	if err != nil {
		return nil, 0, err
	}
	withLabel := make([]GuideWithLabel, 0, len(guides))
	for _, g := range guides {
		w := GuideWithLabel{NxpSopGuide: g}
		if g.FaultCategoryID > 0 {
			w.FaultCategoryLabel = labelMap[g.FaultCategoryID]
		}
		withLabel = append(withLabel, w)
	}
	return withLabel, total, nil
}

// GetStepsByFaultCode 按故障码精确匹配，返回命中的全部步骤（扁平数组，data 直接为步骤对象）。
// 先按 fault_code 查命中指南 id，再统一查这些指南下的步骤；
// 指南不过滤 enabled（返回全部状态）；步骤保留全部（含停用，由前端决定）。
// 排序：按历史维修次数 history_repair_count 降序（次数多者靠前），次数相同再按 id ASC 保证稳定。
func GetStepsByFaultCode(faultCode string) ([]model.NxpSopGuideStep, error) {
	if faultCode == "" {
		return []model.NxpSopGuideStep{}, nil
	}
	var guideIDs []uint
	if err := database.SOPDB().
		Model(&model.NxpSopGuide{}).
		Where("fault_code = ?", faultCode).
		Pluck("id", &guideIDs).Error; err != nil {
		return nil, err
	}
	if len(guideIDs) == 0 {
		return []model.NxpSopGuideStep{}, nil
	}

	var steps []model.NxpSopGuideStep
	if err := database.SOPDB().
		Where("guide_id IN ?", guideIDs).
		Order("history_repair_count DESC, id ASC").
		Find(&steps).Error; err != nil {
		return nil, err
	}
	if steps == nil {
		steps = []model.NxpSopGuideStep{}
	}
	return steps, nil
}

// SaveGuide 创建或更新指南（id 为 0 创建；否则更新主记录，步骤按 ID diff 增删改）
func SaveGuide(guide *model.NxpSopGuide) error {
	// 轻量校验，避免 DTO 批量写入超大/畸形数据：
	// 步骤数量上限 100（与前端单指南步骤规模一致），步骤标题必填且限长。
	const maxSteps = 100
	if len(guide.Steps) > maxSteps {
		return bizerr.NewParam("单个指南步骤数量不能超过 100")
	}
	for i := range guide.Steps {
		if title := strings.TrimSpace(guide.Steps[i].Title); title == "" {
			return bizerr.NewParam("步骤标题不能为空")
		} else if len([]rune(title)) > 200 {
			return bizerr.NewParam("步骤标题长度不能超过 200 字")
		}
	}

	if guide.ID == 0 {
		// 新建指南：步骤全部按新记录插入。前端本地新增步骤带有临时 id，
		// 必须归零交给数据库自增，否则关联 Create 会显式写入 IDENTITY 列而报错。
		for i := range guide.Steps {
			guide.Steps[i].ID = 0
		}
		return database.SOPDB().Create(guide).Error
	}

	var existing model.NxpSopGuide
	if err := database.SOPDB().First(&existing, guide.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrGuideNotFound
		}
		return err
	}

	return database.SOPDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.NxpSopGuide{}).Where("id = ?", guide.ID).Updates(map[string]interface{}{
			"machine_model_id":       guide.MachineModelId,
			"fault_code":             guide.FaultCode,
			"fault_category_id":      guide.FaultCategoryID,
			"question_type_id":       guide.QuestionTypeID,
			"repair_position_id":     guide.RepairPositionID,
			"repair_content_id":      guide.RepairContentID,
			"fault_code_description": guide.FaultCodeDescription,
			"enabled":                guide.Enabled,
			"total_occurrence_count": guide.TotalOccurrenceCount,
		}).Error; err != nil {
			return err
		}

		// 步骤按 ID diff 增删改，保留既有步骤主键：
		// nxp_sop_inquiry.step_id 弱引用步骤（无外键），整体重建会导致历史提问引用悬挂、
		// history_repair_count 统计清零；步骤 ID 稳定同时保证引用不失效。
		var existingIDs []uint
		if err := tx.Model(&model.NxpSopGuideStep{}).
			Where("guide_id = ?", guide.ID).
			Pluck("id", &existingIDs).Error; err != nil {
			return err
		}
		owned := make(map[uint]bool, len(existingIDs))
		for _, id := range existingIDs {
			owned[id] = true
		}

		keep := make(map[uint]bool, len(guide.Steps))
		toCreate := make([]model.NxpSopGuideStep, 0, len(guide.Steps))
		for i := range guide.Steps {
			s := guide.Steps[i]
			if s.ID != 0 && owned[s.ID] {
				keep[s.ID] = true
				upd := s
				upd.GuideID = guide.ID
				upd.ID = 0 // ID 仅作定位条件（Where），不出现在 SET 列中
				// 结构体 + Select 更新：map 写法会绕过 serializer:json，无法正确写入切片字段；
				// 不选取 history_repair_count，服务端统计值不被前端载荷覆盖
				if err := tx.Model(&model.NxpSopGuideStep{}).
					Where("id = ?", s.ID).
					Select(
						"guide_id", "number", "stage", "title", "description", "instruction",
						"image_urls", "video_urls", "pdf_urls",
						"judgment_method", "expert_advice", "safety_warning",
						"enabled",
					).
					Updates(&upd).Error; err != nil {
					return err
				}
				continue
			}
			// ID 为 0 或不属于本指南（如前端本地临时 id）：视为新增，由数据库生成主键
			s.ID = 0
			s.GuideID = guide.ID
			toCreate = append(toCreate, s)
		}
		if len(toCreate) > 0 {
			if err := tx.CreateInBatches(toCreate, 100).Error; err != nil {
				return err
			}
		}

		toDelete := make([]uint, 0, len(existingIDs))
		for _, id := range existingIDs {
			if !keep[id] {
				toDelete = append(toDelete, id)
			}
		}
		if len(toDelete) > 0 {
			if err := tx.Where("id IN ?", toDelete).Delete(&model.NxpSopGuideStep{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateGuideStatus 更新指南启用状态
func UpdateGuideStatus(id uint, enabled bool) error {
	var existing model.NxpSopGuide
	if err := database.SOPDB().First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrGuideNotFound
		}
		return err
	}
	return database.SOPDB().Model(&model.NxpSopGuide{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// DeleteGuide 删除指南（先校验历史引用，再删除步骤与指南，避免弱引用悬挂）
func DeleteGuide(id uint) error {
	return database.SOPDB().Transaction(func(tx *gorm.DB) error {
		// guide_id 为弱引用（无外键）：历史提问与维修记录仍指向该指南，
		// 直接删除会让记录详情页查不到 SOP，故与执行说明选项一致做引用校验。
		var refCount int64
		if err := tx.Model(&model.NxpSopInquiry{}).
			Where("guide_id = ?", id).Count(&refCount).Error; err != nil {
			return err
		}
		if refCount > 0 {
			return bizerr.ErrGuideInUse
		}
		if err := tx.Model(&model.NxpSopRepairRecord{}).
			Where("guide_id = ?", id).Count(&refCount).Error; err != nil {
			return err
		}
		if refCount > 0 {
			return bizerr.ErrGuideInUse
		}

		// 先删除子表步骤，再删除指南，绕开 fk_guide_steps 外键约束
		if err := tx.Where("guide_id = ?", id).Delete(&model.NxpSopGuideStep{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.NxpSopGuide{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return bizerr.ErrGuideNotFound
		}
		return nil
	})
}
