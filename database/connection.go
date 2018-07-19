package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

var database *gorm.DB

func Init() *gorm.DB {
	url := environments.PostgresUrl
	parsed := strings.FieldsFunc(url, Split)
	driver := parsed[0]
	driverArgs := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		parsed[1],
		parsed[2],
		parsed[3],
		parsed[4],
		parsed[5])
	db, err := gorm.Open(driver, driverArgs)
	if err != nil {
		fmt.Println("database err: ", err)
		db = nil
	}
	return db
}

func Split(r rune) bool {
	return r == '@' ||
		r == ':' ||
		r == '/'
}

func Close() {
	database.Close()
}

func GetDB() *gorm.DB {
	if database == nil {
		database = Init()
	}
	return database
}

func Migrate(models ...interface{}) {
	GetDB().AutoMigrate(models)
}
