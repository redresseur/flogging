package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	uxml "github.com/redresseur/utils/xml"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"time"
)

const Content = `content`

// dump: load the data from the body in requests.
// support application/json, application/multipart-format
func dump(ctx *gin.Context) {
	var content string

	contentType, _, err := mime.ParseMediaType(ctx.Request.Header.Get("Content-Type"))
	if err != nil {
		//ctx.AbortWithError(http.StatusBadRequest, err)
		ctx.Next()
	}

	switch contentType {
	case gin.MIMEJSON:
		if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
		} else {
			buff := bytes.NewBuffer(nil)
			err := json.Indent(buff, []byte(data), "", "    ")
			if err != nil {
				content = string(data)
			} else {
				content = buff.String()
			}

			// 此处很重要，否则后面处理请求的函数无法读取到Body数据
			ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}
	case gin.MIMEXML:
		fallthrough
	case gin.MIMEXML2:
		{
			if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
			} else {
				content = uxml.FormatXML(string(data), "", "	")
				// 此处很重要，否则后面处理请求的函数无法读取到Body数据
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			}
		}
	case gin.MIMEPOSTForm:
		{
			if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
			} else {
				// TODO
				content = string(data)
				// 此处很重要，否则后面处理请求的函数无法读取到Body数据
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			}

		}
	case gin.MIMEMultipartPOSTForm:
		{
			// 多重表單中可能包含二進制文件，暫時先不處理
			// TODO
		}
	case gin.MIMEPlain:
		{
			if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
			} else {
				// TODO
				content = string(data)
				// 此处很重要，否则后面处理请求的函数无法读取到Body数据
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			}
		}
	case gin.MIMEYAML:
		{
			if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
			} else {
				// TODO
				content = string(data)
				// 此处很重要，否则后面处理请求的函数无法读取到Body数据
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			}
		}
	case gin.MIMEHTML:
		{
			if data, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
				ctx.AbortWithError(http.StatusBadRequest, err)
			} else {
				// TODO
				content = string(data)
				// 此处很重要，否则后面处理请求的函数无法读取到Body数据
				ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
			}
		}
	default:
	}

	ctx.Request.Header.Set(Content, content)
}

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

		content = params.Request.Header.Get(Content)
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
		dump(ctx)
		loggerFunc(ctx)
	}
}
