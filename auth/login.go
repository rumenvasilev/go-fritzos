package auth

import (
	"context"
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
	"time"
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

// Close will logout from the authenticated device.
// Call is time limited to 30 seconds, after which it will terminate.
func (s *Session) Close() error {
	return s.CloseWithAddress(Address)
}

func (s *Session) CloseWithAddress(address string) error {
	fullAddress := fmt.Sprintf("%s/%s", address, loginPath)
	p := url.Values{}
	p.Set("logout", s.String())

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	resp, err := request.GenericPostRequestWithContext(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("couldn't close session, %w", err)
	}

	return nil
}

// Auth will authenticate to the target FRITZOS device using default address
// and will return either session id, either error.
func Auth(username, password string) (*Session, error) {
	return AuthWithAddress(Address, username, password)
}

func AuthWithAddress(address, username, password string) (*Session, error) {
	if err := validateAuthInput(address, username, password); err != nil {
		return nil, err
	}

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

func validateAuthInput(address, username, password string) error {
	if username == "" {
		return errors.New("please provide username for authentication")
	}

	if password == "" {
		return errors.New("please provide password for authentication")
	}

	if address == "" {
		return errors.New("please provide the address of the target device")
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
	BlockTime int
	Rights    string `xml:"Rights,omitempty"`
	Users     []user `xml:"Users>User"`
}

type user struct {
	Value string `xml:",chardata"`
	Last  int    `xml:"last,attr"`
}

func getChallengeString(address string) (string, error) {
	fullAddress := fmt.Sprintf("%s/%s", address, loginPath)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	resp, err := request.GenericGetRequestWithContext(rctx, fullAddress)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("couldn't get the challenge, %w", err)
	}

	if !request.ValidateHeader(request.HeaderXML, resp.Header) {
		return "", ErrInvalidHeaderContentType
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	var session *sessionInfo
	err = xml.Unmarshal(data, &session)
	return session.Challenge, err
}

func solveChallenge(challenge, password string) (string, error) {
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
		return "", ErrUnsupportedChallenge
	}

	return "", ErrUnsupportedChallenge
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
	fullAddress := fmt.Sprintf("%s/%s", address, loginPath)

	p := url.Values{}
	p.Set("username", username)
	p.Set("response", challenge)

	rctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	resp, err := request.GenericPostRequestWithContext(rctx, fullAddress, strings.NewReader(p.Encode()))
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something went wrong, %w", err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	var session *sessionInfo
	err = xml.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}

	if session.SID == "" || session.SID == "0000000000000000" {
		if session.BlockTime > 0 {
			return nil, &BlockTimeError{
				Duration: int(session.BlockTime),
				Message:  "Login failed. Temporary cooldown for new requests is active",
			}
		}
		return nil, ErrSessionInvalid
	}

	sid := Session(session.SID)
	return &sid, nil
}
