package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	// Try multiple paths
	paths := []string{"../.env", ".env", "../../.env"}
	for _, p := range paths {
		if err := godotenv.Load(p); err == nil {
			log.Printf("Loaded .env from %s", p)
			break
		}
	}

	// Build DSN just like tidb.go
	host := os.Getenv("TIDB_HOST")
	port := os.Getenv("TIDB_PORT")
	user := os.Getenv("TIDB_USER")
	password := os.Getenv("TIDB_PASSWORD")
	database := os.Getenv("TIDB_DATABASE")

	if port == "" {
		port = "4000"
	}

	if database == "" {
		database = "nexuscrm"
	}

	// Register TLS config for TiDB Cloud
	mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
	})

	tlsParam := "&tls=tidb"
	if host == "127.0.0.1" || host == "localhost" {
		tlsParam = ""
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local%s",
		user, password, host, port, database, tlsParam)

	if host == "" || user == "" {
		log.Println("Warning: TIDB_HOST or TIDB_USER not set, connection might fail")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	log.Println("⚠️  Wiping database...")

	// Disable foreign key checks to allow dropping tables in any order
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		log.Fatalf("failed to disable foreign key checks: %v", err)
	}

	// Get all tables
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("failed to list tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			log.Fatalf("failed to scan table: %v", err)
		}
		tables = append(tables, table)
	}

	// Drop all tables
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)); err != nil {
			log.Printf("Failed to drop table %s: %v", table, err)
		} else {
			log.Printf("Dropped table: %s", table)
		}
	}

	// Re-enable foreign key checks
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		log.Fatalf("failed to enable foreign key checks: %v", err)
	}

	log.Println("✅ Database wiped successfully.")
}
