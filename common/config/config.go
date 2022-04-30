package config

import (
	"github.com/Octane0411/qoou/common/logger"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var ServerSetting *ServerSettingS

var EmailSetting *EmailSettingS

func setupSetting() {
	setting, err := NewSetting()
	if err != nil {
		logger.Logger.Panic(err)
	}
	err = setting.ReadSection("Server", &ServerSetting)
	if err != nil {
		logger.Logger.Panic(err)
	}
}

type Setting struct {
	vp *viper.Viper
}

func NewSetting() (*Setting, error) {
	vp := viper.New()
	vp.SetConfigName("config")
	configDir := GetConfigPath()
	vp.AddConfigPath(configDir)
	vp.SetConfigType("yaml")
	err := vp.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return &Setting{
		vp: vp,
	}, nil
}

func GetConfigPath() string {
	ex, err := os.Executable()
	if err != nil {
		logger.Logger.Panic(err)
	}
	exPath := filepath.Dir(ex)
	rootPath, err := filepath.EvalSymlinks(exPath)
	configPath := filepath.Join(rootPath, "config")
	if err != nil {
		logger.Logger.Panic(err)
	}
	return configPath
}
