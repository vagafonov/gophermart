package logger

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	var zl *zap.Logger
	var err error
	if viper.GetString("mode") == "dev" {
		zl, err = zap.NewDevelopment()
	} else {
		zl, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}

	return zl.Sugar()
}
