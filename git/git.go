package git

import (
	"log"
	"os/exec"
	"strings"
)

func execGit(args ...string) ([]byte, error) {
	stdout, err := exec.Command("git", args...).Output()
	if err != nil {
		log.Printf("Error while executing command:\ngit %s\n\n%v\n\n", strings.Join(args, " "), err)
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

type StatusCode byte

const (
	Unmodified         StatusCode = ' '
	Untracked          StatusCode = '?'
	Modified           StatusCode = 'M'
	Added              StatusCode = 'A'
	Deleted            StatusCode = 'D'
	Renamed            StatusCode = 'R'
	Copied             StatusCode = 'C'
	UpdatedButUnmerged StatusCode = 'U'
)

func CurrentRef() (string, string) {
	stdout, err := execGit("rev-parse", "--short", "HEAD")
	ref := ""
	if err != nil {
		log.Fatalf("Failed to get output from \"git rev-parse\": %v\n", err)
	} else {
		ref = strings.Split(string(stdout), "\n")[0]
	}

	stdout, err = execGit("branch", "--show-current")
	branch := ""
	if err != nil {
		log.Fatalf("Failed to get output from \"git branch\": %v\n", err)
	} else {
		branch = strings.Split(string(stdout), "\n")[0]
	}

	return ref, branch
}

func Status() []statusline {
	stdout, err := execGit("status", "--porcelain")
	if err != nil {
		log.Fatalf("Failed to get output from \"git status\": %v\n", err)
	}
	if len(stdout) == 0 {
		return nil
	}

	lines := strings.Split(string(stdout), "\n")
	files := make([]statusline, 0, len(lines))
	for _, v := range lines {
		if v == "" {
			continue
		}

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
		log.Fatalf("Failed to stage files:%v\n", err)
	}
}

func Unstage(paths ...string) {
	args := append([]string{"restore", "--staged"}, paths...)
	_, err := execGit(args...)
	if err != nil {
		log.Fatalf("Error unstaging file: %v\n", err)
	}
}

func Restore(paths ...string) {
	args := append([]string{"restore"}, paths...)
	_, err := execGit(args...)
	if err != nil {
		log.Fatalf("Error restoring file: %v\n", err)
	}
}

func AheadBehind(branch string) (int, int) {
	// TODO: doing this synchronously impacts startup time. Consider a routine-based approach.
	return 0, 0

	// remotebytes, err := execGit("remote")
	// if err != nil {
	// 	return 0, 0
	// }
	// remote := string(remotebytes)
	// compareTo := remote + "/" + branch
	// remoteBranchExists, err := execGit("ls-remote", "--heads", remote, "refs/heads/"+branch)
	// if err != nil || len(remoteBranchExists) == 0 {
	// 	return 0, 0
	// }

	// revlistBytes, err := execGit("rev-list", "--left-right", "--count", branch+"..."+compareTo)
	// if err != nil {
	// 	return 0, 0
	// }
	// ahead, _ := strconv.Atoi(string(revlistBytes[0]))
	// behind, _ := strconv.Atoi(string(revlistBytes[len(revlistBytes)-1]))
	// return ahead, behind
}
