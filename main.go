package main

import (
	"os/exec"
	"bufio"
	"io/ioutil"
	"errors"
	"os"
	"path/filepath"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("gook.yaml")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "/")
	viper.SetDefault("ignore", "/proc/\n/sys/ \n/dev/")

	r := gin.Default()
	r.Use(handler)
	r.Run(":8080")
}

func handler(c *gin.Context) {
	uri := c.Request.RequestURI
	s := strings.Split(uri, "/")
	executor(strings.Join(s[1:len(s)-1], "/"), s[len(s)-1])
}

func executor(path string, key string, args []string) (string, error) {
	webhookPath := filepath.Join(viper.GetString("prefix"), path, ".webhook")
	f, err :=os.Stat(webhookPath)
	if err != nil {
		return "", err
	}
	if f.IsDir() {
		return "", errors.New(".webhook is a directory")
	} 
	
	webhookFile, err := os.Open(webhookPath)
	defer webhookFile.Close()
	if err != nil {
		return "", err
	}
	
	reader := bufio.NewReader(webhookFile)
	
	// Validate first line
	line, prefix, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	if prefix {
		return "", errors.New("couldn't fully buffer first line")
	}
	if !strings.HasPrefix(string(line), "#!") {
		return "", errors.New("not a script")
	}
	
	// Validate second line
	line, prefix, err = reader.ReadLine()
	if err != nil {
		return "", err
	}
	if prefix {
		return "", errors.New("couldn't fully buffer second line")
	}
	if !strings.HasPrefix(string(line), "#@GOOK:") {
		return "", errors.New("not a GOOK-file")
	}
	
	fkey := strings.TrimPrefix(string(line), "#@GOOK:")
	if fkey != key {
		return "", errors.New("GOOK-key doesn't match")
	}

	webhookFile.Close()
	
	cmd := exec.Command(webhookPath)
	cmd.Env = append(os.Environ(), args...)
}

