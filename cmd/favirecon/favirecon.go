package main

import (
	"github.com/edoardottt/favirecon/pkg/favirecon"
	"github.com/edoardottt/favirecon/pkg/input"
)

func main() {
	options := input.ParseOptions()
	runner := favirecon.New(options)
	runner.Run()
}
