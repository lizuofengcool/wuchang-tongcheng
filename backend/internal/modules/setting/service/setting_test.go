package service

import "testing"

func TestParseValue_String(t *testing.T) {
	if got := parseValue("hello", "string"); got != "hello" {
		t.Errorf("parseValue(string) = %v, want hello", got)
	}
	// 未知类型回退为 string
	if got := parseValue("hello", "unknown"); got != "hello" {
		t.Errorf("parseValue(unknown) = %v, want hello", got)
	}
}

func TestParseValue_Number(t *testing.T) {
	got := parseValue("3.14", "number")
	f, ok := got.(float64)
	if !ok {
		t.Fatalf("parseValue(number) returned %T, want float64", got)
	}
	if f != 3.14 {
		t.Errorf("parseValue(number) = %v, want 3.14", f)
	}
	// 非法数字回退为原字符串
	if got := parseValue("abc", "number"); got != "abc" {
		t.Errorf("parseValue(invalid number) = %v, want fallback string", got)
	}
}

func TestParseValue_Bool(t *testing.T) {
	truthy := []string{"true", "TRUE", "1", "yes", "on", "True"}
	for _, v := range truthy {
		if got := parseValue(v, "bool"); got != true {
			t.Errorf("parseValue(%q, bool) = %v, want true", v, got)
		}
	}
	falsy := []string{"false", "0", "no", "off", ""}
	for _, v := range falsy {
		if got := parseValue(v, "bool"); got != false {
			t.Errorf("parseValue(%q, bool) = %v, want false", v, got)
		}
	}
	// 非法 bool 回退为原字符串
	if got := parseValue("maybe", "bool"); got != "maybe" {
		t.Errorf("parseValue(maybe, bool) = %v, want fallback string", got)
	}
}

func TestParseValue_JSON(t *testing.T) {
	// 对象
	got := parseValue(`{"a":1,"b":"x"}`, "json")
	m, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("parseValue(json object) returned %T, want map", got)
	}
	if m["a"] != float64(1) {
		t.Errorf("json a = %v, want 1", m["a"])
	}
	// 数组
	gotArr := parseValue(`[1,2,3]`, "json")
	if _, ok := gotArr.([]interface{}); !ok {
		t.Fatalf("parseValue(json array) returned %T, want slice", gotArr)
	}
	// 非法 JSON 回退为原字符串
	if got := parseValue("{invalid", "json"); got != "{invalid" {
		t.Errorf("parseValue(invalid json) = %v, want fallback string", got)
	}
}

func TestValidateValue_Number(t *testing.T) {
	if err := validateValue("3.14", "number"); err != nil {
		t.Errorf("validateValue(3.14, number) err = %v, want nil", err)
	}
	if err := validateValue("", "number"); err != nil {
		t.Errorf("validateValue(empty, number) err = %v, want nil (empty allowed)", err)
	}
	if err := validateValue("abc", "number"); err == nil {
		t.Error("validateValue(abc, number) should return error")
	}
}

func TestValidateValue_Bool(t *testing.T) {
	for _, v := range []string{"true", "false", "1", "0", "yes", "no", "on", "off", ""} {
		if err := validateValue(v, "bool"); err != nil {
			t.Errorf("validateValue(%q, bool) err = %v, want nil", v, err)
		}
	}
	if err := validateValue("maybe", "bool"); err == nil {
		t.Error("validateValue(maybe, bool) should return error")
	}
}

func TestValidateValue_JSON(t *testing.T) {
	if err := validateValue(`{"a":1}`, "json"); err != nil {
		t.Errorf("validateValue(valid json) err = %v, want nil", err)
	}
	if err := validateValue(`{bad`, "json"); err == nil {
		t.Error("validateValue(invalid json) should return error")
	}
}

func TestValidateValue_String_NoOp(t *testing.T) {
	// string 类型不校验，任意值通过
	if err := validateValue("anything", "string"); err != nil {
		t.Errorf("validateValue(anything, string) err = %v, want nil", err)
	}
}
