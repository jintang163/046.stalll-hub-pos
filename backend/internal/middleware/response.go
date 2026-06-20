package middleware

import (
	"encoding/xml"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	CodeSuccess = 0
	CodeError   = 500
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, args ...interface{}) {
	var statusCode = http.StatusOK
	var code = CodeError
	var message string

	if len(args) == 1 {
		message = args[0].(string)
	} else if len(args) == 2 {
		if sc, ok := args[0].(int); ok {
			statusCode = sc
			message = args[1].(string)
		} else {
			code = args[0].(int)
			message = args[1].(string)
		}
	} else if len(args) >= 3 {
		statusCode = args[0].(int)
		code = args[1].(int)
		message = args[2].(string)
	}

	c.JSON(statusCode, Response{
		Code:    code,
		Message: message,
	})
}

func ErrorWithCode(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func ErrorWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeError,
		Message: message,
		Data:    data,
	})
}

type XMLResponse struct {
	XMLName xml.Name `xml:"xml"`
	Code    string   `xml:"return_code"`
	Message string   `xml:"return_msg"`
}

func XMLResponse(c *gin.Context, statusCode int, code string, message string) {
	c.XML(statusCode, XMLResponse{
		Code:    code,
		Message: message,
	})
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func PageSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}
