package main

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("gook.yaml")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "/")
	viper.SetDefault("ignore", "")

	r := gin.Default()
	r.Use(handler)
}

func handler(c *gin.Context) {
	//c.Request.
}

func executor(path string, key string) {

}
