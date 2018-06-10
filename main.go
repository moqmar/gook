package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moqmar/go-gitignore"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("gook.yaml")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "/")
	viper.SetDefault("ignore", "/proc/\n/sys/\n/dev/")

	gi = gitignore.NewGitIgnoreFromReader("/", strings.NewReader(viper.GetString("ignore")))
	prefix = viper.GetString("prefix")

	r := gin.Default()
	r.Use(handler)
	r.Run(":8080")
}

var gi gitignore.IgnoreMatcher
var prefix string

func handler(c *gin.Context) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	if c.Request.Method != "GET" && c.Request.Method != "POST" {
		c.String(405, "method not allowed")
		return
	}

	uri := c.Request.URL.Path
	path := strings.Split(uri, "/")

	query := c.Request.URL.Query()
	args := []string{}
	for name, value := range query {
		args = append(args, "gook_"+strings.SplitN(name, "=", 1)[0]+"="+value[0])
	}

	output, code, err := executor(strings.Join(path[1:len(path)-1], "/"), path[len(path)-1], args, c.Request.Body)
	if code == 404 {
		fmt.Printf("NO CANDIDATE: [%s] %s\n", uri, err)
		c.String(code, "couldn't find .webhook candidate")
	} else if output == "" && err != nil {
		c.String(code, "%s", err)
	} else {
		if err != nil {
			c.Header("Gook-Error", err.Error())
		}
		c.String(code, "%s", output)
	}

}

func executor(path string, key string, args []string, input io.Reader) (string, int, error) {
	if gi.Match("/"+path, true) {
		return "", 404, errors.New("this path is ignored")
	}

	webhookPath := filepath.Join(prefix, path, ".webhook")
	f, err := os.Stat(webhookPath)
	if err != nil {
		return "", 404, err
	}
	if f.IsDir() {
		return "", 404, errors.New(".webhook is a directory")
	}

	webhookFile, err := os.Open(webhookPath)
	defer webhookFile.Close()
	if err != nil {
		return "", 404, err
	}

	reader := bufio.NewReader(webhookFile)

	// Validate first line
	line, prefix, err := reader.ReadLine()
	if err != nil {
		return "", 404, err
	}
	if prefix {
		return "", 500, errors.New("couldn't fully buffer first line")
	}
	if !strings.HasPrefix(string(line), "#!") {
		return "", 500, errors.New("not a script")
	}

	// Validate second line
	line, prefix, err = reader.ReadLine()
	if err != nil {
		return "", 404, err
	}
	if prefix {
		return "", 500, errors.New("couldn't fully buffer second line")
	}
	if !strings.HasPrefix(string(line), "#@GOOK:") {
		return "", 500, errors.New("no key specified in .webhook")
	}

	fkey := strings.TrimPrefix(string(line), "#@GOOK:")
	if fkey != key {
		return "", 403, errors.New("invalid key")
	}

	webhookFile.Close()

	cmd := exec.Command(webhookPath)
	cmd.Env = append(os.Environ(), args...)

	cmd.Stdin = input

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), 418, err
	}
	return string(out), 200, nil
}
