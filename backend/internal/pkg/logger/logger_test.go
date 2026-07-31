package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"wuchang-tongcheng/internal/pkg/config"

	"go.uber.org/zap"
)

// resetLogger 清除包级 logger/sugar 全局变量，保证用例间状态隔离。
// t.Cleanup 在用例结束后再次清空，避免后续用例受残留影响。
func resetLogger(t *testing.T) {
	t.Helper()
	logger = nil
	sugar = nil
	t.Cleanup(func() {
		logger = nil
		sugar = nil
	})
}

// logFilePath 在临时目录下生成日志文件路径，供文件输出型用例使用。
func logFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.log")
}

// readFileOrEmpty 读取文件全部内容；文件不存在时返回空串（便于断言“未写入”）。
func readFileOrEmpty(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("读取日志文件失败 %s: %v", path, err)
	}
	return string(b)
}

// --- Init 级别解析测试 ---

func TestInit_LevelParsingReturnsNoError(t *testing.T) {
	// 所有合法级别字符串与未知字符串均不应返回错误；
	// 未知级别回退到 info（见 logger.go default 分支）。
	levels := []string{"debug", "info", "warn", "error", "unknown-level"}
	for _, lvl := range levels {
		resetLogger(t)
		cfg := &config.LoggerConfig{Level: lvl, Filename: logFilePath(t), Console: false}
		if err := Init(cfg); err != nil {
			t.Errorf("Init(level=%q) 返回错误: %v", lvl, err)
		}
		if GetLogger() == nil {
			t.Errorf("Init(level=%q) 后 GetLogger() 为 nil", lvl)
		}
	}
}

// --- Init 文件输出测试 ---

