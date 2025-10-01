package main

import (
	"github.com/iamdanielyin/mod"
)

// StaticInfoRequest represents static info request
type StaticInfoRequest struct {
	// Empty struct for static info request
}

// StaticInfoResponse represents static mount configuration
type StaticInfoResponse struct {
	Mounts []MountInfo `json:"mounts"`
}

// MountInfo represents a static mount point
type MountInfo struct {
	URLPrefix  string `json:"url_prefix"`
	LocalPath  string `json:"local_path"`
	Browseable bool   `json:"browseable"`
	IndexFile  string `json:"index_file"`
}

func main() {
	app := mod.New()

	// Register static info service
	app.Register(mod.Service{
		Name:        "static_info",
		DisplayName: "静态文件配置",
		Description: "获取静态文件挂载配置信息",
		Handler: mod.MakeHandler(func(ctx *mod.Context, req *StaticInfoRequest, resp *StaticInfoResponse) error {
			config := ctx.App().GetModConfig()
			if config == nil {
				resp.Mounts = []MountInfo{}
				return nil
			}

			var mounts []MountInfo
			for _, mount := range config.StaticMounts {
				mounts = append(mounts, MountInfo{
					URLPrefix:  mount.URLPrefix,
					LocalPath:  mount.LocalPath,
					Browseable: mount.Browseable,
					IndexFile:  mount.IndexFile,
				})
			}

			resp.Mounts = mounts
			return nil
		}),
		Group:    "静态文件",
		Sort:     1,
		SkipAuth: true,
	})

	app.Run(":8080")
}