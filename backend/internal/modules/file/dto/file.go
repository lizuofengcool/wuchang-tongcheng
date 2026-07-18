// Package dto 文件模块数据传输对象
package dto

// FileInfo 文件信息
type FileInfo struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	FileName  string `json:"file_name"`
	FileURL   string `json:"file_url"`
	FileSize  int64  `json:"file_size"`
	FileType  string `json:"file_type"`
	MimeType  string `json:"mime_type"`
	RegionID  uint   `json:"region_id"`
	CreatedAt string `json:"created_at"`
}

// ListFilesRequest 文件列表查询请求
type ListFilesRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	FileType string `form:"file_type" json:"file_type"` // image/video/doc/archive/audio
	Keyword  string `form:"keyword" json:"keyword"`     // 文件名关键词
}

// PresignUploadRequest 预签名直传申请请求
//
// 前端先调用此接口换取预签名 PUT URL，再直接将文件 PUT 到对象存储，
// 上传完成后调用 POST /file/commit 提交记录。可绕过后端带宽，适合大文件。
type PresignUploadRequest struct {
	FileName string `json:"file_name" binding:"required"` // 原始文件名（含扩展名，用于类型校验）
}

// PresignUploadResponse 预签名直传响应
type PresignUploadResponse struct {
	UploadURL  string `json:"upload_url"`  // 预签名 PUT URL，前端直接 PUT 文件二进制到此地址
	AccessURL  string `json:"access_url"`  // 上传完成后可公开访问的 URL（展示用）
	ObjectName string `json:"object_name"` // 对象名，提交记录时原样回传
	ExpiresIn  int    `json:"expires_in"`   // 上传 URL 有效期（秒）
	FileName   string `json:"file_name"`    // 原始文件名
	FileType   string `json:"file_type"`    // 文件分类（image/video/doc/archive/audio）
}

// CommitUploadRequest 直传完成后提交文件记录请求
//
// 前端 PUT 成功后调用，后端按 object_name 重新拼装访问 URL（避免前端伪造），
// 校验文件类型与大小后写入 file_uploads 表。
type CommitUploadRequest struct {
	FileName   string `json:"file_name" binding:"required"`    // 原始文件名（含扩展名）
	ObjectName string `json:"object_name" binding:"required"` // 预签名阶段返回的对象名
	MimeType   string `json:"mime_type"`                      // MIME 类型（前端可填，后端兜底推断）
	FileSize   int64  `json:"file_size" binding:"required"`    // 文件大小（字节，前端 PUT 后已知）
}

// STSCredentialsResponse STS 临时凭据响应
//
// 前端拿到后用 OSS 浏览器 SDK 或 SigV4 + x-amz-security-token 头直接 PUT 到 OSS。
// 一组凭据可在 expires_in 内上传任意多对象，过期后需重新申请。
// 与预签名直传（/file/presign 单对象一次性 URL）互补，适合批量/大文件场景。
type STSCredentialsResponse struct {
	AccessKeyID     string `json:"access_key_id"`     // 临时 AccessKeyID
	AccessKeySecret string `json:"access_key_secret"` // 临时 AccessKeySecret
	SecurityToken   string `json:"security_token"`    // 安全令牌（OSS 请求头 x-amz-security-token）
	Expiration      string `json:"expiration"`        // 凭据过期时间（ISO8601，阿里云原样透传）
	Bucket          string `json:"bucket"`            // OSS 桶名
	Region          string `json:"region"`            // OSS 区域，如 oss-cn-hangzhou
	Endpoint        string `json:"endpoint"`          // OSS 端点，如 https://oss-cn-hangzhou.aliyuncs.com
	ObjectPrefix    string `json:"object_prefix"`     // 对象 key 前缀，如 uploads/2026/07/
	ExpiresIn       int    `json:"expires_in"`        // 凭据剩余有效期（秒）
}
