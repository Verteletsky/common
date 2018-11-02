package common

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	incorrectData       = 1000
	userAccessDenied    = 1001
	userBanned          = 1002
	userAlreadyBanned   = 1003
	userAlreadyUnbanned = 1004
	userInvalidToken    = 2000

	StatusModeration = 1
	StatusActive     = 2
	StatusDraft      = 3
	StatusCancelled  = 4
	StatusClosed     = 5
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var database *gorm.DB

type ErrorDto struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type Error struct {
	StatusCode int
	Code       int
	Message    string
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
	Ids []int `json:"ids"`
}

func Decode(reader io.ReadCloser, obj interface{}) error {
	return json.NewDecoder(reader).Decode(obj)
}
func Encode(writer io.WriteCloser, obj interface{}) error {
	return json.NewEncoder(writer).Encode(obj)
}
func Init() *gorm.DB {
	postgresUrl := os.Getenv("DB_URL")
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
	return GetDB().Model(bean).Update(bean).Error
}
func Add(bean interface{}) error {
	if !GetDB().NewRecord(bean) {
		return errors.New("unable to create")
	}
	return GetDB().Create(bean).Error
}
func Remove(bean interface{}) error {
	return GetDB().Delete(bean).Error
}
func CustomError(code int, err string) *Error {
	return &Error{StatusCode: http.StatusOK, Code: code, Message: err}
}
func Incorrect(err string) *Error {
	return &Error{StatusCode: http.StatusOK, Code: incorrectData, Message: err}
}
func AccessDenied() *Error {
	return &Error{StatusCode: http.StatusOK, Code: userAccessDenied, Message: "access denied"}
}
func Banned() *Error {
	return &Error{StatusCode: http.StatusOK, Code: userBanned, Message: "user banned"}
}
func AlreadyBanned() *Error {
	return &Error{StatusCode: http.StatusOK, Code: userAlreadyBanned, Message: "user already banned"}
}
func AlreadyUnbanned() *Error {
	return &Error{StatusCode: http.StatusOK, Code: userAlreadyUnbanned, Message: "user already unbanned"}
}
func Forbidden() *Error {
	return &Error{StatusCode: http.StatusForbidden, Code: userInvalidToken, Message: "forbidden"}
}
func Unauthorized() *Error {
	return &Error{StatusCode: http.StatusUnauthorized, Code: userInvalidToken, Message: "unauthorized"}
}

type Response struct {
	Response interface{}
	Error    *Error
}

func MakeResponse(resp interface{}, err *Error) *Response {
	return &Response{Response: resp, Error: err}
}

func Handle(context *gin.Context, f func(*gin.Context, chan *Response)) {
	responseCh := make(chan *Response)
	go f(context, responseCh)
	response := <-responseCh
	if response.Error != nil {
		SendError(context, response.Error)
		return
	}
	SendResponse(context, response.Response)
}

func SendResponse(context *gin.Context, response interface{}) {
	context.JSON(http.StatusOK, gin.H{"response": response})
}
func SendError(context *gin.Context, error *Error) {
	context.JSON(error.StatusCode, gin.H{"error": ErrorDto{error.Code, error.Message}})
}
func SendErrorDto(context *gin.Context, code int, error *ErrorDto) {
	context.JSON(code, gin.H{"error": error})
}
