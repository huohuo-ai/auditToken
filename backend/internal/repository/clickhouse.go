package repository

import (
	"ai-gateway/internal/config"
	"ai-gateway/internal/model"
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ClickHouseConn 全局ClickHouse连接
var ClickHouseConn driver.Conn

// InitClickHouse 初始化ClickHouse连接
func InitClickHouse(cfg *config.ClickHouseConfig) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	// 测试连接
	ctx := context.Background()
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	ClickHouseConn = conn
	return conn, nil
}

// GetClickHouse 获取ClickHouse连接
func GetClickHouse() driver.Conn {
	if ClickHouseConn == nil {
		panic("clickhouse not initialized")
	}
	return ClickHouseConn
}

// CreateTables 创建ClickHouse表
func CreateClickHouseTables() error {
	ctx := context.Background()
	
	// 创建审计日志表
	auditLogTable := `
		CREATE TABLE IF NOT EXISTS audit_logs (
			timestamp DateTime64(3),
			request_id String,
			user_id UInt64,
			user_name String,
			user_email String,
			request_time DateTime64(3),
			request_method String,
			request_path String,
			request_ip String,
			user_agent String,
			request_headers String,
			request_body String,
			model_name String,
			model_provider String,
			response_time DateTime64(3),
			response_status Int32,
			response_body String,
			response_headers String,
			prompt_tokens Int64,
			completion_tokens Int64,
			total_tokens Int64,
			latency_ms Int64,
			is_stream Bool,
			has_error Bool,
			error_message String
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMMDD(timestamp)
		ORDER BY (timestamp, user_id, request_id)
		TTL timestamp + INTERVAL 6 MONTH
		SETTINGS index_granularity = 8192
	`

	if err := ClickHouseConn.Exec(ctx, auditLogTable); err != nil {
		return fmt.Errorf("failed to create audit_logs table: %w", err)
	}

	// 创建风险事件表
	riskEventTable := `
		CREATE TABLE IF NOT EXISTS risk_events (
			timestamp DateTime64(3),
			event_id String,
			request_id String,
			user_id UInt64,
			user_name String,
			risk_level String,
			risk_type String,
			risk_score Float64,
			risk_reason String,
			description String,
			evidence String,
			request_ip String,
			model_name String,
			is_resolved Bool,
			resolved_by String,
			resolved_at DateTime64(3),
			note String
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMMDD(timestamp)
		ORDER BY (timestamp, user_id, risk_level)
		TTL timestamp + INTERVAL 1 YEAR
		SETTINGS index_granularity = 8192
	`

	if err := ClickHouseConn.Exec(ctx, riskEventTable); err != nil {
		return fmt.Errorf("failed to create risk_events table: %w", err)
	}

	// 创建用户行为汇总表
	behaviorSummaryTable := `
		CREATE TABLE IF NOT EXISTS user_behavior_summaries (
			date Date,
			user_id UInt64,
			user_name String,
			total_requests UInt64,
			total_tokens UInt64,
			avg_latency_ms Float64,
			max_latency_ms UInt64,
			off_hours_requests UInt64,
			peak_hour_requests UInt64,
			high_risk_requests UInt64,
			risk_events UInt64,
			models_used String
		) ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (date, user_id)
		TTL date + INTERVAL 2 YEAR
		SETTINGS index_granularity = 8192
	`

	if err := ClickHouseConn.Exec(ctx, behaviorSummaryTable); err != nil {
		return fmt.Errorf("failed to create user_behavior_summaries table: %w", err)
	}

	return nil
}

