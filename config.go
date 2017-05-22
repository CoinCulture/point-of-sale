package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	db   *DbConfig
	http *HttpConfig
}

type HttpConfig struct {
	port int
}

type DbConfig struct {
	net          string
	host         string
	port         int
	databaseName string
	user         string
}

func (c DbConfig) getAddr() string {
	if c.net == "tcp" {
		return fmt.Sprintf("%s:%s", c.host, strconv.Itoa(c.port))
	}
	return ""
}

func LoadConfig() *Config {
	//load the config file
	viper.SetConfigName("config")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file (./config.toml) not found - abort")
		os.Exit(1)
	}

	viper.SetDefault("http.port", 8080)

	return &Config{
		db: &DbConfig{
			net:          viper.GetString("database.net"),
			host:         viper.GetString("database.host"),
			port:         viper.GetInt("database.port"),
			databaseName: viper.GetString("database.name"),
			user:         viper.GetString("database.user"),
		},
		http: &HttpConfig{
			port: viper.GetInt("http.port"),
		},
	}
}
