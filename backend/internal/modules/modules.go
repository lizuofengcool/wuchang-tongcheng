// Package modules 模块聚合包
// 通过导入此包自动注册所有业务模块
// 新增模块时只需在此文件添加一行 import _
package modules

import (
	// 系统模块（基础模块，需先注册）
	_ "wuchang-tongcheng/internal/modules/region"
	_ "wuchang-tongcheng/internal/modules/setting"
	_ "wuchang-tongcheng/internal/modules/file"
	_ "wuchang-tongcheng/internal/modules/permission"
	_ "wuchang-tongcheng/internal/modules/user"

	// 业务模块
	_ "wuchang-tongcheng/internal/modules/category"
	_ "wuchang-tongcheng/internal/modules/news"
	_ "wuchang-tongcheng/internal/modules/shop"

	_ "wuchang-tongcheng/internal/modules/groupbuy"

	// 新增模块在此添加
)
