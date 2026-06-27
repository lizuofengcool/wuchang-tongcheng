// Package service 系统设置业务逻辑层
package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"wuchang-tongcheng/internal/modules/setting/dto"
	"wuchang-tongcheng/internal/modules/setting/model"
	"wuchang-tongcheng/internal/modules/setting/repository"

	"gorm.io/gorm"
)

var (
	ErrSettingNotFound   = errors.New("配置项不存在")
	ErrSettingKeyExists   = errors.New("配置键已存在")
	ErrSettingValueInvalid = errors.New("配置值与值类型不匹配")
)

// SettingService 系统设置业务逻辑接口
type SettingService interface {
	Create(regionID uint, req *dto.CreateSettingRequest) (*dto.SettingInfo, error)
	Update(id uint, req *dto.UpdateSettingRequest) error
	Delete(id uint) error
	GetByID(id uint) (*dto.SettingInfo, error)
	GetByGroup(group string, regionID uint) ([]dto.SettingInfo, error)
	GetAll(regionID uint) (map[string][]dto.SettingInfo, error)
	BatchUpdate(regionID uint, req *dto.BatchUpdateRequest) error
	// GetValue 直接取值（供其他模块读取配置）
	GetValue(group, key string, regionID uint) (string, error)
}

type settingService struct {
	repo repository.SettingRepository
}

// NewSettingService 创建系统设置服务
func NewSettingService(repo repository.SettingRepository) SettingService {
	return &settingService{repo: repo}
}

func toSettingInfo(s *model.Setting) *dto.SettingInfo {
	vt := s.ValueType
	if vt == "" {
		vt = "string"
	}
	return &dto.SettingInfo{
		ID:          s.ID,
		Group:       s.Group,
		Key:         s.Key,
		Value:       s.Value,
		ParsedValue: parseValue(s.Value, vt),
		ValueType:   vt,
		Description: s.Description,
		Sort:        s.Sort,
	}
}

// parseValue 按 valueType 反序列化字符串值为对应的 Go 类型
//   - string: 原样返回
//   - number: float64
//   - bool:   bool
//   - json:   任意 JSON 结构（map/slice/标量）；解析失败回退为原字符串
func parseValue(value, valueType string) interface{} {
	switch valueType {
	case "number":
		f, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return value
		}
		return f
	case "bool":
		v := strings.ToLower(strings.TrimSpace(value))
		switch v {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off", "":
			return false
		}
		return value
	case "json":
		var parsed interface{}
		if err := json.Unmarshal([]byte(value), &parsed); err != nil {
			return value
		}
		return parsed
	default: // string 或未知类型
		return value
	}
}

// validateValue 校验 value 是否符合声明的 valueType，不符合返回错误
func validateValue(value, valueType string) error {
	switch valueType {
	case "number":
		if value == "" {
			return nil
		}
		if _, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err != nil {
			return fmt.Errorf("%w: 期望 number, 实际 %q", ErrSettingValueInvalid, value)
		}
	case "bool":
		if value == "" {
			return nil
		}
		v := strings.ToLower(strings.TrimSpace(value))
		switch v {
		case "true", "false", "1", "0", "yes", "no", "on", "off":
			return nil
		}
		return fmt.Errorf("%w: 期望 bool(true/false/1/0), 实际 %q", ErrSettingValueInvalid, value)
	case "json":
		if value == "" {
			return nil
		}
		var v interface{}
		if err := json.Unmarshal([]byte(value), &v); err != nil {
			return fmt.Errorf("%w: 期望合法 JSON, 解析失败: %v", ErrSettingValueInvalid, err)
		}
	}
	return nil
}

// Create 创建配置
func (s *settingService) Create(regionID uint, req *dto.CreateSettingRequest) (*dto.SettingInfo, error) {
	// 检查key是否已存在
	if _, err := s.repo.FindByKey(req.Group, req.Key, regionID); err == nil {
		return nil, ErrSettingKeyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	valueType := req.ValueType
	if valueType == "" {
		valueType = "string"
	}

	// 校验 value 是否符合声明的 valueType
	if err := validateValue(req.Value, valueType); err != nil {
		return nil, err
	}

	setting := &model.Setting{
		Group:       req.Group,
		Key:         req.Key,
		Value:       req.Value,
		ValueType:   valueType,
		Description: req.Description,
		Sort:        req.Sort,
	}
	setting.RegionID = regionID

	if err := s.repo.Create(setting); err != nil {
		return nil, err
	}
	return toSettingInfo(setting), nil
}

// Update 更新配置
func (s *settingService) Update(id uint, req *dto.UpdateSettingRequest) error {
	setting, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSettingNotFound
		}
		return err
	}
	// 用已存在的 valueType 校验新值
	vt := setting.ValueType
	if vt == "" {
		vt = "string"
	}
	if err := validateValue(req.Value, vt); err != nil {
		return err
	}
	fields := map[string]interface{}{
		"value":       req.Value,
		"description": req.Description,
		"sort":        req.Sort,
	}
	return s.repo.UpdateFields(id, fields)
}

// Delete 删除配置
func (s *settingService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSettingNotFound
		}
		return err
	}
	return s.repo.Delete(id)
}

// GetByID 根据ID获取配置
func (s *settingService) GetByID(id uint) (*dto.SettingInfo, error) {
	setting, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSettingNotFound
		}
		return nil, err
	}
	return toSettingInfo(setting), nil
}

// GetByGroup 根据分组获取配置
func (s *settingService) GetByGroup(group string, regionID uint) ([]dto.SettingInfo, error) {
	list, err := s.repo.FindByGroup(group, regionID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.SettingInfo, 0, len(list))
	for i := range list {
		result = append(result, *toSettingInfo(&list[i]))
	}
	return result, nil
}

// GetAll 获取所有配置，按group分组返回
func (s *settingService) GetAll(regionID uint) (map[string][]dto.SettingInfo, error) {
	list, err := s.repo.FindAll(regionID)
	if err != nil {
		return nil, err
	}
	groupMap := make(map[string][]dto.SettingInfo)
	for i := range list {
		info := *toSettingInfo(&list[i])
		groupMap[info.Group] = append(groupMap[info.Group], info)
	}
	return groupMap, nil
}

// BatchUpdate 批量更新配置值
func (s *settingService) BatchUpdate(regionID uint, req *dto.BatchUpdateRequest) error {
	for _, item := range req.Items {
		// 只更新值，根据key查找（不限group，简化处理）
		// 这里需要遍历查找，可能存在同key不同group的情况，需调用方保证key唯一性
		setting, err := s.repo.FindByKey("", item.Key, regionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue // 跳过不存在的
			}
			return err
		}
		// 用已存在的 valueType 校验新值
		vt := setting.ValueType
		if vt == "" {
			vt = "string"
		}
		if err := validateValue(item.Value, vt); err != nil {
			return err
		}
		if err := s.repo.UpdateFields(setting.ID, map[string]interface{}{"value": item.Value}); err != nil {
			return err
		}
	}
	return nil
}

// GetValue 获取配置值
func (s *settingService) GetValue(group, key string, regionID uint) (string, error) {
	setting, err := s.repo.FindByKey(group, key, regionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrSettingNotFound
		}
		return "", err
	}
	return setting.Value, nil
}
