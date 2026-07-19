// Package service 同城车辆买卖业务逻辑层 - 审核规则
// CarAuditRule 审核规则（全局，BaseModel 无 region_id）
// 提供 M 端规则管理 + 内部审核检查能力（Check 方法供其他 service 调用）
package service

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/modules/car/dto"
	"wuchang-tongcheng/internal/modules/car/model"
	"wuchang-tongcheng/internal/modules/car/repository"
	"wuchang-tongcheng/internal/pkg/utils"

	"gorm.io/gorm"
)

var (
	ErrAuditRuleNotFound       = errors.New("审核规则不存在")
	ErrAuditRuleNoPermission   = errors.New("无权操作此审核规则")
	ErrAuditRuleDuplicate      = errors.New("规则键已存在")
	ErrAuditRulePatternInvalid = errors.New("规则正则表达式无效")
	ErrAuditRuleStatusInvalid  = errors.New("审核规则状态无效")
)

// ===== AuditRuleService 审核规则业务接口 =====

// AuditRuleService 审核规则业务接口
// M 端：规则 CRUD + 状态管理
// 内部：Check 方法供其他 service（如 listing）调用以校验内容是否命中规则
type AuditRuleService interface {
	// M 端管理
	Create(operatorID uint, req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error)
	Update(id uint, operatorID uint, req *dto.UpdateAuditRuleRequest) error
	Delete(id uint, operatorID uint) error
	GetByID(id uint) (*dto.AuditRuleInfo, error)
	List(req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error)
	UpdateStatus(id uint, operatorID uint, status int) error

	// 内部调用
	// Check 根据 AuditCheckRequest 对内容做规则匹配，返回命中信息与执行动作
	// type 取值参考 model.AuditRuleType*；content 为待校验文本；data 为可选结构化数据
	Check(req *dto.AuditCheckRequest) (*dto.AuditCheckResponse, error)
}

type auditRuleService struct {
	repo repository.AuditRuleRepository
}

// NewAuditRuleService 创建审核规则 service 实例
func NewAuditRuleService(repo repository.AuditRuleRepository) AuditRuleService {
	return &auditRuleService{repo: repo}
}

// auditRuleStatusText 审核规则状态文本
// 0 禁用 / 1 启用
func auditRuleStatusText(status int) string {
	switch status {
	case 0:
		return "禁用"
	case 1:
		return "启用"
	}
	return ""
}

