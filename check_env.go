package main

import (
	"database/sql"
	"fmt"
//	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const dbFileLocation = "db.db"

func checkEnv() {
	/*
		if os.Geteuid() != 0 {
			fmt.Println("Only root can run this program")
			os.Exit(1)
		}
	*/

	connStr := "user=postgres password='postgres' dbname=debiancn_package"
	var err error
	if db, err = sql.Open("postgres", connStr); err != nil {
		//fmt.Println("Failed to open", dbFileLocation)
		fmt.Println("Failed to connect to postgres")
	}

	/*
	if _, err := os.Stat(dbFileLocation); os.IsNotExist(err) {
		//fmt.Println(dbFileLocation, "doesn't exists, creating it...")

		if db, err = sql.Open("postgres", connStr); err != nil {
			// fmt.Println("Failed to create", dbFileLocation)
			fmt.Println("Failed to connect to postgres")
		}

		createNeedBuildTableSQL := `
		create table need_build_git_packages (
		package_name text not null,
		version text not null,
		status text );
		`

		createPackageTableSQL := `
		create table packages (
		package_name text primary key,
		latest_version text,
		git_location text );
		`

		if _, err := db.Exec(createNeedBuildTableSQL); err != nil {
			fmt.Println("Failed to create table: need_build_git_packages")
		}

		if _, err = db.Exec(createPackageTableSQL); err != nil {
			fmt.Println("Failed to create table: packages")
		}
	} else {
		if db, err = sql.Open("postgres", connStr); err != nil {
			//fmt.Println("Failed to open", dbFileLocation)
			fmt.Println("Failed to connect to postgres")
		}
	}

	 */
}
