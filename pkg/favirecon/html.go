/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	// MaxFaviconLinks is the maximum number of favicons downloaded from the
	// <link> tags of a page (data URIs do not count, they need no request).
	MaxFaviconLinks = 3
)

var (
	ErrFaviconNotFound        = errors.New("favicon not found")
	ErrFaviconLinkTagNotFound = errors.New("no favicon link tag found")
	ErrHTMLNotFetched         = errors.New("failed to fetch HTML")
	ErrInvalidDataURI         = errors.New("invalid data URI")
	ErrNotAnImage             = errors.New("not an image")
)

// faviconLinks fetches pageURL and returns the favicons referenced by its
// <link rel="icon" ...> tags (icon, shortcut icon, apple-touch-icon...), in
// document order. Relative references are resolved against <base href> and
// the final URL of the page, after redirects.
func faviconLinks(pageURL string, f *fetcher) ([]string, error) {
	resp, err := f.get(pageURL)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrHTMLNotFetched
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	base := resp.Request.URL

	if href, ok := doc.Find("base[href]").First().Attr("href"); ok {
		if u, err := base.Parse(strings.TrimSpace(href)); err == nil {
			base = u
		}
	}

	var links []string

	doc.Find("link[href]").Each(func(_ int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		href, _ := s.Attr("href")
		href = strings.TrimSpace(href)

		if !strings.Contains(strings.ToLower(rel), "icon") || href == "" {
			return
		}

		if isDataURI(href) {
			links = append(links, href)

			return
		}

		if u, err := base.Parse(href); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
			links = append(links, u.String())
		}
	})

	if len(links) == 0 {
		return nil, ErrFaviconLinkTagNotFound
	}

	return links, nil
}

func isDataURI(href string) bool {
	return strings.HasPrefix(strings.ToLower(href), "data:")
}

// hashDataURI returns the hash of the favicon embedded in a data URI.
func hashDataURI(uri string) (string, error) {
	mediaType, content, err := decodeDataURI(uri)
	if err != nil {
		return "", err
	}

	if len(content) == 0 {
		return "", ErrEmptyBody
	}

	if !isImage(mediaType, content) {
		return "", ErrNotAnImage
	}

	return GetFaviconHash(content), nil
}

// decodeDataURI returns the media type and the content of a data URI (RFC 2397),
// either base64 (data:image/png;base64,...) or percent-encoded (data:image/svg+xml,...).
func decodeDataURI(uri string) (string, []byte, error) {
	header, data, ok := strings.Cut(uri[len("data:"):], ",")
	if !ok {
		return "", nil, ErrInvalidDataURI
	}

	data, err := url.PathUnescape(data)
	if err != nil {
		return "", nil, err
	}

	mediaType, isBase64 := strings.CutSuffix(strings.ToLower(header), ";base64")
	if !isBase64 {
		return mediaType, []byte(data), nil
	}

	content, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(data), ""))
	if err != nil {
		return "", nil, err
	}

	return mediaType, content, nil
}
