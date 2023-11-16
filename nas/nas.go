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

// ListDirectory would call FRITZ API and return the response structure with results
// or error.
func (n *NAS) ListDirectory() (*BrowseResponse, error) {
	return n.ListDirectoryWithParams(nil)
}

// ListDirectoryWithParams is the same as ListDirectory, but accepts custom parameters
func (n *NAS) ListDirectoryWithParams(params map[string]string) (*BrowseResponse, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Set("sid", n.session.String())
	p.Set("sorting", "+filename")
	p.Set("c", "files")
	p.Set("a", "browse")

	if len(params) > 0 {
		for k, v := range params {
			p.Set(k, v)
		}
	}

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	res, err := request.GenericPostRequestWithContext(rctx, fullAddress, p)
	if err != nil {
		return nil, err
	}

	if !request.ValidateHeader(request.HeaderJSON, res.Header) {
		return nil, errors.New("incorrect response header content-type received")
	}

	d, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	// parse json
	var result *BrowseResponse
	err = json.Unmarshal(d, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
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
	res, err := request.GenericPostRequestWithContext(rctx, fullAddress, p)
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
		return nil, fmt.Errorf("return status code %d, response: %q", res.StatusCode, string(d))
	}

	// parse json
	var result *CreateDirResponse
	err = json.Unmarshal(d, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetFile downloads an object from FRITZ NAS storage
// Response is the object's data bytes (buffered) or error.
func (n *NAS) GetFile(Session string, path string) (io.ReadCloser, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasFileGetPath)

	p := url.Values{}
	p.Add("sid", Session)
	p.Add("script", fmt.Sprintf("/%s", rootAPI))
	p.Add("c", "files")
	p.Add("a", "get")
	p.Add("path", path)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	resp, err := request.GenericPostRequestWithContext(rctx, fullAddress, p)
	if err != nil {
		return nil, err
	}

	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("return status code %d, response: %q", resp.StatusCode, string(d))
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
	nasUploadFail nasUploadResult = "0"
	nasUploadOK   nasUploadResult = "1"
)

type nasResultCode string

const (
	nasResultOK          nasResultCode = "0" // 0 OK, even if no input is passed
	nasResultNoSession   nasResultCode = "5" // 5 (fileame provided, nothing else)
	nasResultDirNotExist nasResultCode = "9" // 9 (wrong dir)
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
		return nil, fmt.Errorf("return status code %d, response: %q", resp.StatusCode, string(d))
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

// RenameFile will rename a file in the NAS
// `from` takes the full path to the object
// `to` takes only object's new filename, without the extension
// The function returns how many files were updated and an error if any
func (n *NAS) RenameFile(from, to string) (int, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Add("sid", n.session.String())
	p.Add("c", "files")
	p.Add("a", "rename")
	p.Add("paths[1][path]", from)
	p.Add("paths[1][newName]", to)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()
	res, err := request.GenericPostRequestWithContext(rctx, fullAddress, p)
	if err != nil {
		return 0, err
	}

	if !request.ValidateHeader(request.HeaderJSON, res.Header) {
		return 0, errors.New("incorrect response header content-type received")
	}

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("return status code %d, response: %q", res.StatusCode, string(d))
	}

	// parse json
	var result *RenameResponse
	err = json.Unmarshal(d, &result)
	if err != nil {
		return 0, err
	}

	return result.RenameCount, nil
}

type DeleteResponse struct {
	DeleteCount int
}

// DeleteObject will delete a file or directory from the NAS.
// Response contains how many files have been affected and error (if any).
func (n *NAS) DeleteObject(Session, path string) (int, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Add("sid", Session)
	p.Add("c", "files")
	p.Add("a", "delete")
	p.Add("paths[1]", path)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()
	res, err := request.GenericPostRequestWithContext(rctx, fullAddress, p)
	if err != nil {
		return 0, err
	}

	if !request.ValidateHeader(request.HeaderJSON, res.Header) {
		return 0, errors.New("incorrect response header content-type received")
	}

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("return status code %d, response: %q", res.StatusCode, string(d))
	}

	// parse json
	var result *DeleteResponse
	err = json.Unmarshal(d, &result)
	if err != nil {
		return 0, err
	}

	return result.DeleteCount, nil
}

type MoveResponse struct {
	MoveCount int
}

// MoveFile moves a file from source to destination in the NAS
// This is a separate action and is not the same as rename.
func (n *NAS) MoveFile(Session, from, to string) (int, error) {
	fullAddress := fmt.Sprintf("%s/%s", n.address, nasURIPath)

	p := url.Values{}
	p.Add("sid", Session)
	p.Add("c", "files")
	p.Add("a", "move")
	p.Add("paths[1]", from)
	p.Add("target", to)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()
	res, err := request.GenericPostRequestWithContext(rctx, fullAddress, p)
	if err != nil {
		return 0, err
	}

	if !request.ValidateHeader(request.HeaderJSON, res.Header) {
		return 0, errors.New("incorrect response header content-type received")
	}

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("return status code %d, response: %q", res.StatusCode, string(d))
	}

	// parse json
	var result *MoveResponse
	err = json.Unmarshal(d, &result)
	if err != nil {
		return 0, err
	}

	return result.MoveCount, nil
}
