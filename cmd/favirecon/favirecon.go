/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

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
