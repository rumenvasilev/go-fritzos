package nas

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		want        interface{}
		wantElement int
	}{
		{"int-n-dir", "../testfixtures/nas-browse.json", dirMockResponse, 0},
		{"float-n-file", "../testfixtures/nas-browse-float.json", fileMockResponse, 0},
		{"jpg", "../testfixtures/nas-browse-float.json", fileImageMockResponse, 1},
	}

	is := is.New(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var resp *BrowseResponse
			f, err := os.ReadFile(tt.path)
			is.NoErr(err)
			err = json.Unmarshal(f, &resp)
			is.NoErr(err)
			is.True(resp != nil)
			if wantFile, ok := tt.want.(File); ok {
				is.Equal(resp.Files[tt.wantElement], wantFile)
			} else if wantDir, ok := tt.want.(Directory); ok {
				is.Equal(resp.Directories[tt.wantElement], wantDir)
			}
		})
	}
}

var dirMockResponse = Directory{
	Path:        "/Bilder",
	Shared:      false,
	StorageType: "internal_storage",
	Type:        "directory",
	Timestamp:   getMockTimestamp(1700007000),
	Filename:    "Bilder",
}

var fileMockResponse = File{
	Path:        "/Dokumente/FRITZ-NAS.txt",
	Shared:      false,
	StorageType: "internal_storage",
	Type:        "document",
	Timestamp:   getMockTimestamp(1322694000),
	Filename:    "FRITZ-NAS.txt",
	Size:        6640,
}

var fileImageMockResponse = File{
	Path:        "/Dokumente/FRITZ-Picture.jpg",
	Shared:      false,
	Width:       640,
	StorageType: "internal_storage",
	Type:        "picture",
	Height:      400,
	Timestamp:   getMockTimestamp(1700084400),
	Filename:    "FRITZ-Picture.jpg",
	Size:        30192,
}

func getMockTimestamp(ts int) Timestamp {
	return Timestamp{time.Unix(int64(ts), 0)}
}
