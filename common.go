package common

import (
	"github.com/json-iterator/go"
	"io"
	"github.com/jinzhu/gorm"
	"os"
	"strings"
	"fmt"
	"time"
	"net/http"
	"errors"
	"github.com/gin-gonic/gin"
)

const (
	UnknownError  = 999
	IncorrectData = 1000
	InvalidToken  = 1002
)

type ErrorDto struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Error struct {
	StatusCode int
	Code       int
	Error      error
}

type BaseResponse interface {
	ErrorDto() *ErrorDto
}

type IntResponse struct {
	Response uint      `json:"response"`
	Error    *ErrorDto `json:"error"`
}

func (r *IntResponse) ErrorDto() *ErrorDto {
	return r.Error
}

type Ids struct {
	Ids [] int `json:"ids"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var database *gorm.DB

func Decode(reader io.ReadCloser, obj interface{}) error {
	return json.NewDecoder(reader).Decode(obj)
}
func Encode(writer io.WriteCloser, obj interface{}) error {
	return json.NewEncoder(writer).Encode(obj)
}

func Init() *gorm.DB {
	postgresUrl := os.Getenv("POSTGRES_URL")
	parsed := strings.FieldsFunc(postgresUrl, Split)
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

func Unknown(err error) *Error {
	return &Error{StatusCode: http.StatusOK, Code: UnknownError, Error: err}
}
func UnknownWithCode(code int, err error) *Error {
	return &Error{StatusCode: http.StatusOK, Code: UnknownError, Error: err}
}
func Incorrect(err error) *Error {
	return &Error{StatusCode: http.StatusOK, Code: IncorrectData, Error: err}
}
func IncorrectWithCode(code int, err error) *Error {
	return &Error{StatusCode: http.StatusOK, Code: IncorrectData, Error: err}
}
func Forbidden() *Error {
	return &Error{StatusCode: http.StatusForbidden, Code: InvalidToken, Error: errors.New("StatusForbidden")}
}
func Unauthorized() *Error {
	return &Error{StatusCode: http.StatusUnauthorized, Code: InvalidToken, Error: errors.New("StatusUnauthorized")}
}

type Response struct {
	response interface{}
	error    *Error
}

func MakeResponse(resp interface{}, err *Error) *Response{
	return &Response{response:resp, error:err}
}

func Handle(context *gin.Context, f func(*gin.Context, chan *Response)) {
	responseCh := make(chan *Response)
	go f(context, responseCh)
	response := <-responseCh
	if response.error != nil {
		SendError(context, response.error)
		return
	}
	SendResponse(context, response.response)
}

func SendResponse(context *gin.Context, response interface{}) {
	context.JSON(http.StatusOK, gin.H{"response": response})
}
func SendError(context *gin.Context, error *Error) {
	context.JSON(error.StatusCode, gin.H{"error": ErrorDto{error.Code, error.Error.Error()}})
}
func SendErrorDto(context *gin.Context, code int, error *ErrorDto) {
	context.JSON(code, gin.H{"error": error})
}
