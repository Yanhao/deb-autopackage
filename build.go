package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const buildWorkDir = "/var/lib/deb-buildpackage/"

type NewPackage struct {
	packageName string
	version     string
	status      string
}

func addToRepo(packageName, version, codename string) {
	changesFile := buildWorkDir + "/" + packageName + "_" + version + "_amd64.changes"

	if _, err := os.Stat(changesFile); os.IsNotExist(err) {
		fmt.Println("Failed to find the changes File", changesFile)
		return
	}
	debug("changes file:", changesFile)

	return
	err := exec.Command("aptly", "repo", "add", codename, changesFile)
	if err != nil {
		//fmt.Println("Failed:", err.Error()
	}
}

func buildPackage(packageName, version string) {
	packageGitLocation := buildWorkDir + packageName
	if err := os.Chdir(packageGitLocation); err != nil {
		fmt.Println("Failed to change dir to", packageGitLocation)
		return
	}

	syncCommand := exec.Command("git", "pull")
	syncCommand.Start()
	if err := syncCommand.Wait(); err != nil {
		fmt.Println("Failed to sync with github")
		// TODO: remove from  need_build_git_package
		return
	}
	fmt.Println("Sync git repo successfully!")

	buildCommand := exec.Command("gbp", "buildpackage", "--git-ignore-branch", "--git-pbuilder")
	output, err := buildCommand.StdoutPipe()
	if err != nil {
		fmt.Println("Failed:", err.Error())
		return
	}

	err = buildCommand.Start()
	if err != nil {
		fmt.Println("Failed:", err.Error())
		return
	}

	scanner := bufio.NewScanner(output)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err = buildCommand.Wait(); err != nil {
		fmt.Println("Failed:", err.Error())
		return
	}

	updateSQL := strings.Builder{}
	fmt.Fprintf(&updateSQL, `
						update need_build_git_packages
						set status = 'finished'
						where package_name = '%s' and version = '%s';`, packageName, version)

	if _, err = db.Exec(updateSQL.String()); err != nil {
		fmt.Println("Failed to update need_build_package", err.Error())
	}

	addToRepo(packageName, version, "buster")
}

func build() {
	ticker := time.NewTicker(time.Duration(10) * time.Second)
	defer ticker.Stop()

	// 这么做的意义貌似不是很大
	for db.Ping() != nil {
		time.Sleep(time.Second * 5)
	}

	for {
		select {
		case _ = <-ticker.C:
			rows, err := db.Query("select * from need_build_git_packages;")
			if err != nil {
				continue
			}

			for rows.Next() {
				var package_ NewPackage

				if err := rows.Scan(&package_.packageName, &package_.version, &package_.status); err != nil {
					fmt.Println("Failed to read new packages from database", err.Error())
				}

				if package_.status == "new" {
					debug("New package:", package_.packageName, package_.version)
					updateSQL := strings.Builder{}
					fmt.Fprintf(&updateSQL, `
						update need_build_git_packages
						set status = 'building'
						where package_name = '%s' and version = '%s';`, package_.packageName, package_.version)

					if _, err = db.Exec(updateSQL.String()); err != nil {
						fmt.Println("Failed to update need_build_package", err.Error())
						continue
					}

					go buildPackage(package_.packageName, package_.version)
				}
			}

			rows.Close()
		}
	}
}