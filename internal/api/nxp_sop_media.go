package api

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sop/global"
	"sop/internal/middleware"
	"sop/internal/model"
	"sop/internal/service"
	"sop/pkg/response"
)

// MediaApi 媒体资源接口处理器
type MediaApi struct{}

// MediaBaseFields 媒体资源创建/更新共用的业务字段
// 独立 DTO 而非 model.X 类型别名：不暴露 createTime/updateTime 等内部字段，
// 也不接受前端伪造上传者身份（Uploader 一律取自登录态）。
type MediaBaseFields struct {
	FileName    string   `json:"fileName"`
	FileType    string   `json:"fileType"`
	FilePath    string   `json:"filePath"`
	Description string   `json:"description"`
	Size        string   `json:"size"`
	Tags        []string `json:"tags"`
	FileHash    string   `json:"fileHash"`
}

// CreateMediaRequest 创建媒体资源请求
type CreateMediaRequest struct {
	MediaBaseFields
}

// UpdateMediaRequest 更新媒体资源请求（必须带待更新记录的 id）
type UpdateMediaRequest struct {
	MediaBaseFields
	ID uint `json:"id" binding:"required"`
}

// GetMediaList 分页获取媒体资源列表
func (a MediaApi) GetMediaList(c *gin.Context) {
	var req ListMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}
	response.NormalizePage(&req.Page, &req.PageSize, 100)

	list, total, err := service.GetMediaList(service.MediaQuery{
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
		FileType: req.FileType,
	})
	if err != nil {
		response.FailError(c, err)
		return
	}
	response.OKWithPage(c, list, total, req.Page, req.PageSize)
}

