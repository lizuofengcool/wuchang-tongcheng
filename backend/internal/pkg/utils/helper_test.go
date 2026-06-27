package utils

import (
	"strings"
	"testing"
)

func TestMD5(t *testing.T) {
	// 已知 MD5 值
	if got := MD5("hello"); got != "5d41402abc4b2a76b9719d911017c592" {
		t.Errorf("MD5(hello) = %q, want known hash", got)
	}
	if MD5("") != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Error("MD5 of empty string mismatch")
	}
}

func TestIsEmail(t *testing.T) {
	valid := []string{"a@b.com", "user.name+tag@example.co.uk", "test_123@sub.domain.org"}
	for _, e := range valid {
		if !IsEmail(e) {
			t.Errorf("IsEmail(%q) = false, want true", e)
		}
	}
	invalid := []string{"", "abc", "abc@", "@b.com", "a@b", "a@.com"}
	for _, e := range invalid {
		if IsEmail(e) {
			t.Errorf("IsEmail(%q) = true, want false", e)
		}
	}
}

func TestIsMobile(t *testing.T) {
	valid := []string{"13800138000", "15912345678", "17098765432"}
	for _, m := range valid {
		if !IsMobile(m) {
			t.Errorf("IsMobile(%q) = false, want true", m)
		}
	}
	invalid := []string{"", "12345678901", "1380013800", "23800138000", "abc"}
	for _, m := range invalid {
		if IsMobile(m) {
			t.Errorf("IsMobile(%q) = true, want false", m)
		}
	}
}

func TestSubstring(t *testing.T) {
	if got := Substring("hello世界", 0, 5); got != "hello" {
		t.Errorf("Substring = %q, want %q", got, "hello")
	}
	// 中文字符按 rune 计数
	if got := Substring("hello世界", 5, 2); got != "世界" {
		t.Errorf("Substring rune = %q, want %q", got, "世界")
	}
	// 越界自动截断
	if got := Substring("abc", 1, 100); got != "bc" {
		t.Errorf("Substring overflow = %q, want %q", got, "bc")
	}
	// start 越界返回空
	if got := Substring("abc", 10, 2); got != "" {
		t.Errorf("Substring start overflow = %q, want empty", got)
	}
}

func TestFirstUpperFirstLower(t *testing.T) {
	if FirstUpper("hello") != "Hello" {
		t.Error("FirstUpper failed")
	}
	if FirstUpper("") != "" {
		t.Error("FirstUpper of empty string should be empty")
	}
	if FirstLower("Hello") != "hello" {
		t.Error("FirstLower failed")
	}
	if FirstLower("") != "" {
		t.Error("FirstLower of empty string should be empty")
	}
}

func TestInSlice(t *testing.T) {
	if !InSlice(3, []int{1, 2, 3}) {
		t.Error("InSlice should find 3")
	}
	if InSlice(4, []int{1, 2, 3}) {
		t.Error("InSlice should not find 4")
	}
	if !InSlice("a", []string{"a", "b"}) {
		t.Error("InSlice should find 'a'")
	}
}

func TestUniqueSlice(t *testing.T) {
	in := []int{1, 2, 2, 3, 3, 3, 4}
	out := UniqueSlice(in)
	if len(out) != 4 {
		t.Errorf("UniqueSlice len = %d, want 4", len(out))
	}
	want := []int{1, 2, 3, 4}
	for i, v := range want {
		if out[i] != v {
			t.Errorf("UniqueSlice[%d] = %d, want %d", i, out[i], v)
		}
	}
}

func TestMinMaxAbs(t *testing.T) {
	if Min(3, 5) != 3 {
		t.Error("Min failed")
	}
	if Max(3, 5) != 5 {
		t.Error("Max failed")
	}
	if Abs(-5) != 5 {
		t.Error("Abs of negative failed")
	}
	if Abs(5) != 5 {
		t.Error("Abs of positive failed")
	}
	if Abs(0) != 0 {
		t.Error("Abs of zero failed")
	}
}

func TestFormatFileSize(t *testing.T) {
	cases := []struct {
		size int64
		want string
	}{
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
	}
	for _, c := range cases {
		got := FormatFileSize(c.size)
		if got != c.want {
			t.Errorf("FormatFileSize(%d) = %q, want %q", c.size, got, c.want)
		}
	}
}

func TestRandomString(t *testing.T) {
	s := RandomString(16)
	if len(s) != 16 {
		t.Errorf("RandomString(16) len = %d, want 16", len(s))
	}
	// 两次生成应不同（概率上）
	s2 := RandomString(16)
	if s == s2 {
		t.Error("RandomString produced identical output twice (statistically near-impossible)")
	}
}

func TestRandomNumber(t *testing.T) {
	n := RandomNumber(6)
	if len(n) != 6 {
		t.Errorf("RandomNumber(6) len = %d, want 6", len(n))
	}
	if !strings.ContainsAny(n, "0123456789") {
		t.Error("RandomNumber should contain digits")
	}
	for _, c := range n {
		if c < '0' || c > '9' {
			t.Errorf("RandomNumber contains non-digit %q", c)
		}
	}
}

func TestGetFileExt(t *testing.T) {
	if got := GetFileExt("photo.JPG"); got != ".jpg" {
		t.Errorf("GetFileExt = %q, want .jpg (lowercased)", got)
	}
	if got := GetFileExt("noext"); got != "" {
		t.Errorf("GetFileExt of no-extension = %q, want empty", got)
	}
}
