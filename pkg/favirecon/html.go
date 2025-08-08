/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrFaviconNotFound        = errors.New("favicon not found")
	ErrFaviconLinkTagNotFound = errors.New("no favicon link tag found")
	ErrHTMLNotFetched         = errors.New("failed to fetch HTML")
	ErrInvalidDataURI         = errors.New("invalid data URI")
)

func extractFaviconFromHTML(pageURL, ua string, client *http.Client) (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Add("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", ErrHTMLNotFetched
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", err
	}

	var faviconHref string

	doc.Find("link").EachWithBreak(func(i int, s *goquery.Selection) bool {
		rel, _ := s.Attr("rel")
		href, ok := s.Attr("href")

		if ok && strings.Contains(strings.ToLower(rel), "icon") {
			faviconHref = href
			return false // break loop
		}

		return true
	})

	if faviconHref == "" {
		return "", "", ErrFaviconLinkTagNotFound
	}

	// handle base64 data
	if strings.HasPrefix(faviconHref, "data:image") {
		base64Data := strings.SplitN(faviconHref, ",", 2)
		if len(base64Data) != 2 {
			return "", "", ErrInvalidDataURI
		}

		decoded, err := base64.StdEncoding.DecodeString(base64Data[1])
		if err != nil {
			return "", "", err
		}

		return faviconHref, GetFaviconHash(decoded), nil
	}

	faviconURL := resolveURL(pageURL, faviconHref)

	found, favicon, err := getFavicon(faviconURL, ua, client)
	if err != nil {
		return faviconURL, "", err
	}

	if !found {
		return "", "", ErrFaviconNotFound
	}

	return faviconURL, favicon, nil
}
