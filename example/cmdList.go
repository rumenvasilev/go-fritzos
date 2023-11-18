package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/rumenvasilev/go-fritzos/auth"
	"github.com/rumenvasilev/go-fritzos/nas"
)

func NewListFilesCommand() *ListFilesCommand {
	gc := &ListFilesCommand{
		fs: flag.NewFlagSet("list", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.username, "username", "", "Provide username for authentication")
	gc.fs.StringVar(&gc.password, "password", "", "Provide password for authentication")
	gc.fs.StringVar(&gc.path, "path", "", "(Optional) Provide path you want to list")

	return gc
}

type ListFilesCommand struct {
	fs *flag.FlagSet

	username string
	password string
	path     string
}

func (g *ListFilesCommand) Name() string {
	return g.fs.Name()
}

func (g *ListFilesCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *ListFilesCommand) Run() error {
	return exampleList(g)
}

func exampleList(g *ListFilesCommand) error {
	sess, err := auth.Auth(g.username, g.password)
	if err != nil {
		return err
	}
	defer sess.Close()

	log.Println("Login successful! Session ID", sess)

	// Create client
	n := nas.New(sess).WithAddress("http://fritz.box")

	p := "/"
	if g.path != "" {
		p = g.path
	}
	res, err := n.ListDirectory(p)
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(res, "", "  ")
	fmt.Println(string(data))

	// List files at specific directory level
	// params := make(map[string]string)
	// params["path"] = "/Bilder"
	// params["limit"] = "100"
	// res, err := ListDirectoryWithParams(Session, params)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// dirs := res.Directories
	// files := res.Files

	// if len(dirs) > 0 {
	// 	fmt.Println("Listing directories:")
	// }
	// for _, v := range dirs {
	// 	fmt.Println(v.Filename)
	// }

	// if len(files) > 0 {
	// 	fmt.Println("Listing files:")
	// }
	// for _, v := range files {
	// 	fmt.Println(v.Filename)
	// 	fmt.Println(v.Size)
	// 	fmt.Println(v.Path)
	// 	fmt.Println(v.Timestamp)
	// }

	return nil
}
