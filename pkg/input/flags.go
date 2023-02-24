/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package input

import (
	"io"
	"os"
	"strings"

	"github.com/edoardottt/favirecon/pkg/output"
	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

const (
	DefaultTimeout     = 10
	DefaultConcurrency = 100
)

// Options struct specifies how the tool
// will behave.
type Options struct {
	Input       string
	FileInput   string
	FileOutput  string
	Hash        goflags.StringSlice
	Verbose     bool
	Output      io.Writer
	Silent      bool
	Concurrency int
	Timeout     int
}

// configureOutput configures the output on the screen.
func (options *Options) configureOutput() {
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	} else if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
}

// ParseOptions parses the command line options for application.
func ParseOptions() *Options {
	options := &Options{}

	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`Discover new target domains using Content Security Policy.`)

	// Input
	flagSet.CreateGroup("input", "Input",
		flagSet.StringVarP(&options.Input, "url", "u", "", `Input domain`),
		flagSet.StringVarP(&options.FileInput, "list", "l", "", `File containing input domains`),
	)

	flagSet.CreateGroup("configs", "Configurations",
		flagSet.StringSliceVarP(&options.Hash, "hash", "", nil, `Filter results having these favicon hashes (comma separated)`, goflags.CommaSeparatedStringSliceOptions),
		flagSet.IntVarP(&options.Concurrency, "concurrency", "c", DefaultConcurrency, `Concurrency level`),
		flagSet.IntVarP(&options.Timeout, "timeout", "t", DefaultTimeout, `Connection timeout in seconds`),
	)

	// Output
	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&options.FileOutput, "output", "o", "", `File to write output results`),
		flagSet.BoolVarP(&options.Verbose, "verbose", "v", false, `Verbose output`),
		flagSet.BoolVarP(&options.Silent, "silent", "s", false, `Silent output. Print only results`),
	)

	if help() || noArgs() {
		output.ShowBanner()
	}

	if err := flagSet.Parse(); err != nil {
		output.ShowBanner()
		gologger.Fatal().Msgf("%s\n", err)
	}

	// Read the inputs and configure the logging.
	options.configureOutput()

	if err := options.validateOptions(); err != nil {
		output.ShowBanner()
		gologger.Fatal().Msgf("%s\n", err)
	}

	output.ShowBanner()

	return options
}

func help() bool {
	// help usage asked by user.
	for _, arg := range os.Args {
		argStripped := strings.Trim(arg, "-")
		if argStripped == "h" || argStripped == "help" {
			return true
		}
	}

	return false
}

func noArgs() bool {
	// User passed no flag.
	return len(os.Args) < 2
}
