package accrual

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/AlexTerra21/gophermart/internal/app/errs"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

func GetAccrual(order int, endpoint string) (*storage.Accrual, error) {
	accrual := &storage.Accrual{}
	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "application/text").
		SetResult(accrual).
		Get(endpoint + "/api/orders/" + fmt.Sprintf("%d", order))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errs.ErrNoContent
	}

	return accrual, nil
}
