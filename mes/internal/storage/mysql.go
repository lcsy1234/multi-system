package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"os"
	"regexp"
	"strings"
	"time"
)

//go:embed schema.sql
var schema string

func Open() (*sql.DB, error) {
	dsn := os.Getenv("MES_MYSQL_DSN")
	if dsn == "" {
		dsn = "root@tcp(127.0.0.1:3306)/multi_system_mes_demo"
	}
	config, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return nil, errors.New("MySQL DSN 格式无效")
	}
	if config.Params == nil {
		config.Params = map[string]string{}
	}
	config.Params["time_zone"] = "'+00:00'"
	config.ParseTime = true
	config.Loc = time.UTC
	config.Timeout = 3 * time.Second
	config.ReadTimeout = 5 * time.Second
	config.WriteTimeout = 5 * time.Second
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}
func Migrate(ctx context.Context) error {
	db, err := Open()
	if err != nil {
		return err
	}
	defer db.Close()
	// Connect without a schema so the named demo schema can be created explicitly.
	dsn := os.Getenv("MES_MYSQL_DSN")
	if dsn == "" {
		dsn = "root@tcp(127.0.0.1:3306)/multi_system_mes_demo"
	}
	config, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return err
	}
	name := config.DBName
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(name) {
		return errors.New("数据库名只能包含字母、数字和下划线")
	}
	config.DBName = ""
	config.Timeout = 3 * time.Second
	admin, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return err
	}
	defer admin.Close()
	if _, err = admin.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_bin"); err != nil {
		return err
	}
	for _, statement := range strings.Split(schema, ";") {
		if strings.TrimSpace(statement) != "" {
			if _, err = db.ExecContext(ctx, statement); err != nil {
				return err
			}
		}
	}
	return nil
}
