package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	UnknownError  = 999
	IncorrectData = 1000
)

type ErrorDto struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func SendResponse(context *gin.Context, response interface{}) {
	context.JSON(http.StatusOK, gin.H{"response": response})
}
func SendError(context *gin.Context, errorCode int, message string) {
	SendErrorWithStatusCode(context, http.StatusOK, errorCode, message)
}

func SendErrorWithStatusCode(context *gin.Context, status int, errorCode int, message string) {
	context.JSON(status, gin.H{"error": ErrorDto{errorCode, message}})
}