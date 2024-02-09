package storage

type User struct {
	Id       int64
	Name     string `pg:",notnull,unique"`
	Password string `pg:",notnull"`
}