func TestInit_FileOutputWritesInfoLog(t *testing.T) {
	resetLogger(t)
	path := logFilePath(t)
	cfg := &config.LoggerConfig{Level: "debug", Filename: path, Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	Info("file-info-message", zap.String("key", "val"))
	Sync()

	content := readFileOrEmpty(t, path)
	if !strings.Contains(content, "file-info-message") {
		t.Errorf("文件输出未包含 info 消息, got: %q", content)
	}
	if !strings.Contains(content, "val") {
		t.Errorf("文件输出未包含字段值 val, got: %q", content)
	}
}

func TestInit_LevelErrorFiltersLowerLevels(t *testing.T) {
	// level=error 时，Debug/Info/Warn 应被过滤不落盘，Error 应落盘。
	resetLogger(t)
	path := logFilePath(t)
	cfg := &config.LoggerConfig{Level: "error", Filename: path, Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	Debug("filtered-debug")
	Info("filtered-info")
	Warn("filtered-warn")
	Error("kept-error")
	Sync()

	content := readFileOrEmpty(t, path)
	if !strings.Contains(content, "kept-error") {
		t.Errorf("error 级别日志未落盘, got: %q", content)
	}
	if strings.Contains(content, "filtered-debug") {
		t.Errorf("debug 日志在 error 级别下不应落盘, got: %q", content)
	}
	if strings.Contains(content, "filtered-info") {
		t.Errorf("info 日志在 error 级别下不应落盘, got: %q", content)
	}
	if strings.Contains(content, "filtered-warn") {
		t.Errorf("warn 日志在 error 级别下不应落盘, got: %q", content)
	}
}

// --- Init 默认输出测试 ---

func TestInit_DefaultConsoleWhenNoOutputConfigured(t *testing.T) {
	// 既未配置 Filename 也未开启 Console 时，应回退到默认控制台 core，
	// 不应 panic 且 logger 被正确初始化。
	resetLogger(t)
	cfg := &config.LoggerConfig{Level: "info", Filename: "", Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	if GetLogger() == nil {
		t.Fatal("默认控制台回退后 GetLogger() 为 nil")
	}
	// 默认控制台输出不应 panic
	Info("default-console-message")
	Sync()
}

func TestInit_ConsoleAndFileBothActive(t *testing.T) {
	// 同时开启文件与控制台输出，文件应正常落盘。
	resetLogger(t)
	path := logFilePath(t)
	cfg := &config.LoggerConfig{Level: "debug", Filename: path, Console: true}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	Info("dual-output-message")
	Sync()

	content := readFileOrEmpty(t, path)
	if !strings.Contains(content, "dual-output-message") {
		t.Errorf("双输出模式下文件未落盘, got: %q", content)
	}
}

// --- GetLogger / GetSugar 懒加载测试 ---

func TestGetLogger_LazyInitWhenNotInitialized(t *testing.T) {
	resetLogger(t)
	// 全局为 nil 时，GetLogger 应懒加载一个 zap.NewProduction 实例。
	l := GetLogger()
	if l == nil {
		t.Fatal("GetLogger() 懒加载返回 nil")
	}
	// 再次调用应返回同一实例
	if GetLogger() != l {
		t.Error("GetLogger() 两次调用返回不同实例")
	}
}

func TestGetSugar_LazyInitWhenNotInitialized(t *testing.T) {
	resetLogger(t)
	s := GetSugar()
	if s == nil {
		t.Fatal("GetSugar() 懒加载返回 nil")
	}
}

func TestGetLogger_ReturnsInitializedInstanceAfterInit(t *testing.T) {
	resetLogger(t)
	cfg := &config.LoggerConfig{Level: "info", Filename: logFilePath(t), Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	// Init 后 GetLogger 应返回 Init 创建的实例（与 sugar 一致）
	l := GetLogger()
	if l == nil {
		t.Fatal("Init 后 GetLogger() 为 nil")
	}
	s := GetSugar()
	if s == nil {
		t.Fatal("GetSugar() 为 nil")
	}
}

// --- Sync 测试 ---

func TestSync_NoPanicWhenLoggerNil(t *testing.T) {
	resetLogger(t)
	// 全局为 nil 时 Sync 不应 panic
	Sync()
}

func TestSync_NoPanicAfterInit(t *testing.T) {
	resetLogger(t)
	cfg := &config.LoggerConfig{Level: "info", Filename: logFilePath(t), Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	Info("sync-test")
	Sync()
}

// --- 日志函数测试 ---

func TestLogFunctions_AllLevelsWriteToFile(t *testing.T) {
	resetLogger(t)
	path := logFilePath(t)
	cfg := &config.LoggerConfig{Level: "debug", Filename: path, Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	Debug("debug-msg")
	Info("info-msg")
	Warn("warn-msg")
	Error("error-msg")
	Sync()

	content := readFileOrEmpty(t, path)
	for _, want := range []string{"debug-msg", "info-msg", "warn-msg", "error-msg"} {
		if !strings.Contains(content, want) {
			t.Errorf("文件输出缺少 %q, got: %q", want, content)
		}
	}
}

func TestLogfFunctions_FormattedOutput(t *testing.T) {
	resetLogger(t)
	path := logFilePath(t)
	cfg := &config.LoggerConfig{Level: "debug", Filename: path, Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}

	Debugf("hello %s %d", "debug", 1)
	Infof("hello %s %d", "info", 2)
	Warnf("hello %s %d", "warn", 3)
	Errorf("hello %s %d", "error", 4)
	Sync()

	content := readFileOrEmpty(t, path)
	for _, want := range []string{"hello debug 1", "hello info 2", "hello warn 3", "hello error 4"} {
		if !strings.Contains(content, want) {
			t.Errorf("文件输出缺少格式化结果 %q, got: %q", want, content)
		}
	}
}

// --- WithField / WithFields 测试 ---

func TestWithField_ReturnsNonNilSugaredLogger(t *testing.T) {
	resetLogger(t)
	cfg := &config.LoggerConfig{Level: "info", Filename: logFilePath(t), Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	s := WithField("trace_id", "abc123")
	if s == nil {
		t.Fatal("WithField 返回 nil")
	}
}

func TestWithFields_ReturnsNonNilSugaredLogger(t *testing.T) {
	resetLogger(t)
	cfg := &config.LoggerConfig{Level: "info", Filename: logFilePath(t), Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	s := WithFields(map[string]interface{}{
		"trace_id": "abc123",
		"user_id":  42,
		"enabled":  true,
	})
	if s == nil {
		t.Fatal("WithFields 返回 nil")
	}
}

func TestWithFields_EmptyMapReturnsNonNil(t *testing.T) {
	// 空 map 不应 panic，应返回有效的 SugaredLogger。
	resetLogger(t)
	cfg := &config.LoggerConfig{Level: "info", Filename: logFilePath(t), Console: false}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init 失败: %v", err)
	}
	s := WithFields(map[string]interface{}{})
	if s == nil {
		t.Fatal("WithFields(空 map) 返回 nil")
	}
}
