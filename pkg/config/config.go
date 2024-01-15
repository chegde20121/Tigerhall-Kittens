package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Debug("Error occurred while reading env file, might fallback to OS env config")
	}
	viper.AutomaticEnv()
}

// This function can be used to get ENV Var in our App
// Modify this if you change the library to read ENV
func GetEnvVar(name string) string {
	if !viper.IsSet(name) {
		logrus.Errorf("Environment variable %s is not set", name)
		return ""
	}
	value := viper.GetString(name)
	return value
}
