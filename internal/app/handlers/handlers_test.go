package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"

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
	conf.Storage.TestDataSetOrder()
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
			cookie: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDc2NjI2ODYsIlVzZXJJRCI6NX0.gvU226YM6iX7IvfzNP-OHTQ3GveZx9jSatCo_NvWR8c",
			body:   "qwerty",
		},
		{
			name:   "check 422",
			method: http.MethodPost,
			code:   http.StatusUnprocessableEntity,
			cookie: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDc2NjI2ODYsIlVzZXJJRCI6NX0.gvU226YM6iX7IvfzNP-OHTQ3GveZx9jSatCo_NvWR8c",
			body:   "1234567890",
		},
		{
			name:   "check 202",
			method: http.MethodPost,
			code:   http.StatusAccepted,
			cookie: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDc2NjI2ODYsIlVzZXJJRCI6NX0.gvU226YM6iX7IvfzNP-OHTQ3GveZx9jSatCo_NvWR8c",
			body:   "12345678903",
		},
		{
			name:   "check 200",
			method: http.MethodPost,
			code:   http.StatusOK,
			cookie: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDc2NjI2ODYsIlVzZXJJRCI6NX0.gvU226YM6iX7IvfzNP-OHTQ3GveZx9jSatCo_NvWR8c",
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
