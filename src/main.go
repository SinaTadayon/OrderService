package main

import (
	"errors"
	"log"
	"os"

	"github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

var App struct {
	config configuration
}

func main() {}

func init() {
	err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func LoadConfig() error {
	if os.Getenv("APP_ENV") == "dev" {
		//if flag.Lookup("test.v") != nil {
		//	// test mode
		//	err := godotenv.Load("testdata/.env")
		//	if err != nil {
		//		logger.Err("Error loading testdata .env file")
		//	}
		//} else {
		err := godotenv.Load(".env")
		if err != nil {
			return errors.New("Error loading .env file")
		}
		//}
	}

	// Get environment variables for config
	_, err := env.UnmarshalFromEnviron(&App.config)
	if err != nil {
		return err
	}
	return nil
}
