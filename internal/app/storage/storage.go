package storage

import (
	"crypto/rand"
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

func (d *Storage) AddUser(user *User) (int64, error) {
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
	_, err = d.db.Model(user).Insert()
	if pgErr, ok := err.(pg.Error); ok {
		if pgErr.IntegrityViolation() {
			return -1, errs.ErrConflict
		} else {
			return -1, err
		}
	}
	return user.ID, nil
}

func (d *Storage) CheckLoginPassword(user *User) (int64, error) {
	err := d.db.Model(user).Where("name = ?", user.Name).Select()
	if err != nil {
		if err.Error() == pg.ErrNoRows.Error() {
			return -1, nil
		} else {
			return 0, err
		}
	}
	salted := append([]byte(user.Password), user.Salt...)
	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, salted); err != nil {
		return -1, nil
	}
	return user.ID, nil
}

func (d *Storage) GetUserByName(name string) (user *User, err error) {
	user = &User{}
	err = d.db.Model(user).Where("name = ?", name).Select()
	return
}

func (d *Storage) SetOrder(number int, userID int64) (*Order, error) {
	order := &Order{
		Number:     number,
		UserID:     userID,
		UploadedAt: time.Now(),
	}
	_, err := d.db.Model(order).Insert()
	if pgErr, ok := err.(pg.Error); ok {
		if pgErr.IntegrityViolation() {
			_ = d.db.Model(order).Where("number = ?", number).Select()
			return order, errs.ErrConflict
		} else {
			return nil, err
		}
	}
	return order, err
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}
