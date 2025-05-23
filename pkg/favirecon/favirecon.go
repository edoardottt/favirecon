/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"bufio"
	"errors"
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

/*
Try /favicon.ico first. Most common and lightweight check.
Accept it only if:
- Status is 200.
- Content-Type is an image.
- Body length is > 0 (some sites return 200 but empty).
If valid, hash and lookup. ✅ Done.

Fallback to parsing <link rel="icon" ...> in the HTML <head>:
- Fetch original input URL (not with /favicon.ico appended).
- Parse the HTML.
- Extract: rel=icon, rel=shortcut icon, rel=apple-touch-icon
If href is:
- A relative path → resolve against base URL.
- A full URL → use as-is.
- Data URL → decode and hash directly.
*/
func execute(r *Runner) {
	defer r.InWg.Done()

	rl := rateLimiter(r)

	for i := 0; i < r.Options.Concurrency; i++ {
		r.InWg.Add(1)

		go func() {
			defer r.InWg.Done()

			client, err := customClient(&r.Options)
			if err != nil {
				gologger.Error().Msgf("%s", err)

				return
			}

			for value := range r.Input {
				faviconURL, err := PrepareURL(value)
				if err != nil {
					gologger.Error().Msgf("%s", err)

					continue
				}

				rl.Take()

				found, result, err := getFavicon(faviconURL, r.UserAgent, client)
				if err != nil {
					if errors.Is(err, ErrFaviconNotFound) {
						gologger.Debug().Msgf("%s for url %s", err.Error(), value)
					} else {
						gologger.Error().Msgf("%s", err)
					}
				}

				if !found {
					gologger.Debug().Msgf("Fallback to HTML parsing for %s", value)

					faviconURL, result, err = extractFaviconFromHTML(value, r.UserAgent, client)
					if err != nil {
						gologger.Debug().Msgf("Favicon not found for %s: %s", value, err)
						continue
					}
				}

				foundDB, err := CheckFavicon(result, r.Options.Hash, faviconURL)
				if err != nil {
					if r.Options.Verbose {
						gologger.Error().Msgf("%s", err)
					}

					continue
				}

				r.Output <- output.Found{URL: value, Name: foundDB, Hash: result}
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
