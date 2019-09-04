package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var EnvFile = ".env"
var ConfigurationFile = "configuration.go"

func AppendToFile() {
	f, err := os.OpenFile(EnvFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString("\n__CTO__")
	if err != nil {
		log.Fatal(err)
	}
}
func FixEnvFile() {
	f, err := ioutil.ReadFile(EnvFile)
	if err != nil {
		log.Fatal(err)
	}
	newContent := bytes.ReplaceAll(f, []byte("\n__CTO__"), []byte{})
	err = ioutil.WriteFile(EnvFile, newContent, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

//func UpdateConfigurationFile() {
//	f, err := ioutil.ReadFile(ConfigurationFile)
//	if err != nil {
//		log.Fatal(err)
//	}
//	newContent := bytes.ReplaceAll(f, []byte("Port string `env:\"PORT\"`"), []byte("Port int `env:\"PORT\"`"))
//	err = ioutil.WriteFile(ConfigurationFile, newContent, os.ModePerm)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
func FixConfigurationFile() {
	f, err := ioutil.ReadFile(ConfigurationFile)
	if err != nil {
		log.Fatal(err)
	}
	newContent := bytes.ReplaceAll(f, []byte("Port int `env:\"PORT\"`"), []byte("Port string `env:\"PORT\"`"))
	err = ioutil.WriteFile(ConfigurationFile, newContent, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	FixEnvFile()
	FixConfigurationFile()
	os.Exit(m.Run())
}
func TestLoadConfig_AssertTrue(t *testing.T) {
	err := os.Setenv("APP_ENV", "dev")
	assert.Nil(t, err)
	err = LoadConfig()
	assert.Nil(t, err)
}
