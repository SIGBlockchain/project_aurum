package contract

import "fmt"

//seperate file for errors and their structs

type changeError struct {
	change uint64
}

func (e changeError) Error() string {
	return fmt.Sprintf("Claim produces a change of %d", e.change)
}
