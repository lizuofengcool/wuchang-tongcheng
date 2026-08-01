package seed

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// seed_test.go 补齐 pkg/seed 纯数据一致性单元测试（此前 0 覆盖）。
// 与 error_code_test.go（v2.15.0）同风格：plain testing，无 testify/DB/Redis/Docker 依赖，
// 仅校验包级种子数据定义的内部一致性，防止新增条目引入重复码/无效引用/格式错误。
//
// 说明：rolePermMap 中引用的部分权限码（如 ershou:read/order:create/rider:deliver 等）
// 由各业务模块插件路由或独立 SeedXxx 函数注册，并不在 permissionDefs 中定义，
// 故本测试仅校验码格式与角色内去重，不强制要求引用码全部命中 permissionDefs。

// isPermissionCodeFormat 校验权限码格式：非空且形如 "resource:action"。
func isPermissionCodeFormat(code string) bool {
	if code == "" {
		return false
	}
	idx := strings.Index(code, ":")
	if idx <= 0 || idx == len(code)-1 {
		return false
	}
	// 仅允许冒号前后各一段非空子串（不含额外冒号，避免 user:read:extra 类异常码）
	return strings.Index(code[idx+1:], ":") == -1
}

// --- permissionDefs 一致性 ---

func TestPermissionDefs_NoDuplicateCodes(t *testing.T) {
	seen := make(map[string]int, len(permissionDefs))
	for _, p := range permissionDefs {
		seen[p.Code]++
	}
	for code, n := range seen {
		if n > 1 {
			t.Errorf("permissionDefs 存在重复权限码 %q，出现 %d 次", code, n)
		}
	}
}

func TestPermissionDefs_FieldValidity(t *testing.T) {
	if len(permissionDefs) == 0 {
		t.Fatal("permissionDefs 不应为空")
	}
	validTypes := map[int]bool{1: true, 2: true, 3: true}
	for i, p := range permissionDefs {
		if p.Code == "" {
			t.Errorf("permissionDefs[%d].Code 不应为空", i)
		}
		if p.Name == "" {
			t.Errorf("permissionDefs[%d].Name 不应为空（code=%q）", i, p.Code)
		}
		if !isPermissionCodeFormat(p.Code) {
			t.Errorf("permissionDefs[%d].Code %q 不符合 resource:action 格式", i, p.Code)
		}
		if !validTypes[p.Type] {
			t.Errorf("permissionDefs[%d].Type=%d 非法（应为 1菜单/2按钮/3接口，code=%q）", i, p.Type, p.Code)
		}
	}
}

// --- businessRoleDefs 一致性 ---

func TestBusinessRoleDefs_NoDuplicateCodes(t *testing.T) {
	seen := make(map[string]int, len(businessRoleDefs))
	for _, r := range businessRoleDefs {
		seen[r.Code]++
	}
	for code, n := range seen {
		if n > 1 {
			t.Errorf("businessRoleDefs 存在重复角色码 %q，出现 %d 次", code, n)
		}
	}
}

func TestBusinessRoleDefs_FieldValidity(t *testing.T) {
	if len(businessRoleDefs) == 0 {
		t.Fatal("businessRoleDefs 不应为空")
	}
	sortSeen := make(map[int]int, len(businessRoleDefs))
	for i, r := range businessRoleDefs {
		if r.Code == "" {
			t.Errorf("businessRoleDefs[%d].Code 不应为空", i)
		}
		if r.Name == "" {
			t.Errorf("businessRoleDefs[%d].Name 不应为空（code=%q）", i, r.Code)
		}
		if r.Description == "" {
			t.Errorf("businessRoleDefs[%d].Description 不应为空（code=%q）", i, r.Code)
		}
		if r.Sort <= 0 {
			t.Errorf("businessRoleDefs[%d].Sort=%d 应为正数（code=%q）", i, r.Sort, r.Code)
		}
		sortSeen[r.Sort]++
	}
	for s, n := range sortSeen {
		if n > 1 {
			t.Errorf("businessRoleDefs 存在重复 Sort 值 %d，出现 %d 次（Sort 应唯一以保证菜单顺序稳定）", s, n)
		}
	}
}

// --- rolePermMap 一致性 ---

func TestRolePermMap_KeysAreValidBusinessRoles(t *testing.T) {
	// rolePermMap 的所有 key 必须命中 businessRoleDefs（super_admin 由 seedAdminRole
	// 单独处理并直通全部权限，不应出现在此映射中）。
	bizCodes := make(map[string]bool, len(businessRoleDefs))
	for _, r := range businessRoleDefs {
		bizCodes[r.Code] = true
	}
	for roleCode := range rolePermMap {
		if roleCode == "super_admin" {
			t.Errorf("rolePermMap 不应包含 super_admin 键（超管直通全部权限，由 seedAdminRole 单独处理）")
		}
		if !bizCodes[roleCode] {
			t.Errorf("rolePermMap 键 %q 未在 businessRoleDefs 中定义", roleCode)
		}
	}
}

