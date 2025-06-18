package console

import "fmt"

func PrintSimpleWarning(message string) {
	fmt.Printf("\033[33m⚠️  %s\033[0m\n", message)
}
