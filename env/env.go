package env

import (
	"github.com/spf13/viper"
)

type Config struct {
	GrpcPort    string
	HttpPort    string
	Database    *DatabaseConfig
	UserService *UserServiceConfig
}

type DatabaseConfig struct {
	PSQLConfig
}

type PSQLConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DbName   string
	SslMode  string
}

type UserServiceConfig struct {
	Host string
	Port string
}

var Settings *Config

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	viper.BindEnv("GRPC_PORT")
	viper.BindEnv("HTTP_PORT")

	viper.BindEnv("USER_SERVICE_HOST")
	viper.BindEnv("USER_SERVICE_PORT")

	viper.BindEnv("POSTGRES_HOST")
	viper.BindEnv("POSTGRES_PORT")
	viper.BindEnv("POSTGRES_USERNAME")
	viper.BindEnv("POSTGRES_PASSWORD")
	viper.BindEnv("POSTGRES_DBNAME")
	viper.BindEnv("POSTGRES_SSL_MODE")

	Settings = &Config{
		GrpcPort: viper.GetString("GRPC_PORT"),
		HttpPort: viper.GetString("HTTP_PORT"),
		UserService: &UserServiceConfig{
			Host: viper.GetString("USER_SERVICE_HOST"),
			Port: viper.GetString("USER_SERVICE_PORT"),
		},
		Database: &DatabaseConfig{
			PSQLConfig{
				Host:     viper.GetString("POSTGRES_HOST"),
				Port:     viper.GetInt("POSTGRES_PORT"),
				Username: viper.GetString("POSTGRES_USERNAME"),
				Password: viper.GetString("POSTGRES_PASSWORD"),
				DbName:   viper.GetString("POSTGRES_DBNAME"),
				SslMode:  viper.GetString("POSTGRES_SSL_MODE"),
			},
		},
	}
}
