package config

import (
	"log"

	env "github.com/caarlos0/env/v10"
)

var Values = config{}

type config struct {
	Index     indexConf
	DataProxy dataProxyConf
	Mongo     mongoConf
}

type indexConf struct {
	URL string `env:"INDEX_HOST,required"`
}

type dataProxyConf struct {
	URL string `env:"EXTERNAL_DATA_PROXY_URL,required"`
}

type mongoConf struct {
	USERNAME string `env:"MONGO_USERNAME,required"`
	PASSWORD string `env:"MONGO_PASSWORD,required"`
	HOST     string `env:"MONGO_HOST,required"`
	DBName   string `env:"MONGO_DB_NAME,required"`
}

func Init() {
	err := env.Parse(&Values)
	if err != nil {
		log.Fatalf("Failed to decode environment variables: %s", err)
	}
}
