package logicsuite

import (
	"context"
	"time"

	"github.com/AlexTerra21/gophermart/internal/app/errs"
	"github.com/AlexTerra21/gophermart/internal/app/models"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

func CalculateWithdraw(ctx context.Context, userID int64) (*models.Withdrawal, error) {

	sumAccrual, err := storage.GetStorage().GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	sumWithdraw, err := storage.GetStorage().GetWithdrawSum(ctx, userID)
	if err != nil {
		return nil, err
	}

	withdraw := models.Withdrawal{
		Current:   sumAccrual - sumWithdraw,
		Withdrawn: sumWithdraw,
	}

	return &withdraw, nil
}

func RequestWithdrawal(ctx context.Context, userID int64,
	withdrawRequest models.WithdrawRequest) (*models.Withdrawal, error) {
	sumAccrual, err := storage.GetStorage().GetBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	if sumAccrual < withdrawRequest.Sum {
		return nil, errs.ErrPaymentRequired
	}
	withdraw := models.Withdrawal{
		UserID:      userID,
		Order:       withdrawRequest.Order,
		Withdrawn:   withdrawRequest.Sum,
		ProcessedAt: time.Now(),
	}
	return &withdraw, nil
}
