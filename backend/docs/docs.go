// Package docs Swagger 自动生成的 API 文档元数据
// 由 swag init -g cmd/server/main.go -o docs 生成
// 若未执行 swag init，此处提供一份最小可用的占位文档，保证 swagger UI 可访问
package docs

import "github.com/swaggo/swag"

// SwaggerInfo 占位的 Swagger 元数据（推荐执行 swag init 后由生成代码覆盖）
var SwaggerInfo = &swag.Spec{
	Version:          "0.2.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "五常同城本地生活服务平台 API",
	Description:      "五常同城后端服务 API 文档",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

// docTemplate 最小 Swagger JSON 模板
const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "title": "五常同城本地生活服务平台 API",
    "description": "五常同城后端服务 API 文档",
    "version": "0.2.0"
  },
  "basePath": "/api/v1",
  "paths": {},
  "securityDefinitions": {
    "BearerAuth": {
      "type": "apiKey",
      "name": "Authorization",
      "in": "header",
      "description": "JWT Bearer Token，格式：Bearer {token}"
    }
  }
}`

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
