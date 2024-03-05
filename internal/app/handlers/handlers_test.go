package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/AlexTerra21/gophermart/internal/app/accrual/mocks"
	"github.com/AlexTerra21/gophermart/internal/app/async"
	"github.com/AlexTerra21/gophermart/internal/app/auth"
	"github.com/AlexTerra21/gophermart/internal/app/config"
	"github.com/AlexTerra21/gophermart/internal/app/models"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

// ./cmd/gophermart/gophermart.exe -a localhost:8081 -r http://localhost:8091 -l debug -d "postgresql://gophermart:gophermart@localhost/gophermart?sslmode=disable"

func initTestConfig() *config.Config {
	conf := config.NewConfig()
	conf.SetAccrualAddress("localhost:8082")
	conf.SetLogLevel("debug")
	conf.SetDBConnectionString("postgresql://gophermart:gophermart@localhost/gophermart_test?sslmode=disable")
	conf.SetAccrualAddress("http://localhost:8092")
	return conf
}

func SetTestData() error {
	db := storage.GetStorage().GetDB()
	_, err := db.Exec("TRUNCATE orders")
	if err != nil {
		return err
	}
	_, err = db.Exec("TRUNCATE users")
	if err != nil {
		return err
	}
	_, err = db.Exec("TRUNCATE withdrawals")
	if err != nil {
		return err
	}
	return nil
}

func InitTest(t *testing.T, accrual *mocks.Accrual) (*httptest.Server, error) {
	conf := initTestConfig()
	srv := httptest.NewServer(MainRouter(conf))

	if err := storage.Init(conf.GetDBConnectString()); err != nil {
		return nil, err
	}

	if err := SetTestData(); err != nil {
		return nil, err
	}
	doneCh := make(chan struct{})
	async.NewAsync(doneCh, storage.GetStorage(), conf.GetAccrualAddress(), accrual)

	return srv, nil
}

