package a

import "os"

func main() {
	os.Exit(1) // want "os.Exit called within main function"
}

func otherFunc() {
	os.Exit(1) // ok - not in main
}
