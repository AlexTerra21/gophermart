package async

import (
	"context"
	"time"

	"github.com/AlexTerra21/gophermart/internal/app/accrual"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
	"github.com/AlexTerra21/gophermart/internal/app/models"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

type Async struct {
	orderChan      chan *models.Order
	doneCh         chan struct{}
	accrualAddress string
	storage        *storage.Storage
}

var _async *Async

func NewAsync(doneCh chan struct{}, s *storage.Storage, accrualAddress string) {
	instance := &Async{
		orderChan:      make(chan *models.Order, 1024),
		doneCh:         doneCh,
		accrualAddress: accrualAddress,
		storage:        s,
	}

	go instance.orderAccrual()

	_async = instance
}

func GetAsync() *Async {
	return _async
}

func (a *Async) orderAccrual() {
	ticker := time.NewTicker(10 * time.Second)

	var orders []*models.Order

	for {
		select {
		case order := <-a.orderChan:
			orders = append(orders, order)
		case <-a.doneCh:
			logger.Debug("Gorutine finished.")
			return
		case <-ticker.C:
			if len(orders) == 0 {
				continue
			}
			for _, order := range orders {
				if order.Status == models.PROCESSED || order.Status == models.INVALID {
					continue
				}
				accrual, err := accrual.GetAccrual(order.Number, a.accrualAddress)
				if err != nil {
					logger.Debug("Error", logger.Field{Key: "get accrual error", Val: err})
					continue
				}
				order.Status = accrual.Status
				order.Accrual = accrual.Accrual
				err = a.storage.UpdateAccrual(context.Background(), order)
				if err != nil {
					logger.Debug("Error", logger.Field{Key: "cannot update orders", Val: err})
					continue
				}
			}

		}
	}
}

func (a *Async) Push(order *models.Order) {
	a.orderChan <- order
}
