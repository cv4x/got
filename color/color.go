package color

import (
	gloss "github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
)

func fg(c gloss.TerminalColor) func(...string) string {
	return gloss.NewStyle().Foreground(c).Render
}

func bg(c gloss.TerminalColor) func(...string) string {
	return gloss.NewStyle().Background(c).Render
}

var ansi16 = struct {
	black         gloss.TerminalColor
	red           gloss.TerminalColor
	green         gloss.TerminalColor
	yellow        gloss.TerminalColor
	blue          gloss.TerminalColor
	magenta       gloss.TerminalColor
	cyan          gloss.TerminalColor
	white         gloss.TerminalColor
	brightblack   gloss.TerminalColor
	brightred     gloss.TerminalColor
	brightgreen   gloss.TerminalColor
	brightyellow  gloss.TerminalColor
	brightblue    gloss.TerminalColor
	brightmagenta gloss.TerminalColor
	brightcyan    gloss.TerminalColor
	brightwhite   gloss.TerminalColor
}{
	black:         gloss.Color("0"),
	red:           gloss.Color("1"),
	green:         gloss.Color("2"),
	yellow:        gloss.Color("3"),
	blue:          gloss.Color("4"),
	magenta:       gloss.Color("5"),
	cyan:          gloss.Color("6"),
	white:         gloss.Color("7"),
	brightblack:   gloss.Color("8"),
	brightred:     gloss.Color("9"),
	brightgreen:   gloss.Color("10"),
	brightyellow:  gloss.Color("11"),
	brightblue:    gloss.Color("12"),
	brightmagenta: gloss.Color("13"),
	brightcyan:    gloss.Color("14"),
	brightwhite:   gloss.Color("15"),
}

var truecolor = struct {
	middlegray gloss.TerminalColor
}{
	middlegray: gloss.Color("#808080"),
}

type renderer func(...string) string

type renderers struct {
	Foreground renderer
	Background renderer
}

var (
	Black         = renderers{fg(ansi16.black), bg(ansi16.black)}
	Red           = renderers{fg(ansi16.red), bg(ansi16.red)}
	Green         = renderers{fg(ansi16.green), bg(ansi16.green)}
	Yellow        = renderers{fg(ansi16.yellow), bg(ansi16.yellow)}
	Blue          = renderers{fg(ansi16.blue), bg(ansi16.blue)}
	Magenta       = renderers{fg(ansi16.magenta), bg(ansi16.magenta)}
	Cyan          = renderers{fg(ansi16.cyan), bg(ansi16.cyan)}
	White         = renderers{fg(ansi16.white), bg(ansi16.white)}
	BrightBlack   = renderers{fg(ansi16.brightblack), bg(ansi16.brightblack)}
	BrightRed     = renderers{fg(ansi16.red), bg(ansi16.brightred)}
	BrightGreen   = renderers{fg(ansi16.green), bg(ansi16.brightgreen)}
	BrightYellow  = renderers{fg(ansi16.yellow), bg(ansi16.brightyellow)}
	BrightBlue    = renderers{fg(ansi16.blue), bg(ansi16.brightblue)}
	BrightMagenta = renderers{fg(ansi16.magenta), bg(ansi16.brightmagenta)}
	BrightCyan    = renderers{fg(ansi16.cyan), bg(ansi16.brightcyan)}
	BrightWhite   = renderers{fg(ansi16.white), bg(ansi16.brightwhite)}
	MiddleGray    = renderers{fg(truecolor.middlegray), bg(truecolor.middlegray)}
)

var status = map[bool]map[git.StatusCode]renderers{
	true: {
		git.Added:              Green,
		git.Deleted:            Red,
		git.Modified:           Green,
		git.Renamed:            Yellow,
		git.UpdatedButUnmerged: Green,
	},
	false: {
		git.Added:              Red,
		git.Deleted:            Red,
		git.Modified:           Red,
		git.Renamed:            Yellow,
		git.UpdatedButUnmerged: Yellow,
		git.Untracked:          Red,
	},
}

func ByStatus(s string, code git.StatusCode, staged bool) string {
	renderer, ok := status[staged][code]
	if !ok {
		return s
	}
	return renderer.Foreground(s)
}
