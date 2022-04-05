package error

type Error interface {
	error
	ErrCode() int
}

type Err struct {
	Err  error
	Code int
}

func (err *Err) ErrCode() int {
	return err.Code
}

func (err *Err) Error() string {
	return err.Err.Error()
}
