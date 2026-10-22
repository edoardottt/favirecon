/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"errors"
	"net/url"
	"strings"
)

const (
	MinURLLength = 4
)

var (
	ErrMalformedURL = errors.New("malformed input URL")
)

// PrepareURL takes as input a string and prepares
// the input URL in order to get the favicon icon.
func PrepareURL(input string) (string, error) {
	_, faviconURL, err := prepareURLs(input)
	return faviconURL, err
}

// prepareURLs normalizes the input (adding the scheme if missing) and returns
// the URL of the page, where <link> tags are looked for, and the URL of the
// default favicon.
func prepareURLs(input string) (string, string, error) {
	if len(input) < MinURLLength {
		return "", "", ErrMalformedURL
	}

	if !strings.Contains(input, "://") {
		input = "http://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", "", err
	}

	page := url.URL{Scheme: u.Scheme, Host: u.Host, Path: u.Path, RawQuery: u.RawQuery}

	if strings.HasSuffix(u.Path, ".ico") {
		// The input is already a favicon: its page is the directory containing it.
		page.Path = u.Path[:strings.LastIndex(u.Path, "/")+1]
		page.RawQuery = ""
	} else {
		if !strings.HasSuffix(u.Path, "/") {
			u.Path += "/"
		}

		u.Path += "favicon.ico"
	}

	return page.String(), u.Scheme + "://" + u.Host + u.Path, nil
}
