package common

import "fmt"

type HTTPStatusError struct {
	StatusCode int
	Status     string
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("http status error status code: %v status: %v", e.StatusCode, e.Status)
}
