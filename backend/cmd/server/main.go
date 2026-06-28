// Package main 服务入口
// 五常同城本地生活服务平台 - 后端服务入口
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"wuchang-tongcheng/internal/core/middleware"
	"wuchang-tongcheng/internal/core/plugin"
	"wuchang-tongcheng/internal/core/response"
	"wuchang-tongcheng/internal/core/router"
	category "wuchang-tongcheng/internal/modules/category"
	file "wuchang-tongcheng/internal/modules/file"
	news "wuchang-tongcheng/internal/modules/news"
	permission "wuchang-tongcheng/internal/modules/permission"
	region "wuchang-tongcheng/internal/modules/region"
	setting "wuchang-tongcheng/internal/modules/setting"
	user "wuchang-tongcheng/internal/modules/user"
	"wuchang-tongcheng/internal/pkg/config"
	"wuchang-tongcheng/internal/pkg/seed"
	"wuchang-tongcheng/internal/pkg/database"
	espkg "wuchang-tongcheng/internal/pkg/es"
	jwtpkg "wuchang-tongcheng/internal/pkg/jwt"
	"wuchang-tongcheng/internal/pkg/logger"
	mappkg "wuchang-tongcheng/internal/pkg/map"
	mqpkg "wuchang-tongcheng/internal/pkg/mq"
	redispkg "wuchang-tongcheng/internal/pkg/redis"
	wspkg "wuchang-tongcheng/internal/pkg/ws"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	// docs 包含 init() 注册 Swagger 元数据（占位版，swag init 后会覆盖）
	_ "wuchang-tongcheng/docs"
)

