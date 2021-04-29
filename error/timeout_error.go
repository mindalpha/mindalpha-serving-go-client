package error

type TimeoutError struct {
	TOErr error
}

func (err *TimeoutError) Error() string {
	return err.TOErr.Error()
}