// toAuditRuleInfo model -> dto
// Threshold 字段为 JSONB，直接透传；前端按 interface{} 解析
func toAuditRuleInfo(r *model.CarAuditRule) *dto.AuditRuleInfo {
	var threshold interface{}
	if len(r.Threshold) > 0 {
		// JSONB 底层为 []byte，直接赋值给 interface{} 字段
		// json 序列化时会通过 JSONB.MarshalJSON 输出原始 JSON
		threshold = r.Threshold
	}
	return &dto.AuditRuleInfo{
		ID:          r.ID,
		RuleName:    r.RuleName,
		RuleType:    r.RuleType,
		RuleKey:     r.RuleKey,
		Pattern:     r.Pattern,
		Threshold:   threshold,
		Action:      r.Action,
		PenaltyType: r.PenaltyType,
		Severity:    r.Severity,
		Status:      r.Status,
		StatusText:  auditRuleStatusText(r.Status),
		Description: r.Description,
		Sort:        r.Sort,
		CreatedAt:   r.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ===== M 端管理 =====

// Create 创建审核规则
func (s *auditRuleService) Create(operatorID uint, req *dto.CreateAuditRuleRequest) (*dto.AuditRuleInfo, error) {
	// 校验 pattern：若 ruleType 为敏感词/违禁内容/联系方式且 pattern 非空，必须为合法正则
	if needsRegexValidation(req.RuleType) && req.Pattern != "" {
		if _, err := regexp.Compile(req.Pattern); err != nil {
			return nil, ErrAuditRulePatternInvalid
		}
	}

	rule := &model.CarAuditRule{
		RuleName:    req.RuleName,
		RuleType:    req.RuleType,
		RuleKey:     req.RuleKey,
		Pattern:     req.Pattern,
		Action:      req.Action,
		PenaltyType: req.PenaltyType,
		Severity:    req.Severity,
		Status:      req.Status,
		Description: req.Description,
		Sort:        req.Sort,
	}

	// 默认值兜底
	if rule.Action == "" {
		rule.Action = model.AuditActionReject
	}
	if rule.Severity == 0 {
		rule.Severity = 1
	}
	if rule.Status == 0 && req.Status == 0 {
		// binding 中 omitempty 允许传 0；若调用方未指定，默认启用
		rule.Status = 1
	}

	// Threshold JSONB 转换
	if req.Threshold != nil {
		jb, err := model.FromJSON(req.Threshold)
		if err == nil {
			rule.Threshold = jb
		}
	}

	if err := s.repo.Create(rule); err != nil {
		return nil, err
	}
	return toAuditRuleInfo(rule), nil
}

// Update 更新审核规则（仅 M 端）
func (s *auditRuleService) Update(id uint, operatorID uint, req *dto.UpdateAuditRuleRequest) error {
	rule, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	_ = rule

	fields := map[string]interface{}{}
	if req.RuleName != nil {
		fields["rule_name"] = *req.RuleName
	}
	if req.RuleType != nil {
		fields["rule_type"] = *req.RuleType
	}
	if req.RuleKey != nil {
		fields["rule_key"] = *req.RuleKey
	}
	if req.Pattern != nil {
		// 校验正则（若类型需要）
		rt := rule.RuleType
		if req.RuleType != nil {
			rt = *req.RuleType
		}
		if needsRegexValidation(rt) && *req.Pattern != "" {
			if _, err := regexp.Compile(*req.Pattern); err != nil {
				return ErrAuditRulePatternInvalid
			}
		}
		fields["pattern"] = *req.Pattern
	}
	if req.Action != nil {
		fields["action"] = *req.Action
	}
	if req.PenaltyType != nil {
		fields["penalty_type"] = *req.PenaltyType
	}
	if req.Severity != nil {
		fields["severity"] = *req.Severity
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Sort != nil {
		fields["sort"] = *req.Sort
	}

	// Threshold 字段：interface{} 无法用 nil 判断"未提供"，故仅当非 nil 时更新
	// 注意：调用方若希望清空 Threshold，需传入空 map / 空数组而非 nil
	if req.Threshold != nil {
		jb, err := model.FromJSON(req.Threshold)
		if err == nil {
			fields["threshold"] = jb
		}
	}

	if len(fields) == 0 {
		return nil
	}
	return s.repo.Update(id, fields)
}

// Delete 删除审核规则
func (s *auditRuleService) Delete(id uint, operatorID uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 获取审核规则详情
func (s *auditRuleService) GetByID(id uint) (*dto.AuditRuleInfo, error) {
	rule, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAuditRuleNotFound
		}
		return nil, err
	}
	return toAuditRuleInfo(rule), nil
}

// List 审核规则列表
func (s *auditRuleService) List(req *dto.AuditRuleListRequest) (*utils.Pagination, []dto.AuditRuleInfo, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize)
	opts := repository.AuditRuleListQuery{
		RuleType: req.RuleType,
		RuleKey:  req.RuleKey,
		Action:   req.Action,
		Status:   req.Status,
		Keyword:  req.Keyword,
	}

	list, total, err := s.repo.List(opts, pagination)
	if err != nil {
		return nil, nil, err
	}

	infos := make([]dto.AuditRuleInfo, 0, len(list))
	for i := range list {
		infos = append(infos, *toAuditRuleInfo(&list[i]))
	}
	pagination.Total = total
	return pagination, infos, nil
}

// UpdateStatus 更新规则状态（启用/禁用）
func (s *auditRuleService) UpdateStatus(id uint, operatorID uint, status int) error {
	if status != 0 && status != 1 {
		return ErrAuditRuleStatusInvalid
	}
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAuditRuleNotFound
		}
		return err
	}
	return s.repo.Update(id, map[string]interface{}{"status": status})
}

