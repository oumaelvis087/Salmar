package handlers

import (
	"net/http"

	"github.com/elvisouma/salmar-ai/internal/orchestrator"
	"github.com/elvisouma/salmar-ai/pkg/models"
	"github.com/gin-gonic/gin"
)

// Handler contains dependencies for API handlers
type Handler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewHandler creates a new handler instance
func NewHandler(orch *orchestrator.Orchestrator) *Handler {
	return &Handler{
		orchestrator: orch,
	}
}

// RegisterRoutes sets up all API routes
func RegisterRoutes(router *gin.Engine, orch *orchestrator.Orchestrator) {
	handler := NewHandler(orch)

	// API version group
	v1 := router.Group("/api/v1")
	{
		// Health check endpoint
		v1.GET("/health", handler.HealthCheck)
		
		// Main processing endpoints
		v1.POST("/process", handler.ProcessRequest)
		v1.POST("/text", handler.ProcessText)
		v1.POST("/image", handler.ProcessImage)
		v1.POST("/code", handler.ProcessCode)
		
		// Specialized endpoints
		v1.POST("/chat", handler.Chat)
		v1.POST("/generate-image", handler.GenerateImage)
		v1.POST("/analyze-image", handler.AnalyzeImage)
		v1.POST("/generate-code", handler.GenerateCode)
		v1.POST("/explain-code", handler.ExplainCode)
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// ProcessRequest handles multimodal requests
func (h *Handler) ProcessRequest(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ProcessText handles text-only requests
func (h *Handler) ProcessText(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ProcessImage handles image-only requests
func (h *Handler) ProcessImage(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ProcessCode handles code-only requests
func (h *Handler) ProcessCode(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Chat handles conversational requests
func (h *Handler) Chat(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateImage handles image generation requests
func (h *Handler) GenerateImage(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AnalyzeImage handles image analysis requests
func (h *Handler) AnalyzeImage(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GenerateCode handles code generation requests
func (h *Handler) GenerateCode(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ExplainCode handles code explanation requests
func (h *Handler) ExplainCode(c *gin.Context) {
	var req models.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.orchestrator.ProcessRequest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}