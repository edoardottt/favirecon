/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/edoardottt/favirecon/pkg/input"
	"github.com/projectdiscovery/gologger"
	"go.uber.org/ratelimit"
)

const (
	TLSHandshakeTimeout = 10
	KeepAlive           = 30
	MaxIdleConns        = 100
	MaxIdleConnsPerHost = 10
	IdleConnTimeout     = 90
)

// fetcher performs the HTTP requests of all the workers, which share
// the same client (and connection pool) and the same rate limiter.
type fetcher struct {
	client    *http.Client
	limiter   ratelimit.Limiter
	userAgent string
}

func customClient(options *input.Options) (*http.Client, error) {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(options.Timeout) * time.Second,
			KeepAlive: KeepAlive * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: TLSHandshakeTimeout * time.Second,
		MaxIdleConns:        MaxIdleConns,
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
		IdleConnTimeout:     IdleConnTimeout * time.Second,
	}

	if options.Proxy != "" {
		u, err := url.Parse(options.Proxy)
		if err != nil {
			return nil, err
		}

		transport.Proxy = http.ProxyURL(u)

		if options.Verbose {
			gologger.Debug().Msgf("Using Proxy %s", options.Proxy)
		}
	}

	client := http.Client{
		Transport: &transport,
		Timeout:   time.Duration(options.Timeout) * time.Second,
	}

	return &client, nil
}

// get performs a GET request, waiting for the rate limiter first:
// every request counts towards the rate limit, not only the first one of a target.
func (f *fetcher) get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", f.userAgent)

	f.limiter.Take()

	return f.client.Do(req)
}

// getFavicon downloads the favicon at url and returns the URL it was finally
// served from (after redirects) and its hash. The response is accepted only if
// its status is 200 and its body is a non-empty image.
func getFavicon(url string, f *fetcher) (string, string, error) {
	gologger.Debug().Msgf("Checking favicon for %s", url)

	resp, err := f.get(url)
	if err != nil {
		return url, "", err
	}

	defer func() { _ = resp.Body.Close() }()

	finalURL := resp.Request.URL.String()

	if resp.StatusCode != http.StatusOK {
		return finalURL, "", ErrFaviconNotFound
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return finalURL, "", err
	}

	if len(body) == 0 {
		return finalURL, "", ErrEmptyBody
	}

	if !isImage(resp.Header.Get("Content-Type"), body) {
		return finalURL, "", ErrNotAnImage
	}

	return finalURL, GetFaviconHash(body), nil
}

// isImage reports whether body, served with the given Content-Type, is an image.
// Binary formats are recognized by their signature whatever the Content-Type is,
// while HTML pages (soft-404s, SPA index pages, login redirects...) are rejected
// even when they claim to be images.
func isImage(contentType string, body []byte) bool {
	sniffed := http.DetectContentType(body)
	contentType = strings.ToLower(contentType)

	switch {
	case strings.HasPrefix(sniffed, "image/"):
		return true
	case strings.HasPrefix(contentType, "image/svg"):
		// SVG has no binary signature: it sniffs as XML, text or even HTML (leading comment).
		return true
	case strings.HasPrefix(sniffed, "text/html"):
		return false
	default:
		return strings.HasPrefix(contentType, "image/")
	}
}
