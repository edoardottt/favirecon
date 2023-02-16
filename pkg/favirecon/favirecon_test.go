/*
favirecon - Use favicon.ico to improve your target recon phase. Quickly detect technologies, WAF, exposed panels, known services.

This repository is under MIT License https://github.com/edoardottt/favirecon/blob/main/LICENSE
*/

package favirecon_test

import (
	"testing"

	"github.com/edoardottt/favirecon/pkg/favirecon"
	"github.com/stretchr/testify/require"
)

func TestGetFaviconHash(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "Test #1",
			input: []byte("test"),
			want:  "-1541278541",
		},
		{
			name:  "Test #2",
			input: []byte("wiytl8q2yvb58q2y58i34yv38l4yo853ybtv853y4vv38y4ov38y8oyv4348yoylo4"),
			want:  "1897381022",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := favirecon.GetFaviconHash(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPrepareURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		err   error
	}{
		{
			name:  "empty input",
			input: "",
			want:  "",
			err:   favirecon.ErrMalformedURL,
		},
		{
			name:  "URL without protocol without path",
			input: "edoardoottavianelli.it",
			want:  "http://edoardoottavianelli.it/favicon.ico",
			err:   nil,
		},
		{
			name:  "URL without protocol with path",
			input: "edoardoottavianelli.it/",
			want:  "http://edoardoottavianelli.it/favicon.ico",
			err:   nil,
		},
		{
			name:  "URL with protocol without path",
			input: "http://edoardoottavianelli.it",
			want:  "http://edoardoottavianelli.it/favicon.ico",
			err:   nil,
		},
		{
			name:  "URL with protocol and path (no final slash)",
			input: "http://edoardoottavianelli.it/test",
			want:  "http://edoardoottavianelli.it/test/favicon.ico",
			err:   nil,
		},
		{
			name:  "URL with protocol and path (final slash)",
			input: "http://edoardoottavianelli.it/test/",
			want:  "http://edoardoottavianelli.it/test/favicon.ico",
			err:   nil,
		},
		{
			name:  "URL without protocol and path (final slash) with icon",
			input: "edoardoottavianelli.it/test/favicon.ico",
			want:  "http://edoardoottavianelli.it/test/favicon.ico",
			err:   nil,
		},
		{
			name:  "URL with protocol and path (final slash) with icon",
			input: "http://edoardoottavianelli.it/test/favicon.ico",
			want:  "http://edoardoottavianelli.it/test/favicon.ico",
			err:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := favirecon.PrepareURL(tt.input)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.want, got)
		})
	}
}
