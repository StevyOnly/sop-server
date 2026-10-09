package service

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"sop/internal/database"
	"sop/internal/model"
	bizerr "sop/pkg/errors"
)

// MediaQuery 媒体资源列表查询条件
type MediaQuery struct {
	Page     int
	PageSize int
	Keyword  string // 文件名或标签，模糊搜索
	FileType string // 文件类型，精确搜索
}

// GetMediaList 分页查询媒体资源列表
func GetMediaList(q MediaQuery) ([]model.NxpSopMedia, int64, error) {
	db := database.SOPDB().Model(&model.NxpSopMedia{})

	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		// 同一关键字同时匹配文件名或标签。
		// tags 为 JSON 数组字符串，不能用 LIKE 整体匹配（搜 "tag" 会误命中 "tag2"），
		// 改用 OPENJSON 对数组元素逐个精确 LIKE；ISJSON+CASE 兜底 NULL/空串/非法 JSON，避免 OPENJSON 报错。
		db = db.Where("file_name LIKE ? OR EXISTS (SELECT 1 FROM OPENJSON(CASE WHEN ISJSON(nxp_sop_media.tags) = 1 THEN nxp_sop_media.tags ELSE '[]' END) WHERE [value] LIKE ?)", like, like)
	}
	if q.FileType != "" {
		db = db.Where("file_type = ?", q.FileType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.NxpSopMedia
	if err := db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	if list == nil {
		list = []model.NxpSopMedia{}
	}
	return list, total, nil
}

// CreateMedia 创建媒体资源
func CreateMedia(media *model.NxpSopMedia) error {
	return database.SOPDB().Create(media).Error
}

// UpdateMedia 更新媒体资源信息
func UpdateMedia(media *model.NxpSopMedia) error {
	var existing model.NxpSopMedia
	if err := database.SOPDB().First(&existing, media.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ErrMediaNotFound
		}
		return err
	}

	// tags 为 []string，使用 map 更新时 GORM 不识别 serializer:json 标签，
	// 需手动序列化为 JSON 字符串，避免生成非法 SQL
	if media.Tags == nil {
		media.Tags = []string{}
	}
	tagsJSON, err := json.Marshal(media.Tags)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"file_name":   media.FileName,
		"description": media.Description,
		"tags":        string(tagsJSON),
		"file_type":   media.FileType,
		"file_path":   media.FilePath,
		"size":        media.Size,
	}
	// 前端编辑时不回传 fileHash，仅在非空时更新，避免覆盖已有哈希值
	if media.FileHash != "" {
		updates["file_hash"] = media.FileHash
	}
	return database.SOPDB().Model(&model.NxpSopMedia{}).Where("id = ?", media.ID).Updates(updates).Error
}

// DeleteMedia 删除媒体资源
func DeleteMedia(id uint) error {
	result := database.SOPDB().Delete(&model.NxpSopMedia{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizerr.ErrMediaNotFound
	}
	return nil
}