func TestRolePermMap_NoDuplicatePermissionsPerRole(t *testing.T) {
	for roleCode, permCodes := range rolePermMap {
		seen := make(map[string]int, len(permCodes))
		for _, c := range permCodes {
			seen[c]++
		}
		for code, n := range seen {
			if n > 1 {
				t.Errorf("角色 %q 的权限列表存在重复权限码 %q，出现 %d 次", roleCode, code, n)
			}
		}
	}
}

func TestRolePermMap_PermissionCodeFormat(t *testing.T) {
	for roleCode, permCodes := range rolePermMap {
		if len(permCodes) == 0 {
			t.Errorf("角色 %q 的权限列表不应为空（至少应包含基础读取权限）", roleCode)
		}
		for _, code := range permCodes {
			if !isPermissionCodeFormat(code) {
				t.Errorf("角色 %q 引用权限码 %q 不符合 resource:action 格式", roleCode, code)
			}
		}
	}
}

func TestRolePermMap_EveryBusinessRoleHasEntry(t *testing.T) {
	// 每个业务角色都应在 rolePermMap 中有条目（即便为最小权限集），
	// 防止新增业务角色后遗漏权限矩阵配置。
	for _, r := range businessRoleDefs {
		if _, ok := rolePermMap[r.Code]; !ok {
			t.Errorf("业务角色 %q（%s）未在 rolePermMap 中配置权限矩阵", r.Code, r.Name)
		}
	}
}

// --- moduleDefs 一致性 ---

func TestModuleDefs_NoDuplicateNames(t *testing.T) {
	all := append(append([]moduleDef{}, middlewareModuleDefs...), businessModuleDefs...)
	seen := make(map[string]int, len(all))
	for _, m := range all {
		seen[m.Name]++
	}
	for name, n := range seen {
		if n > 1 {
			t.Errorf("moduleDefs 存在重复模块名 %q，出现 %d 次（modules 表 name 唯一）", name, n)
		}
	}
}

func TestModuleDefs_CategoryConsistency(t *testing.T) {
	for i, m := range middlewareModuleDefs {
		if m.Category != "middleware" {
			t.Errorf("middlewareModuleDefs[%d] %q Category=%q，应为 \"middleware\"", i, m.Name, m.Category)
		}
	}
	for i, m := range businessModuleDefs {
		if m.Category != "business" {
			t.Errorf("businessModuleDefs[%d] %q Category=%q，应为 \"business\"", i, m.Name, m.Category)
		}
	}
}

func TestModuleDefs_FieldValidity(t *testing.T) {
	all := append(append([]moduleDef{}, middlewareModuleDefs...), businessModuleDefs...)
	if len(all) == 0 {
		t.Fatal("moduleDefs 不应为空")
	}
	for i, m := range all {
		if m.Name == "" {
			t.Errorf("moduleDefs[%d].Name 不应为空", i)
		}
		if m.DisplayName == "" {
			t.Errorf("moduleDefs[%d].DisplayName 不应为空（name=%q）", i, m.Name)
		}
		if m.Icon == "" {
			t.Errorf("moduleDefs[%d].Icon 不应为空（name=%q）", i, m.Name)
		}
		// Dependencies 必须是合法 JSON 数组（[] 或 ["a","b"]）
		var deps []interface{}
		if err := json.Unmarshal([]byte(m.Dependencies), &deps); err != nil {
			t.Errorf("moduleDefs[%d].Dependencies %q 不是合法 JSON 数组：%v（name=%q）", i, m.Dependencies, err, m.Name)
		}
	}
}

func TestModuleDefs_DependenciesReferenceExistingModules(t *testing.T) {
	// 依赖的模块名应在已知模块清单中存在，防止配置拼写错误导致依赖失效。
	all := append(append([]moduleDef{}, middlewareModuleDefs...), businessModuleDefs...)
	known := make(map[string]bool, len(all))
	for _, m := range all {
		known[m.Name] = true
	}
	for _, m := range all {
		var deps []string
		if err := json.Unmarshal([]byte(m.Dependencies), &deps); err != nil {
			// 格式错误由 FieldValidity 覆盖，此处跳过
			continue
		}
		for _, dep := range deps {
			if dep == "" {
				t.Errorf("模块 %q 的 Dependencies 含空字符串依赖", m.Name)
				continue
			}
			if !known[dep] {
				t.Errorf("模块 %q 依赖 %q 未在已知模块清单中定义", m.Name, dep)
			}
		}
	}
}

