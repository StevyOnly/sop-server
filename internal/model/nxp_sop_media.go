package model

// NxpSopMedia 媒体资源
type NxpSopMedia struct {
	BaseModel
	FileName    string   `json:"fileName" gorm:"column:file_name;index;comment:文件名"`
	FileType    string   `json:"fileType" gorm:"column:file_type;comment:文件类型"`
	FilePath    string   `json:"filePath" gorm:"column:file_path;comment:文件路径"`
	Description string   `json:"description" gorm:"column:description;comment:文件说明"`
	Size        string   `json:"size" gorm:"column:size;comment:文件大小"`
	Tags        []string `json:"tags" gorm:"column:tags;serializer:json;comment:标签"`
	Uploader    string   `json:"uploader" gorm:"column:uploader;comment:上传者"`
	FileHash    string   `json:"fileHash" gorm:"column:file_hash;comment:文件MD5哈希"`
}

// TableName 返回媒体资源表名
func (NxpSopMedia) TableName() string {
	return "nxp_sop_media"
}
