package main

import (
	"github.com/iamdanielyin/mod"
)

// UploadResponse represents upload result
type UploadResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Path     string `json:"path"`
	URL      string `json:"url"`
	Backend  string `json:"backend"`
}

// UploadInfoRequest represents upload info request
type UploadInfoRequest struct {
	// Empty struct for upload info request
}

// UploadInfoResponse represents upload configuration info
type UploadInfoResponse struct {
	LocalEnabled bool     `json:"local_enabled"`
	S3Enabled    bool     `json:"s3_enabled"`
	OSSEnabled   bool     `json:"oss_enabled"`
	MaxSize      string   `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
	AllowedExts  []string `json:"allowed_exts"`
	UploadURL    string   `json:"upload_url"`
	BatchURL     string   `json:"batch_url"`
}

func main() {
	app := mod.New()

	// Register file info service
	app.Register(mod.Service{
		Name:        "upload_info",
		DisplayName: "上传配置信息",
		Description: "获取文件上传配置信息",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *UploadInfoRequest, resp *UploadInfoResponse) error {
			config := ctx.App().GetModConfig()
			if config == nil {
				resp.LocalEnabled = false
				resp.S3Enabled = false
				resp.OSSEnabled = false
				resp.MaxSize = "10MB"
				resp.UploadURL = "/upload"
				resp.BatchURL = "/upload/batch"
				return nil
			}

			fileConfig := config.FileUpload
			resp.LocalEnabled = fileConfig.Local.Enabled
			resp.S3Enabled = fileConfig.S3.Enabled
			resp.OSSEnabled = fileConfig.OSS.Enabled
			resp.MaxSize = fileConfig.Local.MaxSize
			resp.AllowedTypes = fileConfig.Local.AllowedTypes
			resp.AllowedExts = fileConfig.Local.AllowedExts
			resp.UploadURL = "/upload"
			resp.BatchURL = "/upload/batch"

			return nil
		}),
		Group:    "文件管理",
		Sort:     1,
		SkipAuth: true,
	})

	app.Run(":8080")
}
