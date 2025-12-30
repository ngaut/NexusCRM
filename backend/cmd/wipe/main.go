package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	log.Printf("🧹 Wiping database: %s", database)

	// Disable foreign key checks for dropping
	_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		log.Fatalf("Failed to disable FK checks: %v", err)
	}

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("Failed to show tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, table)
	}

	for _, table := range tables {
		log.Printf("🔥 Dropping table: %s", table)
		_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s` CASCADE", table))
		if err != nil {
			log.Printf("⚠️ Warning: Failed to drop table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	_, err = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		log.Fatalf("Failed to enable FK checks: %v", err)
	}

	log.Println("✅ Database wipe complete")
}