// ===== 内部审核检查 =====

// Check 根据 AuditCheckRequest 对内容做规则匹配
// 匹配逻辑：
//   - sensitive_word / prohibited / contact：按 pattern 关键词或正则匹配 content
//   - price_check：从 data 中取 price 字段，与 Threshold.min/max 比对
//   - mileage_check：从 data 中取 mileage 字段，与 Threshold.min/max 比对
//   - vin_check：从 data 中取 vin 字段，校验长度（17 位）与正则
//   - frequency：从 data 中取 count 字段，与 Threshold.max 比对
//   - fake_car：MVP 简化，不自动匹配，需人工审核
//
// 命中规则后：
//   - 取所有命中规则中 severity 最高的 action 作为最终动作
//   - 默认 passed=true；若 action=reject 则 passed=false
func (s *auditRuleService) Check(req *dto.AuditCheckRequest) (*dto.AuditCheckResponse, error) {
	// 获取该类型下所有启用的规则
	rules, err := s.repo.ListByRuleType(req.Type)
	if err != nil {
		return nil, err
	}

	resp := &dto.AuditCheckResponse{
		Passed:  true,
		Action:  "",
		Matched: []dto.AuditMatchedItem{},
		Reason:  "",
	}

	if len(rules) == 0 {
		return resp, nil
	}

	// 提取 data 字段（若存在）
	dataMap := extractDataMap(req.Data)

	for i := range rules {
		rule := &rules[i]
		hit, pattern := matchRule(rule, req.Content, dataMap)
		if !hit {
			continue
		}
		resp.Matched = append(resp.Matched, dto.AuditMatchedItem{
			RuleID:   rule.ID,
			RuleName: rule.RuleName,
			RuleType: rule.RuleType,
			Pattern:  pattern,
			Severity: rule.Severity,
		})
	}

	if len(resp.Matched) == 0 {
		return resp, nil
	}

	// 选取 severity 最高的规则动作
	topSeverity := 0
	topAction := ""
	var topRule *model.CarAuditRule
	for i := range rules {
		rule := &rules[i]
		for _, m := range resp.Matched {
			if m.RuleID == rule.ID && rule.Severity > topSeverity {
				topSeverity = rule.Severity
				topAction = rule.Action
				topRule = rule
			}
		}
	}

	resp.Action = topAction
	if topAction == model.AuditActionReject {
		resp.Passed = false
		if topRule != nil {
			resp.Reason = "命中规则：" + topRule.RuleName
		}
	}

	return resp, nil
}

// ===== 内部辅助函数 =====

// needsRegexValidation 判断规则类型是否需要校验正则表达式
func needsRegexValidation(ruleType string) bool {
	switch ruleType {
	case model.AuditRuleTypeSensitiveWord,
		model.AuditRuleTypeProhibited,
		model.AuditRuleTypeContact,
		model.AuditRuleTypeVINCheck:
		return true
	}
	return false
}

// extractDataMap 从 interface{} 提取 map[string]interface{}
// 若非 map 类型则返回空 map
func extractDataMap(data interface{}) map[string]interface{} {
	if data == nil {
		return map[string]interface{}{}
	}
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// matchRule 对单条规则进行匹配
// 返回：(是否命中, 命中的 pattern 描述)
func matchRule(rule *model.CarAuditRule, content string, data map[string]interface{}) (bool, string) {
	switch rule.RuleType {
	case model.AuditRuleTypeSensitiveWord,
		model.AuditRuleTypeProhibited:
		return matchKeywordOrRegex(rule.Pattern, content)
	case model.AuditRuleTypeContact:
		// 联系方式：手机号 / 微信号 等正则
		return matchKeywordOrRegex(rule.Pattern, content)
	case model.AuditRuleTypePriceCheck:
		return matchThreshold(rule.Threshold, data, "price")
	case model.AuditRuleTypeMileageCheck:
		return matchThreshold(rule.Threshold, data, "mileage")
	case model.AuditRuleTypeFrequency:
		return matchFrequency(rule.Threshold, data)
	case model.AuditRuleTypeVINCheck:
		return matchVIN(rule.Pattern, data)
	case model.AuditRuleTypeFakeCar:
		// MVP 简化：虚假车源由人工审核，规则不自动匹配
		return false, ""
	}
	return false, ""
}

// matchKeywordOrRegex 关键词或正则匹配
// pattern 为空 → 不匹配
// pattern 以 "/" 开头与结尾 → 正则匹配（如 "/\d{11}/"）
// 否则 → 关键词包含匹配（大小写不敏感）
func matchKeywordOrRegex(pattern, content string) (bool, string) {
	if pattern == "" || content == "" {
		return false, ""
	}
	if len(pattern) >= 2 && strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		regexStr := pattern[1 : len(pattern)-1]
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return false, ""
		}
		if re.MatchString(content) {
			return true, pattern
		}
		return false, ""
	}
	// 关键词匹配（支持 | 分隔多关键词）
	for _, kw := range strings.Split(pattern, "|") {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(strings.ToLower(content), strings.ToLower(kw)) {
			return true, kw
		}
	}
	return false, ""
}

