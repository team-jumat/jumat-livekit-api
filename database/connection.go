package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var instance *gorm.DB
var once sync.Once

func GetInstance() (*gorm.DB, error) {
	once.Do(func() {
		err := godotenv.Load("../database/.env")
		if err != nil {
			panic(err)
		}
		host := os.Getenv("MYSQL_HOST")
		port := os.Getenv("MYSQL_PORT")
		user := os.Getenv("MYSQL_USER")
		pass := os.Getenv("MYSQL_PASS")
		dbName := os.Getenv("MYSQL_DB")

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			user, pass, host, port, dbName)

		sqlDb, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		gormDb, err := gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDb,
		}), &gorm.Config{NamingStrategy: &schema.NamingStrategy{
			SingularTable: true,
		},
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			panic(err)
		}
		instance = gormDb
	})
	return instance, nil
}
