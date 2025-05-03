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

func resolveURL(baseURL, ref string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ref // fallback
	}

	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	return base.ResolveReference(u).String()
}

// PrepareURL takes as input a string and prepares
// the input URL in order to get the favicon icon.
func PrepareURL(input string) (string, error) {
	if len(input) < MinURLLength {
		return "", ErrMalformedURL
	}

	if !strings.Contains(input, "://") {
		input = "http://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	if !strings.HasSuffix(u.Path, ".ico") {
		if !strings.HasSuffix(u.Path, "/") {
			u.Path += "/"
		}

		u.Path += "favicon.ico"
	}

	return u.Scheme + "://" + u.Host + u.Path, nil
}
