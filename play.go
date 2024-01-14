//go:build ignore

package main

import (
	"syscall/js"

	"github.com/NicoNex/tau"
)

func run(input string) string {
	out, err := tau.Run("<playground>", input)
	if err != nil {
		return err.Error()
	}
	return out
}

func main() {
	js.Global().Set("runTau", js.FuncOf(func(this js.Value, args []js.Value) any {
		out := run(args[0].String())
		jstr := js.Global().Get("String")
		return jstr.New(out)
	}))
	<-make(chan bool)
}
