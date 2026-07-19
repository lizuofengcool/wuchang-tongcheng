// Package seed 种子数据初始化
// 在服务启动、AutoMigrate 之后执行，幂等：已存在的数据不会重复创建
package seed

import (
	"fmt"

	categoryModel "wuchang-tongcheng/internal/modules/category/model"
	permModel "wuchang-tongcheng/internal/modules/permission/model"
	regionModel "wuchang-tongcheng/internal/modules/region/model"
	settingModel "wuchang-tongcheng/internal/modules/setting/model"
	userModel "wuchang-tongcheng/internal/modules/user/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// permDef 权限定义
type permDef struct {
	Code string
	Name string
	Type int // 1菜单 2按钮 3接口
}

// 全部权限码（与各插件路由上 RequirePermission 使用的 code 一一对应）
var permissionDefs = []permDef{
	// 用户管理
	{"user:read", "用户查看", 3},
	{"user:create", "用户新建", 3},
	{"user:update", "用户更新", 3},
	{"user:delete", "用户删除", 3},
	{"user:reset_password", "重置用户密码", 3},
	// 地区管理
	{"region:read", "地区查看", 3},
	{"region:create", "地区新建", 3},
	{"region:update", "地区更新", 3},
	{"region:delete", "地区删除", 3},
	// 分类管理
	{"category:read", "分类查看", 3},
	{"category:create", "分类新建", 3},
	{"category:update", "分类更新", 3},
	{"category:delete", "分类删除", 3},
	// 同城头条
	{"news:read", "头条查看", 3},
	{"news:create", "头条新建", 3},
	{"news:update", "头条更新", 3},
	{"news:delete", "头条删除", 3},
	// 角色与权限
	{"role:read", "角色查看", 3},
	{"role:create", "角色新建", 3},
	{"role:update", "角色更新", 3},
	{"role:delete", "角色删除", 3},
	{"permission:read", "权限查看", 3},
	{"permission:create", "权限新建", 3},
	{"permission:update", "权限更新", 3},
	{"permission:delete", "权限删除", 3},
	{"permission:assign", "权限分配", 3},
	// 系统设置
	{"setting:read", "设置查看", 3},
	{"setting:create", "设置新建", 3},
	{"setting:update", "设置更新", 3},
	{"setting:delete", "设置删除", 3},
	// 文件管理
	{"file:upload", "文件上传", 3},
	{"file:read", "文件查看", 3},
	{"file:delete", "文件删除", 3},
	// 商家管理
	{"shop:read", "商家查看", 3},
	{"shop:create", "商家新建", 3},
	{"shop:update", "商家更新", 3},
	{"shop:delete", "商家删除", 3},
	{"shop:audit", "商家审核", 3},
	// 团购优惠券
	{"groupbuy:read", "团购查看", 3},
	{"groupbuy:create", "团购新建", 3},
	{"groupbuy:update", "团购更新", 3},
	{"groupbuy:delete", "团购删除", 3},
	{"groupbuy:audit", "团购审核", 3},
	{"groupbuy:verify", "团购核销", 3},
	// 10 角色权限矩阵新增权限码（依据需求文档 1.9.3）
	// C端
	{"ershou:read", "信息浏览（C端）", 3},
	{"ershou:publish", "信息发布（C端）", 3},
	{"order:create", "下单支付（C端）", 3},
	{"vip:privilege", "VIP特权（C端）", 3},
	// B端
	{"shop:manage", "店铺管理（B端）", 3},
	{"product:manage", "商品管理（B端）", 3},
	{"order:verify", "订单核销（B端）", 3},
	{"marketing:manage", "营销活动（B端）", 3},
	// D端
	{"rider:deliver", "接单配送（D端）", 3},
	{"rider:withdraw", "收益提现（D端）", 3},
	// A端
	{"agent:station", "分站管理（A端）", 3},
	{"agent:profit", "分润统计（A端）", 3},
	// M端
	{"content:audit", "内容审核（M端）", 3},
	{"user:manage", "用户管理（M端）", 3},
	{"system:config", "系统设置（M端超管）", 3},
	{"finance:reconcile", "财务对账（M端）", 3},
}

// roleDef 角色定义（依据需求文档 1.9.1 角色定义 - 10 角色）
type roleDef struct {
	Name        string
	Code        string
	Description string
	Sort        int
}

// 全部业务角色（super_admin 由 seedAdminRole 单独处理并赋予全部权限）
var businessRoleDefs = []roleDef{
	{"游客", "guest", "未登录用户，仅可浏览公开内容", 100},
	{"普通用户", "user", "注册用户，可发布信息/下单/支付", 90},
	{"VIP用户", "vip", "付费会员，享会员特权/折扣", 80},
	{"商家", "merchant", "入驻商家，可管理店铺/商品/订单/营销", 70},
	{"店员", "clerk", "商家员工，仅限订单核销/查看", 60},
	{"骑手", "rider", "配送员，接单/配送/收益", 50},
	{"代理商", "agent", "城市代理，分站管理/分润", 40},
	{"内容审核员", "auditor", "内容审核/举报处理", 30},
	{"平台运营", "operator", "除超管外全部管理权限", 20},
}

// rolePermMap 角色-权限映射（依据需求文档 1.9.2 权限矩阵）
// super_admin 直通全部权限（middleware.IsSuperAdmin），无需在此分配
var rolePermMap = map[string][]string{
	"guest":    {"ershou:read"},
	"user":     {"ershou:read", "ershou:publish", "order:create", "file:upload", "file:read"},
	"vip":      {"ershou:read", "ershou:publish", "order:create", "vip:privilege", "file:upload", "file:read"},
	"merchant": {"ershou:read", "ershou:publish", "order:create", "shop:manage", "product:manage", "order:verify", "marketing:manage", "shop:read", "file:upload", "file:read"},
	"clerk":    {"ershou:read", "order:verify", "shop:read", "file:read"},
	"rider":    {"ershou:read", "rider:deliver", "rider:withdraw", "file:read"},
	"agent":    {"ershou:read", "agent:station", "agent:profit", "shop:audit", "region:read", "category:read"},
	"auditor":  {"ershou:read", "content:audit", "news:read", "shop:read"},
	"operator": {
		// 除 system:config 外的全部权限
		"ershou:read", "ershou:publish", "order:create", "vip:privilege",
		"shop:manage", "product:manage", "order:verify", "marketing:manage",
		"rider:deliver", "rider:withdraw",
		"agent:station", "agent:profit",
		"content:audit", "user:manage", "shop:audit", "finance:reconcile",
		"user:read", "user:create", "user:update", "user:delete", "user:reset_password",
		"region:read", "region:create", "region:update", "region:delete",
		"category:read", "category:create", "category:update", "category:delete",
		"news:read", "news:create", "news:update", "news:delete",
		"role:read", "permission:read",
		"setting:read", "setting:create", "setting:update",
		"file:upload", "file:read", "file:delete",
		"shop:read", "shop:create", "shop:update", "shop:delete",
		"groupbuy:read", "groupbuy:create", "groupbuy:update", "groupbuy:delete", "groupbuy:audit", "groupbuy:verify",
	},
}

// Run 执行种子数据初始化（幂等）
func Run(db *gorm.DB) error {
	if err := seedRegions(db); err != nil {
		return err
	}
	if err := seedPermissions(db); err != nil {
		return err
	}
	// 迁移历史 admin 角色 code → super_admin（依据需求文档 1.9.1）
	if err := migrateAdminToSuperAdmin(db); err != nil {
		return err
	}
	if err := seedAdminRole(db); err != nil {
		return err
	}
	if err := seedBusinessRoles(db); err != nil {
		return err
	}
	if err := seedRolePermissions(db); err != nil {
		return err
	}
	if err := seedAdminUser(db); err != nil {
		return err
	}
	if err := seedCategories(db); err != nil {
		return err
	}
	if err := seedSettings(db); err != nil {
		return err
	}
	// P0 新增：模块注册表 + 定时任务调度中心
	// 依赖 001_p0_baseline.sql 已执行（modules / cron_jobs 表已存在）
	if err := SeedModules(db); err != nil {
		return err
	}
	if err := SeedCronJobs(db); err != nil {
		return err
	}
	return nil
}

// migrateAdminToSuperAdmin 迁移历史 admin 角色 code 为 super_admin
// 一次性迁移：如果存在 code="admin" 的角色，将其 code 改为 "super_admin"
// 幂等：若已存在 code="super_admin"，则删除重复的 admin 角色（保留 super_admin）
func migrateAdminToSuperAdmin(db *gorm.DB) error {
	var adminRole permModel.Role
	err := db.Where("code = ?", "admin").First(&adminRole).Error
	if err == gorm.ErrRecordNotFound {
		// 无历史 admin 角色，无需迁移
		return nil
	}
	if err != nil {
		return err
	}

	// 检查是否已存在 super_admin
	var superAdminRole permModel.Role
	err = db.Where("code = ?", "super_admin").First(&superAdminRole).Error
	if err == nil {
		// super_admin 已存在，删除重复的 admin 角色（其 user_role 关联需迁移到 super_admin）
		// 把 user_roles 中指向 adminRole.ID 的记录迁移到 superAdminRole.ID
		if err := db.Model(&permModel.UserRole{}).
			Where("role_id = ?", adminRole.ID).
			Update("role_id", superAdminRole.ID).Error; err != nil {
			return err
		}
		// 把 role_permissions 中指向 adminRole.ID 的记录迁移到 superAdminRole.ID
		if err := db.Model(&permModel.RolePermission{}).
			Where("role_id = ?", adminRole.ID).
			Update("role_id", superAdminRole.ID).Error; err != nil {
			return err
		}
		// 删除重复的 admin 角色
		return db.Delete(&adminRole).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// 不存在 super_admin，直接把 admin 改名为 super_admin
	return db.Model(&adminRole).Update("code", "super_admin").Error
}

// seedBusinessRoles 创建 9 个业务角色（super_admin 由 seedAdminRole 处理）
// 幂等：按 code 唯一判断是否已存在
func seedBusinessRoles(db *gorm.DB) error {
	for _, r := range businessRoleDefs {
		var found permModel.Role
		err := db.Where("code = ?", r.Code).First(&found).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		role := permModel.Role{
			Name:        r.Name,
			Code:        r.Code,
			Description: r.Description,
			Sort:        r.Sort,
			Status:      1,
		}
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedRolePermissions 按 1.9.2 权限矩阵为各角色分配权限
// 幂等：已存在的 role-permission 关联跳过
func seedRolePermissions(db *gorm.DB) error {
	for roleCode, permCodes := range rolePermMap {
		// 查角色
		var role permModel.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			return err
		}
		// 查权限
		var perms []permModel.Permission
		if err := db.Where("code IN ?", permCodes).Find(&perms).Error; err != nil {
			return err
		}
		// 分配（幂等）
		for _, p := range perms {
			var rp permModel.RolePermission
			err := db.Where("role_id = ? AND permission_id = ?", role.ID, p.ID).First(&rp).Error
			if err == nil {
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			rp = permModel.RolePermission{RoleID: role.ID, PermissionID: p.ID}
			if err := db.Create(&rp).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// 默认地区ID（五常市，与 seedRegions 中保持一致）
const defaultRegionID = 3

// seedCategories 默认分类：10 个标准顶级分类，全部归属五常市
// 幂等：按 (region_id, name) 唯一判断是否已存在
func seedCategories(db *gorm.DB) error {
	categories := []categoryModel.Category{
		{Name: "二手转让", Icon: "ShoppingBag", ParentID: 0, Level: 1, Sort: 100, Status: 1},
		{Name: "招聘求职", Icon: "Briefcase", ParentID: 0, Level: 1, Sort: 90, Status: 1},
		{Name: "房屋租售", Icon: "House", ParentID: 0, Level: 1, Sort: 80, Status: 1},
		{Name: "同城交友", Icon: "User", ParentID: 0, Level: 1, Sort: 70, Status: 1},
		{Name: "生活服务", Icon: "Service", ParentID: 0, Level: 1, Sort: 60, Status: 1},
		{Name: "车辆买卖", Icon: "Van", ParentID: 0, Level: 1, Sort: 50, Status: 1},
		{Name: "宠物服务", Icon: "Charm", ParentID: 0, Level: 1, Sort: 40, Status: 1},
		{Name: "教育培训", Icon: "Reading", ParentID: 0, Level: 1, Sort: 30, Status: 1},
		{Name: "美食外卖", Icon: "Food", ParentID: 0, Level: 1, Sort: 20, Status: 1},
		{Name: "休闲娱乐", Icon: "Football", ParentID: 0, Level: 1, Sort: 10, Status: 1},
	}
	for i := range categories {
		c := categories[i]
		c.RegionID = defaultRegionID
		var found categoryModel.Category
		err := db.Where("region_id = ? AND name = ?", c.RegionID, c.Name).First(&found).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedSettings 默认系统设置：站点基本信息（site 组），归属五常市
// 幂等：按 (region_id, group, key) 唯一判断是否已存在
// 注意：group/key 是 PostgreSQL 保留字，GORM Where 中需用双引号引用（非反引号）
func seedSettings(db *gorm.DB) error {
	settings := []settingModel.Setting{
		{Group: "site", Key: "site_name", Value: "近享同城", ValueType: "string", Description: "站点名称", Sort: 100},
		{Group: "site", Key: "site_description", Value: "五常本地生活服务平台，提供分类信息、同城头条、商家服务、团购优惠等一站式服务", ValueType: "string", Description: "站点描述", Sort: 90},
		{Group: "site", Key: "site_address", Value: "黑龙江省哈尔滨市五常市", ValueType: "string", Description: "联系地址", Sort: 80},
		{Group: "site", Key: "site_phone", Value: "", ValueType: "string", Description: "联系电话", Sort: 70},
		{Group: "site", Key: "site_email", Value: "", ValueType: "string", Description: "联系邮箱", Sort: 60},
		{Group: "site", Key: "site_icp", Value: "", ValueType: "string", Description: "ICP备案号", Sort: 50},
	}
	for i := range settings {
		s := settings[i]
		s.RegionID = defaultRegionID
		var found settingModel.Setting
		// 使用 map[string]interface{} 让 GORM 自动处理列名引用，避免 PostgreSQL 保留字 group/key 的语法问题
		err := db.Where(map[string]interface{}{
			"region_id": s.RegionID,
			"group":     s.Group,
			"key":       s.Key,
		}).First(&found).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := db.Create(&s).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedRegions 地区：按顺序写入，确保五常市 id=3（与 DefaultRegionID 对应）
// 结构：黑龙江省(id=1) → 哈尔滨市(id=2) → 五常市(id=3) → 4乡镇(id=4-7)
func seedRegions(db *gorm.DB) error {
	regions := []regionModel.Region{
		// Level 1 省
		{Name: "黑龙江省", Code: "230000", Level: 1, Sort: 1, Status: 1},
		// Level 2 市
		{Name: "哈尔滨市", Code: "230100", Level: 2, Sort: 1, Status: 1},
		// Level 3 县级市
		{Name: "五常市", Code: "230184", Level: 3, Sort: 1, Status: 1},
		// Level 4 乡镇
		{Name: "五常镇", Code: "23018401", Level: 4, Sort: 1, Status: 1},
		{Name: "拉林满族镇", Code: "23018402", Level: 4, Sort: 2, Status: 1},
		{Name: "山河镇", Code: "23018403", Level: 4, Sort: 3, Status: 1},
		{Name: "小山子镇", Code: "23018404", Level: 4, Sort: 4, Status: 1},
	}
	// 按 Level 逐级写入，确保父级先于子级创建
	for level := 1; level <= 4; level++ {
		for i, r := range regions {
			if r.Level != level {
				continue
			}
			// 查找父级 ID
			if level > 1 {
				parentIdx := -1
				for j, p := range regions {
					if p.Level == level-1 {
						parentIdx = j
						break
					}
				}
				if parentIdx >= 0 {
					r.ParentID = regions[parentIdx].ID
				}
			}
			if err := firstOrCreateRegion(db, &r); err != nil {
				return err
			}
			regions[i] = r
		}
	}
	return nil
}

func firstOrCreateRegion(db *gorm.DB, r *regionModel.Region) error {
	var found regionModel.Region
	err := db.Where("code = ?", r.Code).First(&found).Error
	if err == nil {
		r.ID = found.ID
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	return db.Create(r).Error
}

// seedPermissions 权限码
func seedPermissions(db *gorm.DB) error {
	for _, p := range permissionDefs {
		var found permModel.Permission
		err := db.Where("code = ?", p.Code).First(&found).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		perm := permModel.Permission{
			Name:   p.Name,
			Code:   p.Code,
			Type:   p.Type,
			Status: 1,
		}
		if err := db.Create(&perm).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedAdminRole 创建 super_admin 超级管理员角色并赋予全部权限
// 依据需求文档 1.9.1 角色定义（原 admin 已由 migrateAdminToSuperAdmin 迁移）
func seedAdminRole(db *gorm.DB) error {
	var role permModel.Role
	err := db.Where("code = ?", "super_admin").First(&role).Error
	if err == gorm.ErrRecordNotFound {
		role = permModel.Role{
			Name:        "超级管理员",
			Code:        "super_admin",
			Description: "系统超级管理员，拥有全部权限",
			Sort:        0,
			Status:      1,
		}
		if err := db.Create(&role).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 收集全部权限ID
	var perms []permModel.Permission
	if err := db.Find(&perms).Error; err != nil {
		return err
	}

	// 为 super_admin 角色分配全部权限（幂等：跳过已存在的关联）
	for _, p := range perms {
		var rp permModel.RolePermission
		err := db.Where("role_id = ? AND permission_id = ?", role.ID, p.ID).First(&rp).Error
		if err == gorm.ErrRecordNotFound {
			rp = permModel.RolePermission{RoleID: role.ID, PermissionID: p.ID}
			if err := db.Create(&rp).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

// seedAdminUser 创建默认管理员账号 admin / admin123 并分配 super_admin 角色
// 用户名保留 admin（登录账号），仅角色 code 从 admin 改为 super_admin
func seedAdminUser(db *gorm.DB) error {
	var user userModel.User
	err := db.Where("username = ?", "admin").First(&user).Error
	if err == gorm.ErrRecordNotFound {
		hash, herr := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if herr != nil {
			return herr
		}
		user = userModel.User{
			Username: "admin",
			Password: string(hash),
			Nickname: "超级管理员",
			Email:    "admin@wuchang.local",
			Gender:   0,
			Status:   1,
		}
		// 默认地区 五常市(id=3)
		user.RegionID = 3
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// 分配 super_admin 角色
	var adminRole permModel.Role
	if err := db.Where("code = ?", "super_admin").First(&adminRole).Error; err != nil {
		return err
	}
	var ur permModel.UserRole
	err = db.Where("user_id = ? AND role_id = ?", user.ID, adminRole.ID).First(&ur).Error
	if err == gorm.ErrRecordNotFound {
		ur = permModel.UserRole{UserID: user.ID, RoleID: adminRole.ID}
		return db.Create(&ur).Error
	}
	return err
}

// ============================================================================
// P0 新增：模块注册表 + 定时任务调度中心种子数据
// 依据 v3.2.1 架构方案：12 中台 + 15 垂直业务 + P0 基线定时任务
// 表结构由 backend/migrations/001_p0_baseline.sql 创建，此处仅插入数据
// 幂等：使用 INSERT ... ON CONFLICT ... DO UPDATE
// ============================================================================

// moduleDef 模块定义（对应 modules 表）
type moduleDef struct {
	Name         string // 模块标识（唯一）
	DisplayName  string // 显示名称
	Category     string // middleware（中台）/ business（垂直业务）
	Description  string // 模块描述
	Dependencies string // 依赖模块 JSON 数组，如 `["user","pay"]`
	Icon         string // 图标
	Author       string // 作者
	Homepage     string // 主页
}

// 12 中台模块清单（category=middleware）
var middlewareModuleDefs = []moduleDef{
	{Name: "user", DisplayName: "用户账号中台", Category: "middleware", Description: "账号、认证、VIP、第三方绑定", Dependencies: `[]`, Icon: "User", Author: "wuchang", Homepage: ""},
	{Name: "pay", DisplayName: "支付财务中台", Category: "middleware", Description: "支付订单、退款、对账、提现", Dependencies: `["user"]`, Icon: "Wallet", Author: "wuchang", Homepage: ""},
	{Name: "im", DisplayName: "IM消息中台", Category: "middleware", Description: "会话、消息、群组、未读数", Dependencies: `["user"]`, Icon: "Chat", Author: "wuchang", Homepage: ""},
	{Name: "merchant", DisplayName: "商家商户中台", Category: "middleware", Description: "店铺、商品、订单核销", Dependencies: `["user"]`, Icon: "Shop", Author: "wuchang", Homepage: ""},
	{Name: "distribution", DisplayName: "分销合伙人中台", Category: "middleware", Description: "代理商、分润、分站", Dependencies: `["user","pay"]`, Icon: "Share", Author: "wuchang", Homepage: ""},
	{Name: "marketing", DisplayName: "营销活动中台", Category: "middleware", Description: "优惠券、满减、签到、拼团", Dependencies: `["user","merchant"]`, Icon: "Gift", Author: "wuchang", Homepage: ""},
	{Name: "risk", DisplayName: "风控审核中台", Category: "middleware", Description: "内容审核、举报、黑名单", Dependencies: `["user"]`, Icon: "Shield", Author: "wuchang", Homepage: ""},
	{Name: "lbs", DisplayName: "LBS地图中台", Category: "middleware", Description: "POI、行政区域、地理围栏", Dependencies: `[]`, Icon: "Location", Author: "wuchang", Homepage: ""},
	{Name: "ai", DisplayName: "AI智能中台", Category: "middleware", Description: "智能推荐、内容生成、向量库", Dependencies: `["user"]`, Icon: "Magic", Author: "wuchang", Homepage: ""},
	{Name: "tenant", DisplayName: "多租户分站中台", Category: "middleware", Description: "分站、租户、配置中心", Dependencies: `["user"]`, Icon: "Building", Author: "wuchang", Homepage: ""},
	{Name: "material", DisplayName: "素材存储中台", Category: "middleware", Description: "图片、视频、文件、素材库", Dependencies: `[]`, Icon: "Folder", Author: "wuchang", Homepage: ""},
	{Name: "diy", DisplayName: "DIY页面中台", Category: "middleware", Description: "可视化页面、组件、模板", Dependencies: `["material"]`, Icon: "Layout", Author: "wuchang", Homepage: ""},
}

// 15 垂直业务模块清单（category=business）
var businessModuleDefs = []moduleDef{
	{Name: "info", DisplayName: "分类信息核心", Category: "business", Description: "通用分类信息发布（兜底）", Dependencies: `["user","risk"]`, Icon: "List", Author: "wuchang", Homepage: ""},
	{Name: "ers", DisplayName: "同城二手", Category: "business", Description: "二手商品发布、交易、行为", Dependencies: `["user","pay","lbs","risk"]`, Icon: "ShoppingBag", Author: "wuchang", Homepage: ""},
	{Name: "job", DisplayName: "同城招聘+零工", Category: "business", Description: "招聘信息、零工任务", Dependencies: `["user","risk"]`, Icon: "Briefcase", Author: "wuchang", Homepage: ""},
	{Name: "house", DisplayName: "同城房产", Category: "business", Description: "房屋租售、小区、经纪人", Dependencies: `["user","lbs","risk"]`, Icon: "House", Author: "wuchang", Homepage: ""},
	{Name: "shop", DisplayName: "同城黄页114商圈", Category: "business", Description: "商户黄页、商圈", Dependencies: `["user","merchant","lbs"]`, Icon: "PhoneBook", Author: "wuchang", Homepage: ""},
	{Name: "daojia", DisplayName: "同城到家预约服务", Category: "business", Description: "上门服务、预约", Dependencies: `["user","merchant","pay","lbs"]`, Icon: "Service", Author: "wuchang", Homepage: ""},
	{Name: "mall", DisplayName: "同城商城电商", Category: "business", Description: "商品、订单、购物车", Dependencies: `["user","merchant","pay","risk"]`, Icon: "ShoppingCart", Author: "wuchang", Homepage: ""},
	{Name: "toutiao", DisplayName: "同城头条资讯", Category: "business", Description: "资讯、文章、订阅", Dependencies: `["user","risk"]`, Icon: "Newspaper", Author: "wuchang", Homepage: ""},
	{Name: "quan", DisplayName: "同城圈子社群", Category: "business", Description: "圈子、帖子、成员", Dependencies: `["user","im","risk"]`, Icon: "Users", Author: "wuchang", Homepage: ""},
	{Name: "huodong", DisplayName: "同城活动", Category: "business", Description: "线下活动、报名、签到", Dependencies: `["user","lbs"]`, Icon: "Calendar", Author: "wuchang", Homepage: ""},
	{Name: "love", DisplayName: "同城婚恋相亲", Category: "business", Description: "交友、相亲、匹配", Dependencies: `["user","im"]`, Icon: "Heart", Author: "wuchang", Homepage: ""},
	{Name: "car", DisplayName: "同城汽车", Category: "business", Description: "车辆买卖、车务", Dependencies: `["user","risk"]`, Icon: "Car", Author: "wuchang", Homepage: ""},
	{Name: "edu", DisplayName: "同城教育培训", Category: "business", Description: "课程、机构、报名", Dependencies: `["user","risk"]`, Icon: "Reading", Author: "wuchang", Homepage: ""},
	{Name: "zhuangxiu", DisplayName: "同城装修", Category: "business", Description: "装修案例、公司、报价", Dependencies: `["user","lbs","risk"]`, Icon: "Tool", Author: "wuchang", Homepage: ""},
	{Name: "zhibo", DisplayName: "同城直播", Category: "business", Description: "直播间、观众、礼物", Dependencies: `["user","im","pay"]`, Icon: "VideoCamera", Author: "wuchang", Homepage: ""},
}

// SeedModules 注册 12 中台 + 15 垂直业务到 modules 表（enabled=true，version=1.0.0）
// 幂等：使用 INSERT ... ON CONFLICT (name) DO UPDATE
// 依赖：modules 表已由 001_p0_baseline.sql 创建
func SeedModules(db *gorm.DB) error {
	const sql = `INSERT INTO modules (name, display_name, category, description, version, dependencies, icon, author, homepage, enabled, installed_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, '1.0.0', $5::jsonb, $6, $7, $8, TRUE, NOW(), NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    version = EXCLUDED.version,
    dependencies = EXCLUDED.dependencies,
    icon = EXCLUDED.icon,
    author = EXCLUDED.author,
    homepage = EXCLUDED.homepage,
    enabled = EXCLUDED.enabled,
    updated_at = NOW();`

	allModules := make([]moduleDef, 0, len(middlewareModuleDefs)+len(businessModuleDefs))
	allModules = append(allModules, middlewareModuleDefs...)
	allModules = append(allModules, businessModuleDefs...)

	for _, m := range allModules {
		if err := db.Exec(sql,
			m.Name, m.DisplayName, m.Category, m.Description,
			m.Dependencies, m.Icon, m.Author, m.Homepage,
		).Error; err != nil {
			return fmt.Errorf("seed module %s failed: %w", m.Name, err)
		}
	}
	return nil
}

// cronJobDef 定时任务定义（对应 cron_jobs 表）
type cronJobDef struct {
	ModuleName      string // 模块名
	JobName         string // 任务名（模块内唯一）
	CronExpr        string // cron 表达式（6 字段：秒 分 时 日 月 周）
	Handler         string // 处理函数标识
	Params          string // 参数 JSON
	MaxRetry        int    // 最大重试次数
	TimeoutSeconds  int    // 超时秒数
}

// P0 基线定时任务（module_name=system 表示平台基础设施任务）
var p0CronJobDefs = []cronJobDef{
	{
		ModuleName:     "system",
		JobName:        "cron_scheduler",
		CronExpr:       "* * * * * *", // 每秒扫描 cron_jobs 表调度到期任务
		Handler:        "system.CronScheduler",
		Params:         `{}`,
		MaxRetry:       0,
		TimeoutSeconds: 30,
	},
	{
		ModuleName:     "system",
		JobName:        "metrics_collector",
		CronExpr:       "0 * * * * *", // 每分钟采集模块指标到 module_metrics
		Handler:        "system.MetricsCollector",
		Params:         `{"interval_seconds":60}`,
		MaxRetry:       3,
		TimeoutSeconds: 60,
	},
	{
		ModuleName:     "system",
		JobName:        "message_queue_scanner",
		CronExpr:       "*/10 * * * * *", // 每 10 秒扫描 message_queue 表投递待发消息
		Handler:        "system.MessageQueueScanner",
		Params:         `{"batch_size":100}`,
		MaxRetry:       3,
		TimeoutSeconds: 30,
	},
	{
		ModuleName:     "system",
		JobName:        "message_queue_retry",
		CronExpr:       "0 */1 * * * *", // 每 1 分钟重试失败消息（next_retry_at 已到）
		Handler:        "system.MessageQueueRetry",
		Params:         `{"batch_size":50}`,
		MaxRetry:       3,
		TimeoutSeconds: 120,
	},
	{
		ModuleName:     "system",
		JobName:        "dead_letter_cleaner",
		CronExpr:       "0 0 3 * * *", // 每天 03:00 清理 7 天前的 dead 消息
		Handler:        "system.DeadLetterCleaner",
		Params:         `{"retain_days":7}`,
		MaxRetry:       3,
		TimeoutSeconds: 300,
	},
	{
		ModuleName:     "system",
		JobName:        "metrics_aggregator",
		CronExpr:       "0 0 * * * *", // 每小时聚合 module_metrics 指标
		Handler:        "system.MetricsAggregator",
		Params:         `{"window":"1h"}`,
		MaxRetry:       3,
		TimeoutSeconds: 300,
	},
}

// SeedCronJobs 注册 P0 阶段定时任务到 cron_jobs 表
// 幂等：使用 INSERT ... ON CONFLICT (module_name, job_name) DO UPDATE
// 依赖：cron_jobs 表已由 001_p0_baseline.sql 创建
func SeedCronJobs(db *gorm.DB) error {
	const sql = `INSERT INTO cron_jobs (module_name, job_name, cron_expr, handler, params, enabled, max_retry, timeout_seconds, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5::jsonb, TRUE, $6, $7, NOW(), NOW())
ON CONFLICT (module_name, job_name) DO UPDATE SET
    cron_expr = EXCLUDED.cron_expr,
    handler = EXCLUDED.handler,
    params = EXCLUDED.params,
    max_retry = EXCLUDED.max_retry,
    timeout_seconds = EXCLUDED.timeout_seconds,
    updated_at = NOW();`

	for _, j := range p0CronJobDefs {
		maxRetry := j.MaxRetry
		if maxRetry <= 0 {
			maxRetry = 3
		}
		timeout := j.TimeoutSeconds
		if timeout <= 0 {
			timeout = 300
		}
		if err := db.Exec(sql,
			j.ModuleName, j.JobName, j.CronExpr, j.Handler, j.Params, maxRetry, timeout,
		).Error; err != nil {
			return fmt.Errorf("seed cron job %s.%s failed: %w", j.ModuleName, j.JobName, err)
		}
	}
	return nil
}
