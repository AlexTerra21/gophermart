package storage

type User struct {
	ID             int64
	Name           string `pg:",notnull,unique" json:"login"`
	Password       string `pg:"-" json:"password"`
	HashedPassword []byte `json:"-"`
	Salt           []byte `json:"-"`
}
