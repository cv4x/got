package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cv4x/got/commands"
	"github.com/cv4x/got/git"
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

	state := git.CurrentRef()
	if state.Ref == "" {
		panic("panic: no git repository")
	}

	// Output the colorized status before exiting
	printstatus := func() {
		stdout, err := exec.Command("git", "-c", "color.ui=always", "status").Output()
		if err == nil {
			fmt.Println(string(stdout))
		}
	}

	args := flags()
	if len(args) == 0 {
		commands.Status(state, args)
		printstatus()
	} else {
		switch strings.ToLower(args[0]) {
		case status:
			commands.Status(state, args[1:])
			printstatus()
		default:
			fmt.Printf("%s is not a known command\n\n", args[0])
			flag.Usage()
		}
	}
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
