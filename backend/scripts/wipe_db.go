package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Connect to TiDB without specifying DB initially to allow dropping it
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true",
		"root", "", "127.0.0.1", "4000")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	log.Println("💣 Wiping Database 'test'...")

	_, err = db.Exec("DROP DATABASE IF EXISTS test")
	if err != nil {
		log.Fatalf("Failed to drop database: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE test")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	log.Println("✅ Database 'test' recreated successfully.")
}
