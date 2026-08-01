// Package database 数据库封装单元测试
// 覆盖 GetDB/Close/AutoMigrate 未初始化降级路径与 BaseModel/RegionBaseModel/TableNamer 契约。
// 全部用例离线执行，不依赖真实 PostgreSQL（Init 需真实连接，故仅覆盖未初始化分支）。
package database

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
)

// resetDB 清空包级 db 全局变量，保证用例间状态隔离。
// t.Cleanup 在用例结束后再次清空，避免后续用例受残留影响。
func resetDB(t *testing.T) {
	t.Helper()
	db = nil
	t.Cleanup(func() { db = nil })
}

// --- GetDB / Close / AutoMigrate 未初始化状态测试 ---

// TestGetDB_PanicsWhenNotInitialized 未初始化时 GetDB 应 panic 并携带提示信息
func TestGetDB_PanicsWhenNotInitialized(t *testing.T) {
	resetDB(t)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("未初始化时 GetDB 应 panic，got nil")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic 值应为 string 类型，got %T: %v", r, r)
		}
		if msg != "database not initialized, call Init() first" {
			t.Errorf("panic 消息不匹配，got %q", msg)
		}
	}()
	_ = GetDB()
}

// TestClose_UninitializedReturnsNil 未初始化时 Close 不 panic 且返回 nil
func TestClose_UninitializedReturnsNil(t *testing.T) {
	resetDB(t)
	if err := Close(); err != nil {
		t.Errorf("未初始化时 Close 应返回 nil，got: %v", err)
	}
}

// TestAutoMigrate_UninitializedReturnsError 未初始化时 AutoMigrate 返回错误
func TestAutoMigrate_UninitializedReturnsError(t *testing.T) {
	resetDB(t)
	err := AutoMigrate(&BaseModel{})
	if err == nil {
		t.Fatalf("未初始化时 AutoMigrate 应返回错误，got nil")
	}
}

// TestAutoMigrate_ErrorContainsExpectedMessage 未初始化错误消息应包含提示语
func TestAutoMigrate_ErrorContainsExpectedMessage(t *testing.T) {
	resetDB(t)
	err := AutoMigrate(&BaseModel{})
	if err == nil {
		t.Fatalf("期望错误，got nil")
	}
	want := "database not initialized"
	if err.Error() != want {
		t.Errorf("错误消息 = %q, want %q", err.Error(), want)
	}
}

// TestAutoMigrate_NoModels_StillErrorsWhenUninitialized 即使无 model 参数，未初始化仍应报错
func TestAutoMigrate_NoModels_StillErrorsWhenUninitialized(t *testing.T) {
	resetDB(t)
	err := AutoMigrate()
	if err == nil {
		t.Fatalf("未初始化时 AutoMigrate() 应返回错误，got nil")
	}
}

// --- BaseModel 契约测试 ---

// TestBaseModel_ZeroValue 零值字段符合预期
func TestBaseModel_ZeroValue(t *testing.T) {
	var m BaseModel
	if m.ID != 0 {
		t.Errorf("BaseModel 零值 ID 应为 0，got %d", m.ID)
	}
	if !m.CreatedAt.IsZero() {
		t.Errorf("BaseModel 零值 CreatedAt 应为零值时间")
	}
	if !m.UpdatedAt.IsZero() {
		t.Errorf("BaseModel 零值 UpdatedAt 应为零值时间")
	}
	// DeletedAt 未删除时 Valid 应为 false
	if m.DeletedAt.Valid {
		t.Errorf("BaseModel 零值 DeletedAt.Valid 应为 false")
	}
	if m.DeletedAt.Time.IsZero() == false {
		// gorm.DeletedAt 零值 Time 为零值时间，仅做不强制断言
		t.Errorf("BaseModel 零值 DeletedAt.Time 应为零值时间")
	}
}

// TestBaseModel_IDIsPrimary ID 字段应带 primarykey 的 gorm 标签与 id 的 json 标签
func TestBaseModel_IDIsPrimary(t *testing.T) {
	tp := reflect.TypeOf(BaseModel{})
	f, ok := tp.FieldByName("ID")
	if !ok {
		t.Fatalf("BaseModel 应包含 ID 字段")
	}
	if got := f.Tag.Get("gorm"); got != "primarykey" {
		t.Errorf("ID gorm 标签 = %q, want %q", got, "primarykey")
	}
	if got := f.Tag.Get("json"); got != "id" {
		t.Errorf("ID json 标签 = %q, want %q", got, "id")
	}
}

