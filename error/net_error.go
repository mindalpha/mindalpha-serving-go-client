package error

type NetError struct {
	Nerr error
}

func (netErr *NetError) Error() string {
	return netErr.Nerr.Error()
}
