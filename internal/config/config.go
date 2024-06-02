package config

import (
	"fmt"
	"gophermart/pkg/utils/fs"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Init(lr *zap.SugaredLogger, name string) {
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fs.Config(""))
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func ParseFlags(lr *zap.SugaredLogger) {
	cmd := &cobra.Command{
		Use:   "gophermart",
		Short: "Gophermart",
	}
	cmd.Flags().StringP("run_address", "a", "", "address and port for starting the service")
	cmd.Flags().StringP("database_uri", "d", "", "database connection address")
	cmd.Flags().StringP("accrual_system_address", "r", "", "address of accrual calculation system")
	err := viper.BindPFlag("run_address", cmd.Flags().Lookup("run_address"))
	if err != nil {
		panic(fmt.Errorf("failed to bind flag run_address: %w", err))
	}
	err = viper.BindPFlag("database_uri", cmd.Flags().Lookup("database_uri"))
	if err != nil {
		panic(fmt.Errorf("failed to bind flag database_uri: %w", err))
	}
	err = viper.BindPFlag("accrual_system_address", cmd.Flags().Lookup("accrual_system_address"))
	if err != nil {
		panic(fmt.Errorf("failed to bind flag accrual_system_address: %w", err))
	}

	// Связывание переменных среды с переменными конфигурации
	viper.AutomaticEnv()

	// Получение значений из флага или переменной среды

	if err := cmd.Execute(); err != nil {
		lr.Panic(fmt.Sprintf("failed to execute cobra command: %s", err.Error()))

		return
	}
}

func DebugConfig(lr *zap.SugaredLogger) {
	lr.Info("configuration variables: ",
		zap.String("run_address", viper.GetString("run_address")),
		zap.String("database_uri", viper.GetString("database_uri")),
		zap.String("accrual_system_address", viper.GetString("accrual_system_address")),
	)
}
