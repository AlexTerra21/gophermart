package accrual

import (
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/AlexTerra21/gophermart/internal/app/errs"
	"github.com/AlexTerra21/gophermart/internal/app/models"
)

func GetAccrual(order string, endpoint string) (*models.Accrual, error) {
	accrual := &models.Accrual{}
	client := resty.New()

	resp, err := client.R().
		SetHeader("Content-Type", "application/text").
		SetResult(accrual).
		Get(endpoint + "/api/orders/" + order)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errs.ErrNoContent
	}

	return accrual, nil
}
