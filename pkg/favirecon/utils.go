/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twmb/murmur3"
)

const (
	TLSHandshakeTimeout = 10
	KeepAlive           = 30
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

func getFavicon(url, ua string, client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return GetFaviconHash(body), nil
}

func prepareURL(input string) (string, error) {
	if !strings.Contains(input, "://") {
		input = "http://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	if len(u.Path) == 0 || u.Path[len(u.Path)-1:] != "/" {
		u.Path += "/"
	}

	return u.Scheme + "://" + u.Host + u.Path, nil
}

// base64Content : RFC2045.
func base64Content(input []byte) []byte {
	inputEncoded := base64.StdEncoding.EncodeToString(input)
	buffer := bytes.Buffer{}

	for i := 0; i < len(inputEncoded); i++ {
		ch := inputEncoded[i]
		buffer.WriteByte(ch)

		if (i+1)%76 == 0 { // 76 bytes.
			buffer.WriteByte('\n')
		}
	}

	buffer.WriteByte('\n')

	return buffer.Bytes()
}

func GetFaviconHash(input []byte) string {
	b64 := base64Content(input)
	return fmt.Sprint(int32(murmur3.Sum32(b64)))
}

func customClient(timeout int) *http.Client {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(timeout) * time.Second,
			KeepAlive: KeepAlive * time.Second,
		}).Dial,
		TLSHandshakeTimeout: TLSHandshakeTimeout * time.Second,
	}

	client := http.Client{
		Transport: &transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	return &client
}
