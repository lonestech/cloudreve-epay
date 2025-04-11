package server

import (
	"html/template"
	"io/fs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/topjohncian/cloudreve-pro-epay/internal/appconf"
)

func CreateHttp(conf *appconf.Config, templateFS fs.FS) *gin.Engine {
	r := gin.Default()

	gin.SetMode(gin.ReleaseMode)
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	}

	// 列出模板文件
	entries, err := fs.ReadDir(templateFS, "templates")
	if err != nil {
		logrus.WithError(err).Error("无法读取模板目录")
	} else {
		for _, entry := range entries {
			logrus.Infof("找到模板文件: %s", entry.Name())
		}
	}

	// 尝试加载模板
	tmpl, err := template.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		logrus.WithError(err).Error("无法加载模板文件")
		// 使用一个简单的默认模板
		tmpl = template.Must(template.New("default").Parse(`<!DOCTYPE html>
<html><body><h1>模板加载失败</h1><p>请检查模板文件是否存在。</p></body></html>`))
	}

	r.SetHTMLTemplate(tmpl)

	r.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": conf.Listen})
	})

	return r
}
