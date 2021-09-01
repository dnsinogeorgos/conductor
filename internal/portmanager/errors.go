package portmanager

import "fmt"

type PortInUseError struct {
	p int32
	n string
}

func (e PortInUseError) Error() string {
	return fmt.Sprintf("port %d is currently in use by %s", e.p, e.n)
}

type PortsExhaustedError struct {
	np int
}

func (e PortsExhaustedError) Error() string {
	return fmt.Sprintf("configured range of ports is exhausted")
}

type PortNotFoundError struct {
	p int32
}

func (e PortNotFoundError) Error() string {
	return fmt.Sprintf("port %d not found in list of configured ports", e.p)
}
