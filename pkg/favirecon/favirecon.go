/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/edoardottt/favirecon/pkg/input"
	"github.com/edoardottt/favirecon/pkg/output"
	"github.com/edoardottt/golazy"
	"github.com/projectdiscovery/gologger"
	fileutil "github.com/projectdiscovery/utils/file"
)

type Runner struct {
	Input     chan string
	Output    chan output.Found
	Result    output.Result
	UserAgent string
	InWg      *sync.WaitGroup
	OutWg     *sync.WaitGroup
	Options   input.Options
	OutMutex  *sync.Mutex
}

// New takes as input the options and returns
// a new runner.
func New(options *input.Options) Runner {
	if options.FileOutput != "" {
		_, err := os.Create(options.FileOutput)
		if err != nil {
			gologger.Error().Msgf("%s", err)
		}
	}

	return Runner{
		Input:     make(chan string, options.Concurrency),
		Output:    make(chan output.Found, options.Concurrency),
		Result:    output.New(),
		UserAgent: golazy.GenerateRandomUserAgent(),
		InWg:      &sync.WaitGroup{},
		OutWg:     &sync.WaitGroup{},
		Options:   *options,
		OutMutex:  &sync.Mutex{},
	}
}

// Run takes the input and executes all the tasks
// specified in the options.
func (r *Runner) Run() {
	r.InWg.Add(1)

	go pushInput(r)
	r.InWg.Add(1)

	go execute(r)
	r.OutWg.Add(1)

	go pullOutput(r)
	r.InWg.Wait()

	close(r.Output)
	r.OutWg.Wait()
}

func pushInput(r *Runner) {
	defer r.InWg.Done()

	if fileutil.HasStdin() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if r.Options.Cidr {
				ips, err := handleCidrInput(scanner.Text())
				if err != nil {
					gologger.Error().Msg(err.Error())
				} else {
					for _, ip := range ips {
						r.Input <- ip
					}
				}
			} else {
				r.Input <- scanner.Text()
			}
		}
	}

	if r.Options.FileInput != "" {
		for _, line := range golazy.RemoveDuplicateValues(golazy.ReadFileLineByLine(r.Options.FileInput)) {
			if r.Options.Cidr {
				ips, err := handleCidrInput(line)
				if err != nil {
					gologger.Error().Msg(err.Error())
				} else {
					for _, ip := range ips {
						r.Input <- ip
					}
				}
			} else {
				r.Input <- line
			}
		}
	}

	if r.Options.Input != "" {
		if r.Options.Cidr {
			ips, err := handleCidrInput(r.Options.Input)
			if err != nil {
				gologger.Error().Msg(err.Error())
			} else {
				for _, ip := range ips {
					r.Input <- ip
				}
			}
		} else {
			r.Input <- r.Options.Input
		}
	}

	close(r.Input)
}

func execute(r *Runner) {
	defer r.InWg.Done()

	rl := rateLimiter(r)

	for i := 0; i < r.Options.Concurrency; i++ {
		r.InWg.Add(1)

		go func() {
			defer r.InWg.Done()

			for value := range r.Input {
				targetURL, err := PrepareURL(value)
				if err != nil {
					gologger.Error().Msgf("%s", err)

					return
				}

				rl.Take()

				client, err := customClient(&r.Options)
				if err != nil {
					gologger.Error().Msgf("%s", err)

					return
				}

				result, err := getFavicon(targetURL, r.UserAgent, client)
				if err != nil {
					gologger.Error().Msgf("%s", err)

					return
				}

				found, err := CheckFavicon(result, r.Options.Hash, targetURL)
				if err != nil {
					if r.Options.Verbose {
						gologger.Error().Msgf("%s", err)
					}
				} else {
					r.Output <- output.Found{URL: targetURL, Name: found, Hash: result}
				}
			}
		}()
	}
}

func pullOutput(r *Runner) {
	defer r.OutWg.Done()

	for o := range r.Output {
		if !r.Result.Printed(o.URL) {
			r.OutWg.Add(1)

			go writeOutput(r.OutWg, r.OutMutex, &r.Options, o)
		}
	}
}

func writeOutput(wg *sync.WaitGroup, m *sync.Mutex, options *input.Options, o output.Found) {
	defer wg.Done()

	if options.FileOutput != "" && options.Output == nil {
		file, err := os.OpenFile(options.FileOutput, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		if err != nil {
			gologger.Fatal().Msg(err.Error())
		}

		options.Output = file
	}

	m.Lock()

	out := o.Format()

	if options.JSON {
		outJSON, err := o.FormatJSON()
		if err != nil {
			gologger.Fatal().Msg(err.Error())
		}

		out = string(outJSON)
	}

	if options.Output != nil {
		if _, err := options.Output.Write([]byte(out + "\n")); err != nil && options.Verbose {
			gologger.Fatal().Msg(err.Error())
		}
	}

	m.Unlock()

	fmt.Println(out)
}
