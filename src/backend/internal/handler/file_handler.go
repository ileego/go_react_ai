package handler

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ileego/go_react_ai/internal/domain"
	"github.com/ileego/go_react_ai/internal/middleware"
	"github.com/ileego/go_react_ai/internal/service"
	"github.com/ileego/go_react_ai/pkg/response"
)

// FileHandler 文件相关 HTTP 接口。
type FileHandler struct {
	svc service.FileService
}

// NewFileHandler 创建 FileHandler。
func NewFileHandler(svc service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// Upload 上传文件。
// POST /api/files
func (h *FileHandler) Upload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}

	// 限制大小在 Handler 层先做一层保护
	if fileHeader.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小超过 10MB 限制")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		response.FromError(c, fmt.Errorf("打开文件失败: %w", err))
		return
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		response.FromError(c, fmt.Errorf("读取文件失败: %w", err))
		return
	}

	userID := middleware.GetUserID(c)
	file, err := h.svc.Upload(c.Request.Context(), userID, fileHeader.Filename, fileHeader.Header.Get("Content-Type"), data)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, toFileResponse(file))
}

// List 获取文件列表。
// GET /api/files?page=1&page_size=20
func (h *FileHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID := middleware.GetUserID(c)
	files, total, err := h.svc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.List(c, files, total, page, pageSize)
}

// Get 获取文件元数据。
// GET /api/files/:id
func (h *FileHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	file, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, toFileResponse(file))
}

// Download 获取文件下载 URL。
// GET /api/files/:id/download
func (h *FileHandler) Download(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	url, err := h.svc.GetDownloadURL(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, gin.H{"download_url": url, "expires_in": int64(15 * time.Minute / time.Second)})
}

// Delete 删除文件。
// DELETE /api/files/:id
func (h *FileHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c)
}

// PresignedUpload 获取客户端直传的预签名上传 URL。
// GET /api/files/presigned-upload?name=xxx&content_type=xxx
func (h *FileHandler) PresignedUpload(c *gin.Context) {
	name := c.Query("name")
	contentType := c.Query("content_type")
	if name == "" || contentType == "" {
		response.BadRequest(c, "name 和 content_type 不能为空")
		return
	}

	userID := middleware.GetUserID(c)
	url, file, err := h.svc.PresignedUploadURL(c.Request.Context(), userID, name, contentType)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Data(c, gin.H{
		"upload_url": url,
		"file":       toFileResponse(file),
		"expires_in": int64(15 * time.Minute / time.Second),
	})
}

// FileResponse 文件元数据响应。
type FileResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Bucket      string `json:"bucket"`
	CreatedAt   string `json:"created_at"`
}

func toFileResponse(file *domain.File) FileResponse {
	return FileResponse{
		ID:          file.ID,
		Name:        file.Name,
		ContentType: file.ContentType,
		Size:        file.Size,
		Bucket:      file.Bucket,
		CreatedAt:   file.CreatedAt.Format(time.RFC3339),
	}
}
