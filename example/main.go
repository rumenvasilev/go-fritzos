package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

type Runner interface {
	Init([]string) error
	Run() error
	Name() string
}

func root(args []string) error {
	if len(args) < 1 {
		return errors.New("You must pass a sub-command")
	}

	cmds := []Runner{
		NewListFilesCommand(),
		NewGetFileCommand(),
		NewPutFileCommand(),
		NewRenameCommand(),
		NewDeleteCommand(),
		NewMoveCommand(),
		NewCreateDirCommand(),
	}

	subcommand := os.Args[1]

	for _, cmd := range cmds {
		if cmd.Name() == subcommand {
			err := cmd.Init(os.Args[2:])
			if err != nil {
				if errors.Is(err, flag.ErrHelp) {
					return nil
				}
				return err
			}
			return cmd.Run()
		}
	}

	return fmt.Errorf("Unknown subcommand: %s", subcommand)
}

func main() {
	if err := root(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
}
