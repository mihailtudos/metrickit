package helpers

import (
	"errors"
	"fmt"
)

func ExampleErrAttr() {
	err := errors.New("example error")
	attr := ErrAttr(err)
	fmt.Println(attr)

	// Output:
	// error=example error
}
