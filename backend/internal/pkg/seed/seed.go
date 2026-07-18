// Package seed 种子数据初始化
// 在服务启动、AutoMigrate 之后执行，幂等：已存在的数据不会重复创建
package seed

import (
	permModel "wuchang-tongcheng/internal/modules/permission/model"
	regionModel "wuchang-tongcheng/internal/modules/region/model"
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
}

// Run 执行种子数据初始化（幂等）
func Run(db *gorm.DB) error {
	if err := seedRegions(db); err != nil {
		return err
	}
	if err := seedPermissions(db); err != nil {
		return err
	}
	if err := seedAdminRole(db); err != nil {
		return err
	}
	if err := seedAdminUser(db); err != nil {
		return err
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

// seedAdminRole 创建 admin 超级管理员角色并赋予全部权限
func seedAdminRole(db *gorm.DB) error {
	var role permModel.Role
	err := db.Where("code = ?", "admin").First(&role).Error
	if err == gorm.ErrRecordNotFound {
		role = permModel.Role{
			Name:        "超级管理员",
			Code:        "admin",
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

	// 为 admin 角色分配全部权限（幂等：跳过已存在的关联）
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

// seedAdminUser 创建默认管理员账号 admin / admin123 并分配 admin 角色
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

	// 分配 admin 角色
	var adminRole permModel.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
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
