package internal

type Queue interface {
	Next() bool
	Get() (interface{}, error)
	Add(item interface{}) error
}
