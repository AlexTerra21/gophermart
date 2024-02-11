package storage

import "time"

type Status string

// Статусы заказов
const (
	NEW        Status = "NEW"
	PROCESSING Status = "PROCESSING"
	INVALID    Status = "INVALID"
	PROCESSED  Status = "PROCESSED"
)

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
	Status     Status    `pg:",notnull" json:"status"`
	Accrual    int64     `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int64     `json:"-"`
	// User       *User     `json:"-" pg:"rel:has-one"`
}

type Accrual struct {
	Number  int    `json:"-"`
	Order   string `json:"order"`
	Status  Status `json:"status"`
	Accrual int64  `json:"accrual"`
}
