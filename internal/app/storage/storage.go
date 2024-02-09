package storage

import (
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
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
	}

	for _, model := range models {
		err := d.db.Model(model).CreateTable(&orm.CreateTableOptions{
			// Temp:        false,
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

func (d *Storage) AddUser(user *User) (err error) {

	return
}

func (d *Storage) GetUserByName(name string) (user *User, err error) {

	return
}
