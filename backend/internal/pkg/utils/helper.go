package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// MD5 MD5加密
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// RandomNumber 生成随机数字字符串
func RandomNumber(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// IsEmail 验证邮箱格式
func IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsMobile 验证手机号格式（中国大陆）
func IsMobile(mobile string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, mobile)
	return matched
}

// IsIDCard 验证身份证号格式（中国大陆）
func IsIDCard(idCard string) bool {
	pattern := `^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`
	matched, _ := regexp.MatchString(pattern, idCard)
	return matched
}

// IsURL 验证URL格式
func IsURL(url string) bool {
	pattern := `^(https?://)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*/?$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// IsIP 验证IP地址格式
func IsIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// Substring 字符串截取
func Substring(str string, start, length int) string {
	runes := []rune(str)
	if start < 0 {
		start = 0
	}
	if start >= len(runes) {
		return ""
	}
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// Contains 判断字符串是否包含子串
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// StartsWith 判断字符串是否以子串开头
func StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith 判断字符串是否以子串结尾
func EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Trim 去除字符串首尾空格
func Trim(str string) string {
	return strings.TrimSpace(str)
}

// Join 字符串拼接
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// Split 字符串分割
func Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// ToUpper 转大写
func ToUpper(str string) string {
	return strings.ToUpper(str)
}

// ToLower 转小写
func ToLower(str string) string {
	return strings.ToLower(str)
}

// FirstUpper 首字母大写
func FirstUpper(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + str[1:]
}

// FirstLower 首字母小写
func FirstLower(str string) string {
	if str == "" {
		return ""
	}
	return strings.ToLower(str[:1]) + str[1:]
}

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return t.Format(layout)
}

// ParseTime 解析时间
func ParseTime(str, layout string) (time.Time, error) {
	if layout == "" {
		layout = "2006-01-02 15:04:05"
	}
	return time.ParseInLocation(layout, str, time.Local)
}

// Now 当前时间
func Now() time.Time {
	return time.Now()
}

// NowString 当前时间字符串
func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Today 今天日期字符串
func Today() string {
	return time.Now().Format("2006-01-02")
}

// GetFileExt 获取文件扩展名
func GetFileExt(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

// GetFileName 获取文件名（不含扩展名）
func GetFileName(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

// FileExists 判断文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

// IsDir 判断是否是目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileSize 获取文件大小（字节）
func FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// UniqueID 生成唯一ID（基于时间戳+随机数）
func UniqueID() string {
	return fmt.Sprintf("%d%s", time.Now().UnixNano(), RandomString(6))
}

// SnowflakeID 生成雪花ID（简化版）
// 注意：生产环境建议使用真正的雪花算法实现
func SnowflakeID() int64 {
	return time.Now().UnixNano()
}

// InSlice 判断元素是否在切片中
func InSlice[T comparable](item T, slice []T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// UniqueSlice 切片去重
func UniqueSlice[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// ReverseSlice 反转切片
func ReverseSlice[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Min 返回较小值
func Min[T int | int64 | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 返回较大值
func Max[T int | int64 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Abs 返回绝对值
func Abs[T int | int64 | float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
