package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

	"github.com/AlexTerra21/gophermart/internal/app/auth"
	"github.com/AlexTerra21/gophermart/internal/app/config"
)

// ./cmd/gophermart/gophermart.exe -a localhost:8081 -r http://localhost:8091 -l debug -d "postgresql://gophermart:gophermart@localhost/gophermart?sslmode=disable"

func initTestConfig() *config.Config {
	conf := config.NewConfig()
	conf.SetAccrualAddress("localhost:8082")
	conf.SetLogLevel("debug")
	conf.SetDBConnectionString("postgresql://gophermart:gophermart@localhost/gophermart_test?sslmode=disable")
	conf.SetAccrualAddress("http://localhost:8092")
	conf.InitStorage()
	return conf
}

func SetTestData(conf *config.Config) {

}

func Test_addOrder(t *testing.T) {
	conf := initTestConfig()
	srv := httptest.NewServer(MainRouter(conf))
	defer srv.Close()
	conf.InitAsync()
	conf.Storage.TestDataSetOrder()
	token5, _ := auth.BuildJWTString(5)
	token7, _ := auth.BuildJWTString(7)
	tests := []struct {
		name   string
		method string
		code   int
		cookie string
		body   string
	}{
		{
			name:   "check 401",
			method: http.MethodPost,
			code:   http.StatusUnauthorized,
			cookie: "",
			body:   "",
		},
		{
			name:   "check 400",
			method: http.MethodPost,
			code:   http.StatusBadRequest,
			cookie: token5,
			body:   "qwerty",
		},
		{
			name:   "check 422",
			method: http.MethodPost,
			code:   http.StatusUnprocessableEntity,
			cookie: token5,
			body:   "1234567890",
		},
		{
			name:   "check 202",
			method: http.MethodPost,
			code:   http.StatusAccepted,
			cookie: token5,
			body:   "12345678903",
		},
		{
			name:   "check 200",
			method: http.MethodPost,
			code:   http.StatusOK,
			cookie: token5,
			body:   "12345678903",
		},
		{
			name:   "check 409",
			method: http.MethodPost,
			code:   http.StatusConflict,
			cookie: token7,
			body:   "12345678903",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL + "/api/user/orders"
			if len(tt.body) > 0 {
				req.SetHeader("Content-Type", "text/plain")
				req.SetBody(tt.body)
			}
			req.SetCookie(&http.Cookie{
				Name:  "Authorization",
				Value: tt.cookie,
			})
			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.code, resp.StatusCode(), "Response code didn't match expected")
		})
	}
}