// matchThreshold 阈值匹配（price/mileage 等）
// Threshold JSON 结构：{"min": 1000, "max": 100000}
// 命中条件：value < min 或 value > max
func matchThreshold(threshold model.JSONB, data map[string]interface{}, field string) (bool, string) {
	if len(threshold) == 0 {
		return false, ""
	}
	val, ok := data[field]
	if !ok {
		return false, ""
	}
	value, err := toFloat64(val)
	if err != nil {
		return false, ""
	}

	var th struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	}
	if err := threshold.Parse(&th); err != nil {
		return false, ""
	}

	if th.Min > 0 && value < th.Min {
		return true, "min=" + strconv.FormatFloat(th.Min, 'f', -1, 64)
	}
	if th.Max > 0 && value > th.Max {
		return true, "max=" + strconv.FormatFloat(th.Max, 'f', -1, 64)
	}
	return false, ""
}

// matchFrequency 频率匹配
// Threshold JSON 结构：{"max": 10, "window": 3600}（window 单位秒，匹配时仅校验 max）
// 命中条件：data["count"] > max
func matchFrequency(threshold model.JSONB, data map[string]interface{}) (bool, string) {
	if len(threshold) == 0 {
		return false, ""
	}
	val, ok := data["count"]
	if !ok {
		return false, ""
	}
	count, err := toFloat64(val)
	if err != nil {
		return false, ""
	}

	var th struct {
		Max    float64 `json:"max"`
		Window int     `json:"window"`
	}
	if err := threshold.Parse(&th); err != nil {
		return false, ""
	}

	if th.Max > 0 && count > th.Max {
		return true, "max=" + strconv.FormatFloat(th.Max, 'f', -1, 64)
	}
	return false, ""
}

// matchVIN 车架号校验
// 优先使用 rule.Pattern 作为正则；为空则使用默认 17 位字母数字正则
// data["vin"] 必须存在
func matchVIN(pattern string, data map[string]interface{}) (bool, string) {
	vin, ok := data["vin"]
	if !ok {
		return false, ""
	}
	vinStr, ok := vin.(string)
	if !ok {
		return false, ""
	}

	regexStr := pattern
	if regexStr == "" {
		// 默认 VIN：17 位字母数字（不含 I/O/Q）
		regexStr = `^[A-HJ-NPR-Z0-9]{17}$`
	} else if len(regexStr) >= 2 && strings.HasPrefix(regexStr, "/") && strings.HasSuffix(regexStr, "/") {
		regexStr = regexStr[1 : len(regexStr)-1]
	}

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return false, ""
	}
	// 命中规则 = VIN 不合法
	if !re.MatchString(vinStr) {
		return true, "vin_invalid"
	}
	return false, ""
}

// toFloat64 将 interface{} 转为 float64
// 支持 int / int64 / float32 / float64 / json.Number / string
func toFloat64(v interface{}) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case string:
		return strconv.ParseFloat(x, 64)
	case json.Number:
		return x.Float64()
	}
	return 0, errors.New("invalid number type")
}
