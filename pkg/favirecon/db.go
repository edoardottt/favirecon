/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	_ "embed"

	"github.com/projectdiscovery/goflags"
)

//nolint:gochecknoglobals
var (
	//go:embed db.json
	dbJSON string

	db                 map[string]string
	ErrHashNotFound    = errors.New("hash not found")
	ErrHashNotMatching = errors.New("hash not matching hash provided")
)

func init() {
	if err := json.Unmarshal([]byte(dbJSON), &db); err != nil {
		log.Fatal("error while unmarshaling db")
	}
}

// CheckFavicon checks if faviconHash is present in the database. If hash (slice) is not empty,
// it checks also if that faviconHash is one of the inputted hashes.
// If faviconHash is not found, an error is returned.
func CheckFavicon(faviconHash string, hash goflags.StringSlice, url ...string) (string, error) {
	if k, ok := db[faviconHash]; ok {
		if len(hash) != 0 {
			if contains(hash, faviconHash) {
				return k, nil
			}

			return "", fmt.Errorf("[%s] %s %w", faviconHash, url, ErrHashNotMatching)
		}

		return k, nil
	}

	if len(url) == 0 {
		return "", fmt.Errorf("%w", ErrHashNotFound)
	}

	return "", fmt.Errorf("[%s] %s %w", faviconHash, url, ErrHashNotFound)
}
