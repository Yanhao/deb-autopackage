package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	go_version "github.com/hashicorp/go-version"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"net/http"
	"regexp"
	"strings"
)

type Package struct {
	packageName string
	latestVersion string
	gitLocation string
}

func validate() bool {
	return true
}

func packageLatestVersin(package_name string) string {
	s := strings.Builder{}
	fmt.Fprintf(&s, `select * from packages where package_name = '%s'`, package_name)

	rows, err := db.Query(s.String())
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

	eventType := gjson.Get(string(body), "ref_type").String()

	if eventType != "tag" {
		debug("Not a tag event, event:", eventType)
		return
	}

	tag := gjson.Get(string(body), "ref").String()

	debug("Tag:", tag)

	validateTag := regexp.MustCompile(`debian/(.*)`)
	m := validateTag.FindStringSubmatch(tag)
	if len(m) < 2 {
		debug("Tag is not comply to format 'debian/X.X.X', tag:", tag)
		return
	}
	version := m[1]

	debug("version:", version)

	packageName := gjson.Get(string(body), "repository.name").String()
	debug("package name:", packageName)

	if checkVersion(packageName, version) != nil {
		return
	}

	insertSQL := strings.Builder{}
	fmt.Fprintf(&insertSQL,
		`insert into need_build_git_packages values('%s', '%s', '%s');`,
		packageName, version, "new")

	debug("insertSql:", insertSQL.String())
	if _, err = db.Exec(insertSQL.String()); err != nil {
		fmt.Println("Failed to insert package to need_build_git_package table", err.Error())
	}
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func main() {
	flag.BoolVar(&enableDebugOutput, "debug", true, "enable debug output")
	flag.Parse()

	//gin.SetMode(gin.ReleaseMode)

	checkEnv()
	go build()
	defer db.Close()

	r := gin.Default()
	r.GET("ping", handlePing)
	r.POST("push_event", handlePushEvent)
	r.Run("127.0.0.1:4567")
}
