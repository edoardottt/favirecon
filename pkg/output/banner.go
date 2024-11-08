/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package output

import "github.com/projectdiscovery/gologger"

// nolint: gochecknoglobals
var printed = false

const (
	Version = "v0.1.2"
	banner  = `    ____            _                          
   / __/___  __  __(_)_______  _________  ____ 
  / /_/ __ ` + `\/ | / / / ___/ _ \/ ___/ __ \/ __ \
 / __/ /_/ /| |/ / / /  /  __/ /__/ /_/ / / / /
/_/  \__,_/ |___/_/_/   \___/\___/\____/_/ /_/ 
                                               `
)

func ShowBanner() {
	if !printed {
		gologger.Print().Msgf("%s%s\n\n", banner, Version)
		gologger.Print().Msgf("\t\t@edoardottt, https://edoardottt.com/\n")
		gologger.Print().Msgf("\t\t             https://github.com/edoardottt/\n\n")

		printed = true
	}
}
