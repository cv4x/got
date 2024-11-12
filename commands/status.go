package commands

import (
	"flag"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	gloss "github.com/charmbracelet/lipgloss"
	"github.com/claire-west/got/color"
	"github.com/go-git/go-git/v5"
)

var (
	headerStyle = func() gloss.Style {
		b := gloss.RoundedBorder()
		b.Right = "├"
		b.Left = "┤"
		return gloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()
	leftPad  = gloss.NewStyle().PaddingLeft(1)
	rightPad = gloss.NewStyle().PaddingRight(1)
)

type head struct {
	name     string
	ref      string
	isbranch bool
}

type category string

const (
	Staged    category = "Staged"
	Unstaged  category = "Unstaged"
	Untracked category = "Untracked"
)

type file struct {
	category category
	align    gloss.Position
	path     string
	status   git.StatusCode
	extra    string
}

type dimensions struct {
	width  int
	height int
}

type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Submit key.Binding
	Quit   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left},
		{k.Down},
		{k.Up},
		{k.Right},
		{k.Submit},
		{k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up   "),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down   "),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "unstage   "),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "add   "),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter", "y"),
		key.WithHelp("y/ent", "confirm   "),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type model struct {
	worktree *git.Worktree
	xy       dimensions
	viewport viewport.Model
	keys     keyMap
	help     help.Model
	ready    bool
	clean    bool
	head     head
	files    []file
	selected int
}

func Status(r *git.Repository, args []string) {
	args = flags(args)

	model := prepare(r)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) process(files []file) {
	for _, v := range files {
		if v.category == Staged && v.align == gloss.Left {
			opts := &git.RestoreOptions{
				Files:  []string{v.path},
				Staged: true,
			}
			if err := m.worktree.Restore(opts); err != nil {
				log.Fatalf("Error unstaging file: %v", err)
			}
		} else if (v.category == Unstaged || v.category == Untracked) && v.align == gloss.Right {
			_, err := m.worktree.Add(v.path)
			if err != nil {
				log.Fatalf("Error staging file: %v", err)
			}
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	selectedFile := m.files[m.selected]
	up := func() {
		m.selected = (m.selected + len(m.files) - 1) % len(m.files)
	}
	down := func() {
		m.selected = (m.selected + 1) % len(m.files)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			up()
		case key.Matches(msg, keys.Down):
			down()
		case key.Matches(msg, keys.Left):
			if selectedFile.align == gloss.Right {
				m.files[m.selected].align = gloss.Left
				down()
			}
		case key.Matches(msg, keys.Right):
			if selectedFile.align == gloss.Left {
				m.files[m.selected].align = gloss.Right
				down()
			}
		case key.Matches(msg, keys.Submit):
			m.process(m.files)
			return m, tea.Quit
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		newWidth := min(76, msg.Width)
		viewportWidth := newWidth - 4

		m.xy.width = newWidth
		m.xy.height = msg.Height

		headerHeight := gloss.Height(m.viewHeader())
		footerHeight := gloss.Height(m.viewFooter())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, msg.Height-verticalMarginHeight)

			m.viewport.YPosition = headerHeight
			m.viewport.Style = m.viewport.Style.Padding(0, 2)
			m.viewport.SetContent(m.viewContent())
			m.ready = true
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.help.Width = newWidth - 10
		}
	}

	// Handle keyboard and mouse events in the viewport
	// m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.xy.width < 40 || m.xy.height < 10 {
		return gloss.NewStyle().Width(m.xy.width).Height(m.xy.height).Align(gloss.Center, gloss.Center).Render("Your terminal is too small.\nResize the terminal to proceed\nor press q/esc/ctrl+c to exit.")
	}

	m.viewport.SetContent(m.viewContent())
	main := fmt.Sprintf("%s\n%s\n%s",
		m.viewHeader(),
		m.viewport.View(),
		m.viewFooter())

	return gloss.JoinHorizontal(gloss.Center, leftPad.Render(m.viewSideBorder("╭", "│", "╰")), main, rightPad.Render(m.viewSideBorder("╮", "│", "╯")))
}

func (m model) viewHeader() string {
	var titleText string
	if m.head.isbranch {
		titleText = fmt.Sprintf("On branch %s (%s)", color.Blue.Foreground(m.head.name), color.Cyan.Foreground(m.head.ref[:7]))
	} else {
		titleText = "Detached at " + color.Yellow.Foreground(m.head.ref[:7])
	}

	var subtitleText string

	title := headerStyle.Render(titleText)
	subtitle := headerStyle.Render(subtitleText)

	// if title overflows, truncate, reset color, add elipses, and re-render
	maxTitleWidth := m.viewport.Width - gloss.Width(subtitle)
	if gloss.Width(title) > maxTitleWidth {
		titleText = titleText[:max(0, maxTitleWidth-2)] + color.White.Foreground("") + "…"
		title = headerStyle.Render(titleText)
	}

	fill := strings.Repeat("─", max(0, m.viewport.Width-gloss.Width(title)-gloss.Width(subtitle)-2))
	return gloss.JoinHorizontal(gloss.Center, "─", title, fill, subtitle, "─")
}

