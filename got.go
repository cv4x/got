package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/claire-west/got/commands"
	"github.com/go-git/go-git/v5"
)

const (
	status = "status"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-interrupt
		os.Exit(0)
	}()

	r, err := wdRepo()
	if err != nil {
		panic("panic: no git repository")
	}

	args := flags()
	if len(args) == 0 {
		commands.Status(r, args)
	} else {
		switch strings.ToLower(args[0]) {
		case status:
			commands.Status(r, args[1:])
		default:
			fmt.Printf("%s is not a known command\n\n", args[0])
			flag.Usage()
		}
	}
}

func wdRepo() (*git.Repository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return git.PlainOpenWithOptions(wd, &git.PlainOpenOptions{DetectDotGit: true})
}

func flags() []string {
	ex := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage of %[1]s:
	%[1]s [common_flags] <command> [command_flags]

Commands:
	status      View worktree status and add/restore files.

Common Flags:
	None yet.
`, ex)
		flag.PrintDefaults()
	}

	// TODO add flags

	flag.Parse()
	return flag.Args()
}
