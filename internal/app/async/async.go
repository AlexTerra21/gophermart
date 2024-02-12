package async

import (
	"context"
	"time"

	"github.com/AlexTerra21/gophermart/internal/app/accrual"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
	"go.uber.org/zap"
)

type Async struct {
	orderChan      chan *storage.Order
	accrualAddress string
	storage        *storage.Storage
}

func NewAsync(s *storage.Storage, accrualAddress string) *Async {
	instance := &Async{
		orderChan:      make(chan *storage.Order, 1024),
		accrualAddress: accrualAddress,
		storage:        s,
	}

	go instance.orderAccrual()

	return instance
}

func (a *Async) orderAccrual() {
	ticker := time.NewTicker(10 * time.Second)

	var orders []*storage.Order

	for {
		select {
		case order := <-a.orderChan:
			orders = append(orders, order)
		case <-ticker.C:
			if len(orders) == 0 {
				continue
			}
			for _, order := range orders {
				if order.Status == storage.PROCESSED || order.Status == storage.INVALID {
					continue
				}
				accrual, err := accrual.GetAccrual(order.Number, a.accrualAddress)
				if err != nil {
					logger.Log().Debug("get accrual error", zap.Error(err))
					continue
				}
				order.Status = accrual.Status
				order.Accrual = accrual.Accrual
				err = a.storage.UpdateAccrual(context.Background(), order)
				if err != nil {
					logger.Log().Debug("cannot update orders", zap.Error(err))
					continue
				}
			}

		}
	}
}

func (a *Async) Push(order *storage.Order) {
	a.orderChan <- order
}