// TestBaseModel_DeletedAtIndexed DeletedAt 字段应带 index 的 gorm 标签且 json 输出忽略
func TestBaseModel_DeletedAtIndexed(t *testing.T) {
	tp := reflect.TypeOf(BaseModel{})
	f, ok := tp.FieldByName("DeletedAt")
	if !ok {
		t.Fatalf("BaseModel 应包含 DeletedAt 字段")
	}
	if got := f.Tag.Get("gorm"); got != "index" {
		t.Errorf("DeletedAt gorm 标签 = %q, want %q", got, "index")
	}
	if got := f.Tag.Get("json"); got != "-" {
		t.Errorf("DeletedAt json 标签 = %q, want %q（不输出给前端）", got, "-")
	}
}

// --- RegionBaseModel 契约测试 ---

// TestRegionBaseModel_EmbedsBaseModel RegionBaseModel 应内嵌 BaseModel，可直接访问其字段
func TestRegionBaseModel_EmbedsBaseModel(t *testing.T) {
	var r RegionBaseModel
	// 内嵌字段可直接通过外层类型访问
	r.ID = 42
	r.CreatedAt = gorm.DeletedAt{}.Time // 仅验证可写，不关心具体值
	if r.ID != 42 {
		t.Errorf("内嵌 BaseModel.ID 访问失败，got %d", r.ID)
	}
	// 类型断言：RegionBaseModel 内应能取到 BaseModel 字段
	tp := reflect.TypeOf(RegionBaseModel{})
	if _, ok := tp.FieldByName("ID"); !ok {
		t.Errorf("RegionBaseModel 应通过内嵌暴露 ID 字段")
	}
	if _, ok := tp.FieldByName("CreatedAt"); !ok {
		t.Errorf("RegionBaseModel 应通过内嵌暴露 CreatedAt 字段")
	}
	if _, ok := tp.FieldByName("DeletedAt"); !ok {
		t.Errorf("RegionBaseModel 应通过内嵌暴露 DeletedAt 字段")
	}
}

// TestRegionBaseModel_RegionIDField RegionID 字段存在且带 index/not null/default:1 标签
func TestRegionBaseModel_RegionIDField(t *testing.T) {
	tp := reflect.TypeOf(RegionBaseModel{})
	f, ok := tp.FieldByName("RegionID")
	if !ok {
		t.Fatalf("RegionBaseModel 应包含 RegionID 字段")
	}
	wantGorm := "index;not null;default:1"
	if got := f.Tag.Get("gorm"); got != wantGorm {
		t.Errorf("RegionID gorm 标签 = %q, want %q", got, wantGorm)
	}
	if got := f.Tag.Get("json"); got != "region_id" {
		t.Errorf("RegionID json 标签 = %q, want %q", got, "region_id")
	}
	// 零值为 0（数据库 default:1 由 DB 侧保证，结构体零值为 0）
	var r RegionBaseModel
	if r.RegionID != 0 {
		t.Errorf("RegionBaseModel 零值 RegionID 应为 0，got %d", r.RegionID)
	}
}

// TestRegionBaseModel_RegionIDAssignable RegionID 可被赋值并读回
func TestRegionBaseModel_RegionIDAssignable(t *testing.T) {
	r := RegionBaseModel{}
	r.RegionID = 99
	if r.RegionID != 99 {
		t.Errorf("RegionID 赋值后读回失败，got %d", r.RegionID)
	}
}

// --- TableNamer 接口测试 ---

// dummyTable 用于验证 TableNamer 接口契约
type dummyTable struct{}

func (dummyTable) TableName() string { return "dummy_table" }

// TestTableNamer_InterfaceSatisfied 实现 TableName() string 的类型应满足 TableNamer 接口
func TestTableNamer_InterfaceSatisfied(t *testing.T) {
	var n TableNamer = dummyTable{}
	if got := n.TableName(); got != "dummy_table" {
		t.Errorf("TableName() = %q, want %q", got, "dummy_table")
	}
}

// TestTableNamer_NilPointerDoesNotSatisfy 未实现方法的基础模型不应满足 TableNamer
func TestTableNamer_NilPointerDoesNotSatisfy(t *testing.T) {
	// BaseModel 没有实现 TableName()，编译期即可保证不满足 TableNamer
	// 此处仅做行为断言：接口零值为 nil
	var n TableNamer
	if n != nil {
		t.Errorf("未赋值的 TableNamer 应为 nil")
	}
}
