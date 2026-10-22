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
	Input      chan string
	Output     chan output.Found
	Result     output.Result
	UserAgent  string
	InWg       *sync.WaitGroup
	OutWg      *sync.WaitGroup
	Options    input.Options
	outputFile *os.File
}

// New takes as input the options and returns
// a new runner.
func New(options *input.Options) Runner {
	r := Runner{
		Input:     make(chan string, options.Concurrency),
		Output:    make(chan output.Found, options.Concurrency),
		Result:    output.New(),
		UserAgent: golazy.GenerateRandomUserAgent(),
		InWg:      &sync.WaitGroup{},
		OutWg:     &sync.WaitGroup{},
		Options:   *options,
	}

	if r.Options.FileOutput != "" && r.Options.Output == nil {
		file, err := os.Create(r.Options.FileOutput)
		if err != nil {
			gologger.Fatal().Msgf("%s", err)
		}

		r.Options.Output = file
		r.outputFile = file
	}

	return r
}

// Run takes the input and executes all the tasks
// specified in the options.
func (r *Runner) Run() {
	r.OutWg.Add(1)

	go pullOutput(r)

	r.InWg.Add(1)

	go execute(r)

	r.InWg.Add(1)

	go pushInput(r)

	r.InWg.Wait()

	close(r.Output)
	r.OutWg.Wait()

	if r.outputFile != nil {
		if err := r.outputFile.Close(); err != nil {
			gologger.Error().Msgf("%s", err)
		}
	}
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

		// Scan stops at the first error (e.g. a line longer than 64 KiB):
		// report it, otherwise the rest of the input is silently ignored.
		if err := scanner.Err(); err != nil {
			gologger.Error().Msgf("error while reading stdin: %s", err)
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
- Body length is > 0 (some sites return 200 but empty).
- Body is an image (soft-404s, SPA index pages and login redirects are not).
If valid, hash and lookup.

Otherwise fallback to parsing <link rel="icon" ...> in the HTML <head>,
unless the request failed altogether (e.g. host unreachable):
- Fetch the page (the input URL, with the scheme added if missing).
- Parse the HTML.
- Extract: rel=icon, rel=shortcut icon, rel=apple-touch-icon...
If href is:
- A relative path → resolve against <base href> and the final page URL (after redirects).
- A full URL → use as-is.
- Data URL → decode and hash directly.
Try them in document order, skipping the URLs already tried, until one is found
in the database (downloading at most MaxFaviconLinks of them).
*/
func execute(r *Runner) {
	defer r.InWg.Done()

	client, err := customClient(&r.Options)
	if err != nil {
		gologger.Fatal().Msgf("%s", err)
	}

	f := &fetcher{client: client, limiter: rateLimiter(r), userAgent: r.UserAgent}

	for i := 0; i < r.Options.Concurrency; i++ {
		r.InWg.Add(1)

		go func() {
			defer r.InWg.Done()

			for value := range r.Input {
				if found, ok := lookup(r, f, value); ok {
					r.Output <- found
				}
			}
		}()
	}
}

// lookup returns the first favicon of the target found in the database
// (and matching the input hashes, if any).
func lookup(r *Runner, f *fetcher, value string) (output.Found, bool) {
	pageURL, faviconURL, err := prepareURLs(value)
	if err != nil {
		gologger.Error().Msgf("%s", err)

		return output.Found{}, false
	}

	finalURL, hash, err := getFavicon(faviconURL, f)

	switch {
	case err == nil:
		if name, ok := checkDB(r, hash, faviconURL); ok {
			return output.Found{URL: value, Name: name, Hash: hash}, true
		}

		return output.Found{}, false
	case errors.Is(err, ErrFaviconNotFound), errors.Is(err, ErrEmptyBody), errors.Is(err, ErrNotAnImage):
		gologger.Debug().Msgf("%s for url %s", err.Error(), value)
	default:
		// The request itself failed: the HTML page would not be served either.
		gologger.Error().Msgf("%s", err)

		return output.Found{}, false
	}

	gologger.Debug().Msgf("Fallback to HTML parsing for %s", value)

	links, err := faviconLinks(pageURL, f)
	if err != nil {
		gologger.Debug().Msgf("Favicon not found for %s: %s", value, err)

		return output.Found{}, false
	}

	tried := map[string]struct{}{faviconURL: {}, finalURL: {}}

	return lookupLinks(r, f, value, links, tried)
}

// lookupLinks checks the favicons referenced by the <link> tags of the page,
// skipping the URLs already tried (also as redirect targets).
func lookupLinks(r *Runner, f *fetcher, value string, links []string, tried map[string]struct{}) (output.Found, bool) {
	downloads := 0

	for _, link := range links {
		if _, ok := tried[link]; ok {
			continue
		}

		tried[link] = struct{}{}

		var (
			source = link
			hash   string
			err    error
		)

		switch {
		case isDataURI(link):
			source = "data URI in " + value
			hash, err = hashDataURI(link)
		case downloads < MaxFaviconLinks:
			downloads++

			var finalURL string

			finalURL, hash, err = getFavicon(link, f)
			tried[finalURL] = struct{}{}
		default:
			continue
		}

		if err != nil {
			gologger.Debug().Msgf("Favicon not found for %s: %s", source, err)

			continue
		}

		if name, ok := checkDB(r, hash, source); ok {
			return output.Found{URL: value, Name: name, Hash: hash}, true
		}
	}

	return output.Found{}, false
}

// checkDB looks the hash up in the database, honoring the -hash filter.
func checkDB(r *Runner, hash, faviconURL string) (string, bool) {
	name, err := CheckFavicon(hash, r.Options.Hash, faviconURL)
	if err != nil {
		if r.Options.Verbose {
			gologger.Error().Msgf("%s", err)
		}

		return "", false
	}

	return name, true
}

// pullOutput is the only goroutine writing results, so that
// writes never interleave and need no locking.
func pullOutput(r *Runner) {
	defer r.OutWg.Done()

	for o := range r.Output {
		if !r.Result.Printed(o.URL) {
			writeOutput(&r.Options, o)
		}
	}
}

func writeOutput(options *input.Options, o output.Found) {
	out := o.Format()

	if options.JSON {
		outJSON, err := o.FormatJSON()
		if err != nil {
			gologger.Fatal().Msg(err.Error())
		}

		out = string(outJSON)
	}

	if options.Output != nil {
		if _, err := options.Output.Write([]byte(out + "\n")); err != nil {
			gologger.Error().Msg(err.Error())
		}
	}

	fmt.Println(out)
}
