package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/nas"
)

func NewGetFileCommand() *GetFileCommand {
	gc := &GetFileCommand{
		fs: flag.NewFlagSet("getfile", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.path, "path", "", "(Optional) Provide full path to the file you want to get")

	return gc
}

type GetFileCommand struct {
	fs *flag.FlagSet

	username string
	password string
	path     string
}

func (g *GetFileCommand) Name() string {
	return g.fs.Name()
}

func (g *GetFileCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *GetFileCommand) Run() error {
	return exampleGetFile(g)
}

func exampleGetFile(g *GetFileCommand) error {
	if g.path == "" {
		return errors.New("Please specify -path and full path to a filename available in the target device.")
	}

	sess, err := auth.Auth(g.username, g.password)
	if err != nil {
		return err
	}
	defer sess.Close()

	log.Println("Login successful! Session ID", sess)

	// Create client
	n := nas.New(sess).WithAddress("http://fritz.box")

	// Get specific object from the NAS
	d, err := n.GetFile(g.path)
	if err != nil {
		return err
	}

	data, _ := io.ReadAll(d)
	defer d.Close()

	f := strings.Split(g.path, "/")
	err = os.WriteFile(f[len(f)-1], data, 0755)
	if err != nil {
		return fmt.Errorf("failed writing the resulting file to disk, %w", err)
	}

	return nil
}
