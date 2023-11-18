package nas

import (
	"encoding/json"
	"testing"

	"github.com/matryer/is"
)

var sampleError = `{
	"error": {
		"message": "[file_system_controller] The folder with the name \"blabla\" already exists and therefore cannot be created.",
		"data": {
			"message": "The folder with the name \"blabla\" already exists and therefore cannot be created.",
			"path": "/Dokumente",
			"code": 5
		},
		"code": 400
	}
}`

var sampleErrorStruct = SystemError{
	Message: "[file_system_controller] The folder with the name \"blabla\" already exists and therefore cannot be created.",
	Data:    nil,
	Code:    400,
}

var sampleFileSystemControllerErrorStruct = FileSystemControllerError{
	Message: "The folder with the name \"blabla\" already exists and therefore cannot be created.",
	Path:    "/Dokumente",
	Code:    5,
}

func TestSystemError(t *testing.T) {
	is := is.New(t)

	var e struct {
		Err *SystemError `json:"error"`
	}
	err := json.Unmarshal([]byte(sampleError), &e)
	is.NoErr(err)

	t.Run("SystemError tests", func(t *testing.T) {
		is.True(e.Err != nil)
		is.Equal(e.Err.Error(), (&sampleErrorStruct).Error())
	})
	t.Run("FileSystemControllerError tests", func(t *testing.T) {
		is.True(e.Err.Data != nil)
		is.True(e.Err.Unwrap() != nil)
		is.Equal(e.Err.Unwrap(), &sampleFileSystemControllerErrorStruct)
		is.Equal(e.Err.Unwrap().Error(), "Code: 5, Path: /Dokumente, Msg: The folder with the name \"blabla\" already exists and therefore cannot be created.")
	})
}