// --- p0CronJobDefs 一致性 ---

func TestP0CronJobDefs_NoDuplicateJobKeys(t *testing.T) {
	// cron_jobs 表 (module_name, job_name) 唯一约束，种子数据不应重复。
	type jobKey struct{ module, job string }
	seen := make(map[jobKey]int, len(p0CronJobDefs))
	for _, j := range p0CronJobDefs {
		seen[jobKey{j.ModuleName, j.JobName}]++
	}
	for k, n := range seen {
		if n > 1 {
			t.Errorf("p0CronJobDefs 存在重复任务键 (module=%q, job=%q)，出现 %d 次", k.module, k.job, n)
		}
	}
}

func TestP0CronJobDefs_FieldValidity(t *testing.T) {
	if len(p0CronJobDefs) == 0 {
		t.Fatal("p0CronJobDefs 不应为空")
	}
	for i, j := range p0CronJobDefs {
		if j.ModuleName == "" {
			t.Errorf("p0CronJobDefs[%d].ModuleName 不应为空", i)
		}
		if j.JobName == "" {
			t.Errorf("p0CronJobDefs[%d].JobName 不应为空", i)
		}
		if j.Handler == "" {
			t.Errorf("p0CronJobDefs[%d].Handler 不应为空（job=%q）", i, j.JobName)
		}
		// CronExpr 6 字段：秒 分 时 日 月 周
		fields := strings.Fields(j.CronExpr)
		if len(fields) != 6 {
			t.Errorf("p0CronJobDefs[%d].CronExpr %q 应为 6 字段（秒 分 时 日 月 周），实际 %d 字段（job=%q）",
				i, j.CronExpr, len(fields), j.JobName)
		}
		// Params 必须是合法 JSON（{} 或 {...}）
		var params interface{}
		if err := json.Unmarshal([]byte(j.Params), &params); err != nil {
			t.Errorf("p0CronJobDefs[%d].Params %q 不是合法 JSON：%v（job=%q）", i, j.Params, err, j.JobName)
		}
		if j.TimeoutSeconds <= 0 {
			t.Errorf("p0CronJobDefs[%d].TimeoutSeconds=%d 应为正数（job=%q）", i, j.TimeoutSeconds, j.JobName)
		}
		if j.MaxRetry < 0 {
			t.Errorf("p0CronJobDefs[%d].MaxRetry=%d 不应为负数（job=%q）", i, j.MaxRetry, j.JobName)
		}
	}
}

// --- 跨结构聚合校验 ---

func TestSeedDataCounts_NonZero(t *testing.T) {
	// 防御性断言：种子数据规模非零，避免重构时意外清空关键定义。
	if len(permissionDefs) < 30 {
		t.Errorf("permissionDefs 数量 %d 过少，预期至少 30 个权限码", len(permissionDefs))
	}
	if len(businessRoleDefs) != 9 {
		t.Errorf("businessRoleDefs 数量=%d，预期 9 个业务角色（super_admin 单独处理）", len(businessRoleDefs))
	}
	if len(rolePermMap) != len(businessRoleDefs) {
		t.Errorf("rolePermMap 数量=%d 与 businessRoleDefs 数量=%d 不一致，每个业务角色都应有权限矩阵条目",
			len(rolePermMap), len(businessRoleDefs))
	}
	if len(middlewareModuleDefs) != 12 {
		t.Errorf("middlewareModuleDefs 数量=%d，预期 12 个中台模块", len(middlewareModuleDefs))
	}
	if len(businessModuleDefs) != 15 {
		t.Errorf("businessModuleDefs 数量=%d，预期 15 个垂直业务模块", len(businessModuleDefs))
	}
}

// TestSeedData_StableSortedBySort 提取角色 Sort 值并校验降序排列，
// 保证菜单展示顺序与设计一致（Sort 越大越靠前）。
func TestBusinessRoleDefs_SortDescending(t *testing.T) {
	sorts := make([]int, 0, len(businessRoleDefs))
	for _, r := range businessRoleDefs {
		sorts = append(sorts, r.Sort)
	}
	sortedDesc := make([]int, len(sorts))
	copy(sortedDesc, sorts)
	sort.Sort(sort.Reverse(sort.IntSlice(sortedDesc)))
	for i := range sorts {
		if sorts[i] != sortedDesc[i] {
			t.Errorf("businessRoleDefs Sort 未按降序排列，got %v，期望 %v", sorts, sortedDesc)
			break
		}
	}
}
