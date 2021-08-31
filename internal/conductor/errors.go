package conductor

import "fmt"

type CastAlreadyExistsError struct {
	c string
}

func (e CastAlreadyExistsError) Error() string {
	return fmt.Sprintf("cast %s already exists", e.c)
}

type CastNotFoundError struct {
	c string
}

func (e CastNotFoundError) Error() string {
	return fmt.Sprintf("cast %s not found", e.c)
}

type CastNotEmpty struct {
	c string
}

func (e CastNotEmpty) Error() string {
	return fmt.Sprintf("cast %s contains replicas", e.c)
}

type ReplicaAlreadyExistsError struct {
	c string
	r string
}

func (e ReplicaAlreadyExistsError) Error() string {
	return fmt.Sprintf("replica %s already exists in cast %s", e.r, e.c)
}

type ReplicaNotFoundError struct {
	c string
	r string
}

func (e ReplicaNotFoundError) Error() string {
	return fmt.Sprintf("replica %s not found in cast %s", e.r, e.c)
}
