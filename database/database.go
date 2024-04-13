package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/les-cours/organization-service/env"
	_ "github.com/lib/pq"
)

func StartDatabase() (*sql.DB, error) {
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		env.Settings.Database.PSQLConfig.Host,
		env.Settings.Database.PSQLConfig.Port,
		env.Settings.Database.PSQLConfig.Username,
		env.Settings.Database.PSQLConfig.Password,
		env.Settings.Database.PSQLConfig.DbName,
		env.Settings.Database.PSQLConfig.SslMode,
	)
	log.Println(dbinfo)
	db, err := sql.Open("postgres", dbinfo)
	log.Println("staging dbinfo", dbinfo)

	if err != nil {
		log.Fatalf("Failed to connect to postgres database: %v", err)
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(7)

	return db, nil
}
