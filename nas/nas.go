package nas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/request"
)

const (
	nasURIPath        = "nas/api/data.lua"
	nasFileGetPath    = "nas/cgi-bin/luacgi_notimeout"
	nasFileUploadPath = "nas/cgi-bin/nasupload_notimeout"
	rootAPI           = "api/data.lua"
)

type NAS struct {
	session *auth.Session
	address string
}

func New(s *auth.Session) *NAS {
	return &NAS{
		session: s,
		address: auth.Address,
	}
}

func (n *NAS) WithAddress(addr string) *NAS {
	n.address = addr
	return n
}

type BrowseResponse struct {
	DiskInfo    DiskInfo
	Files       []File
	Directories []Directory
	WriteRight  bool
	Browse      Browse
}

type DiskInfo struct {
	Used  float64
	Total float64
	Free  float64
}

type File struct {
	Path        string
	Shared      bool
	Width       int
	StorageType string //internal_storage, external_storage?
	Type        string
	Height      int
	Timestamp   Timestamp
	Filename    string
	Size        int
}

type Directory struct {
	Path        string
	Shared      bool
	StorageType string
	Type        string
	Timestamp   Timestamp
	Filename    string
}

type Browse struct {
	Path       string
	Index      int
	TotalCount int
	Finished   bool
	Mode       string
	Limit      int
	Sorting    string
}

type Timestamp struct {
	time.Time
}

// UnmarshalJSON decodes te timestamp incorrectly, so we help
func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
	var i64 int64
	err := json.Unmarshal(bytes, &i64)
	if err != nil {
		return fmt.Errorf("error decoding timestamp: %w", err)
	}

	p.Time = time.Unix(i64, 0)
	return nil
}

