/*
csprecon - Discover new target domains using Content Security Policy

This repository is under MIT License https://github.com/edoardottt/csprecon/blob/main/LICENSE
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