func (m model) viewSideBorder(top, mid, bot string) string {
	borderHeight := m.xy.height - 3

	mid = mid + "\n"
	// Not enough content to scroll, therefore no scrollbar
	if m.viewport.TotalLineCount() == m.viewport.VisibleLineCount() {
		return fmt.Sprintf("%s\n%s%s\n",
			top,
			strings.Repeat(mid, max(0, borderHeight)),
			bot)
	}

	scrollBar := "█\n█\n"
	scrollBarHeight := gloss.Height(scrollBar)
	scrollPercent := m.viewport.ScrollPercent()

	var scroll int
	if scrollPercent == 0 {
		scroll = 0
	} else {
		scroll = int(math.Floor(scrollPercent*float64(borderHeight-scrollBarHeight-1))) + 1
	}

	return fmt.Sprintf("%s\n%s%s%s%s\n",
		top,
		strings.Repeat(mid, max(0, scroll)),
		scrollBar,
		strings.Repeat(mid, max(0, borderHeight-scroll-scrollBarHeight)),
		bot)
}

func (m model) getContentSeparator(s string) string {
	s = " " + s + " "

	contentWidth := m.viewport.Width - m.viewport.Style.GetHorizontalPadding()
	fill := contentWidth - 10
	leftFill := int(math.Floor(float64(fill) / 2.0))
	rightFill := contentWidth - gloss.Width(s) - leftFill - 2

	separator := "╾" + strings.Repeat("─", leftFill) + s + strings.Repeat("─", rightFill) + "╼"
	return color.MiddleGray.Foreground(separator) + "\n"
}

func (m model) viewContent() string {
	contentWidth := m.viewport.Width - m.viewport.Style.GetHorizontalPadding()

	var out string
	seenCategories := make(map[category]struct{})

	for i, v := range m.files {
		_, ok := seenCategories[v.category]
		if !ok {
			seenCategories[v.category] = struct{}{}
			out += m.getContentSeparator(string(v.category))
		}

		var line string
		text := v.path
		if gloss.Width(text) > contentWidth-8 {
			text = "…" + text[max(0, gloss.Width(text)-contentWidth+8):]
		}
		text = color.ByStatus(text, v.status)

		switch v.align {
		case gloss.Left:
			if i == m.selected {
				text = text + color.Magenta.Foreground(" ◈")
			}
			line = text
		case gloss.Right:
			if i == m.selected {
				text = color.Magenta.Foreground("◈ ") + text
			}
			line = strings.Repeat(" ", contentWidth-gloss.Width(text)) + text
		}

		out += line + "\n"
	}

	return out
}

func (m model) viewFooter() string {
	helpText := color.MiddleGray.Foreground(m.help.View(m.keys))
	footerContent := headerStyle.Render(helpText)

	footerFill := m.viewport.Width - gloss.Width(footerContent)
	leftFill := int(math.Floor(float64(footerFill) / 2.0))
	rightFill := m.viewport.Width - gloss.Width(footerContent) - leftFill

	return gloss.JoinHorizontal(gloss.Center,
		strings.Repeat("─", max(1, leftFill)),
		footerContent,
		strings.Repeat("─", max(1, rightFill)))
}

func prepare(r *git.Repository) *model {
	h, _ := r.Head()
	w, _ := r.Worktree()
	s, _ := w.Status()

	model := &model{
		worktree: w,
		clean:    s.IsClean(),
		keys:     keys,
		help:     help.New(),
		head: head{
			name:     h.Name().Short(),
			ref:      h.Hash().String(),
			isbranch: h.Name().IsBranch(),
		},
		files: make([]file, 0, len(s)),
	}

	emptyStyle := gloss.NewStyle()
	model.help.Styles.FullKey = emptyStyle
	model.help.Styles.FullDesc = emptyStyle
	model.help.Styles.ShortKey = emptyStyle
	model.help.Styles.ShortDesc = emptyStyle
	model.help.ShowAll = true

	if model.clean {
		return model
	}

	for k, v := range s {
		if s.IsUntracked(k) {
			model.files = append(model.files, file{
				category: Untracked,
				align:    gloss.Left,
				path:     k,
				status:   git.Untracked,
			})
			continue
		}
		if v.Staging != git.Unmodified {
			model.files = append(model.files, file{
				category: Staged,
				align:    gloss.Right,
				path:     k,
				status:   v.Staging,
				extra:    v.Extra,
			})
		}
		if v.Worktree != git.Unmodified {
			model.files = append(model.files, file{
				category: Unstaged,
				align:    gloss.Left,
				path:     k,
				status:   v.Worktree,
				extra:    v.Extra,
			})
		}
	}

	slices.SortFunc(model.files, func(a file, b file) int {
		if a.category != b.category {
			return strings.Compare(string(a.category), string(b.category))
		}
		return strings.Compare(a.path, b.path)
	})

	for i, v := range model.files {
		if v.category == Unstaged {
			model.selected = i
			break
		}
	}

	return model
}

func flags(args []string) []string {
	flagset := flag.NewFlagSet("got status", flag.ExitOnError)
	if err := flagset.Parse(args); err != nil {
		flagset.Usage()
	}
	return flagset.Args()
}
