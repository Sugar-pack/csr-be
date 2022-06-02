package repositories

type Transaction interface {
	Commit() error
	Rollback() error
}
