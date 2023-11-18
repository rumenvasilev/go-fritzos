package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/nas"
)

func NewCreateDirCommand() *CreateDirCommand {
	gc := &CreateDirCommand{
		fs: flag.NewFlagSet("createdir", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.name, "name", "", "Provide name for the new directory you want to create")
	gc.fs.StringVar(&gc.remotePath, "remote-path", "", "Provide full path where the new directory will be created on the remote target")

	return gc
}

type CreateDirCommand struct {
	fs *flag.FlagSet

	username   string
	password   string
	name       string
	remotePath string
}

func (g *CreateDirCommand) Name() string {
	return g.fs.Name()
}

func (g *CreateDirCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *CreateDirCommand) Run() error {
	return exampleCreateDir(g)
}

func exampleCreateDir(g *CreateDirCommand) error {
	if g.name == "" {
		return errors.New("Please specify -name")
	}
	if g.remotePath == "" {
		return errors.New("Please specify -remote-path")
	}

	sess, err := auth.Auth(g.username, g.password)
	if err != nil {
		return err
	}
	// defer sess.Close()

	log.Println("Login successful! Session ID", sess)

	// Create Dir
	n := nas.New(sess).WithAddress("http://fritz.box")
	r, err := n.CreateDir(g.name, g.remotePath)
	if err != nil {
		var fse *nas.SystemError
		if errors.As(err, &fse) {
			return fse.Unwrap()
		}
	}
	s, _ := json.Marshal(r)
	fmt.Println(string(s))

	return nil
}
