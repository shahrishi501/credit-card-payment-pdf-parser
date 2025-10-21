package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shahrishi501/creditcardpaymentpdfparser/models"
	"github.com/shahrishi501/creditcardpaymentpdfparser/utils"
)

func StartServer() {
	r := gin.Default()
	r.POST("/api/parse-pdf", parsePDFHandler)
	port := "8080"
	r.Run(":" + port)
}


func parsePDFHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<20) // 32 MB

	password := c.PostForm("password")
	if password == "" {
		c.JSON(http.StatusBadRequest, models.ParseResponse{
			Success: false,
			Error:   "Password is required",
		})
		return
	}

	// Retrieve file
	file, err := c.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ParseResponse{
			Success: false,
			Error:   "PDF file is required",
		})
		return
	}

	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, models.ParseResponse{
			Success: false,
			Error:   "Only PDF files are allowed",
		})
		return
	}

	// Save file temporarily
	tempDir := "./tmp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ParseResponse{
			Success: false,
			Error:   "Failed to create tmp directory",
		})
		return
	}
	safeName := filepath.Base(file.Filename)
	temp, err := os.CreateTemp(tempDir, fmt.Sprintf("upload_*_%s", safeName))

	if err != nil {
        // log actual error and return
        fmt.Printf("failed to create temp file: %v\n", err)
        c.JSON(http.StatusInternalServerError, models.ParseResponse{
            Success: false,
            Error:   "Failed to create temporary file",
        })
        return
    }
	
	tempPath := temp.Name()
	temp.Close()

	 if err := c.SaveUploadedFile(file, tempPath); err != nil {
        // log the real error to server output for debugging
        fmt.Printf("failed to save uploaded file to %s: %v\n", tempPath, err)
        c.JSON(http.StatusInternalServerError, models.ParseResponse{
            Success: false,
            Error:   fmt.Sprintf("Failed to save uploaded file: %v", err),
        })
        return
    }
    defer os.Remove(tempPath) // Clean up

	// Process PDF
	result, err := utils.ProcessPDF(tempPath, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to process PDF: %v", err),
		})
		return
	}

	c.Header("Content-Type", "application/json")

	// 🧹 Clean Gemini output (remove ```json fences)
	clean := strings.TrimSpace(result)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```JSON")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	// Try to decode
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		// If parsing fails, return raw text
		c.JSON(http.StatusOK, gin.H{"success": true, "data": clean})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    parsed,
	})
}