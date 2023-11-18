package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/nas"
)

func NewDeleteCommand() *DeleteCommand {
	gc := &DeleteCommand{
		fs: flag.NewFlagSet("delete", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.remotePath, "remote-path", "", "Provide path (remote) to the file you want to delete")

	return gc
}

type DeleteCommand struct {
	fs *flag.FlagSet

	username   string
	password   string
	remotePath string
}

func (g *DeleteCommand) Name() string {
	return g.fs.Name()
}

func (g *DeleteCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *DeleteCommand) Run() error {
	return exampleDelete(g)
}

func exampleDelete(g *DeleteCommand) error {
	if g.remotePath == "" {
		return errors.New("Please specify -path")
	}

	sess, err := auth.Auth(g.username, g.password)
	if err != nil {
		return err
	}
	defer sess.Close()

	log.Println("Login successful! Session ID", sess)

	// Create client
	n := nas.New(sess).WithAddress("http://fritz.box")

	// Delete File
	r, err := n.DeleteObject(g.remotePath)
	if err != nil {
		return fmt.Errorf("Delete failed, %v", err)
	}
	fmt.Println(r)

	return nil
}