// 版本信息
var (
	Version   = "0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// @title           五常同城本地生活服务平台 API
// @version         0.2.0
// @description     五常同城后端服务 API 文档
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in   header
// @name Authorization
// @description JWT Bearer Token，格式：Bearer {token}
func main() {
	// 解析命令行参数
	configPath := flag.String("config", "./configs/config.yaml", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本信息")
	flag.Parse()

	// 显示版本信息
	if *showVersion {
		fmt.Printf("五常同城服务 v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	// 1. 加载配置
	fmt.Println("正在加载配置...")
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("配置加载成功")

	// 2. 初始化日志
	fmt.Println("正在初始化日志...")
	if err := logger.Init(&cfg.Logger); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	logger.Info("日志初始化成功")

	// 3. 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 4. 初始化数据库
	logger.Info("正在初始化数据库...")
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	defer database.Close()
	logger.Info("数据库初始化成功")

	// 5. 初始化Redis
	logger.Info("正在初始化Redis...")
	if err := redispkg.Init(&cfg.Redis); err != nil {
		logger.Fatal("Redis初始化失败", zap.Error(err))
	}
	defer redispkg.Close()
	logger.Info("Redis初始化成功")

	// 5.1 初始化RabbitMQ（可选：连不上不阻塞服务启动）
	logger.Info("正在初始化RabbitMQ...")
	if err := mqpkg.Init(&cfg.RabbitMQ); err != nil {
		logger.Warn("RabbitMQ初始化失败，业务可降级运行", zap.Error(err))
	} else {
		defer mqpkg.Close()
		logger.Info("RabbitMQ初始化成功")
	}

	// 5.2 初始化Elasticsearch（可选：连不上不阻塞服务启动）
	logger.Info("正在初始化Elasticsearch...")
	if err := espkg.Init(&cfg.ES); err != nil {
		logger.Warn("Elasticsearch初始化失败，业务可降级运行", zap.Error(err))
	} else {
		defer espkg.Close()
		logger.Info("Elasticsearch初始化成功")
	}

	// 5.3 初始化WebSocket Hub（无外部依赖，总是启动，用于实时通知推送）
	wspkg.Init()
	defer wspkg.Close()
	logger.Info("WebSocket Hub初始化成功")

	// 5.4 初始化高德地图客户端（可选：key 未配置则降级不可用）
	mappkg.Init(&cfg.Map)
	if mappkg.IsAvailable() {
		logger.Info("高德地图客户端初始化成功")
	} else {
		logger.Warn("高德地图未配置 key，地图服务不可用（业务可降级）")
	}

	// 6. 初始化JWT
	logger.Info("正在初始化JWT...")
	jwtpkg.Init(cfg.JWT.Secret, cfg.JWT.Expire)
	logger.Info("JWT初始化成功")

	// 7. 初始化路由
	logger.Info("正在初始化路由...")
	r := router.NewRouter()

	// 注册全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger(logger.GetLogger()))
	r.Use(middleware.Recovery(logger.GetLogger()))
	r.Use(middleware.Region())
	r.Use(middleware.Auth())

	// 注册健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{
			"status":  "ok",
			"version": Version,
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		}))
	})

	// 静态文件服务（上传的文件访问）
	r.Engine().Static("/uploads", "./uploads")

	// WebSocket 实时通知端点（鉴权：?token=<JWT>）
	r.GET("/ws", middleware.WebSocketHandler())

	// 注册根路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Success(gin.H{
			"name":    "五常同城本地生活服务平台",
			"version": Version,
			"docs":    "/api/v1/docs",
		}))
	})

	// 8. 初始化插件
	logger.Info("正在初始化插件...")
	pluginManager := plugin.GetManager()

	// 注册业务模块插件
	pluginManager.Register(user.NewPlugin())
	pluginManager.Register(region.NewPlugin())
	pluginManager.Register(category.NewPlugin())
	pluginManager.Register(news.NewPlugin())
	pluginManager.Register(permission.NewPlugin())
	pluginManager.Register(file.NewPlugin())
	pluginManager.Register(setting.NewPlugin())

	// 初始化所有插件
	ctx := context.Background()
	if err := pluginManager.InitAll(ctx); err != nil {
		logger.Fatal("插件初始化失败", zap.Error(err))
	}
	logger.Infof("已加载 %d 个插件", len(pluginManager.List()))

	// 初始化种子数据（幂等）：默认地区、权限码、admin 角色、admin 账号
	if err := seed.Run(database.GetDB()); err != nil {
		logger.Error("种子数据初始化失败", zap.Error(err))
	} else {
		logger.Info("种子数据初始化完成")
	}

	// 注册插件路由
	rootGroup := r.Group("")
	pluginManager.RegisterAllRoutes(rootGroup)

	// Swagger 文档路由（/api/v1/docs 与 /swagger/* 均可访问）
	// 生成文档命令：swag init -g cmd/server/main.go -o docs
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/api/v1/docs/*any", swaggerHandler)
	r.GET("/swagger/*any", swaggerHandler)

	// 地图服务路由（需登录，key 未配置时返回 503）
	mapAuth := middleware.AuthRequired()
	mapLimiter := middleware.RateLimit(30, 60, "map")
	apiV1 := r.Engine().Group("/api/v1")
	apiV1.GET("/map/regeocode", mapAuth, mapLimiter, func(c *gin.Context) {
		if !mappkg.IsAvailable() {
			c.JSON(http.StatusOK, response.Fail(503, "地图服务未配置"))
			return
		}
		lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
		lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
		if lng == 0 || lat == 0 {
			c.JSON(http.StatusOK, response.Fail(400, "缺少经纬度参数 lng/lat"))
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()
		result, err := mappkg.Regeocode(ctx, lng, lat)
		if err != nil {
			c.JSON(http.StatusOK, response.Fail(500, "逆地理编码失败: "+err.Error()))
			return
		}
		c.JSON(http.StatusOK, response.Success(result))
	})
	apiV1.GET("/map/geocode", mapAuth, mapLimiter, func(c *gin.Context) {
		if !mappkg.IsAvailable() {
			c.JSON(http.StatusOK, response.Fail(503, "地图服务未配置"))
			return
		}
		address := c.Query("address")
		if address == "" {
			c.JSON(http.StatusOK, response.Fail(400, "缺少地址参数 address"))
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()
		result, err := mappkg.Geocode(ctx, address)
		if err != nil {
			c.JSON(http.StatusOK, response.Fail(500, "地理编码失败: "+err.Error()))
			return
		}
		c.JSON(http.StatusOK, response.Success(result))
	})
	apiV1.GET("/map/around", mapAuth, mapLimiter, func(c *gin.Context) {
		if !mappkg.IsAvailable() {
			c.JSON(http.StatusOK, response.Fail(503, "地图服务未配置"))
			return
		}
		lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
		lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
		if lng == 0 || lat == 0 {
			c.JSON(http.StatusOK, response.Fail(400, "缺少经纬度参数 lng/lat"))
			return
		}
		radius, _ := strconv.Atoi(c.Query("radius"))
		if radius <= 0 {
			radius = 1000
		}
		limit, _ := strconv.Atoi(c.Query("limit"))
		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()
		pois, err := mappkg.Around(ctx, lng, lat, radius, c.Query("types"), limit)
		if err != nil {
			c.JSON(http.StatusOK, response.Fail(500, "周边搜索失败: "+err.Error()))
			return
		}
		c.JSON(http.StatusOK, response.Success(gin.H{"pois": pois}))
	})

	// 404处理
	r.Engine().NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, response.NotFound("请求的接口不存在"))
	})

	// 8. 启动服务
	addr := cfg.Server.GetAddr()
	logger.Infof("服务启动中，监听地址: %s", addr)
	logger.Infof("服务版本: v%s", Version)

	srv := &http.Server{
		Addr:    addr,
		Handler: r.Engine(),
	}

	// 优雅启动
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	logger.Info("服务启动成功！")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务...")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭插件
	if err := pluginManager.CloseAll(); err != nil {
		logger.Error("插件关闭失败", zap.Error(err))
	}

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("服务强制关闭", zap.Error(err))
	}

	logger.Info("服务已正常关闭")
}
