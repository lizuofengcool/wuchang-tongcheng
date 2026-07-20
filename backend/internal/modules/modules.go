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

	_ "wuchang-tongcheng/internal/modules/ershou"

	_ "wuchang-tongcheng/internal/modules/groupbuy"

	// P1 阶段2 第1批：同城业务垂直模块（job/house/car）
	_ "wuchang-tongcheng/internal/modules/job"
	_ "wuchang-tongcheng/internal/modules/house"
	_ "wuchang-tongcheng/internal/modules/car"

	// P1 阶段3 第2批：同城业务垂直模块（love/pinche/linggong/dh114）
	_ "wuchang-tongcheng/internal/modules/love"
	_ "wuchang-tongcheng/internal/modules/pinche"
	_ "wuchang-tongcheng/internal/modules/linggong"
	_ "wuchang-tongcheng/internal/modules/dh114"

	// P1 中台精简版（ershou 模块依赖）
	_ "wuchang-tongcheng/internal/modules/pay"
	_ "wuchang-tongcheng/internal/modules/im"
	_ "wuchang-tongcheng/internal/modules/material"
	_ "wuchang-tongcheng/internal/modules/risk"
	_ "wuchang-tongcheng/internal/modules/ai"

	// 招聘求职模块
	_ "wuchang-tongcheng/internal/modules/job"

	// 模块管理（注册表 + 开关 + 元信息同步，需最后注册，确保同步时其他模块已就绪）
	_ "wuchang-tongcheng/internal/modules/module"

	// 新增模块在此添加
)