// ListMediaRequest 媒体资源列表请求
type ListMediaRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`  // 文件名或标签，模糊搜索
	FileType string `json:"fileType"` // 文件类型，精确搜索
}

// UploadMedia 上传文件（multipart），计算 MD5 并与前端提交的 fileHash 比对后保存
func (a MediaApi) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "未获取到上传文件")
		return
	}

	// 读取前端上传时携带的 fileHash 字段（用于完整性校验）
	clientHash := c.PostForm("fileHash")

	// 必须校验：前端未传或传空的 hash 值直接拒绝
	if clientHash == "" {
		response.Fail(c, "缺少文件哈希值 fileHash，请重新上传")
		return
	}

	// 创建上传目录（媒体资源目录），未配置时回退到根目录
	// 配置可能随时被热重载替换，取一次快照保证本请求内使用同一路径配置
	srvCfg := global.GetConfig().Server
	uploadDir := srvCfg.UploadMediaDir
	if uploadDir == "" {
		uploadDir = srvCfg.UploadDir
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Fail(c, "创建上传目录失败")
		return
	}

	// 计算文件 MD5
	src, err := file.Open()
	if err != nil {
		response.Fail(c, "读取文件失败")
		return
	}
	defer src.Close()

	h := md5.New()
	if _, err := io.Copy(h, src); err != nil {
		response.Fail(c, "计算文件哈希失败")
		return
	}
	serverHash := fmt.Sprintf("%x", h.Sum(nil))

	// 校验：前端 hash 与后端计算的一致才允许上传
	if !strings.EqualFold(clientHash, serverHash) {
		response.Fail(c, "文件哈希值校验失败，文件可能已损坏，请重新上传")
		return
	}

	// 安全保存文件名
	origName := filepath.Base(file.Filename)
	ext := filepath.Ext(origName)
	base := strings.TrimSuffix(origName, ext)

	// 用毫秒级时间戳命名（纯数字无分隔符），避免同名文件互相覆盖
	ts := time.Now().Format("20060102150405000")
	saveName := fmt.Sprintf("%s-%s%s", base, ts, ext)
	savePath := filepath.Join(uploadDir, saveName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Fail(c, "保存文件失败: "+err.Error())
		return
	}

	response.Result(c, gin.H{
		"hash": serverHash,
		"path": "/" + strings.ReplaceAll(savePath, "\\", "/"),
	}, "上传成功")
}

// UploadScenePhotoRequest 现场照片 base64 上传请求
type UploadScenePhotoRequest struct {
	// File 图片 base64 字符串，支持 data URI（data:image/png;base64,xxx）或纯 base64
	File     string `json:"file" binding:"required"`
	FileHash string `json:"fileHash"`
	// Extension 前端显式声明的图片扩展名（可选，优先使用；格式如 jpg/png）
	Extension string `json:"extension"`
}

// UploadScenePhoto 上传现场照片（JSON base64），保存到 uploadSceneDir 目录
func (a MediaApi) UploadScenePhoto(c *gin.Context) {
	var req UploadScenePhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	// 必须校验：前端未传或传空的 hash 值直接拒绝
	clientHash := req.FileHash
	if clientHash == "" {
		response.Fail(c, "缺少文件哈希值 fileHash，请重新上传")
		return
	}

	// 去掉 data URI 前缀（data:image/png;base64, 等），并解析 base64 字节
	b64 := req.File
	if idx := strings.Index(b64, ","); idx >= 0 && strings.Contains(b64, ";base64") {
		b64 = b64[idx+1:]
	}
	// 防超大 base64 一次性进内存：
	// 上限从配置 maxScenePhotoSizeMB 读取（MB），默认 20MB；base64 字符串约为原图的 4/3 倍，
	// 因此解码前先按 len(b64) 截断，解码后再按原字节数复核，双保险。
	maxScenePhotoSizeMB := global.GetConfig().Server.MaxScenePhotoSizeMB
	if maxScenePhotoSizeMB <= 0 {
		maxScenePhotoSizeMB = 20 // 默认 20MB
	}
	maxScenePhotoBytes := int64(maxScenePhotoSizeMB) << 20
	maxBase64Len := int64(((maxScenePhotoBytes + 2) / 3) * 4)
	if n := int64(len(b64)); n > maxBase64Len {
		response.Fail(c, fmt.Sprintf("图片大小超出上限（最大 %dMB）", maxScenePhotoSizeMB))
		return
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		response.Fail(c, "base64 解码失败，文件格式不正确")
		return
	}
	if len(data) == 0 {
		response.Fail(c, "上传内容为空")
		return
	}
	if int64(len(data)) > maxScenePhotoBytes {
		response.Fail(c, fmt.Sprintf("图片大小超出上限（最大 %dMB）", maxScenePhotoSizeMB))
		return
	}

	// 计算文件 MD5
	h := md5.New()
	if _, err := h.Write(data); err != nil {
		response.Fail(c, "计算文件哈希失败")
		return
	}
	serverHash := fmt.Sprintf("%x", h.Sum(nil))

	// 校验：前端 hash 与后端计算的一致才允许上传
	if !strings.EqualFold(clientHash, serverHash) {
		response.Fail(c, "文件哈希值校验失败，文件可能已损坏，请重新上传")
		return
	}

	// 创建上传目录（现场照片目录），未配置时回退到根目录
	// 配置可能随时被热重载替换，取一次快照保证本请求内使用同一路径配置
	srvCfg := global.GetConfig().Server
	uploadDir := srvCfg.UploadSceneDir
	if uploadDir == "" {
		uploadDir = srvCfg.UploadDir
	}
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Fail(c, "创建上传目录失败")
		return
	}

	// 扩展名优先级：前端白名单 > 字节嗅探真实 MIME > 兜底 .png
	ext := ""
	if raw := strings.TrimPrefix(strings.ToLower(req.Extension), "."); raw != "" {
		for _, e := range []string{"jpg", "jpeg", "png", "webp", "gif", "bmp"} {
			if raw == e {
				ext = "." + e
				break
			}
		}
	}
	if ext == "" {
		// 基于解码后的真实字节嗅探 MIME（覆盖 data URI 与纯 base64 两种形式）
		if sniffed := http.DetectContentType(data); sniffed != "" {
			if exts, err := mime.ExtensionsByType(sniffed); err == nil && len(exts) > 0 {
				ext = exts[0]
			}
		}
		if ext == "" {
			ext = ".png"
		}
	}

	// 用毫秒级时间戳命名（纯数字无分隔符），避免同名文件互相覆盖
	ts := time.Now().Format("20060102150405000")
	saveName := fmt.Sprintf("scene-%s%s", ts, ext)
	savePath := filepath.Join(uploadDir, saveName)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		response.Fail(c, "保存文件失败: "+err.Error())
		return
	}

	response.Result(c, gin.H{
		"hash": serverHash,
		"path": "/" + strings.ReplaceAll(savePath, "\\", "/"),
	}, "上传成功")
}

// CreateMedia 创建媒体资源信息
// 用户启用状态校验由路由上的 middleware.RequireActiveUser() 完成
func (a MediaApi) CreateMedia(c *gin.Context) {
	var req CreateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	// 前端把 filePath 传为 '#' 占位符，映射为空
	path := req.FilePath
	if path == "#" {
		path = ""
	}

	// 上传者统一取登录态用户名（RequireActiveUser 已反查），拒绝前端伪造
	uploader := ""
	if current := middleware.GetActiveUser(c); current != nil && current.User.Name != "" {
		uploader = current.User.Name
	}

	media := model.NxpSopMedia{
		FileName:    req.FileName,
		FileType:    req.FileType,
		FilePath:    path,
		Description: req.Description,
		Size:        req.Size,
		Tags:        req.Tags,
		Uploader:    uploader,
		FileHash:    req.FileHash,
	}
	if err := service.CreateMedia(&media); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// UpdateMedia 更新媒体资源信息
func (a MediaApi) UpdateMedia(c *gin.Context) {
	var req UpdateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	path := req.FilePath
	if path == "#" {
		path = ""
	}

	media := model.NxpSopMedia{
		FileName:    req.FileName,
		FileType:    req.FileType,
		FilePath:    path,
		Description: req.Description,
		Size:        req.Size,
		Tags:        req.Tags,
	}
	media.ID = req.ID
	if err := service.UpdateMedia(&media); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}

// DeleteMedia 删除媒体资源
func (a MediaApi) DeleteMedia(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		bindFail(c, err)
		return
	}

	if err := service.DeleteMedia(req.ID); err != nil {
		response.FailError(c, err)
		return
	}
	response.OK(c)
}