// InsertAuditLog 插入审计日志
func InsertAuditLog(log *model.AuditLog) error {
	ctx := context.Background()
	
	query := `
		INSERT INTO audit_logs (
			timestamp, request_id, user_id, user_name, user_email,
			request_time, request_method, request_path, request_ip, user_agent,
			request_headers, request_body, model_name, model_provider,
			response_time, response_status, response_body, response_headers,
			prompt_tokens, completion_tokens, total_tokens, latency_ms,
			is_stream, has_error, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	err := ClickHouseConn.Exec(ctx, query,
		log.Timestamp, log.RequestID, log.UserID, log.UserName, log.UserEmail,
		log.RequestTime, log.RequestMethod, log.RequestPath, log.RequestIP, log.UserAgent,
		log.RequestHeaders, log.RequestBody, log.ModelName, log.ModelProvider,
		log.ResponseTime, log.ResponseStatus, log.ResponseBody, log.ResponseHeaders,
		log.PromptTokens, log.CompletionTokens, log.TotalTokens, log.LatencyMs,
		log.IsStream, log.HasError, log.ErrorMessage,
	)
	
	return err
}

// InsertRiskEvent 插入风险事件
func InsertRiskEvent(event *model.RiskEvent) error {
	ctx := context.Background()
	
	query := `
		INSERT INTO risk_events (
			timestamp, event_id, request_id, user_id, user_name,
			risk_level, risk_type, risk_score, risk_reason, description,
			evidence, request_ip, model_name, is_resolved, resolved_by,
			resolved_at, note
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	err := ClickHouseConn.Exec(ctx, query,
		event.Timestamp, event.EventID, event.RequestID, event.UserID, event.UserName,
		event.RiskLevel, event.RiskType, event.RiskScore, event.RiskReason, event.Description,
		event.Evidence, event.RequestIP, event.ModelName, event.IsResolved, event.ResolvedBy,
		event.ResolvedAt, event.Note,
	)
	
	return err
}

// QueryAuditLogs 查询审计日志
func QueryAuditLogs(req *model.AuditQueryRequest) (*model.AuditQueryResponse, error) {
	ctx := context.Background()
	
	// 构建查询条件
	whereClause := "1=1"
	var args []interface{}
	
	if req.StartTime != nil {
		whereClause += " AND timestamp >= ?"
		args = append(args, *req.StartTime)
	}
	if req.EndTime != nil {
		whereClause += " AND timestamp <= ?"
		args = append(args, *req.EndTime)
	}
	if req.UserID != nil {
		whereClause += " AND user_id = ?"
		args = append(args, *req.UserID)
	}
	if req.ModelName != "" {
		whereClause += " AND model_name = ?"
		args = append(args, req.ModelName)
	}
	if req.RequestIP != "" {
		whereClause += " AND request_ip = ?"
		args = append(args, req.RequestIP)
	}
	
	// 分页
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	
	// 查询总数
	countQuery := fmt.Sprintf("SELECT count() FROM audit_logs WHERE %s", whereClause)
	var total int64
	if err := ClickHouseConn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}
	
	// 查询数据
	query := fmt.Sprintf(`
		SELECT * FROM audit_logs 
		WHERE %s 
		ORDER BY timestamp DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, pageSize, (page-1)*pageSize)
	
	rows, err := ClickHouseConn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var logs []model.AuditLog
	for rows.Next() {
		var log model.AuditLog
		err := rows.Scan(
			&log.Timestamp, &log.RequestID, &log.UserID, &log.UserName, &log.UserEmail,
			&log.RequestTime, &log.RequestMethod, &log.RequestPath, &log.RequestIP, &log.UserAgent,
			&log.RequestHeaders, &log.RequestBody, &log.ModelName, &log.ModelProvider,
			&log.ResponseTime, &log.ResponseStatus, &log.ResponseBody, &log.ResponseHeaders,
			&log.PromptTokens, &log.CompletionTokens, &log.TotalTokens, &log.LatencyMs,
			&log.IsStream, &log.HasError, &log.ErrorMessage,
		)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	
	return &model.AuditQueryResponse{
		Data:     logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetRiskEvents 获取风险事件
func GetRiskEvents(startTime, endTime time.Time, riskLevel string, page, pageSize int) ([]model.RiskEvent, int64, error) {
	ctx := context.Background()
	
	whereClause := "timestamp >= ? AND timestamp <= ?"
	args := []interface{}{startTime, endTime}
	
	if riskLevel != "" {
		whereClause += " AND risk_level = ?"
		args = append(args, riskLevel)
	}
	
	// 查询总数
	countQuery := fmt.Sprintf("SELECT count() FROM risk_events WHERE %s", whereClause)
	var total int64
	if err := ClickHouseConn.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	query := fmt.Sprintf(`
		SELECT * FROM risk_events 
		WHERE %s 
		ORDER BY timestamp DESC 
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, pageSize, (page-1)*pageSize)
	
	rows, err := ClickHouseConn.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var events []model.RiskEvent
	for rows.Next() {
		var event model.RiskEvent
		err := rows.Scan(
			&event.Timestamp, &event.EventID, &event.RequestID, &event.UserID, &event.UserName,
			&event.RiskLevel, &event.RiskType, &event.RiskScore, &event.RiskReason, &event.Description,
			&event.Evidence, &event.RequestIP, &event.ModelName, &event.IsResolved, &event.ResolvedBy,
			&event.ResolvedAt, &event.Note,
		)
		if err != nil {
			continue
		}
		events = append(events, event)
	}
	
	return events, total, nil
}

// GetUserStatistics 获取用户统计
func GetUserStatistics(userID uint64, startDate, endDate string) ([]model.UserBehaviorSummary, error) {
	ctx := context.Background()
	
	query := `
		SELECT * FROM user_behavior_summaries 
		WHERE user_id = ? AND date >= ? AND date <= ?
		ORDER BY date DESC
	`
	
	rows, err := ClickHouseConn.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var summaries []model.UserBehaviorSummary
	for rows.Next() {
		var s model.UserBehaviorSummary
		err := rows.Scan(
			&s.Date, &s.UserID, &s.UserName, &s.TotalRequests, &s.TotalTokens,
			&s.AvgLatencyMs, &s.MaxLatencyMs, &s.OffHoursRequests, &s.PeakHourRequests,
			&s.HighRiskRequests, &s.RiskEvents, &s.ModelsUsed,
		)
		if err != nil {
			continue
		}
		summaries = append(summaries, s)
	}
	
	return summaries, nil
}