// ListDirectory would call FRITZ API and return the response structure with results
// or error.
func (n *NAS) ListDirectory(path string) (*BrowseResponse, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Set("sid", n.session.String())
	p.Set("sorting", "+filename")
	p.Set("c", "files")
	p.Set("a", "browse")
	p.Set("limit", "100")

	// If no path is provided, we list the root of the storage
	if path == "" {
		path = "/"
	}
	p.Set("path", path)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	d, err := execute(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil {
		return nil, err
	}

	// parse json
	var result *BrowseResponse
	err = json.Unmarshal(d, &result)
	return result, err
}

type CreateDirResponse struct {
	Directory `json:"directory"`
}

func (n *NAS) CreateDir(name, path string) (*CreateDirResponse, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Set("sid", n.session.String())
	p.Set("path", path)
	p.Set("name", name)
	p.Set("parents", "false") // todo, find out
	p.Set("c", "files")
	p.Set("a", "create_dir")

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	d, err := execute(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil {
		return nil, err
	}

	// parse json
	var result *CreateDirResponse
	err = json.Unmarshal(d, &result)
	return result, err
}

// GetFile downloads an object from FRITZ NAS storage
// Response is the object's data bytes (buffered) or error.
func (n *NAS) GetFile(path string) (io.ReadCloser, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasFileGetPath)

	p := url.Values{}
	p.Add("sid", n.session.String())
	p.Add("script", fmt.Sprintf("/%s", rootAPI))
	p.Add("c", "files")
	p.Add("a", "get")
	p.Add("path", path)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	resp, err := request.GenericPostRequestWithContext(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil {
		return nil, err
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// extract the error from the body
		var e struct {
			Err SystemError `json:"error"`
		}
		errU := json.Unmarshal(d, &e)
		if errU != nil {
			return nil, fmt.Errorf("couldn't unmarshal error response, %w", err)
		}

		return nil, &e.Err
	}

	return io.NopCloser(bytes.NewReader(d)), nil
}

// Example:
//
//	{
//	    "sid": "e5363f0a80aacd5a",
//	    "dir": "/Dokumente",
//	    "Filename": "FRITZ-Picture.jpg",
//	    "SuccessfulUploads": "1",
//	    "ResultCode": "0"
//	}
type PutFileResponse struct {
	Sid               string
	Dir               string
	Filename          string
	SuccessfulUploads nasUploadResult
	ResultCode        nasResultCode
}

type nasUploadResult string

const (
	UploadResultFail nasUploadResult = "0"
	UploadResultOK   nasUploadResult = "1"
)

type nasResultCode string

const (
	ResultCodeOK          nasResultCode = "0" // 0 OK, even if no input is passed
	ResultCodeNoSession   nasResultCode = "5" // 5 (fileame provided, nothing else)
	ResultCodeDirNotExist nasResultCode = "9" // 9 (wrong dir)
)

func (n *NAS) PutFile(path string, data io.Reader) (*PutFileResponse, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasFileUploadPath)

	// Parse path into file and dir
	p := strings.Split(path, "/")
	file := p[len(p)-1]
	dir := strings.Join(p[:len(p)-1], "/")

	// Create multipart writer
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	// We need to insert this into the content type
	params := make(map[string]string)
	params["sid"] = n.session.String()
	params["dir"] = dir
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	// Add the file metadata
	w, err := writer.CreateFormFile("UploadFile", file)
	if err != nil {
		return nil, err
	}

	// Finally append the data
	_, err = io.Copy(w, data)
	if err != nil {
		return nil, err
	}

	writer.Close()

	// Send the request to the API
	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(rctx, http.MethodPost, fullAddress, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := request.HttpRequest(req)
	if err != nil {
		return nil, err
	}

	if !request.ValidateHeader(request.HeaderJSON, resp.Header) {
		return nil, errors.New("incorrect response header content-type received")
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// extract the error from the body
		var e struct {
			Err SystemError `json:"error"`
		}
		errU := json.Unmarshal(d, &e)
		if errU != nil {
			return nil, fmt.Errorf("couldn't unmarshal error response, %w", err)
		}

		return nil, &e.Err
	}

	// parse JSON response
	var result *PutFileResponse
	err = json.Unmarshal(d, &result)
	return result, err
}

type RenameResponse struct {
	RenameCount int
}

// sid=7857ff83dfb20d3b&
// paths%5B1%5D%5Bpath%5D=%2FDokumente%2F2022-0006_Anmeldung_fr_den_Schwimmkurs_Einverstndniserklrung.pdf&
// paths%5B1%5D%5BnewName%5D=2022-0006.pdf&
// c=files&
// a=rename

// RenameInput is the input struct holding parameters for RenameObject.
type RenameInput struct {
	From string
	To   string
}

// RenameObject renames files and/or directories in the NAS.
// It takes variadic slice of RenameInput{}, where its fields
// could represent file and/or directory.
// Response contains how many files have been affected and an error (if any).
func (n *NAS) RenameObject(params []*RenameInput) (int, error) {
	if len(params) == 0 {
		return 0, errors.New("no parameters supplied, cannot execute RenameObject command")
	}

	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Add("sid", n.session.String())
	p.Add("c", "files")
	p.Add("a", "rename")

	for k, v := range params {
		p.Add(fmt.Sprintf("paths[%d][path]", k+1), v.From)
		p.Add(fmt.Sprintf("paths[%d][newName]", k+1), v.To)
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	d, err := execute(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil {
		return 0, err
	}

	// parse json
	var result RenameResponse
	err = json.Unmarshal(d, &result)
	return result.RenameCount, err
}

type DeleteResponse struct {
	DeleteCount int
}

// DeleteObject deletes files and/or directories from the NAS.
// It takes variadic `paths` string parameter, representing file(s) and/or directory(ies)
// that will be deleted from the storage of the NAS.
// Response contains how many files have been affected and an error (if any).
func (n *NAS) DeleteObject(paths ...string) (int, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Add("sid", n.session.String())
	p.Add("c", "files")
	p.Add("a", "delete")

	for k, v := range paths {
		p.Add(fmt.Sprintf("paths[%d]", k+1), v)
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	d, err := execute(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil {
		return 0, err
	}

	// parse json
	var result DeleteResponse
	err = json.Unmarshal(d, &result)
	return result.DeleteCount, err
}

type MoveResponse struct {
	MoveCount int
}

// MoveObject moves files and or directories from source to destination within the NAS.
// It takes `dest` and a variadic `paths` (a.k.a the source paths to move) string parameter, representing file(s) and/or directory(ies).
// This is a separate action from rename.
func (n *NAS) MoveObject(dest string, paths ...string) (int, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Add("sid", n.session.String())
	p.Add("c", "files")
	p.Add("a", "move")
	p.Add("target", dest)

	for k, v := range paths {
		p.Add(fmt.Sprintf("paths[%d]", k+1), v)
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	d, err := execute(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil {
		return 0, err
	}

	// parse json
	var result *MoveResponse
	err = json.Unmarshal(d, &result)
	if err != nil {
		return 0, err
	}

	return result.MoveCount, nil
}

// execute will call the API server
func execute(ctx context.Context, addr string, body io.Reader) ([]byte, error) {
	res, err := request.GenericPostRequestWithContext(ctx, addr, body)
	if err != nil {
		return nil, err
	}

	if !request.ValidateHeader(request.HeaderJSON, res.Header) {
		return nil, errors.New("incorrect response header content-type received")
	}

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		// extract the error from the body
		var e struct {
			Err SystemError `json:"error"`
		}
		errU := json.Unmarshal(d, &e)
		if errU != nil {
			return nil, fmt.Errorf("couldn't unmarshal error response, %w", err)
		}

		return nil, &e.Err
	}

	return d, nil
}
