package common

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/json-iterator/go"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	unknownError         = 500
	connectionError      = 501
	accessDenied         = 1000
	banned               = 1001
	alreadyBanned        = 1002
	alreadyUnbanned      = 1003
	alreadyChanged       = 1004
	statusIsNotExist     = 1005
	objectIsNotExist     = 1006
	incorrectField       = 1007
	emptyField           = 1008
	incorrectContentType = 1009
	incorrectImageType   = 1010
	userInvalidToken     = 2000

	StatusModeration = 1
	StatusActive     = 2
	StatusDraft      = 3
	StatusCancelled  = 4
	StatusClosed     = 5
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var database *gorm.DB

func StatusAlreadyChanged() *Error {
	return &Error{StatusCode: http.StatusOK, Code: alreadyChanged, Message: "status already changed"}
}
func StatusIsNotExist() *Error {
	return &Error{StatusCode: http.StatusOK, Code: statusIsNotExist, Message: "status is not exist"}
}
func IsNotExist(object string) *Error {
	return &Error{StatusCode: http.StatusOK, Code: objectIsNotExist, Message: object + " is not exist"}
}
func IncorrectField(field string) *Error {
	return &Error{StatusCode: http.StatusOK, Code: incorrectField, Message: "incorrect " + field}
}
func EmptyField(field string) *Error {
	return &Error{StatusCode: http.StatusOK, Code: emptyField, Message: field + "can not be empty"}
}
func IncorrectContentType() *Error {
	return &Error{StatusCode: http.StatusOK, Code: incorrectContentType, Message: "incorrect content type"}
}
func IncorrectImageType() *Error {
	return &Error{StatusCode: http.StatusOK, Code: incorrectImageType, Message: "incorrect image type"}
}
func AccessDenied() *Error {
	return &Error{StatusCode: http.StatusOK, Code: accessDenied, Message: "access denied"}
}
func Banned() *Error {
	return &Error{StatusCode: http.StatusOK, Code: banned, Message: "user banned"}
}
func Incorrect(err error) *Error {
	switch err.(type) {
	case *url.Error:
		return &Error{StatusCode: http.StatusOK, Code: connectionError, Message: "connection error"}
	}
	return &Error{StatusCode: http.StatusOK, Code: unknownError, Message: err.Error()}
}
func AlreadyBanned() *Error {
	return &Error{StatusCode: http.StatusOK, Code: alreadyBanned, Message: "user already banned"}
}
func AlreadyUnbanned() *Error {
	return &Error{StatusCode: http.StatusOK, Code: alreadyUnbanned, Message: "user already unbanned"}
}
func Forbidden() *Error {
	return &Error{StatusCode: http.StatusForbidden, Code: userInvalidToken, Message: "forbidden"}
}
func Unauthorized() *Error {
	return &Error{StatusCode: http.StatusUnauthorized, Code: userInvalidToken, Message: "unauthorized"}
}

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
	db := GetDB().Model(bean).Update(bean)
	defer db.Close()
	return db.Error
}
func Add(bean interface{}) error {
	db := GetDB()
	if !db.NewRecord(bean) {
		defer db.Close()
		return errors.New("unable to create")
	}
	db.Create(bean)
	defer db.Close()
	return db.Error
}
func Remove(bean interface{}) error {
	db := GetDB().Delete(bean)
	defer db.Close()
	return db.Error
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
