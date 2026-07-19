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
	// v3.2.1 新增：ershou 二手交易模块完整功能样板
	// 依赖 003_ershou_full.sql 已执行（ers_* 19 张子表已创建）
	if err := SeedErshouFull(db); err != nil {
		return err
	}
	// P1 中台精简版种子数据（pay/im/risk/ai，material 无需种子）
	// 依赖 005_p1_middlewares.sql 已执行（5 个中台 26 张表已创建）
	if err := SeedP1Middlewares(db); err != nil {
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

// ============================================================================
// v3.2.1 新增：ershou 二手交易模块完整功能样板种子数据
// 依据 v3.2.1 架构方案第二章 2.2.A 节，对标闲鱼/转转/58同城/瓜子/贝壳/趣店
// 表结构由 backend/migrations/003_ershou_full.sql 创建，此处仅插入数据
// 幂等：使用 INSERT ... ON CONFLICT ... DO UPDATE / DO NOTHING
// ============================================================================

// SeedErshouFull ershou 模块完整功能样板种子数据入口
// 依赖 003_ershou_full.sql 已执行（ers_* 19 张子表已创建）
func SeedErshouFull(db *gorm.DB) error {
	// 1. 品牌（20 个）
	if err := seedErshouBrands(db); err != nil {
		return err
	}
	// 2. 型号（依赖品牌）
	if err := seedErshouModels(db); err != nil {
		return err
	}
	// 3. 审核规则（10 敏感词 + 3 价格 + 2 频率）
	if err := seedErshouAuditRules(db); err != nil {
		return err
	}
	// 4. 分类属性配置（数码类 + 服装类）
	if err := seedErshouCategoryAttrs(db); err != nil {
		return err
	}
	// 5. 运营标签（精选/新品/爆款/特价/包邮/急转）
	if err := seedErshouTags(db); err != nil {
		return err
	}
	return nil
}

// seedErshouBrands 测试品牌库（20 个）
// 幂等：ON CONFLICT (name) DO UPDATE
func seedErshouBrands(db *gorm.DB) error {
	const sql = `INSERT INTO ers_brands (name, logo, english_name, description, country, official_verified, official_url, category_ids, status, sort, use_count, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, 1, $9, 0, NOW(), NOW())
ON CONFLICT (name) DO UPDATE SET
    logo = EXCLUDED.logo,
    english_name = EXCLUDED.english_name,
    description = EXCLUDED.description,
    country = EXCLUDED.country,
    official_verified = EXCLUDED.official_verified,
    official_url = EXCLUDED.official_url,
    category_ids = EXCLUDED.category_ids,
    sort = EXCLUDED.sort,
    status = 1,
    updated_at = NOW();`

	type brandSeed struct {
		Name, Logo, EnglishName, Description, Country, OfficialURL, CategoryIDs string
		OfficialVerified                                                        bool
		Sort                                                                    int
	}
	brands := []brandSeed{
		{"Apple", "", "Apple Inc.", "美国科技巨头，iPhone/Mac/iPad 等产品制造商", "美国", "https://www.apple.com", `["digital","mobile","computer"]}`, true, 100},
		{"Huawei", "", "HUAWEI", "中国科技领军企业，手机/通讯设备/芯片", "中国", "https://www.huawei.com", `["digital","mobile"]}`, true, 95},
		{"Xiaomi", "", "Xiaomi", "小米，手机/智能家居/生态链", "中国", "https://www.mi.com", `["digital","mobile","smart_home"]}`, true, 90},
		{"Samsung", "", "SAMSUNG", "韩国电子巨头，手机/家电/芯片", "韩国", "https://www.samsung.com", `["digital","mobile","appliance"]}`, true, 88},
		{"OPPO", "", "OPPO", "广东欧珀，手机/智能终端", "中国", "https://www.oppo.com", `["digital","mobile"]}`, true, 85},
		{"vivo", "", "vivo", "维沃移动通信，手机/智能终端", "中国", "https://www.vivo.com", `["digital","mobile"]}`, true, 84},
		{"Nike", "", "NIKE", "耐克，运动鞋/服装/装备", "美国", "https://www.nike.com", `["clothing","sports"]}`, true, 80},
		{"Adidas", "", "ADIDAS", "阿迪达斯，运动鞋/服装", "德国", "https://www.adidas.com", `["clothing","sports"]}`, true, 78},
		{"Lining", "", "LI-NING", "李宁，运动品牌", "中国", "https://www.lining.com", `["clothing","sports"]}`, true, 75},
		{"Anta", "", "ANTA", "安踏，运动品牌", "中国", "https://www.anta.com", `["clothing","sports"]}`, true, 73},
		{"Sony", "", "SONY", "索尼，相机/耳机/游戏机", "日本", "https://www.sony.com", `["digital","camera","audio"]}`, true, 70},
		{"Canon", "", "Canon", "佳能，相机/打印机", "日本", "https://www.canon.com", `["digital","camera"]}`, true, 68},
		{"Nikon", "", "Nikon", "尼康，相机/光学仪器", "日本", "https://www.nikon.com", `["digital","camera"]}`, true, 65},
		{"DJI", "", "DJI", "大疆创新，无人机/相机稳定器", "中国", "https://www.dji.com", `["digital","drone"]}`, true, 62},
		{"Dyson", "", "Dyson", "戴森，吸尘器/吹风机", "英国", "https://www.dyson.com", `["appliance","home"]}`, true, 60},
		{"Gucci", "", "GUCCI", "古驰，奢侈品/箱包/服装", "意大利", "https://www.gucci.com", `["luxury","clothing"]}`, true, 58},
		{"Louis Vuitton", "", "LV", "路易威登，奢侈品/箱包", "法国", "https://www.louisvuitton.com", `["luxury","bag"]}`, true, 55},
		{"Chanel", "", "CHANEL", "香奈儿，奢侈品/化妆品", "法国", "https://www.chanel.com", `["luxury","cosmetics"]}`, true, 53},
		{"Rolex", "", "ROLEX", "劳力士，瑞士手表", "瑞士", "https://www.rolex.com", `["luxury","watch"]}`, true, 50},
		{"Lego", "", "LEGO", "乐高，玩具/积木", "丹麦", "https://www.lego.com", `["toy"]}`, true, 48},
	}
	for i := range brands {
		b := brands[i]
		if err := db.Exec(sql,
			b.Name, b.Logo, b.EnglishName, b.Description, b.Country,
			b.OfficialVerified, b.OfficialURL, b.CategoryIDs, b.Sort,
		).Error; err != nil {
			return fmt.Errorf("seed ershou brand %s failed: %w", b.Name, err)
		}
	}
	return nil
}

// seedErshouModels 测试型号库
// 幂等：ON CONFLICT (brand_id, name) DO UPDATE
func seedErshouModels(db *gorm.DB) error {
	const sql = `INSERT INTO ers_models (brand_id, name, full_name, specifications, image, release_date, status, sort, use_count, reference_price, created_at, updated_at)
VALUES ($1, $2, $3, $4::jsonb, $5, $6, 1, $7, 0, $8, NOW(), NOW())
ON CONFLICT (brand_id, name) DO UPDATE SET
    full_name = EXCLUDED.full_name,
    specifications = EXCLUDED.specifications,
    image = EXCLUDED.image,
    release_date = EXCLUDED.release_date,
    sort = EXCLUDED.sort,
    reference_price = EXCLUDED.reference_price,
    status = 1,
    updated_at = NOW();`

	type modelSeed struct {
		BrandName   string
		Name        string
		FullName    string
		SpecJSON    string
		Image       string
		ReleaseDate string
		Sort        int
		RefPrice    float64
	}
	models := []modelSeed{
		{"Apple", "iPhone 15 Pro Max 256G", "Apple iPhone 15 Pro Max 256GB 深空黑", `{"内存":"256GB","屏幕":"6.7寸","处理器":"A17 Pro","颜色":"深空黑","电池":"4422mAh","像素":"4800万"}`, "", "2023-09", 100, 9999.00},
		{"Apple", "iPhone 15 128G", "Apple iPhone 15 128GB 蓝色", `{"内存":"128GB","屏幕":"6.1寸","处理器":"A16","颜色":"蓝色","电池":"3349mAh","像素":"4800万"}`, "", "2023-09", 98, 5999.00},
		{"Apple", "MacBook Pro 14 M3", "Apple MacBook Pro 14 M3 512GB", `{"内存":"16GB","硬盘":"512GB","屏幕":"14寸","处理器":"M3","颜色":"深空黑"}`, "", "2023-10", 95, 14999.00},
		{"Huawei", "Mate 60 Pro 256G", "Huawei Mate 60 Pro 256GB 雅丹黑", `{"内存":"256GB","屏幕":"6.82寸","处理器":"麒麟9000S","颜色":"雅丹黑","电池":"5000mAh","像素":"5000万"}`, "", "2023-08", 90, 6999.00},
		{"Huawei", "MateBook X Pro", "Huawei MateBook X Pro 2024", `{"内存":"16GB","硬盘":"1TB","屏幕":"14.2寸","处理器":"Intel i7","颜色":"深空灰"}`, "", "2024-04", 88, 11999.00},
		{"Xiaomi", "14 Pro 256G", "Xiaomi 14 Pro 256GB 黑色", `{"内存":"256GB","屏幕":"6.73寸","处理器":"骁龙8 Gen3","颜色":"黑色","电池":"4880mAh","像素":"5000万"}`, "", "2023-10", 85, 4999.00},
		{"Xiaomi", "Redmi Note 13 Pro", "Xiaomi Redmi Note 13 Pro 8+256GB", `{"内存":"256GB","屏幕":"6.67寸","处理器":"骁龙7s Gen2","颜色":"子夜黑","电池":"5100mAh","像素":"2亿"}`, "", "2023-09", 80, 1399.00},
		{"Samsung", "S24 Ultra 512G", "Samsung Galaxy S24 Ultra 512GB 钛黑", `{"内存":"512GB","屏幕":"6.8寸","处理器":"骁龙8 Gen3","颜色":"钛黑","电池":"5000mAh","像素":"2亿"}`, "", "2024-01", 78, 10999.00},
		{"OPPO", "Find X7 256G", "OPPO Find X7 256GB 海阔天空", `{"内存":"256GB","屏幕":"6.78寸","处理器":"天玑9300","颜色":"海阔天空","电池":"5000mAh","像素":"5000万"}`, "", "2024-01", 75, 3999.00},
		{"vivo", "X100 Pro 256G", "vivo X100 Pro 256GB 辰夜黑", `{"内存":"256GB","屏幕":"6.78寸","处理器":"天玑9300","颜色":"辰夜黑","电池":"5400mAh","像素":"5000万"}`, "", "2023-11", 73, 4999.00},
		{"Sony", "A7M4", "Sony Alpha A7M4 单机身", `{"像素":"3300万","传感器":"全画幅","视频":"4K 60p","防抖":"5轴"}`, "", "2021-11", 70, 16999.00},
		{"Canon", "EOS R6 Mark II", "Canon EOS R6 Mark II 单机身", `{"像素":"2420万","传感器":"全画幅","视频":"4K 60p","防抖":"5轴"}`, "", "2022-11", 68, 15999.00},
		{"Nikon", "Z6 III", "Nikon Z6 III 单机身", `{"像素":"2450万","传感器":"全画幅","视频":"6K 60p","防抖":"5轴"}`, "", "2024-06", 65, 17999.00},
		{"DJI", "Mini 4 Pro", "DJI Mini 4 Pro 长续航版", `{"像素":"4800万","续航":"34分钟","重量":"249g","视频":"4K 100p"}`, "", "2023-09", 60, 4788.00},
		{"Nike", "Air Jordan 1 Mid", "Nike Air Jordan 1 Mid 黑红", `{"尺码":"42","颜色":"黑红","材质":"皮革"}`, "", "2023-01", 55, 899.00},
		{"Adidas", "Yeezy 350 V2", "Adidas Yeezy Boost 350 V2 黑白", `{"尺码":"42","颜色":"黑白","材质":"Primeknit"}`, "", "2022-01", 50, 1599.00},
		{"Rolex", "Submariner 126610LN", "Rolex Submariner Date 126610LN 黑水鬼", `{"表壳":"41mm","材质":"904L钢","防水":"300m","机芯":"3235"}`, "", "2020-01", 45, 98000.00},
		{"Lego", "Technic 42115", "Lego Technic 42115 Lamborghini Sián", `{"零件数":"3696","比例":"1:8","遥控":"否"}`, "", "2020-06", 40, 3299.00},
	}
	// 查询品牌 ID
	for _, m := range models {
		var brandID uint
		if err := db.Table("ers_brands").Where("name = ?", m.BrandName).Select("id").Scan(&brandID).Error; err != nil {
			return fmt.Errorf("query brand %s failed: %w", m.BrandName, err)
		}
		if brandID == 0 {
			continue // 品牌未找到，跳过
		}
		if err := db.Exec(sql,
			brandID, m.Name, m.FullName, m.SpecJSON, m.Image, m.ReleaseDate, m.Sort, m.RefPrice,
		).Error; err != nil {
			return fmt.Errorf("seed ershou model %s failed: %w", m.Name, err)
		}
	}
	return nil
}

// seedErshouAuditRules 审核规则种子（10 敏感词 + 3 价格 + 2 频率 + 1 违禁品 + 1 内容）
// 幂等：使用 INSERT ... ON CONFLICT DO NOTHING（无唯一约束时改为 WHERE NOT EXISTS）
func seedErshouAuditRules(db *gorm.DB) error {
	const sql = `INSERT INTO ers_audit_rules (rule_name, rule_type, rule_key, pattern, threshold, action, penalty_type, severity, status, description, sort, created_at, updated_at)
SELECT $1, $2, $3, $4, $5::jsonb, $6, $7, $8, 1, $9, $10, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM ers_audit_rules WHERE rule_name = $1);`

	type ruleSeed struct {
		Name, Type, Key, Pattern, Threshold, Action, Penalty string
		Severity                                             int
		Description                                          string
		Sort                                                 int
	}
	rules := []ruleSeed{
		// 10 条敏感词规则
		{"政治敏感词", "sensitive_word", "politics", `["习近平","毛泽东","邓小平","江泽民","胡锦涛","温家宝","李克强","李强","国务院总理","国家主席"]`, `{"level":"high"}`, "reject", "ban7d", 5, "政治类敏感词，触发即拒绝并封禁 7 天", 100},
		{"色情敏感词", "sensitive_word", "porn", `["约炮","一夜情","裸聊","色情","黄色","av","援交","外围","小姐服务","包夜"]`, `{"level":"high"}`, "reject", "banForever", 5, "色情类敏感词，触发即拒绝并永久封禁", 98},
		{"暴力敏感词", "sensitive_word", "violence", `["杀人","炸弹","枪支","弹药","管制刀具","打架斗殴","恐怖袭击","毒品","大麻","海洛因"]`, `{"level":"high"}`, "reject", "ban7d", 5, "暴力/违禁品类敏感词", 95},
		{"诈骗敏感词", "sensitive_word", "scam", `["中奖","免费领","刷单","兼职日结","高返利","零投入","躺赚","快速致富","内部消息","稳赚不赔"]`, `{"level":"medium"}`, "approval", "warning", 3, "疑似诈骗关键词，触发转人工审核", 90},
		{"广告敏感词", "sensitive_word", "ad", `["加微信","加QQ","扫码","二维码","加群","私聊","咨询详情","代理","招商","加盟"]`, `{"level":"low"}`, "approval", "warning", 2, "疑似广告关键词，转人工审核", 85},
		{"赌博敏感词", "sensitive_word", "gamble", `["赌博","赌场","博彩","外围彩","六合彩","时时彩","北京赛车","幸运飞艇","押注","下注"]`, `{"level":"high"}`, "reject", "banForever", 5, "赌博类敏感词", 80},
		{"违禁药品词", "sensitive_word", "drug", `["摇头丸","K粉","冰毒","麻古","摇头丸","止咳水","联邦止咳露","曲马多","三唑仑","麻黄素"]`, `{"level":"high"}`, "reject", "banForever", 5, "违禁药品关键词", 78},
		{"野生动物词", "sensitive_word", "wildlife", `["象牙","犀角","虎骨","熊掌","穿山甲","象牙制品","玳","玳瑁","砗磲","红豆杉","兰花"]`, `{"level":"high"}`, "reject", "ban7d", 4, "野生动物及其制品交易关键词", 75},
		{"管制刀具词", "sensitive_word", "weapon", `["弹簧刀","蝴蝶刀","三棱刀","匕首","军刀","武术刀","弩","电击器","防身喷雾","甩棍"]`, `{"level":"high"}`, "reject", "ban7d", 4, "管制刀具/武器关键词", 73},
		{"虚假宣传词", "sensitive_word", "fake_ad", `["包治百病","神药","特效药","祖传","秘方","百年秘方","包好","根治","100%有效","无副作用"]`, `{"level":"medium"}`, "approval", "warning", 3, "虚假宣传类关键词", 70},
		// 3 条价格规则
		{"低价异常检测", "price_check", "low_price", "", `{"min_ratio":0.3,"description":"低于同类商品均价 30% 触发审核"}`, "approval", "warning", 3, "发布价格低于同类商品均价 30%，转人工审核", 60},
		{"高价异常检测", "price_check", "high_price", "", `{"max_ratio":3.0,"description":"高于同类商品均价 3 倍触发审核"}`, "approval", "warning", 2, "发布价格高于同类商品均价 3 倍，转人工审核", 58},
		{"原价倒挂检测", "price_check", "price_inverted", "", `{"check":"price > original_price"}`, "reject", "warning", 2, "售价高于原价，直接拒绝", 55},
		// 2 条频率规则
		{"发布频率限制_分钟", "frequency", "publish_per_min", "", `{"max_count_per_min":3,"window_seconds":60}`, "reject", "warning", 2, "1 分钟内最多发布 3 条，超出拒绝", 50},
		{"发布频率限制_小时", "frequency", "publish_per_hour", "", `{"max_count_per_hour":10,"window_seconds":3600}`, "reject", "limit", 3, "1 小时内最多发布 10 条，超出限流 1 小时", 48},
		// 1 条违禁品规则
		{"违禁品总规则", "prohibited", "all", "", `{"categories":["药品","烟草","野生动物","管制刀具","假币","盗版","医疗器械","危险化学品"]}`, "reject", "banForever", 5, "违禁品类总规则，触发即永久封禁", 30},
		// 1 条内容规则
		{"二维码图片检测", "content", "qr_code", "", `{"scan_image":true,"block_if_qr":true}`, "approval", "warning", 2, "图片中包含二维码，转人工审核", 20},
	}
	for _, r := range rules {
		if err := db.Exec(sql,
			r.Name, r.Type, r.Key, r.Pattern, r.Threshold, r.Action, r.Penalty,
			r.Severity, r.Description, r.Sort,
		).Error; err != nil {
			return fmt.Errorf("seed ershou audit rule %s failed: %w", r.Name, err)
		}
	}
	return nil
}

// seedErshouCategoryAttrs 分类属性配置种子
// 数码类（手机）：内存/屏幕/处理器/像素/电池
// 服装类：尺码/颜色/面料/季节
// 幂等：使用 WHERE NOT EXISTS（按 category_id + attr_name 判重）
func seedErshouCategoryAttrs(db *gorm.DB) error {
	const sql = `INSERT INTO ers_category_attrs (category_id, attr_name, attr_key, attr_type, options, unit, is_required, is_filterable, is_searchable, default_value, placeholder, description, status, sort, created_at, updated_at)
SELECT $1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9, $10, $11, $12, 1, $13, NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM ers_category_attrs WHERE category_id = $1 AND attr_name = $2);`

	type attrSeed struct {
		CategoryID                                            int
		AttrName, AttrKey, AttrType, Options, Unit           string
		IsRequired, IsFilterable, IsSearchable                bool
		DefaultValue, Placeholder, Description                string
		Sort                                                    int
	}
	// 假设分类 ID：1二手转让 2数码 3服装 4家具 5家电 6图书 7运动户外 8美妆 9母婴 10奢侈品
	// 实际使用时由 category 表的实际 ID 决定，此处使用约定 ID
	attrs := []attrSeed{
		// 数码类（category_id=2）
		{2, "内存", "memory", "select", `["64GB","128GB","256GB","512GB","1TB","2TB"]`, "GB", true, true, true, "", "请选择内存容量", "手机/平板/电脑内存容量", 100},
		{2, "屏幕尺寸", "screen_size", "string", "", "英寸", false, true, true, "", "请输入屏幕尺寸", "屏幕对角线长度", 95},
		{2, "处理器", "cpu", "string", "", "", false, false, true, "", "请输入处理器型号", "CPU 型号", 90},
		{2, "像素", "pixel", "string", "", "万像素", false, false, false, "", "请输入主摄像素", "主摄像头像素", 85},
		{2, "电池容量", "battery", "string", "", "mAh", false, false, false, "", "请输入电池容量", "电池容量", 80},
		{2, "网络制式", "network", "multi_select", `["2G","3G","4G","5G"]`, "", false, true, false, "", "请选择网络制式", "支持的网络制式", 75},
		// 服装类（category_id=3）
		{3, "尺码", "size", "select", `["XXS","XS","S","M","L","XL","XXL","XXXL","均码"]`, "", true, true, true, "", "请选择尺码", "服装尺码", 100},
		{3, "颜色", "color", "select", `["黑","白","灰","红","橙","黄","绿","蓝","紫","粉","棕","花色"]`, "", true, true, true, "", "请选择颜色", "服装颜色", 95},
		{3, "面料", "fabric", "select", `["棉","麻","丝","毛","涤纶","锦纶","氨纶","混纺","皮革","其他"]`, "", false, true, false, "", "请选择面料", "服装面料材质", 90},
		{3, "季节", "season", "multi_select", `["春","夏","秋","冬","四季"]`, "", false, true, false, "", "请选择适合季节", "适用季节", 85},
		{3, "风格", "style", "select", `["休闲","商务","运动","复古","时尚","简约","街头","甜美","其他"]`, "", false, true, false, "", "请选择风格", "服装风格", 80},
		{3, "品牌", "brand", "string", "", "", false, true, true, "", "请输入品牌", "服装品牌", 75},
	}
	for _, a := range attrs {
		if err := db.Exec(sql,
			a.CategoryID, a.AttrName, a.AttrKey, a.AttrType, a.Options, a.Unit,
			a.IsRequired, a.IsFilterable, a.IsSearchable,
			a.DefaultValue, a.Placeholder, a.Description, a.Sort,
		).Error; err != nil {
			return fmt.Errorf("seed ershou category attr %s.%s failed: %w", string(rune(a.CategoryID+'0')), a.AttrName, err)
		}
	}
	return nil
}

// ============================================================================
// P1 中台精简版种子数据（pay / im / risk / ai）
// 依据 v3.2.1 架构方案 + ershou 模块依赖
// 表结构由 backend/migrations/005_p1_middlewares.sql 创建，此处仅插入数据
// 幂等：使用 INSERT ... ON CONFLICT ... DO UPDATE / DO NOTHING
// ============================================================================

// SeedP1Middlewares P1 中台精简版种子数据入口
// 依赖 005_p1_middlewares.sql 已执行（5 个中台 26 张表已创建）
func SeedP1Middlewares(db *gorm.DB) error {
	// 1. pay 支付方式（4 个：微信/支付宝/余额/银行卡）
	if err := seedPayMethods(db); err != nil {
		return err
	}
	// 2. im 消息模板（10 个系统通知模板）
	if err := seedIMTemplates(db); err != nil {
		return err
	}
	// 3. risk 敏感词（200+：政治/色情/暴力/广告/诈骗/赌博/违禁品）
	if err := seedRiskSensitiveWords(db); err != nil {
		return err
	}
	// 4. risk 审核规则（10 条：敏感词/价格/频率/违禁品）
	if err := seedRiskAuditRules(db); err != nil {
		return err
	}
	// 5. ai 模型配置（5 个：阿里云/腾讯云/通义千问/文心一言/讯飞星火）
	if err := seedAIModels(db); err != nil {
		return err
	}
	return nil
}

// seedPayMethods pay 支付方式种子（4 个）
// 幂等：ON CONFLICT (method_code) DO UPDATE
func seedPayMethods(db *gorm.DB) error {
	const sql = `INSERT INTO pay_methods (region_id, method_code, method_name, icon, description, config, fee_rate, fee_fixed, sort, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9, 1, NOW(), NOW())
ON CONFLICT (method_code) DO UPDATE SET
    method_name = EXCLUDED.method_name,
    icon = EXCLUDED.icon,
    description = EXCLUDED.description,
    config = EXCLUDED.config,
    fee_rate = EXCLUDED.fee_rate,
    fee_fixed = EXCLUDED.fee_fixed,
    sort = EXCLUDED.sort,
    status = 1,
    updated_at = NOW();`

	type methodSeed struct {
		RegionID                               int
		Code, Name, Icon, Desc, Config         string
		FeeRate                                float64
		FeeFixed                               float64
		Sort                                   int
	}
	methods := []methodSeed{
		{3, "wechat", "微信支付", "IconWechat", "微信扫码/JSAPI/H5/小程序支付", `{"app_id":"","mch_id":"","api_key":"","notify_url":"/api/v1/pay/callback/wechat"}`, 0.0060, 0.00, 100},
		{3, "alipay", "支付宝", "IconAlipay", "支付宝 App/网页/扫码支付", `{"app_id":"","private_key":"","public_key":"","notify_url":"/api/v1/pay/callback/alipay"}`, 0.0060, 0.00, 95},
		{3, "balance", "余额支付", "IconWallet", "账户可用余额支付（担保交易）", `{"check_password":true,"min_amount":0.01}`, 0.0000, 0.00, 90},
		{3, "bank_card", "银行卡", "IconCard", "银行卡快捷支付（银联渠道）", `{"channel":"unionpay","timeout_seconds":300}`, 0.0030, 0.00, 85},
	}
	for _, m := range methods {
		if err := db.Exec(sql,
			m.RegionID, m.Code, m.Name, m.Icon, m.Desc, m.Config, m.FeeRate, m.FeeFixed, m.Sort,
		).Error; err != nil {
			return fmt.Errorf("seed pay method %s failed: %w", m.Code, err)
		}
	}
	return nil
}

// seedIMTemplates im 消息模板种子（10 个系统通知模板）
// 幂等：ON CONFLICT (template_code) DO UPDATE
func seedIMTemplates(db *gorm.DB) error {
	const sql = `INSERT INTO im_templates (region_id, template_code, template_name, template_type, title, content, variables, jump_url, description, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, 1, NOW(), NOW())
ON CONFLICT (template_code) DO UPDATE SET
    template_name = EXCLUDED.template_name,
    template_type = EXCLUDED.template_type,
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    variables = EXCLUDED.variables,
    jump_url = EXCLUDED.jump_url,
    description = EXCLUDED.description,
    status = 1,
    updated_at = NOW();`

	type tplSeed struct {
		RegionID                                                int
		Code, Name, Type, Title, Content, Variables, JumpURL, Desc string
	}
	templates := []tplSeed{
		{3, "system_welcome", "欢迎注册", "welcome", "欢迎来到近享同城", "亲爱的{{username}}，欢迎注册近享同城！完善个人信息即可发布信息、享受同城服务。", `["username"]`, "/user/profile", "新用户注册成功通知"},
		{3, "order_paid", "订单支付成功", "order", "订单支付成功", "您的订单 {{order_no}} 已支付 {{amount}} 元，商家将尽快为您处理。", `["order_no","amount"]`, "/order/{{order_no}}", "订单支付成功通知"},
		{3, "order_shipped", "订单已发货", "order", "订单已发货", "您的订单 {{order_no}} 已发货，物流单号：{{tracking_no}}，请注意查收。", `["order_no","tracking_no"]`, "/order/{{order_no}}", "订单发货通知"},
		{3, "order_completed", "订单已完成", "order", "订单已完成", "您的订单 {{order_no}} 已确认完成，欢迎评价交易体验。", `["order_no"]`, "/order/{{order_no}}/review", "订单完成通知"},
		{3, "order_cancelled", "订单已取消", "order", "订单已取消", "您的订单 {{order_no}} 已取消，退款将在 {{refund_days}} 个工作日内到账。", `["order_no","refund_days"]`, "/order/{{order_no}}", "订单取消通知"},
		{3, "refund_processed", "退款已受理", "refund", "退款已受理", "您的退款申请（单号 {{refund_no}}）已受理，金额 {{amount}} 元，预计 1-3 个工作日到账。", `["refund_no","amount"]`, "/refund/{{refund_no}}", "退款受理通知"},
		{3, "refund_completed", "退款已完成", "refund", "退款已完成", "您的退款（单号 {{refund_no}}）已到账 {{amount}} 元，请查收。", `["refund_no","amount"]`, "/refund/{{refund_no}}", "退款完成通知"},
		{3, "activity_invite", "活动邀请", "activity", "同城活动邀请", "{{inviter}} 邀请您参加活动「{{activity_name}}」，时间：{{activity_time}}，地点：{{location}}。", `["inviter","activity_name","activity_time","location"]`, "/activity/{{activity_id}}", "活动邀请通知"},
		{3, "vip_expiring", "VIP 即将到期", "system", "VIP 会员即将到期", "您的 VIP 会员将在 {{expire_date}} 到期，续费可继续享受专属特权。", `["expire_date"]`, "/vip/renew", "VIP 到期提醒"},
		{3, "account_security", "账号安全提醒", "system", "账号安全提醒", "您的账号于 {{login_time}} 在 {{device}} 登录，如非本人操作请及时修改密码。", `["login_time","device"]`, "/user/security", "账号安全提醒"},
	}
	for _, t := range templates {
		if err := db.Exec(sql,
			t.RegionID, t.Code, t.Name, t.Type, t.Title, t.Content, t.Variables, t.JumpURL, t.Desc,
		).Error; err != nil {
			return fmt.Errorf("seed im template %s failed: %w", t.Code, err)
		}
	}
	return nil
}

// seedRiskSensitiveWords risk 敏感词种子（200+：政治/色情/暴力/广告/诈骗/赌博/违禁品）
// 幂等：ON CONFLICT (word) DO NOTHING
func seedRiskSensitiveWords(db *gorm.DB) error {
	const sql = `INSERT INTO risk_sensitive_words (word, word_type, category, replacement, status, created_at, updated_at)
VALUES ($1, $2, $3, '***', 1, NOW(), NOW())
ON CONFLICT (word) DO NOTHING;`

	type wordSeed struct {
		Word, WordType, Category string
	}
	words := []wordSeed{
		// === 政治敏感词（politics）50 个 ===
		{"习近平", "politics", "leaders"}, {"毛泽东", "politics", "leaders"}, {"邓小平", "politics", "leaders"},
		{"江泽民", "politics", "leaders"}, {"胡锦涛", "politics", "leaders"}, {"温家宝", "politics", "leaders"},
		{"李克强", "politics", "leaders"}, {"李强", "politics", "leaders"}, {"国务院总理", "politics", "leaders"},
		{"国家主席", "politics", "leaders"}, {"总理", "politics", "leaders"}, {"党中央", "politics", "organization"},
		{"中南海", "politics", "organization"}, {"政治局", "politics", "organization"}, {"政法委", "politics", "organization"},
		{" 六四 ", "politics", "event"}, {"天安门事件", "politics", "event"}, {"学潮", "politics", "event"},
		{"法轮功", "politics", "cult"}, {"全能神", "politics", "cult"}, {"邪教", "politics", "cult"},
		{"东突", "politics", "separatism"}, {"藏独", "politics", "separatism"}, {"疆独", "politics", "separatism"},
		{"台独", "politics", "separatism"}, {"港独", "politics", "separatism"}, {"蒙独", "politics", "separatism"},
		{"颜色革命", "politics", "subversion"}, {"颠覆国家", "politics", "subversion"}, {"推翻共产党", "politics", "subversion"},
		{"反共", "politics", "subversion"}, {"反华", "politics", "subversion"}, {"辱华", "politics", "subversion"},
		{"分裂国家", "politics", "subversion"}, {"煽动颠覆", "politics", "subversion"}, {"政治迫害", "politics", "subversion"},
		{"维权律师", "politics", "sensitive"}, {"上访", "politics", "sensitive"}, {"群体事件", "politics", "sensitive"},
		{"游行示威", "politics", "sensitive"}, {"罢工罢课", "politics", "sensitive"}, {"静坐抗议", "politics", "sensitive"},
		{"言论自由", "politics", "sensitive"}, {"新闻自由", "politics", "sensitive"}, {"民主运动", "politics", "sensitive"},
		{"宪政", "politics", "sensitive"}, {"多党制", "politics", "sensitive"}, {"普选", "politics", "sensitive"},
		{" referendums ", "politics", "sensitive"}, {"公民投票", "politics", "sensitive"}, {"一国两制", "politics", "sensitive"},

		// === 色情敏感词（porn）40 个 ===
		{"约炮", "porn", "sex"}, {"一夜情", "porn", "sex"}, {"裸聊", "porn", "sex"}, {"色情", "porn", "sex"},
		{"黄色视频", "porn", "sex"}, {"av女优", "porn", "sex"}, {"援交", "porn", "sex"}, {"外围女", "porn", "sex"},
		{"小姐服务", "porn", "sex"}, {"包夜", "porn", "sex"}, {"上门服务", "porn", "sex"}, {"特殊服务", "porn", "sex"},
		{"裸体", "porn", "nudity"}, {"裸照", "porn", "nudity"}, {"走光", "porn", "nudity"}, {"偷拍", "porn", "nudity"},
		{"成人视频", "porn", "content"}, {"成人电影", "porn", "content"}, {"三级片", "porn", "content"},
		{"色情网站", "porn", "content"}, {"黄色网站", "porn", "content"}, {"成人网站", "porn", "content"},
		{"情趣用品", "porn", "product"}, {"充气娃娃", "porn", "product"}, {"飞机杯", "porn", "product"},
		{"迷药", "porn", "illegal"}, {"催情药", "porn", "illegal"}, {"春药", "porn", "illegal"},
		{"伟哥", "porn", "drug"}, {"印度神油", "porn", "drug"}, {"人造处女", "porn", "illegal"},
		{"代孕", "porn", "illegal"}, {"卖卵", "porn", "illegal"}, {"色情直播", "porn", "content"},
		{"裸体直播", "porn", "content"}, {"脱衣舞", "porn", "content"}, {"色情表演", "porn", "content"},
		{"黄赌毒", "porn", "illegal"}, {"涉黄", "porn", "illegal"}, {"扫黄", "porn", "sensitive"},

		// === 暴力敏感词（violence）35 个 ===
		{"杀人", "violence", "killing"}, {"杀人方法", "violence", "killing"}, {"投毒", "violence", "killing"},
		{"爆炸", "violence", "killing"}, {"炸弹", "violence", "weapon"}, {"自制炸弹", "violence", "weapon"},
		{"手枪", "violence", "weapon"}, {"枪支", "violence", "weapon"}, {"弹药", "violence", "weapon"},
		{"子弹", "violence", "weapon"}, {"雷管", "violence", "weapon"}, {"导火索", "violence", "weapon"},
		{"管制刀具", "violence", "weapon"}, {"弹簧刀", "violence", "weapon"}, {"蝴蝶刀", "violence", "weapon"},
		{"三棱刀", "violence", "weapon"}, {"匕首", "violence", "weapon"}, {"军刀", "violence", "weapon"},
		{"武术刀", "violence", "weapon"}, {"弩", "violence", "weapon"}, {"电击器", "violence", "weapon"},
		{"防身喷雾", "violence", "weapon"}, {"甩棍", "violence", "weapon"}, {"指虎", "violence", "weapon"},
		{"打架斗殴", "violence", "behavior"}, {"恐怖袭击", "violence", "terrorism"}, {"恐怖组织", "violence", "terrorism"},
		{"极端组织", "violence", "terrorism"}, {"ISIS", "violence", "terrorism"}, {"基地组织", "violence", "terrorism"},
		{"塔利班", "violence", "terrorism"}, {"自焚", "violence", "extreme"}, {"自杀方法", "violence", "extreme"},
		{"割腕", "violence", "extreme"}, {"上吊", "violence", "extreme"},

		// === 广告敏感词（ad）40 个 ===
		{"加微信", "ad", "contact"}, {"加QQ", "ad", "contact"}, {"加Q群", "ad", "contact"},
		{"加微信群", "ad", "contact"}, {"扫码加好友", "ad", "contact"}, {"扫码进群", "ad", "contact"},
		{"私聊", "ad", "contact"}, {"咨询详情", "ad", "contact"}, {"联系客服", "ad", "contact"},
		{"加我微信", "ad", "contact"}, {"加我QQ", "ad", "contact"}, {"扫码咨询", "ad", "contact"},
		{"二维码", "ad", "qr"}, {"微信二维码", "ad", "qr"}, {"QQ二维码", "ad", "qr"},
		{"收款码", "ad", "qr"}, {"收款二维码", "ad", "qr"}, {"付款码", "ad", "qr"},
		{"代理", "ad", "business"}, {"招商", "ad", "business"}, {"加盟", "ad", "business"},
		{"招商加盟", "ad", "business"}, {"招代理", "ad", "business"}, {"全国招商", "ad", "business"},
		{"诚招代理", "ad", "business"}, {"火爆招商", "ad", "business"}, {"限时加盟", "ad", "business"},
		{"免费试用", "ad", "promotion"}, {"免费领取", "ad", "promotion"}, {"免费送", "ad", "promotion"},
		{"0元购", "ad", "promotion"}, {"一分钱", "ad", "promotion"}, {"超低价", "ad", "promotion"},
		{"全网最低", "ad", "promotion"}, {"史上最低", "ad", "promotion"}, {"亏本清仓", "ad", "promotion"},
		{"跳楼价", "ad", "promotion"}, {"吐血甩卖", "ad", "promotion"}, {"清仓处理", "ad", "promotion"},

		// === 诈骗敏感词（fraud）40 个 ===
		{"中奖了", "fraud", "lottery"}, {"恭喜中奖", "fraud", "lottery"}, {"您已中奖", "fraud", "lottery"},
		{"免费领奖", "fraud", "lottery"}, {"领奖通知", "fraud", "lottery"}, {"幸运用户", "fraud", "lottery"},
		{"刷单", "fraud", "scam"}, {"刷销量", "fraud", "scam"}, {"刷信誉", "fraud", "scam"},
		{"刷好评", "fraud", "scam"}, {"兼职刷单", "fraud", "scam"}, {"刷单返现", "fraud", "scam"},
		{"日结兼职", "fraud", "scam"}, {"在家兼职", "fraud", "scam"}, {"手机兼职", "fraud", "scam"},
		{"高返利", "fraud", "scam"}, {"高佣金", "fraud", "scam"}, {"零投入", "fraud", "scam"},
		{"零风险", "fraud", "scam"}, {"躺赚", "fraud", "scam"}, {"快速致富", "fraud", "scam"},
		{"内部消息", "fraud", "scam"}, {"稳赚不赔", "fraud", "scam"}, {"百分百赚钱", "fraud", "scam"},
		{"一天赚千", "fraud", "scam"}, {"月入十万", "fraud", "scam"}, {"日入过万", "fraud", "scam"},
		{"资金盘", "fraud", "scam"}, {"庞氏骗局", "fraud", "scam"}, {"传销", "fraud", "scam"},
		{"拉人头", "fraud", "scam"}, {"发展下线", "fraud", "scam"}, {"团队计酬", "fraud", "scam"},
		{"解冻资金", "fraud", "scam"}, {"安全账户", "fraud", "scam"}, {"公检法转账", "fraud", "scam"},
		{"验证账户", "fraud", "scam"}, {"退还保费", "fraud", "scam"}, {"贷款手续费", "fraud", "scam"},

		// === 赌博敏感词（gamble）30 个 ===
		{"赌博", "gamble", "casino"}, {"赌场", "gamble", "casino"}, {"博彩", "gamble", "casino"},
		{"网络赌博", "gamble", "casino"}, {"线上赌场", "gamble", "casino"}, {"真人赌场", "gamble", "casino"},
		{"外围彩", "gamble", "lottery"}, {"六合彩", "gamble", "lottery"}, {"时时彩", "gamble", "lottery"},
		{"北京赛车", "gamble", "lottery"}, {"幸运飞艇", "gamble", "lottery"}, {"重庆时时彩", "gamble", "lottery"},
		{"押注", "gamble", "bet"}, {"下注", "gamble", "bet"}, {"投注", "gamble", "bet"},
		{"赔率", "gamble", "bet"}, {"盘口", "gamble", "bet"}, {"庄家", "gamble", "bet"},
		{"走地", "gamble", "bet"}, {"滚球", "gamble", "bet"}, {"让球", "gamble", "bet"},
		{"赌球", "gamble", "bet"}, {"赌马", "gamble", "bet"}, {"赌狗", "gamble", "bet"},
		{"百家乐", "gamble", "game"}, {"牛牛", "gamble", "game"}, {"炸金花", "gamble", "game"},
		{"斗地主", "gamble", "game"}, {"德州扑克", "gamble", "game"}, {"龙虎斗", "gamble", "game"},

		// === 违禁品敏感词（contraband）30 个 ===
		{"摇头丸", "contraband", "drug"}, {"K粉", "contraband", "drug"}, {"冰毒", "contraband", "drug"},
		{"麻古", "contraband", "drug"}, {"海洛因", "contraband", "drug"}, {"大麻", "contraband", "drug"},
		{"可卡因", "contraband", "drug"}, {"摇头丸出售", "contraband", "drug"}, {"止咳水", "contraband", "drug"},
		{"联邦止咳露", "contraband", "drug"}, {"曲马多", "contraband", "drug"}, {"三唑仑", "contraband", "drug"},
		{"麻黄素", "contraband", "drug"}, {"罂粟", "contraband", "drug"}, {"鸦片", "contraband", "drug"},
		{"象牙", "contraband", "wildlife"}, {"犀角", "contraband", "wildlife"}, {"虎骨", "contraband", "wildlife"},
		{"熊掌", "contraband", "wildlife"}, {"穿山甲", "contraband", "wildlife"}, {"玳瑁", "contraband", "wildlife"},
		{"砗磲", "contraband", "wildlife"}, {"红豆杉", "contraband", "wildlife"}, {"兰花", "contraband", "wildlife"},
		{"假币", "contraband", "illegal"}, {"假钞", "contraband", "illegal"}, {"伪造货币", "contraband", "illegal"},
		{"盗版", "contraband", "illegal"}, {"盗版软件", "contraband", "illegal"}, {"盗版光盘", "contraband", "illegal"},

		// === 其他敏感词（other）10 个 ===
		{"假发票", "other", "illegal"}, {"代开发票", "other", "illegal"}, {"出售发票", "other", "illegal"},
		{"办证", "other", "illegal"}, {"办假证", "other", "illegal"}, {"刻章", "other", "illegal"},
		{"私刻公章", "other", "illegal"}, {"办学历", "other", "illegal"}, {"办文凭", "other", "illegal"},
		{"假文凭", "other", "illegal"},
	}

	for _, w := range words {
		if err := db.Exec(sql, w.Word, w.WordType, w.Category).Error; err != nil {
			return fmt.Errorf("seed risk sensitive word %s failed: %w", w.Word, err)
		}
	}
	return nil
}

// seedRiskAuditRules risk 审核规则种子（10 条）
// 幂等：ON CONFLICT (rule_name) DO UPDATE
func seedRiskAuditRules(db *gorm.DB) error {
	const sql = `INSERT INTO risk_audit_rules (rule_name, rule_type, config, description, status, created_at, updated_at)
VALUES ($1, $2, $3::jsonb, $4, 1, NOW(), NOW())
ON CONFLICT (rule_name) DO UPDATE SET
    rule_type = EXCLUDED.rule_type,
    config = EXCLUDED.config,
    description = EXCLUDED.description,
    status = 1,
    updated_at = NOW();`

	type ruleSeed struct {
		Name, Type, Config, Desc string
	}
	rules := []ruleSeed{
		{"政治敏感词检测", "sensitive_word", `{"word_types":["politics"],"action":"reject","penalty":"ban_7d"}`, "政治类敏感词触发即拒绝并封禁 7 天"},
		{"色情敏感词检测", "sensitive_word", `{"word_types":["porn"],"action":"reject","penalty":"ban_forever"}`, "色情类敏感词触发即永久封禁"},
		{"暴力敏感词检测", "sensitive_word", `{"word_types":["violence"],"action":"reject","penalty":"ban_7d"}`, "暴力类敏感词触发拒绝并封禁 7 天"},
		{"诈骗敏感词检测", "sensitive_word", `{"word_types":["fraud"],"action":"approval","penalty":"warning"}`, "诈骗类敏感词转人工审核"},
		{"广告敏感词检测", "sensitive_word", `{"word_types":["ad"],"action":"approval","penalty":"warning"}`, "广告类敏感词转人工审核"},
		{"赌博敏感词检测", "sensitive_word", `{"word_types":["gamble"],"action":"reject","penalty":"ban_forever"}`, "赌博类敏感词触发即永久封禁"},
		{"违禁品敏感词检测", "sensitive_word", `{"word_types":["contraband"],"action":"reject","penalty":"ban_forever"}`, "违禁品类敏感词触发即永久封禁"},
		{"低价异常检测", "price", `{"min_ratio":0.3,"compare":"category_avg","action":"approval"}`, "发布价格低于同类商品均价 30% 转人工审核"},
		{"高价异常检测", "price", `{"max_ratio":3.0,"compare":"category_avg","action":"approval"}`, "发布价格高于同类商品均价 3 倍转人工审核"},
		{"发布频率限制", "frequency", `{"max_per_min":3,"max_per_hour":10,"max_per_day":30,"action":"reject"}`, "1 分钟内最多 3 条、1 小时内最多 10 条、1 天内最多 30 条"},
	}
	for _, r := range rules {
		if err := db.Exec(sql, r.Name, r.Type, r.Config, r.Desc).Error; err != nil {
			return fmt.Errorf("seed risk audit rule %s failed: %w", r.Name, err)
		}
	}
	return nil
}

// seedAIModels ai 模型配置种子（5 个：阿里云/腾讯云/通义千问/文心一言/讯飞星火）
// 幂等：ON CONFLICT (model_name) DO UPDATE
func seedAIModels(db *gorm.DB) error {
	const sql = `INSERT INTO ai_models (model_name, provider, model_type, api_key, endpoint, config, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, 1, NOW(), NOW())
ON CONFLICT (model_name) DO UPDATE SET
    provider = EXCLUDED.provider,
    model_type = EXCLUDED.model_type,
    api_key = EXCLUDED.api_key,
    endpoint = EXCLUDED.endpoint,
    config = EXCLUDED.config,
    status = 1,
    updated_at = NOW();`

	type modelSeed struct {
		Name, Provider, ModelType, APIKey, Endpoint, Config string
	}
	models := []modelSeed{
		{"aliyun-bailian", "aliyun", "llm", "", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation", `{"model":"qwen-max","temperature":0.7,"max_tokens":2000,"timeout":30}`, },
		{"tencent-hunyuan", "tencent", "llm", "", "https://hunyuan.tencentcloudapi.com/v1/chat/completions", `{"model":"hunyuan-pro","temperature":0.7,"max_tokens":2000,"timeout":30}`, },
		{"qwen-turbo", "qwen", "llm", "", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation", `{"model":"qwen-turbo","temperature":0.7,"max_tokens":1500,"timeout":15}`, },
		{"wenxin-ernie", "wenxin", "llm", "", "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/ernie-bot-4", `{"model":"ernie-bot-4","temperature":0.7,"max_tokens":2000,"timeout":30}`, },
		{"xfyun-spark", "xfyun", "llm", "", "https://spark-api.xf-yun.com/v3.5/chat", `{"model":"spark-v3.5","temperature":0.7,"max_tokens":2000,"timeout":30}`, },
	}
	for _, m := range models {
		if err := db.Exec(sql,
			m.Name, m.Provider, m.ModelType, m.APIKey, m.Endpoint, m.Config,
		).Error; err != nil {
			return fmt.Errorf("seed ai model %s failed: %w", m.Name, err)
		}
	}
	return nil
}

// seedErshouTags 运营标签种子
// 6 个运营标签 + 4 个智能标签 + 4 个自定义标签示例
// 幂等：ON CONFLICT (name, type) DO UPDATE
func seedErshouTags(db *gorm.DB) error {
	const sql = `INSERT INTO ers_tags (name, type, color, icon, background, status, sort, use_count, is_hot, creator_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, 1, $6, 0, $7, 0, NOW(), NOW())
ON CONFLICT (name, type) DO UPDATE SET
    color = EXCLUDED.color,
    icon = EXCLUDED.icon,
    background = EXCLUDED.background,
    sort = EXCLUDED.sort,
    is_hot = EXCLUDED.is_hot,
    status = 1,
    updated_at = NOW();`

	type tagSeed struct {
		Name, Type, Color, Icon, Background string
		Sort                              int
		IsHot                             bool
	}
	tags := []tagSeed{
		// 6 个运营标签
		{"精选", "operation", "#F56C6C", "Star", "#F56C6C", 100, true},
		{"新品", "operation", "#409EFF", "Plus", "#409EFF", 95, true},
		{"爆款", "operation", "#E6A23C", "TrendCharts", "#E6A23C", 90, true},
		{"特价", "operation", "#F56C6C", "Discount", "#F56C6C", 85, true},
		{"包邮", "operation", "#67C23A", "Box", "#67C23A", 80, true},
		{"急转", "operation", "#E6A23C", "AlarmClock", "#E6A23C", 75, true},
		// 4 个智能标签（AI 自动识别）
		{"95新", "smart", "#67C23A", "Medal", "#67C23A", 70, false},
		{"官方验真", "smart", "#409EFF", "Select", "#409EFF", 65, false},
		{"极速发货", "smart", "#409EFF", "Van", "#409EFF", 60, false},
		{"同城自提", "smart", "#67C23A", "Location", "#67C23A", 55, false},
		// 4 个自定义标签示例（用户使用频次较高）
		{"个人闲置", "custom", "#909399", "User", "#909399", 50, false},
		{"送原装盒", "custom", "#67C23A", "Box", "#67C23A", 45, false},
		{"支持验货", "custom", "#409EFF", "Select", "#409EFF", 40, false},
		{"小刀", "custom", "#E6A23C", "Discount", "#E6A23C", 35, false},
	}
	for _, t := range tags {
		if err := db.Exec(sql,
			t.Name, t.Type, t.Color, t.Icon, t.Background, t.Sort, t.IsHot,
		).Error; err != nil {
			return fmt.Errorf("seed ershou tag %s failed: %w", t.Name, err)
		}
	}
	return nil
}
