package appentry

import (
	"time"

	"github.com/imroc/req/v3"
	"github.com/topjohncian/cloudreve-pro-epay/internal/appconf"
	"github.com/topjohncian/cloudreve-pro-epay/internal/cache"
	"github.com/topjohncian/cloudreve-pro-epay/internal/controller"
	"github.com/topjohncian/cloudreve-pro-epay/internal/server"
	"github.com/topjohncian/cloudreve-pro-epay/internal/usdtmore"
	"go.uber.org/fx"
)

func AppEntry() []fx.Option {
	return []fx.Option{
		fx.Provide(appconf.Parse),
		fx.Provide(Log),
		fx.WithLogger(FxLogger),

		cache.Cache(),
		fx.Provide(server.CreateHttp),
		fx.Provide(func(c *appconf.Config) *req.Client {
			if c.Debug {
				req.DevMode()
			}
			return req.C()
		}),
		controller.Module(),

		// 添加 USDTMore 模块
		usdtmore.Module,

		fx.StartTimeout(1 * time.Second),
		fx.StopTimeout(5 * time.Minute),
	}
}
