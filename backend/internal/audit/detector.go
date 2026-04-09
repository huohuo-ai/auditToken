package audit

import (
	"ai-gateway/internal/config"
	"ai-gateway/internal/model"
	"ai-gateway/internal/repository"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RiskDetector 风险检测器
type RiskDetector struct {
	cfg      *config.AuditConfig
	patterns []model.PromptPattern
}

// NewRiskDetector 创建新的风险检测器
func NewRiskDetector() *RiskDetector {
	cfg := config.GetConfig().Audit
	
	// 从数据库加载检测模式
	db := repository.GetDB()
	var patterns []model.PromptPattern
	db.Where("is_enabled = ?", true).Find(&patterns)
	
	return &RiskDetector{
		cfg:      &cfg,
		patterns: patterns,
	}
}

// DetectRisk 检测风险
func (d *RiskDetector) DetectRisk(auditLog *model.AuditLog) *model.RiskEvent {
	var riskEvents []*model.RiskEvent
	
	// 1. Token滥用检测
	if event := d.detectTokenAbuse(auditLog); event != nil {
		riskEvents = append(riskEvents, event)
	}
	
	// 2. 非工作时间访问检测
	if event := d.detectOffHoursAccess(auditLog); event != nil {
		riskEvents = append(riskEvents, event)
	}
	
	// 3. 敏感信息获取尝试检测
	if event := d.detectSensitiveInfo(auditLog); event != nil {
		riskEvents = append(riskEvents, event)
	}
	
	// 4. 异常请求频率检测
	if event := d.detectAbnormalFrequency(auditLog); event != nil {
		riskEvents = append(riskEvents, event)
	}
	
	// 5. IP异常检测
	if event := d.detectIPAnomaly(auditLog); event != nil {
		riskEvents = append(riskEvents, event)
	}
	
	// 6. 异常请求模式检测
	if event := d.detectAbnormalPattern(auditLog); event != nil {
		riskEvents = append(riskEvents, event)
	}
	
	// 合并风险事件，取最高风险等级
	if len(riskEvents) > 0 {
		return d.mergeRiskEvents(auditLog, riskEvents)
	}
	
	return nil
}

// detectTokenAbuse 检测Token滥用
func (d *RiskDetector) detectTokenAbuse(auditLog *model.AuditLog) *model.RiskEvent {
	// 检查单次请求token数是否异常
	if auditLog.TotalTokens > d.cfg.TokenThresholdHourly {
		return &model.RiskEvent{
			Timestamp:  time.Now(),
			EventID:    uuid.New().String(),
			RequestID:  auditLog.RequestID,
			UserID:     auditLog.UserID,
			UserName:   auditLog.UserName,
			RiskLevel:  string(model.RiskLevelHigh),
			RiskType:   string(model.RiskTypeTokenAbuse),
			RiskScore:  0.8,
			RiskReason: "单次请求token数超过阈值",
			Description: "用户单次请求使用了大量token",
			Evidence:   jsonString(map[string]interface{}{
				"total_tokens": auditLog.TotalTokens,
				"threshold":    d.cfg.TokenThresholdHourly,
			}),
			RequestIP: auditLog.RequestIP,
			ModelName: auditLog.ModelName,
		}
	}
	
	// 检查用户最近一小时的token使用情况
	// 这里可以查询ClickHouse获取更准确的统计
	
	return nil
}

// detectOffHoursAccess 检测非工作时间访问
func (d *RiskDetector) detectOffHoursAccess(auditLog *model.AuditLog) *model.RiskEvent {
	hour := auditLog.RequestTime.Hour()
	
	// 检查是否在非工作时间
	isOffHours := false
	if d.cfg.OffHoursEnd < d.cfg.OffHoursStart {
		// 跨午夜的情况，如 22:00 - 06:00
		isOffHours = hour >= d.cfg.OffHoursStart || hour < d.cfg.OffHoursEnd
	} else {
		isOffHours = hour >= d.cfg.OffHoursStart && hour < d.cfg.OffHoursEnd
	}
	
	if isOffHours {
		return &model.RiskEvent{
			Timestamp:  time.Now(),
			EventID:    uuid.New().String(),
			RequestID:  auditLog.RequestID,
			UserID:     auditLog.UserID,
			UserName:   auditLog.UserName,
			RiskLevel:  string(model.RiskLevelLow),
			RiskType:   string(model.RiskTypeOffHoursAccess),
			RiskScore:  0.3,
			RiskReason: "非工作时间访问",
			Description: "用户在非工作时间使用AI服务",
			Evidence:   jsonString(map[string]interface{}{
				"request_hour":     hour,
				"off_hours_start":  d.cfg.OffHoursStart,
				"off_hours_end":    d.cfg.OffHoursEnd,
			}),
			RequestIP: auditLog.RequestIP,
			ModelName: auditLog.ModelName,
		}
	}
	
	return nil
}

// detectSensitiveInfo 检测敏感信息获取尝试
func (d *RiskDetector) detectSensitiveInfo(auditLog *model.AuditLog) *model.RiskEvent {
	requestBody := strings.ToLower(auditLog.RequestBody)
	
	// 定义敏感关键词
	sensitiveKeywords := []string{
		"password", "secret", "key", "token", "credential",
		"密码", "密钥", "机密", "隐私",
		"身份证", "手机号", "银行卡", "信用卡",
		"工资", "薪资", "salary", "income",
		"内部文件", "internal document", "confidential",
	}
	
	// 检查是否包含敏感关键词
	var matchedKeywords []string
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(requestBody, keyword) {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}
	
	// 检查是否匹配配置的敏感模式
	var matchedPatterns []string
	for _, pattern := range d.patterns {
		if pattern.PatternType == "sensitive_info" {
			matched, _ := regexp.MatchString(pattern.Pattern, requestBody)
			if matched {
				matchedPatterns = append(matchedPatterns, pattern.Description)
			}
		}
	}
	
	if len(matchedKeywords) > 0 || len(matchedPatterns) > 0 {
		riskLevel := model.RiskLevelMedium
		score := 0.5
		
		// 如果匹配多个敏感词或敏感模式，提高风险等级
		if len(matchedKeywords) >= 3 || len(matchedPatterns) >= 2 {
			riskLevel = model.RiskLevelHigh
			score = 0.8
		}
		
		return &model.RiskEvent{
			Timestamp:  time.Now(),
			EventID:    uuid.New().String(),
			RequestID:  auditLog.RequestID,
			UserID:     auditLog.UserID,
			UserName:   auditLog.UserName,
			RiskLevel:  string(riskLevel),
			RiskType:   string(model.RiskTypeSensitiveInfo),
			RiskScore:  score,
			RiskReason: "尝试获取敏感信息",
			Description: "用户请求中可能包含敏感信息获取意图",
			Evidence: jsonString(map[string]interface{}{
				"matched_keywords": matchedKeywords,
				"matched_patterns": matchedPatterns,
			}),
			RequestIP: auditLog.RequestIP,
			ModelName: auditLog.ModelName,
		}
	}
	
	return nil
}

// detectAbnormalFrequency 检测异常请求频率
func (d *RiskDetector) detectAbnormalFrequency(auditLog *model.AuditLog) *model.RiskEvent {
	// 获取用户最近5分钟的请求数
	// 这里简化处理，实际应该查询ClickHouse或Redis
	
	// 使用Redis统计用户请求频率
	key := "freq:user:" + strconv.FormatUint(auditLog.UserID, 10)
	count, _ := repository.GetRedis().Incr(context.Background(), key).Result()
	repository.GetRedis().Expire(context.Background(), key, 5*time.Minute)
	
	// 如果5分钟内请求超过100次，判定为异常
	if count > 100 {
		riskLevel := model.RiskLevelMedium
		if count > 300 {
			riskLevel = model.RiskLevelHigh
		}
		
		return &model.RiskEvent{
			Timestamp:  time.Now(),
			EventID:    uuid.New().String(),
			RequestID:  auditLog.RequestID,
			UserID:     auditLog.UserID,
			UserName:   auditLog.UserName,
			RiskLevel:  string(riskLevel),
			RiskType:   string(model.RiskTypeAbnormalFrequency),
			RiskScore:  0.7,
			RiskReason: "异常请求频率",
			Description: "用户短时间内发送了大量请求",
			Evidence: jsonString(map[string]interface{}{
				"requests_in_5min": count,
				"threshold":        100,
			}),
			RequestIP: auditLog.RequestIP,
			ModelName: auditLog.ModelName,
		}
	}
	
	return nil
}

// detectIPAnomaly 检测IP异常
func (d *RiskDetector) detectIPAnomaly(auditLog *model.AuditLog) *model.RiskEvent {
	// 检查IP是否在可疑IP列表中
	for _, ip := range d.cfg.SuspiciousIPList {
		if auditLog.RequestIP == ip {
			return &model.RiskEvent{
				Timestamp:  time.Now(),
				EventID:    uuid.New().String(),
				RequestID:  auditLog.RequestID,
				UserID:     auditLog.UserID,
				UserName:   auditLog.UserName,
				RiskLevel:  string(model.RiskLevelHigh),
				RiskType:   string(model.RiskTypeIPAnomaly),
				RiskScore:  0.9,
				RiskReason: "可疑IP访问",
				Description: "请求来自可疑IP地址",
				Evidence: jsonString(map[string]interface{}{
					"ip":              auditLog.RequestIP,
					"suspicious_list": d.cfg.SuspiciousIPList,
				}),
				RequestIP: auditLog.RequestIP,
				ModelName: auditLog.ModelName,
			}
		}
	}
	
	// TODO: 检查IP地理位置是否异常
	// TODO: 检查IP是否频繁变动
	
	return nil
}

// detectAbnormalPattern 检测异常请求模式
func (d *RiskDetector) detectAbnormalPattern(auditLog *model.AuditLog) *model.RiskEvent {
	requestBody := strings.ToLower(auditLog.RequestBody)
	
	// 检查是否匹配异常模式
	var matchedPatterns []string
	for _, pattern := range d.patterns {
		if pattern.PatternType == "abnormal_pattern" || pattern.PatternType == "injection" {
			matched, _ := regexp.MatchString(pattern.Pattern, requestBody)
			if matched {
				matchedPatterns = append(matchedPatterns, pattern.Description)
			}
		}
	}
	
	// 检测提示词注入攻击
	injectionPatterns := []string{
		`ignore previous instructions`,
		`ignore all prior instructions`,
		`disregard previous`,
		`system prompt`,
		`you are now`,
		`new role`,
		`developer mode`,
		`DAN mode`,
	}
	
	for _, pattern := range injectionPatterns {
		if strings.Contains(requestBody, pattern) {
			matchedPatterns = append(matchedPatterns, "提示词注入: "+pattern)
		}
	}
	
	if len(matchedPatterns) > 0 {
		return &model.RiskEvent{
			Timestamp:  time.Now(),
			EventID:    uuid.New().String(),
			RequestID:  auditLog.RequestID,
			UserID:     auditLog.UserID,
			UserName:   auditLog.UserName,
			RiskLevel:  string(model.RiskLevelHigh),
			RiskType:   string(model.RiskTypeAbnormalPattern),
			RiskScore:  0.85,
			RiskReason: "异常请求模式",
			Description: "检测到异常的请求模式或潜在的提示词注入",
			Evidence: jsonString(map[string]interface{}{
				"matched_patterns": matchedPatterns,
			}),
			RequestIP: auditLog.RequestIP,
			ModelName: auditLog.ModelName,
		}
	}
	
	return nil
}

// mergeRiskEvents 合并风险事件
func (d *RiskDetector) mergeRiskEvents(auditLog *model.AuditLog, events []*model.RiskEvent) *model.RiskEvent {
	if len(events) == 1 {
		return events[0]
	}
	
	// 计算最高风险等级
	maxScore := 0.0
	var riskTypes []string
	var riskReasons []string
	
	for _, e := range events {
		if e.RiskScore > maxScore {
			maxScore = e.RiskScore
		}
		riskTypes = append(riskTypes, e.RiskType)
		riskReasons = append(riskReasons, e.RiskReason)
	}
	
	// 确定最终风险等级
	var finalRiskLevel string
	switch {
	case maxScore >= 0.8:
		finalRiskLevel = string(model.RiskLevelHigh)
	case maxScore >= 0.5:
		finalRiskLevel = string(model.RiskLevelMedium)
	default:
		finalRiskLevel = string(model.RiskLevelLow)
	}
	
	return &model.RiskEvent{
		Timestamp:   time.Now(),
		EventID:     uuid.New().String(),
		RequestID:   auditLog.RequestID,
		UserID:      auditLog.UserID,
		UserName:    auditLog.UserName,
		RiskLevel:   finalRiskLevel,
		RiskType:    "multiple",
		RiskScore:   maxScore,
		RiskReason:  "多个风险因素",
		Description: "检测到多个风险因素",
		Evidence: jsonString(map[string]interface{}{
			"risk_types":   riskTypes,
			"risk_reasons": riskReasons,
			"events":       events,
		}),
		RequestIP: auditLog.RequestIP,
		ModelName: auditLog.ModelName,
	}
}

// jsonString 将对象转换为JSON字符串
func jsonString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// ProcessAuditLog 处理审计日志并进行风险检测
func ProcessAuditLog(auditLog *model.AuditLog) {
	detector := NewRiskDetector()
	
	riskEvent := detector.DetectRisk(auditLog)
	if riskEvent != nil {
		// 保存风险事件
		if err := repository.InsertRiskEvent(riskEvent); err != nil {
			logrus.WithError(err).Error("failed to insert risk event")
		}
		
		// 记录日志
		logrus.WithFields(logrus.Fields{
			"event_id":   riskEvent.EventID,
			"user_id":    riskEvent.UserID,
			"risk_level": riskEvent.RiskLevel,
			"risk_type":  riskEvent.RiskType,
		}).Warn("Risk detected")
		
		// TODO: 发送告警通知（邮件、钉钉、企业微信等）
		if riskEvent.RiskLevel == string(model.RiskLevelHigh) || riskEvent.RiskLevel == string(model.RiskLevelCritical) {
			// sendAlert(riskEvent)
		}
	}
}
