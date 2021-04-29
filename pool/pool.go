package pool

type Pool interface {
	Get() (interface{}, error)
	Put(interface{}, error) error
	Release()
}
