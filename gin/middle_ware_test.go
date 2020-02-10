package gin

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"
)

func ExampleGinLoggerHandler() {
	router := gin.New()
	router.Use(GinLoggerHandler(os.Stdout), gin.Recovery())
	// TODO: 绑定路由
	// 启动
	router.Run()
}

func TestGinLoggerHandler_Json(t *testing.T) {
	GinLoggerHandler(os.Stdout)
	router := gin.New()
	router.Use(GinLoggerHandler(os.Stdout), gin.Recovery())

	router.POST("/test", func(context *gin.Context) {

	})

	data, _ := json.Marshal(struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	}{
		Year:  2020,
		Month: 02,
		Day:   10,
	})

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/json")
	if assert.NoError(t, err) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

func TestGinLoggerHandler_xml(t *testing.T) {
	GinLoggerHandler(os.Stdout)
	router := gin.New()
	router.Use(GinLoggerHandler(os.Stdout), gin.Recovery())

	router.POST("/test", func(context *gin.Context) {

	})

	date := struct {
		XMLName xml.Name `xml:"date"`
		Year    int      `xml:"year"`
		Month   int      `xml:"month"`
		Day     int      `xml:"day"`
	}{
		Year:  2020,
		Month: 2,
		Day:   10,
	}

	data, err := xml.Marshal(&date)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(data))
	req.Header.Add("Content-Type", "application/xml")
	if assert.NoError(t, err) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}

// Deprecated
func TestGinLoggerHandler_form(t *testing.T) {
	GinLoggerHandler(os.Stdout)
	router := gin.New()
	router.Use(GinLoggerHandler(os.Stdout), gin.Recovery())

	router.POST("/test", func(context *gin.Context) {

	})

	buff := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(buff)
	header := textproto.MIMEHeader{}
	header.Add("Content-Type", "application/json")
	iw, err := mw.CreatePart(header)
	if assert.NoError(t, err) {
		data, _ := json.Marshal(struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		}{
			Year:  2020,
			Month: 02,
			Day:   10,
		})
		iw.Write(data)
	}

	mw.Close()

	req, err := http.NewRequest(http.MethodPost, "/test", buff)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if assert.NoError(t, err) {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}
}
