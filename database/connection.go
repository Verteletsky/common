package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"github.com/Verteletsky/common/environments"
	"time"
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
	GetDB().Close()
}

func GetDB() *gorm.DB {
	if database == nil {
		database = Init()
		var sleep = time.Duration(1)
		for database == nil {
			sleep = sleep * 2
			fmt.Printf("database is unavailable. wait for %d sec.\n", sleep)
			time.Sleep(sleep * time.Second)
			database = Init()
		}
	}
	return database
}

func Update(bean interface{}) error {
	query := GetDB().Model(bean).Update(bean)
	return query.Error
}

func Add(bean interface{}) error {
	GetDB().NewRecord(bean)
	db := GetDB().Create(bean)
	return db.Error
}

func Remove(bean interface{}) error {
	query := GetDB().Delete(bean)
	return query.Error
}

func GetPageDB(page, limit int) *gorm.DB {
	db := GetDB()
	if limit > 0 {
		db = db.Limit(limit)
	}
	db = db.Offset(page * limit)
	return db
}
