package portmanager

type PortInUseError struct {
	s string
}

func (e PortInUseError) Error() string {
	return e.s
}

type PortsExhaustedError struct {
	s string
}

func (e PortsExhaustedError) Error() string {
	return e.s
}

type PortNotFoundError struct {
	s string
}

func (e PortNotFoundError) Error() string {
	return e.s
}
