package config

import "fmt"

type MissingConfigurationVariableError struct {
	t string
	n string
}

func (e MissingConfigurationVariableError) Error() string {
	return fmt.Sprintf("missing %s configuration variable %s", e.t, e.n)
}
