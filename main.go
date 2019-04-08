package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	go_version "github.com/hashicorp/go-version"
	_ "github.com/lib/pq"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

type Package struct {
	packageName   string
	latestVersion string
	gitLocation   string
}

var secretToken string

const tokenFile = ".deb-buildpackage.token"

func validate(c *gin.Context, body []byte) error {
	hashStr := c.GetHeader("X-Hub-Signature")
	if hashStr == "" {
		return errors.New("")
	}

	h := hmac.New(sha1.New, []byte(secretToken))
	_, err := h.Write(body)
	if err != nil {
		return errors.New(err.Error())
	}

	realHash := "sha1=" + hex.EncodeToString(h.Sum(nil))
	if realHash != hashStr {
		return errors.New("")
	}

	return nil
}

func packageLatestVersin(package_name string) string {
	s := fmt.Sprintf(`select * from packages where package_name = '%s'`, package_name)

	rows, err := db.Query(s)
	if err != nil {
		return ""
	}

	var package_ Package
	for rows.Next() {
		if err = rows.Scan(&package_.packageName, &package_.latestVersion, &package_.gitLocation); err != nil {
			return ""
		}

		return package_.latestVersion
	}

	return ""
}

func checkVersion(package_name, version string) error {
	latestVersion := packageLatestVersin(package_name)
	if latestVersion == "" {
		return nil
	}

	latestVersion_compare, _ := go_version.NewVersion(latestVersion)
	newVersion_compare, _ := go_version.NewVersion(version)
	if latestVersion_compare.LessThan(newVersion_compare) {
		return nil
	}

	return errors.New("")
}

func handlePushEvent(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		fmt.Println("Failed to get request body")
		return
	}

	if validate(c, body) != nil {
		return
	}

	eventType := gjson.Get(string(body), "ref_type").String()

	if eventType != "tag" {
		debug("Not a tag event, event:", eventType)
		return
	}

	tag := gjson.Get(string(body), "ref").String()

	debug("Tag:", tag)

	validateTag := regexp.MustCompile(`debiancn/(.*)`)
	m := validateTag.FindStringSubmatch(tag)
	if len(m) < 2 {
		debug("Tag is not comply to format 'debiancn/X.X.X', tag:", tag)
		return
	}
	version := m[1]

	debug("version:", version)

	packageName := gjson.Get(string(body), "repository.name").String()
	debug("package name:", packageName)

	if checkVersion(packageName, version) != nil {
		return
	}

	insertSQL := fmt.Sprintf(
		`insert into need_build_git_packages values('%s', '%s', '%s');`,
		packageName, version, "new")

	debug("insertSql:", insertSQL)
	if _, err = db.Exec(insertSQL); err != nil {
		fmt.Println("Failed to insert package to need_build_git_package table", err.Error())
	}
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func getSecretToken(token *string) error {
	home, _ := os.UserHomeDir()
	absTokenFile := path.Join(home, tokenFile)
	t, err := ioutil.ReadFile(absTokenFile)
	if err != nil {
		fmt.Println(err.Error())
		return errors.New(err.Error())
	}
	*token = strings.TrimSpace(string(t))

	return nil
}

func main() {
	flag.BoolVar(&enableDebugOutput, "debug", false, "enable debug output")
	flag.Parse()

	if getSecretToken(&secretToken) != nil {
		fmt.Println("Failed to initilize secret token")
		return
	}

	if !enableDebugOutput {
		gin.SetMode(gin.ReleaseMode)
	}

	checkEnv()
	go build()
	defer db.Close()

	r := gin.Default()
	r.GET("ping", handlePing)
	r.POST("push_event", handlePushEvent)
	r.Run("0.0.0.0:4567")
}
