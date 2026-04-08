package handler

import (
	"ai-gateway/internal/service"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计处理器
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler 创建审计处理器
func NewAuditHandler() *AuditHandler {
	return &AuditHandler{
		auditService: service.NewAuditService(),
	}
}

// QueryAuditLogs 查询审计日志
func (h *AuditHandler) QueryAuditLogs(c *gin.Context) {
	var req service.QueryAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.auditService.QueryAuditLogs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetRiskEvents 获取风险事件
func (h *AuditHandler) GetRiskEvents(c *gin.Context) {
	var req service.GetRiskEventsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	events, total, err := h.auditService.GetRiskEvents(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      events,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// GetUserStatistics 获取用户统计
func (h *AuditHandler) GetUserStatistics(c *gin.Context) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var userID uint64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	stats, err := h.auditService.GetUserStatistics(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// GetDashboardStats 获取仪表盘统计
func (h *AuditHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.auditService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ResolveRiskEvent 解决风险事件
func (h *AuditHandler) ResolveRiskEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "event_id is required"})
		return
	}

	var req struct {
		Note string `json:"note,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户
	username, _ := c.Get("username")
	resolvedBy := ""
	if username != nil {
		resolvedBy = username.(string)
	}

	if err := h.auditService.ResolveRiskEvent(eventID, resolvedBy, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "risk event resolved successfully"})
}
