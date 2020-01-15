package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func GinLoggerHandler(w io.Writer) gin.HandlerFunc {
	formatter := func(params gin.LogFormatterParams) (output string) {
		var (
			statusColor, methodColor, resetColor string
			content, contentType                 string
			header, body, errMessage             string
		)

		if params.IsOutputColor() {
			statusColor = params.StatusCodeColor()
			methodColor = params.MethodColor()
			resetColor = params.ResetColor()
		}

		if params.Latency > time.Minute {
			// Truncate in a golang < 1.8 safe way
			params.Latency = params.Latency - params.Latency%time.Second
		}

		content = params.Request.Header.Get("Request")
		header = fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s\n",
			params.TimeStamp.Format("2006/01/02 - 15:04:05"),
			statusColor, params.StatusCode, resetColor,
			params.Latency,
			params.ClientIP,
			methodColor, params.Method, resetColor,
			params.Path,
		)

		contentType = params.Request.Header.Get("Content-Type")
		if content != "" {
			body = fmt.Sprintf("[BODY] Content-Type %s - %s\n", contentType, content)
		}

		if params.ErrorMessage != "" {
			errMessage = fmt.Sprintf("[ERROR] %s\n", params.ErrorMessage)
		}

		if gin.Mode() == gin.DebugMode {
			output = header + body + errMessage
		} else {
			output = header + errMessage
		}

		return output
	}

	loggerFunc := gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: formatter,
		Output:    w,
	})

	return func(ctx *gin.Context) {
		var content string
		contentType := ctx.Request.Header.Get("Content-Type")
		switch contentType {
		case gin.MIMEJSON:
			if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
			} else {
				content = string(data)
				str := bytes.NewBuffer(nil)
				json.Indent(str, []byte(data), "", "    ")
				content = str.String()

				// 此处很重要，否则后面处理请求的函数无法读取到Body数据
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			}
		default:
		}

		ctx.Request.Header.Set("Request", content)
		loggerFunc(ctx)
	}
}
