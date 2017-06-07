package utils

import (
	"fmt"
)

func Warning(params string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(params, args...))
}

func Error(params string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(params, args...))
}
