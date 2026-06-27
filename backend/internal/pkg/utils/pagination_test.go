package utils

import "testing"

func TestNewPagination_Defaults(t *testing.T) {
	p := NewPagination(0, 0)
	if p.Page != DefaultPage {
		t.Errorf("expected default page %d, got %d", DefaultPage, p.Page)
	}
	if p.PageSize != DefaultPageSize {
		t.Errorf("expected default page size %d, got %d", DefaultPageSize, p.PageSize)
	}
}

func TestNewPagination_NegativeValuesUseDefaults(t *testing.T) {
	p := NewPagination(-1, -5)
	if p.Page != DefaultPage {
		t.Errorf("expected default page for negative input, got %d", p.Page)
	}
	if p.PageSize != DefaultPageSize {
		t.Errorf("expected default page size for negative input, got %d", p.PageSize)
	}
}

func TestNewPagination_PageSizeCapped(t *testing.T) {
	p := NewPagination(1, 500)
	if p.PageSize != MaxPageSize {
		t.Errorf("expected page size capped to %d, got %d", MaxPageSize, p.PageSize)
	}
}

func TestPagination_Offset(t *testing.T) {
	cases := []struct {
		page, pageSize, want int
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 20, 40},
	}
	for _, c := range cases {
		p := NewPagination(c.page, c.pageSize)
		if got := p.Offset(); got != c.want {
			t.Errorf("page=%d size=%d: expected offset %d, got %d", c.page, c.pageSize, c.want, got)
		}
	}
}

func TestPagination_TotalPages(t *testing.T) {
	cases := []struct {
		total int64
		size  int
		want  int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{25, 10, 3},
	}
	for _, c := range cases {
		p := NewPagination(1, c.size)
		p.Total = c.total
		if got := p.TotalPages(); got != c.want {
			t.Errorf("total=%d size=%d: expected %d pages, got %d", c.total, c.size, c.want, got)
		}
	}
}

func TestPagination_HasNextHasPrev(t *testing.T) {
	p := NewPagination(2, 10)
	p.Total = 25 // 3 pages
	if !p.HasPrev() {
		t.Error("expected HasPrev=true for page 2")
	}
	if !p.HasNext() {
		t.Error("expected HasNext=true when page < total pages")
	}

	p2 := NewPagination(1, 10)
	p2.Total = 25
	if p2.HasPrev() {
		t.Error("expected HasPrev=false for page 1")
	}

	p3 := NewPagination(3, 10)
	p3.Total = 25
	if p3.HasNext() {
		t.Error("expected HasNext=false on last page")
	}
}

func TestParsePagination(t *testing.T) {
	p := ParsePagination("2", "20")
	if p.Page != 2 || p.PageSize != 20 {
		t.Errorf("expected page=2 size=20, got page=%d size=%d", p.Page, p.PageSize)
	}

	// 非法字符串退回默认值
	p2 := ParsePagination("abc", "")
	if p2.Page != DefaultPage || p2.PageSize != DefaultPageSize {
		t.Errorf("expected defaults for invalid input, got page=%d size=%d", p2.Page, p2.PageSize)
	}
}

func TestSortParams_OrderString(t *testing.T) {
	s := NewSortParams("created_at", "desc")
	if got := s.OrderString(); got != "created_at desc" {
		t.Errorf("expected 'created_at desc', got %q", got)
	}

	// 非法 order 应回退为 desc
	s2 := NewSortParams("id", "invalid")
	if s2.Order != "desc" {
		t.Errorf("expected order fallback to desc, got %q", s2.Order)
	}

	// 空 field 返回空串
	s3 := NewSortParams("", "asc")
	if got := s3.OrderString(); got != "" {
		t.Errorf("expected empty order string for empty field, got %q", got)
	}
}
