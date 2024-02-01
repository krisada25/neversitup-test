package config

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

var DbPostgres *sqlx.DB

func PostgresConn() {
	hostEnv := os.Getenv("DB_HOST")
	portEnv := os.Getenv("DB_PORT")
	userEnv := os.Getenv("DB_USER")
	passEnv := os.Getenv("DB_PASS")
	dbnameEnv := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSL_MODE")
	dbClientCertPath := os.Getenv("DB_CLIENT_CERT_PATH")
	dbClientKeyPath := os.Getenv("DB_CLIENT_KEY_PATH")
	dbServerCAPath := os.Getenv("DB_SERVER_CA_PATH")

	DBStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", hostEnv, portEnv, userEnv, passEnv, dbnameEnv, dbSSLMode, dbClientCertPath, dbClientKeyPath, dbServerCAPath)
	db, err := sqlx.Connect("postgres", DBStr)

	CheckError(err)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	err = db.Ping()
	CheckError(err)
	
	DbPostgres = db
	fmt.Println("Postgres Connected!")
}

func CheckError(err error) {
	if err != nil {
		fmt.Printf("Error DB")
		panic(err)
	}
}
