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

func NewRenameCommand() *RenameCommand {
	gc := &RenameCommand{
		fs: flag.NewFlagSet("rename", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.from, "from", "", "Provide (remote) path to the object (file or dir) you want to rename")
	gc.fs.StringVar(&gc.to, "to", "", "Provide new name for the object (file or dir) you want to rename")

	return gc
}

type RenameCommand struct {
	fs *flag.FlagSet

	username string
	password string
	from     string
	to       string
}

func (g *RenameCommand) Name() string {
	return g.fs.Name()
}

func (g *RenameCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *RenameCommand) Run() error {
	return exampleRename(g)
}

func exampleRename(g *RenameCommand) error {
	if g.from == "" {
		return errors.New("Please specify -from")
	}
	if g.to == "" {
		return errors.New("Please specify -to")
	}

	sess, err := auth.Auth(g.username, g.password)
	if err != nil {
		return err
	}
	defer sess.Close()

	log.Println("Login successful! Session ID", sess)

	// Create client
	n := nas.New(sess).WithAddress("http://fritz.box")

	// Rename File
	params := []*nas.RenameInput{{From: g.from, To: g.to}}
	r, err := n.RenameObject(params)
	if err != nil {
		return fmt.Errorf("Rename failed, %v", err)
	}
	s, _ := json.Marshal(r)
	fmt.Println(string(s))

	return nil
}
