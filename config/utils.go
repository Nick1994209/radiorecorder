package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)


func getEnv(key, fallback string) string {
	loadDotEnv()

	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var envfileIsLoaded = false
func loadDotEnv()  {
	if envfileIsLoaded {
		return
	}
	dotEnvFile, err := filepath.Abs(".env")
	if err != nil {
		log.Warning("Can not get abs .env path")
	}

	if err = godotenv.Load(dotEnvFile); err == nil {
		log.Infof("Variables from file=%s were loaded", dotEnvFile)
	} else {
		log.Warningf("Can not load secure variables from file=%s", dotEnvFile)
	}
	envfileIsLoaded = true
}
