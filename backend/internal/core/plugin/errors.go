package plugin

import "fmt"

// 插件系统错误定义
var (
	// ErrPluginAlreadyExists 插件已存在错误
	ErrPluginAlreadyExists = fmt.Errorf("plugin already exists")
	// ErrPluginNotFound 插件未找到错误
	ErrPluginNotFound = fmt.Errorf("plugin not found")
)

// PluginInitError 插件初始化错误
type PluginInitError struct {
	PluginName string
	Err        error
}

func (e *PluginInitError) Error() string {
	return fmt.Sprintf("plugin [%s] init failed: %v", e.PluginName, e.Err)
}

func (e *PluginInitError) Unwrap() error {
	return e.Err
}

// PluginCloseError 插件关闭错误
type PluginCloseError struct {
	PluginName string
	Err        error
}

func (e *PluginCloseError) Error() string {
	return fmt.Sprintf("plugin [%s] close failed: %v", e.PluginName, e.Err)
}

func (e *PluginCloseError) Unwrap() error {
	return e.Err
}
