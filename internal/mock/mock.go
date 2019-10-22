/*
Project Aurum mocking package

Usage:
m := SomeMockObject{}
m.When("MethodToBeMocked").Given("ParametersForMethod").Return("ReturnValuesFromMethod")

Designing Mock Objects:
Step 1 - Identify interface to be mocked
Step 2 - Identify functions required to implement interface
Step 3 - For each function, identify parameters and return values.
Step 4 - Store parameters and values as variables in Mock struct
Step 5 - Implement interface functions, checking for matching parameters and returning return values
Step 6 - Add When function (documentation below)
Step 7 - Add Given function (documentation below)
Step 8 - Add Return function (documentation below)
OPTIONAL - Create Check function (documentation below)
*/

package mock

import (
	"fmt"
	"io/ioutil"
)

type MockIoReader struct {
	Buffer []byte // Desired contents of buffer that will be read in
	NRead  int    // Desired number of bytes read to be returned
	Error  error  // Desired error to be returned
}

/*
When an interface requires multiple implemetations, the When function
will designate which method returns which struct members. It is also
used to describe to a reviewer what behavior this object is attempting
to mock out.
*/
func (mocker *MockIoReader) When(s string) *MockIoReader {
	return mocker
}

/*
Stores the parameters that will be passed to the mocked out method,
e.g. the byte slice passed into Read
*/
func (mocker *MockIoReader) Given(b []byte) *MockIoReader {
	mocker.Buffer = b
	return mocker
}

/*
Stores the return variables from the mocked out method,
e.g. the int and error returned from Read
*/
func (mocker *MockIoReader) Return(n int, e error) {
	mocker.NRead = n
	mocker.Error = e
}

/*
Does a simple check to see if complex behavior can be mocked
Good to use if to check if your mock object will satisfy the interface
*/
func (mocker *MockIoReader) Check() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panicked")
		}
	}()
	_, err := ioutil.ReadAll(mocker)
	return err
}

/*
The implementation of the interface for the mock object,
e.g. io.Reader
*/
func (mock MockIoReader) Read(b []byte) (int, error) {
	copy(b, mock.Buffer)
	return mock.NRead, mock.Error
}

// =============================================================
// ====================== KEEP CODE BELOW ======================
// =============================================================
// =============================================================
// =============================================================

// type IMocker interface {
// 	Returns(...interface{}) IMocker
// 	Given(...interface{}) IMocker
// }

// type Mocker struct{}

// func (mocker MockIoReader) Returns(ifaces ...interface{}) IMocker {
// 	for _, iface := range ifaces {
// 		if n, ok := iface.(int); ok {
// 			mocker.NRead = n
// 		} else if e, ok := iface.(error); ok {
// 			mocker.Error = e
// 		}
// 	}
// 	return mocker
// }

// func (mocker MockIoReader) Given(ifaces ...interface{}) IMocker {
// 	for _, iface := range ifaces {
// 		if b, ok := iface.([]byte); ok {
// 			mocker.Buffer = b
// 		}
// 	}
// 	return mocker
// }

// func (mocker Mocker) When(s string) IMocker {
// 	switch s {
// 	case "Read":
// 		return MockIoReader{}
// 	default:
// 		return nil
// 	}
// }