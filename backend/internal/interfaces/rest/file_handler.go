package rest

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/backend/internal/application/services"
	"github.com/nexuscrm/shared/pkg/constants"
)

type FileHandler struct {
	svcMgr *services.ServiceManager
}

func NewFileHandler(svcMgr *services.ServiceManager) *FileHandler {
	return &FileHandler{svcMgr: svcMgr}
}

// Upload handles file uploads
func (h *FileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Create uploads directory if not exists
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}
	}

	// Generate safe filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), "file", ext)
	path := filepath.Join(uploadDir, filename)

	// Save
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return path (relative)
	c.JSON(http.StatusOK, gin.H{
		"path":                           path,
		constants.FieldName:              file.Filename,
		constants.FieldSysFile_SizeBytes: file.Size,
		constants.FieldSysFile_MimeType:  file.Header.Get("Content-Type"),
	})
}
