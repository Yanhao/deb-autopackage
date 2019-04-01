package main

import (
	"fmt"
	"strings"
)

var enableDebugOutput bool

func debug(a ...interface{}) {
	if !enableDebugOutput {
		return
	}

	outLine := strings.Builder{}

	for _, v := range a {
		fmt.Fprint(&outLine, v)
		fmt.Fprint(&outLine, " ")
	}

	debugMsg := outLine.String()
	fmt.Println(debugMsg[:len(debugMsg)-1])
}
