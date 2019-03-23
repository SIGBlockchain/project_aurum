package contract

import "fmt"

//seperate file for errors and their structs

type ChangeError struct {
	Change uint64
}

func (e ChangeError) Error() string {
	return fmt.Sprintf("Claim produces a change of %d", e.Change)
}
