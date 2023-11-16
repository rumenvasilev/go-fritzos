package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"int", "./testfixtures/nas-browse.json"},
		{"float", "./testfixtures/nas-browse-float.json"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var resp *NASBrowseResponse
			f, err := os.ReadFile(tt.path)
			require.NoError(t, err)
			err = json.Unmarshal(f, &resp)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}

}
