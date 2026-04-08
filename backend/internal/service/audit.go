package service

import (
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"time"
)

// AuditService 审计服务
type AuditService struct{}

// NewAuditService 创建审计服务
func NewAuditService() *AuditService {
	return &AuditService{}
}

// QueryAuditLogsRequest 查询审计日志请求
type QueryAuditLogsRequest struct {
	StartTime  *time.Time `form:"start_time"`
	EndTime    *time.Time `form:"end_time"`
	UserID     *uint64    `form:"user_id"`
	ModelName  string     `form:"model_name"`
	RiskLevel  string     `form:"risk_level"`
	RequestIP  string     `form:"request_ip"`
	Page       int        `form:"page"`
	PageSize   int        `form:"page_size"`
}

// QueryAuditLogs 查询审计日志
func (s *AuditService) QueryAuditLogs(req *QueryAuditLogsRequest) (*model.AuditQueryResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	
	queryReq := &model.AuditQueryRequest{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		UserID:    req.UserID,
		ModelName: req.ModelName,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
	
	return repository.QueryAuditLogs(queryReq)
}

// GetRiskEventsRequest 获取风险事件请求
type GetRiskEventsRequest struct {
	StartTime  string `form:"start_time"`
	EndTime    string `form:"end_time"`
	RiskLevel  string `form:"risk_level"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

// GetRiskEvents 获取风险事件
func (s *AuditService) GetRiskEvents(req *GetRiskEventsRequest) ([]model.RiskEvent, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}
	
	// 解析时间
	startTime, _ := time.Parse("2006-01-02", req.StartTime)
	if startTime.IsZero() {
		startTime = time.Now().AddDate(0, 0, -7)
	}
	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())
	
	endTime, _ := time.Parse("2006-01-02", req.EndTime)
	if endTime.IsZero() {
		endTime = time.Now()
	}
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.Location())
	
	return repository.GetRiskEvents(startTime, endTime, req.RiskLevel, req.Page, req.PageSize)
}

// GetUserStatistics 获取用户统计
func (s *AuditService) GetUserStatistics(userID uint64, startDate, endDate string) ([]model.UserBehaviorSummary, error) {
	return repository.GetUserStatistics(userID, startDate, endDate)
}

// GetDashboardStats 获取仪表盘统计
func (s *AuditService) GetDashboardStats() (map[string]interface{}, error) {
	ctx := repository.GetClickHouse().Context()
	
	stats := make(map[string]interface{})
	
	// 今日请求数
	var todayRequests uint64
	err := repository.GetClickHouse().QueryRow(ctx, `
		SELECT count() FROM audit_logs 
		WHERE toDate(timestamp) = today()
	`).Scan(&todayRequests)
	if err == nil {
		stats["today_requests"] = todayRequests
	}
	
	// 今日token使用量
	var todayTokens uint64
	err = repository.GetClickHouse().QueryRow(ctx, `
		SELECT sum(total_tokens) FROM audit_logs 
		WHERE toDate(timestamp) = today()
	`).Scan(&todayTokens)
	if err == nil {
		stats["today_tokens"] = todayTokens
	}
	
	// 活跃用户数（今日）
	var activeUsers uint64
	err = repository.GetClickHouse().QueryRow(ctx, `
		SELECT uniqExact(user_id) FROM audit_logs 
		WHERE toDate(timestamp) = today()
	`).Scan(&activeUsers)
	if err == nil {
		stats["active_users"] = activeUsers
	}
	
	// 风险事件数（今日）
	var riskEvents uint64
	err = repository.GetClickHouse().QueryRow(ctx, `
		SELECT count() FROM risk_events 
		WHERE toDate(timestamp) = today()
	`).Scan(&riskEvents)
	if err == nil {
		stats["risk_events"] = riskEvents
	}
	
	// 最近7天趋势
	rows, err := repository.GetClickHouse().Query(ctx, `
		SELECT 
			toDate(timestamp) as date,
			count() as requests,
			sum(total_tokens) as tokens
		FROM audit_logs 
		WHERE timestamp >= now() - INTERVAL 7 DAY
		GROUP BY date
		ORDER BY date
	`)
	if err == nil {
		defer rows.Close()
		var trends []map[string]interface{}
		for rows.Next() {
			var date time.Time
			var requests, tokens uint64
			if err := rows.Scan(&date, &requests, &tokens); err == nil {
				trends = append(trends, map[string]interface{}{
					"date":     date.Format("2006-01-02"),
					"requests": requests,
					"tokens":   tokens,
				})
			}
		}
		stats["trends"] = trends
	}
	
	// 模型使用排行
	rows, err = repository.GetClickHouse().Query(ctx, `
		SELECT 
			model_name,
			count() as requests,
			sum(total_tokens) as tokens
		FROM audit_logs 
		WHERE toDate(timestamp) = today()
		GROUP BY model_name
		ORDER BY requests DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		var modelStats []map[string]interface{}
		for rows.Next() {
			var modelName string
			var requests, tokens uint64
			if err := rows.Scan(&modelName, &requests, &tokens); err == nil {
				modelStats = append(modelStats, map[string]interface{}{
					"model_name": modelName,
					"requests":   requests,
					"tokens":     tokens,
				})
			}
		}
		stats["model_stats"] = modelStats
	}
	
	return stats, nil
}

// ResolveRiskEvent 解决风险事件
func (s *AuditService) ResolveRiskEvent(eventID string, resolvedBy, note string) error {
	ctx := repository.GetClickHouse().Context()
	return repository.GetClickHouse().Exec(ctx, `
		ALTER TABLE risk_events 
		UPDATE is_resolved = 1, resolved_by = ?, resolved_at = now(), note = ?
		WHERE event_id = ?
	`, resolvedBy, note, eventID)
}
