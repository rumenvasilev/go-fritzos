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

func NewMoveCommand() *MoveCommand {
	gc := &MoveCommand{
		fs: flag.NewFlagSet("move", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.from, "from", "", "Provide (remote) path to the object (file or dir) you want to move")
	gc.fs.StringVar(&gc.to, "to", "", "Provide new directory path for the object (file or dir) you want to move")

	return gc
}

type MoveCommand struct {
	fs *flag.FlagSet

	username string
	password string
	from     string
	to       string
}

func (g *MoveCommand) Name() string {
	return g.fs.Name()
}

func (g *MoveCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *MoveCommand) Run() error {
	return exampleMove(g)
}

func exampleMove(g *MoveCommand) error {
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

	// Move File
	r, err := n.MoveObject(g.to, g.from)
	if err != nil {
		return fmt.Errorf("Move failed, %v", err)
	}
	s, _ := json.Marshal(r)
	fmt.Println(string(s))

	return nil
}
