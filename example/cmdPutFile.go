package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/nas"
)

func NewPutFileCommand() *PutFileCommand {
	gc := &PutFileCommand{
		fs: flag.NewFlagSet("putfile", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.path, "path", "", "Provide full path to the file you want to upload")
	gc.fs.StringVar(&gc.remotePath, "remote-path", "", "Provide full path where your file will be placed on the remote target")

	return gc
}

type PutFileCommand struct {
	fs *flag.FlagSet

	username   string
	password   string
	path       string
	remotePath string
}

func (g *PutFileCommand) Name() string {
	return g.fs.Name()
}

func (g *PutFileCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *PutFileCommand) Run() error {
	return examplePutFile(g)
}

func examplePutFile(g *PutFileCommand) error {
	if g.path == "" {
		return errors.New("Please specify -path")
	}
	if g.remotePath == "" {
		return errors.New("Please specify -remote-path")
	}

	sess, err := auth.Auth(g.username, g.password)
	if err != nil {
		return err
	}
	defer sess.Close()

	log.Println("Login successful! Session ID", sess)

	// Create client
	n := nas.New(sess).WithAddress("http://fritz.box")

	// Put File
	data, err := os.Open(g.path)
	if err != nil {
		return err
	}

	r, err := n.PutFile(g.remotePath, data)
	if err != nil {
		return fmt.Errorf("failed uploading the file, %w", err)
	}

	if r.ResultCode != nas.ResultCodeOK || r.SuccessfulUploads == nas.UploadResultFail {
		return errors.New("Upload failed")
	}

	s, _ := json.Marshal(r)
	fmt.Println(string(s))
	fmt.Println("Success!")
	return nil
}
