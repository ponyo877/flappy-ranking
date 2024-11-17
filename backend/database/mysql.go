package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func NewDBConfig() *DBConfig {
	return &DBConfig{
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}
}

func NewMySQL() (*sql.DB, error) {
	config := NewDBConfig()
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	c := mysql.Config{
		DBName:    config.Database,
		User:      config.User,
		Passwd:    config.Password,
		Addr:      addr,
		Net:       "tcp",
		ParseTime: true,
		Collation: "utf8mb4_unicode_ci",
		Loc:       jst,
	}
	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
