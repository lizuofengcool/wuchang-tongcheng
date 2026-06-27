package service

import "testing"

func TestHashPassword_AndCheckPassword(t *testing.T) {
	pw := "mySecret123"
	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword err = %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	// 哈希不能等于明文
	if hash == pw {
		t.Fatal("hash should not equal plaintext")
	}
	// 正确密码校验通过
	if !CheckPassword(pw, hash) {
		t.Error("CheckPassword with correct password should return true")
	}
	// 错误密码校验失败
	if CheckPassword("wrongPassword", hash) {
		t.Error("CheckPassword with wrong password should return false")
	}
}

func TestCheckPassword_EmptyInputs(t *testing.T) {
	// 空明文 + 任意哈希应失败
	hash, _ := HashPassword("real")
	if CheckPassword("", hash) {
		t.Error("CheckPassword with empty password should return false")
	}
	// 空哈希应失败（不 panic）
	if CheckPassword("anything", "") {
		t.Error("CheckPassword with empty hash should return false")
	}
}

func TestHashPassword_DifferentSalts(t *testing.T) {
	// bcrypt 每次哈希使用不同 salt，相同密码生成不同哈希
	h1, _ := HashPassword("samepassword")
	h2, _ := HashPassword("samepassword")
	if h1 == h2 {
		t.Error("two hashes of same password should differ (bcrypt salt)")
	}
	// 但两个哈希都能校验通过同一密码
	if !CheckPassword("samepassword", h1) || !CheckPassword("samepassword", h2) {
		t.Error("both hashes should validate against same password")
	}
}
