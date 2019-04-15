package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moqmar/go-gitignore"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("gook")
	viper.AddConfigPath("/etc/")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath(".")

	viper.SetDefault("prefix", "/")
	viper.SetDefault("ignore", "/proc/\n/sys/\n/dev/\n.git/\nnode_modules/")

	err := viper.ReadInConfig()
	if err != nil {
		if strings.HasPrefix(err.Error(), "Config File \"gook\" Not Found in ") {
			fmt.Printf("No config file found, using default.\n")
		} else {
			fmt.Printf("Couldn't parse config: %s\n", err)
			os.Exit(1)
		}
	}

	gi = gitignore.NewGitIgnoreFromReader("/", strings.NewReader(viper.GetString("ignore")))
	prefix = viper.GetString("prefix")

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(handler)

	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(host + ":" + port)
}

var gi gitignore.IgnoreMatcher
var prefix string
var gookline = regexp.MustCompile(`^#@[gG][oO][oO][kK]((?:\+[^\+:]+)*):(.*)$`)

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
		c.String(code, "couldn't find .webhook candidate\n")
	} else if output == "" && err != nil {
		c.String(code, "%s\n", err)
	} else {
		if err != nil && strings.HasPrefix(err.Error(), "exit status ") {
			c.Header("Gook-Status", strings.TrimPrefix(err.Error(), "exit status "))
			code = 424
		} else if err != nil {
			c.Header("Gook-Error", err.Error())
		}
		c.String(code, "%s\n", output)
	}

}

type flagsType struct {
	stdin bool
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
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		return "", 404, err
	}
	if isPrefix {
		return "", 500, errors.New("couldn't fully buffer first line")
	}
	if !strings.HasPrefix(string(line), "#!") {
		return "", 500, errors.New("not a script")
	}

	// Validate second line
	line, isPrefix, err = reader.ReadLine()
	if err != nil {
		return "", 404, err
	}
	if isPrefix {
		return "", 500, errors.New("couldn't fully buffer second line")
	}

	flags := flagsType{}

	// Validate key
	if m := gookline.FindStringSubmatch(string(line)); m != nil {
		flagsString := strings.Split(m[1], "+")
		for _, flag := range flagsString {
			switch strings.ToLower(flag) {
			// Parse flags
			case "stdin":
				flags.stdin = true
			case "":
			default:
				return "", 500, errors.New("invalid flag in .webhook: " + flag)
			}
		}
		if m[2] != key {
			return "", 403, errors.New("invalid key")
		}
	} else {
		return "", 404, errors.New(".webhook is not a valid webhook script (check syntax of second line)")
	}

	webhookFile.Close()

	cmd := exec.Command(webhookPath)
	cmd.Env = append(os.Environ(), args...)
	cmd.Dir = filepath.Join(prefix, path)

	if flags.stdin {
		cmd.Stdin = input
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), 418, err
	}
	return string(out), 200, nil
}
