package models

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
	Number     string    `pg:",notnull,unique" json:"number"`
	Status     Status    `pg:",notnull" json:"status"`
	Accrual    float32   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int64     `json:"-"`
	// User       *User     `json:"-" pg:"rel:has-one"`
}

type Accrual struct {
	// Number  int64   `json:"-"`
	Order   string  `json:"order"`
	Status  Status  `json:"status"`
	Accrual float32 `json:"accrual"`
}

type Withdrawal struct {
	ID          int64     `json:"-"`
	UserID      int64     `json:"-"`
	Order       string    `pg:",notnull,unique" json:"order,omitempty"`
	Withdrawn   float32   `json:"withdrawn,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	Current     float32   `pg:"-" json:"current,omitempty"`
}

type WithdrawRequest struct {
	Order       string    `json:"order,omitempty"`
	Sum         float32   `json:"sum,omitempty"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
}
