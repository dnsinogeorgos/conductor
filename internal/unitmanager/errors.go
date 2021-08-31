package unitmanager

type StopMainError struct {
	s string
}

func (e StopMainError) Error() string {
	return e.s
}

type StartMainError struct {
	s string
}

func (e StartMainError) Error() string {
	return e.s
}
