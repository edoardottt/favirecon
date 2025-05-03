/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net"

	"github.com/projectdiscovery/mapcidr"
	"github.com/twmb/murmur3"
)

var (
	ErrCidrBadFormat = errors.New("malformed input CIDR")
	ErrEmptyBody     = errors.New("empty body")
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
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

// GetFaviconHash computes the murmur3 hash.
func GetFaviconHash(input []byte) string {
	b64 := base64Content(input)
	return fmt.Sprint(int32(murmur3.Sum32(b64)))
}

func handleCidrInput(inputCidr string) ([]string, error) {
	if !isCidr(inputCidr) {
		return nil, ErrCidrBadFormat
	}

	ips, err := mapcidr.IPAddresses(inputCidr)
	if err != nil {
		return nil, err
	}

	return ips, nil
}

// isCidr determines if the given ip is a cidr range.
func isCidr(inputCidr string) bool {
	_, _, err := net.ParseCIDR(inputCidr)
	return err == nil
}
