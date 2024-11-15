# got

This project exists because I got tired of using the mouse to copy files from the `git status` output into my `git add` commands. The primary use case for got is to easily stage/unstage files without needing to move your hands off the keyboard.

The TUI is created using [Bubble Tea](https://github.com/charmbracelet/bubbletea). [go-git](https://github.com/go-git/go-git) was originally used to access and modify the repository, however due to functional limitations and performance issues new features will be based on the Git CLI instead, and the go-git dependency will eventually be removed from the project.
