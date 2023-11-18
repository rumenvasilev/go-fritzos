package nas

import (
	"encoding/json"
	"fmt"
)

type SystemError struct {
	Message string
	Data    *json.RawMessage
	Code    int // http response code
}

type FileSystemControllerError struct {
	Message string
	Path    string
	Code    int // internal NAS response code
}

func (e *SystemError) Error() string {
	return e.Message
}

func (e *SystemError) Unwrap() error {
	var fse *FileSystemControllerError
	_ = json.Unmarshal(*e.Data, &fse)
	return fse
}

func (e *FileSystemControllerError) Error() string {
	return fmt.Sprintf("Code: %d, Path: %s, Msg: %s", e.Code, e.Path, e.Message)
}
