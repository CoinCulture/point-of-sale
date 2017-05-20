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
	password     string
}

func (c DbConfig) getAddr() string {
	if c.net == "tcp" {
		return c.host + ":" + strconv.Itoa(c.port)
	}
	return ""
}

func LoadConfig(env *string) *Config {
	//load the config file
	viper.SetConfigName("app")
	viper.AddConfigPath("config")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file (./config/app.[json|yaml|toml]) not found - abort")
		os.Exit(1)
	}

	viper.SetDefault("env."+*env+".http.port", 8080)

	return &Config{
		db: &DbConfig{
			net:          viper.GetString("env." + *env + ".db.net"),
			host:         viper.GetString("env." + *env + ".db.host"),
			port:         viper.GetInt("env." + *env + ".db.port"),
			databaseName: viper.GetString("env." + *env + ".db.name"),
			user:         viper.GetString("env." + *env + ".db.user"),
			password:     viper.GetString("env." + *env + ".db.pass"),
		},
		http: &HttpConfig{
			port: viper.GetInt("env." + *env + ".http.port"),
		},
	}
}
