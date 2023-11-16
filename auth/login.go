package auth

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/rumenvasilev/go-fritzos/request"
	"golang.org/x/crypto/pbkdf2"
)

const (
	Address   = "http://fritz.box"
	loginPath = "login_sid.lua?version=2"
)

type Session string

func (s *Session) String() string {
	return string(*s)
}

// Auth will authenticate to the target FRITZOS device using default address
// and will return either session id, either error.
func Auth(username, password string) (*Session, error) {
	return AuthWithAddress(Address, username, password)
}

func AuthWithAddress(address, username, password string) (*Session, error) {
	challenge, err := getChallengeString(address)
	if err != nil {
		return nil, err
	}

	answer, err := solveChallenge(challenge, password)
	if err != nil {
		return nil, err
	}

	return authenticate(address, answer, username)
}

// Close will logout from the authenticated device
func Close(sid *Session) error {
	return CloseWithAddress(Address, sid)
}

func CloseWithAddress(address string, sid *Session) error {
	fullAddress := fmt.Sprintf("%s/%s", address, loginPath)
	data := url.Values{}
	data.Set("logout", sid.String())

	resp, err := http.PostForm(fullAddress, data)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("couln't logout session id %q, %w", sid.String(), err)
	}

	return nil
}

// <?xml version="1.0" encoding="utf-8"?>
// <SessionInfo>
//
//	<SID>0000000000000000</SID>
//	<Challenge>2$60000$492c53ea76b6789a7c40301272276aca$6000$16a80a6246d0ec41a816cc9e03e0e01f</Challenge>
//	<BlockTime>0</BlockTime>
//	<Rights></Rights>
//	<Users>
//		<User last="1">some</User>
//	</Users>
//
// </SessionInfo>
type sessionInfo struct {
	SID       string
	Challenge string
	BlockTime uint8
	Rights    string `xml:"Rights,omitempty"`
	Users     []user `xml:"Users>User"`
}

type user struct {
	Value string `xml:",chardata"`
	Last  int    `xml:"last,attr"`
}

func getChallengeString(address string) (string, error) {
	url := fmt.Sprintf("%s/%s", address, loginPath)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("couldn't get the challenge, response code was %d", resp.StatusCode)
	}

	if !request.ValidateHeader(request.HeaderXML, resp.Header) {
		return "", fmt.Errorf("expected %s response, but got something else", request.HeaderXML)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var session *sessionInfo
	err = xml.Unmarshal(data, &session)
	if err != nil {
		return "", err
	}

	return session.Challenge, err
}

func solveChallenge(challenge, password string) (string, error) {
	noMatch := errors.New("cannot solve challenge, input string is not in the expected format")
	parts := strings.Split(challenge, "$")

	switch len(parts) {
	case 1:
		// version 1
		// MD5 encryption
		return calculateMD5Response(challenge, password), nil
	case 5:
		// version 2
		// pbkdf encryption FRITZ!OS 7.24 and later
		if parts[0] == "2" {
			return calculatePBKDF2Response(parts, password)
		}
		return "", noMatch
	}

	return "", noMatch
}

// calculatePBKDF2Response uses pbkdf encryption and is supported by
// FRITZ!OS 7.24 and later.
// Format of challenge string: 	2$<iter1>$<salt1>$<iter2>$<salt2>
// Example: 2$60000$492c53ea76b6789a7c40301272276aca$6000$3c52023c5900a6381ef790b915941c5a
func calculatePBKDF2Response(parts []string, password string) (string, error) {
	// pbkdf encryption
	iter1, _ := strconv.Atoi(parts[1])
	salt1, err := hex.DecodeString(parts[2])
	if err != nil {
		return "", err
	}

	iter2, _ := strconv.Atoi(parts[3])
	salt2, err := hex.DecodeString(parts[4])
	if err != nil {
		return "", err
	}

	key1 := pbkdf2.Key([]byte(password), salt1, iter1, 32, sha256.New)
	key2 := pbkdf2.Key([]byte(key1), salt2, iter2, 32, sha256.New)

	return fmt.Sprintf("%s$%x", parts[4], key2), nil
}

// The MD5 hash is generated from the byte sequence of the UTF-16LE coding of this string (without
// BOM and without terminating 0 bytes).
func calculateMD5Response(challenge, password string) string {
	codes := utf16.Encode([]rune(fmt.Sprintf("%s-%s", challenge, password)))
	b := make([]byte, len(codes)*2)
	for i, r := range codes {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8)
	}
	return fmt.Sprintf("%s-%x", challenge, md5.Sum(b))

}

func authenticate(address, challenge, username string) (*Session, error) {
	loginURL := fmt.Sprintf("%s/%s", address, loginPath)
	v := url.Values{}
	v.Set("username", username)
	v.Set("response", challenge)

	resp, err := http.PostForm(loginURL, v)
	if err != nil {
		return nil, fmt.Errorf("something went wrong, %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned non-200 status code -> %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)

	var session *sessionInfo
	err = xml.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}

	if session.SID == "" || session.SID == "0000000000000000" {
		return nil, errors.New("login failed, session id is wrong")
	}

	sid := Session(session.SID)
	return &sid, nil
}
