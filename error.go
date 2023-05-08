package misc

type WarpError struct {
	msg string
	err error
}

func (e *WarpError) Error() string {
	return e.err.Error() + " warp by " + e.msg
}

func (e *WarpError) Unwrap() error {
	return e.err
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &WarpError{err: err, msg: msg}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
