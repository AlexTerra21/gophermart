package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"golang.org/x/crypto/bcrypt"

	"github.com/AlexTerra21/gophermart/internal/app/errs"
)

type Storage struct {
	db *pg.DB
}

func (d *Storage) New(dbURI string) error {
	opt, err := pg.ParseURL(dbURI)
	if err != nil {
		return err
	}
	d.db = pg.Connect(opt)
	err = d.createSchema()
	return err
}

func (d *Storage) createSchema() error {
	models := []interface{}{
		(*User)(nil),
		(*Order)(nil),
		(*Withdrawal)(nil),
	}

	for _, model := range models {
		err := d.db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Storage) Close() {
	d.db.Close()
}

func (d *Storage) AddUser(ctx context.Context, user *User) (int64, error) {
	salt, err := generateSalt()
	if err != nil {
		return -1, err
	}
	toHash := append([]byte(user.Password), salt...)
	hashedPassword, err := bcrypt.GenerateFromPassword(toHash, bcrypt.DefaultCost)
	if err != nil {
		return -1, err
	}
	user.Salt = salt
	user.HashedPassword = hashedPassword
	_, err = d.db.ModelContext(ctx, user).Insert()
	if pgErr, ok := err.(pg.Error); ok {
		if pgErr.IntegrityViolation() {
			return -1, errs.ErrConflict
		} else {
			return -1, err
		}
	}
	return user.ID, nil
}

// Проверка на совпадение логина и пароля. Возвращает userID в случае совпадения и -1 в противном случае.
func (d *Storage) CheckLoginPassword(ctx context.Context, user *User) (int64, error) {
	// Проверка, что логин присутствует в базе ...
	err := d.db.ModelContext(ctx, user).Where("name = ?", user.Name).Select()
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return -1, nil
		} else {
			return 0, err
		}
	}
	// ... логин есть. Теперь проверим совпадение паролей.
	salted := append([]byte(user.Password), user.Salt...)
	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, salted); err != nil {
		return -1, nil
	}
	return user.ID, nil
}

// func (d *Storage) GetUserByName(ctx context.Context, name string) (user *User, err error) {
// 	user = &User{}
// 	err = d.db.ModelContext(ctx, user).Where("name = ?", name).Select()
// 	return
// }

func (d *Storage) SetOrder(ctx context.Context, number int64, userID int64) (*Order, error) {
	order := &Order{
		Number:     fmt.Sprintf("%d", number),
		UserID:     userID,
		Status:     NEW,
		Accrual:    0,
		UploadedAt: time.Now(),
	}
	tx, err := d.db.BeginContext(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Close()

	_, err = tx.ModelContext(ctx, order).Insert()
	if pgErr, ok := err.(pg.Error); ok {
		if pgErr.IntegrityViolation() {
			_ = d.db.ModelContext(ctx, order).Where("number = ?", fmt.Sprintf("%d", number)).Select()
			return order, errs.ErrConflict
		} else {
			return nil, err
		}
	}
	return order, tx.Commit()
}

func (d *Storage) GetOrders(ctx context.Context, userID int64) ([]Order, error) {
	orders := make([]Order, 0)
	err := d.db.ModelContext(ctx, &orders).Where("user_id = ?", userID).Order("uploaded_at ASC").Select()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (d *Storage) UpdateAccrual(ctx context.Context, order *Order) error {
	tx, err := d.db.BeginContext(ctx)
	if err != nil {
		return err
	}
	defer tx.Close()
	_, err = tx.ModelContext(ctx, order).Column("status", "accrual").Where("number = ?", order.Number).Update()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (d *Storage) GetBalance(ctx context.Context, userID int64) (float32, error) {
	var sumAccrual float32
	err := d.db.ModelContext(ctx, (*Order)(nil)).ColumnExpr("sum(?)", pg.Ident("accrual")).Where("user_id = ?", userID).Select(&sumAccrual)
	if err != nil {
		return -1, err
	}
	return sumAccrual, nil
}

func (d *Storage) GetWithdrawSum(ctx context.Context, userID int64) (float32, error) {
	var sumWithdraw float32
	err := d.db.ModelContext(ctx, (*Withdrawal)(nil)).ColumnExpr("sum(?)", pg.Ident("withdrawn")).Where("user_id = ?", userID).Select(&sumWithdraw)
	if err != nil {
		return -1, err
	}
	return sumWithdraw, nil
}

func (d *Storage) SetWithdraw(ctx context.Context, withdraw Withdrawal) error {
	tx, err := d.db.BeginContext(ctx)
	if err != nil {
		return err
	}
	defer tx.Close()

	_, err = tx.ModelContext(ctx, &withdraw).Insert()
	if err != nil {
		if pgErr, ok := err.(pg.Error); ok {
			if pgErr.IntegrityViolation() {
				return errs.ErrConflict
			}
		} else {
			return err
		}
	}
	return tx.Commit()
}

func (d *Storage) GetWithdrawals(ctx context.Context, userID int64) ([]Withdrawal, error) {
	withdrawals := make([]Withdrawal, 0)
	err := d.db.ModelContext(ctx, &withdrawals).Where("user_id = ?", userID).Order("processed_at DESC").Select()
	if err != nil {
		return nil, err
	}
	return withdrawals, nil
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Для тестирования
func (d *Storage) TestDataSetOrder() {
	d.db.Exec("TRUNCATE orders")
}
