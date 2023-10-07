package commands

import "fmt"

func NotValidArgs(validArgs []string) error {
	return fmt.Errorf("Not a valid Argument, ValidArgs are [%s]", validArgs)
}
