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
	"time"

	"github.com/edoardottt/favirecon/pkg/input"
	"github.com/projectdiscovery/gologger"
)

const (
	TLSHandshakeTimeout = 10
	KeepAlive           = 30
	MaxIdleConns        = 100
	MaxIdleConnsPerHost = 10
	IdleConnTimeout     = 90
)

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

func getFavicon(url, ua string, client *http.Client) (bool, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, "", err
	}

	gologger.Debug().Msgf("Checking favicon for %s", url)

	req.Header.Add("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return false, "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, "", ErrFaviconNotFound
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	if len(body) == 0 {
		return false, "", ErrEmptyBody
	}

	return true, GetFaviconHash(body), nil
}
