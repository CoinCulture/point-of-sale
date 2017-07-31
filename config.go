package main

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	db    *DbConfig
	http  *HttpConfig
	items *ItemsConfig
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

type ItemsConfig struct {
	dontPrint  []string
	dontNotify []string
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
		fmt.Println("[NOTICE] Config file (./config.toml) not found - assuming defaults for all params")
	}

	viper.SetDefault("database.net", "tcp")
	viper.SetDefault("database.host", "127.0.0.1")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.name", "myBusiness")
	viper.SetDefault("database.user", "root")
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
		items: &ItemsConfig{
			dontPrint:  viper.GetStringSlice("items.dont_print"),
			dontNotify: viper.GetStringSlice("items.dont_notify"),
		},
	}
}
