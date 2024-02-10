package storage

import "time"

type User struct {
	ID             int64
	Name           string `pg:",notnull,unique" json:"login"`
	Password       string `pg:"-" json:"password"`
	HashedPassword []byte `json:"-"`
	Salt           []byte `json:"-"`
}

type Order struct {
	ID         int64     `json:"-"`
	Number     int       `pg:",notnull,unique" json:"number"`
	Status     string    `json:"status"`
	Accrual    int64     `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int64     `json:"-"`
	User       *User     `json:"-" pg:"rel:has-one"`
}
