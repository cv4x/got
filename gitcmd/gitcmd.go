package gitcmd

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func execGit(args ...string) ([]byte, error) {
	stdout, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, err
	}
	if len(stdout) > 0 {
		// trim trailing newline
		stdout = stdout[:len(stdout)-1]
	}
	return stdout, nil
}

type statusline struct {
	Path    string
	Staged  byte
	Tracked byte
	Extra   string
}

func Status() []statusline {
	stdout, err := execGit("status", "--porcelain")
	if err != nil {
		log.Fatalf("Failed to get output from \"git status\": %v", err)
	}
	lines := strings.Split(string(stdout), "\n")
	files := make([]statusline, 0, len(lines))
	for _, v := range lines {
		file := statusline{
			Path:    v[3:],
			Staged:  v[0],
			Tracked: v[1],
		}

		// if renamed (R) set "extra" to the previous name and "path" to the new name
		if file.Staged == 82 {
			parts := strings.Split(file.Path, " -> ")
			file.Path = parts[1]
			file.Extra = parts[0]
		}

		files = append(files, file)
	}
	return files
}

func Add(paths ...string) {
	args := append([]string{"add"}, paths...)
	_, err := execGit(args...)
	if err != nil {
		log.Fatalf("Error staging file: %v", err)
	}
}

func Restore(paths ...string) {
	args := append([]string{"restore", "--staged"}, paths...)
	_, err := execGit(args...)
	if err != nil {
		log.Fatalf("Error unstaging file: %v", err)
	}
}

func AheadBehind(branch string) (int, int) {
	remotebytes, err := execGit("remote")
	if err != nil {
		return 0, 0
	}
	remote := string(remotebytes)
	revlistBytes, err := execGit("rev-list", "--left-right", "--count", branch+"..."+remote+"/"+branch)
	if err != nil {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(string(revlistBytes[0]))
	behind, _ := strconv.Atoi(string(revlistBytes[len(revlistBytes)-1]))
	return ahead, behind
}
