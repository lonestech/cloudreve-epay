package appconf

import (
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../..")
)

func Parse() (*Config, error) {
	// Try to load .env file if it exists, but don't fail if it doesn't
	if err := godotenv.Load(".env"); err != nil {
		logrus.Debugln("No .env file found in current directory")
	}
	if err := godotenv.Load(filepath.Join(Root, ".env")); err != nil {
		logrus.Debugln("No .env file found in project root")
	}

	var config Config
	err := envconfig.Process("cr_epay", &config)

	if err != nil {
		envconfig.Usage("cr_epay", &config)
		logrus.WithError(err).Fatalln("无法加载配置")
		return nil, err
	}

	return &config, nil
}