func Test_addOrder(t *testing.T) {
	srv, err := InitTest(t, nil)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()

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

func Test_register_login(t *testing.T) {
	srv, err := InitTest(t, nil)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()

	registerBody := "{\"login\": \"user\",\"password\": \"password\"}"
	t.Run("test_register_login_returns_equal_token", func(t *testing.T) {
		req := resty.New().R()
		req.Method = http.MethodPost
		req.URL = srv.URL + "/api/user/register"
		req.SetHeader("Content-Type", "application/json")
		req.SetBody(registerBody)
		resp, err := req.Send()
		assert.NoError(t, err, "error making HTTP request")
		cookiesSave := resp.Cookies()

		req.URL = srv.URL + "/api/user/login"
		resp, err = req.Send()
		assert.NoError(t, err, "error making HTTP request")
		cookies := resp.Cookies()

		assert.Equal(t, cookiesSave, cookies)
	})
}

func Test_register(t *testing.T) {
	srv, err := InitTest(t, nil)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()

	tests := []struct {
		name   string
		method string
		code   int
		cookie string
		body   string
	}{
		{
			name:   "check 500",
			method: http.MethodPost,
			code:   http.StatusInternalServerError,
			cookie: "",
			body:   "qwerty",
		},
		{
			name:   "check 200",
			method: http.MethodPost,
			code:   http.StatusOK,
			cookie: "",
			body:   "{\"login\": \"user\",\"password\": \"password\"}",
		},
		{
			name:   "check 409",
			method: http.MethodPost,
			code:   http.StatusConflict,
			cookie: "",
			body:   "{\"login\": \"user\",\"password\": \"password\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL + "/api/user/register"
			if len(tt.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
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

func Test_balance(t *testing.T) {
	srv, err := InitTest(t, nil)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()

	// Добавить тестовые данные
	db := storage.GetStorage().GetDB()
	_, err = db.Exec(`INSERT INTO orders (number, status, accrual, user_id) 
					  VALUES( '123456789031', 'PROCESSED', 9.55, 5)`)
	if err != nil {
		t.Error(err)
	}
	_, err = db.Exec(`INSERT INTO public.withdrawals (user_id, "order", withdrawn)
					  VALUES( 5, '123456789031', 3.14)`)
	if err != nil {
		t.Error(err)
	}

	token5, _ := auth.BuildJWTString(5)
	tests := []struct {
		name   string
		method string
		code   int
		cookie string
		body   string
	}{
		{
			name:   "check 401",
			method: http.MethodGet,
			code:   http.StatusUnauthorized,
			cookie: "",
			body:   "",
		},
		{
			name:   "check 200",
			method: http.MethodGet,
			code:   http.StatusOK,
			cookie: token5,
			body:   `{"withdrawn":3.14,"processed_at":"0001-01-01T00:00:00Z","current":6.41}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL + "/api/user/balance"

			req.SetCookie(&http.Cookie{
				Name:  "Authorization",
				Value: tt.cookie,
			})
			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.code, resp.StatusCode(), "Response code didn't match expected")
			if len(tt.body) > 0 {
				assert.Equal(t, tt.body, string(resp.Body()))
			}
		})
	}
}

func Test_withdraw(t *testing.T) {
	srv, err := InitTest(t, nil)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()

	// Добавить тестовые данные
	db := storage.GetStorage().GetDB()
	_, err = db.Exec(`INSERT INTO orders (number, status, accrual, user_id) 
					  VALUES('1234567890', 'PROCESSED', 10.55, 5),
					        ('12345678903', 'PROCESSED', 4.18, 5)`)
	if err != nil {
		t.Error(err)
	}

	token5, _ := auth.BuildJWTString(5)
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
			name:   "check 402",
			method: http.MethodPost,
			code:   http.StatusPaymentRequired,
			cookie: token5,
			body:   `{"order": "2377225624","sum": 251}`,
		},
		{
			name:   "check 200",
			method: http.MethodPost,
			code:   http.StatusOK,
			cookie: token5,
			body:   `{"order": "2377225624","sum": 10}`,
		},
		{
			name:   "check 422",
			method: http.MethodPost,
			code:   http.StatusUnprocessableEntity,
			cookie: token5,
			body:   `{"order": "1","sum": 10}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL + "/api/user/balance/withdraw"
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

func Test_withdrawals(t *testing.T) {
	srv, err := InitTest(t, nil)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()

	// Добавить тестовые данные

	db := storage.GetStorage().GetDB()
	_, err = db.Exec(`INSERT INTO public.withdrawals (user_id, "order", withdrawn)
					  VALUES(5, '12345678903', 3.14),
					        (5, '1234567890', 6.28)`)
	if err != nil {
		t.Error(err)
	}

	token5, _ := auth.BuildJWTString(5)
	tests := []struct {
		name   string
		method string
		code   int
		cookie string
		body   string
	}{
		{
			name:   "check 401",
			method: http.MethodGet,
			code:   http.StatusUnauthorized,
			cookie: "",
			body:   "",
		},
		{
			name:   "check 200",
			method: http.MethodGet,
			code:   http.StatusOK,
			cookie: token5,
			body:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tt.method
			req.URL = srv.URL + "/api/user/withdrawals"
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

func Test_addOrder_getOrders(t *testing.T) {
	// Тестовые данные
	order := "12345678903"
	status := models.Status("PROCESSED")
	accrual := float32(3.14)

	accrual_mock := mocks.NewAccrual(t)
	// GetAccrual(order string, endpoint string) (*models.Accrual, error)
	accrual_mock.
		On("GetAccrual", mock.Anything, mock.Anything).
		Return(&models.Accrual{
			Order:   order,
			Status:  status,
			Accrual: accrual,
		}, nil)
	srv, err := InitTest(t, accrual_mock)
	if err != nil {
		t.Log(err)
		return
	}
	defer srv.Close()
	defer storage.GetStorage().Close()
	token5, _ := auth.BuildJWTString(5)
	t.Run("test_addOrder_getOrders", func(t *testing.T) {
		req := resty.New().R()
		req.Method = http.MethodPost
		req.URL = srv.URL + "/api/user/orders"
		req.SetHeader("Content-Type", "text/plain")
		req.SetBody(order)
		req.SetCookie(&http.Cookie{
			Name:  "Authorization",
			Value: token5,
		})
		_, err := req.Send()
		assert.NoError(t, err, "error making HTTP request")
		// Ждать 10 сек пока выполнится горутина
		time.Sleep(10 * time.Second)

		req.Method = http.MethodGet
		resp, err := req.Send()
		assert.NoError(t, err, "error making HTTP request")
		orders := make([]models.Order, 0)
		if err := json.Unmarshal(resp.Body(), &orders); err != nil {
			t.Error(err)
		}

		assert.Equal(t, order, orders[0].Number)
		assert.Equal(t, status, orders[0].Status)
		assert.Equal(t, accrual, orders[0].Accrual)

	})
}
