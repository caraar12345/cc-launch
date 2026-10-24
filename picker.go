package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const maxRows = 8

// filterDirs keeps configured directories matching every whitespace-separated
// term of query (case-insensitive substring); basename matches rank first.
func filterDirs(dirs []string, query string) []string {
	terms := strings.Fields(strings.ToLower(query))
	type scored struct {
		path  string
		score int
	}
	var out []scored
	for _, d := range dirs {
		full := strings.ToLower(d)
		base := strings.ToLower(filepath.Base(d))
		score := 0
		ok := true
		for _, t := range terms {
			switch {
			case strings.HasPrefix(base, t):
				score += 0
			case strings.Contains(base, t):
				score += 1
			case strings.Contains(full, t):
				score += 2
			default:
				ok = false
			}
		}
		if ok {
			out = append(out, scored{d, score})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score < out[j].score })
	res := make([]string, len(out))
	for i, s := range out {
		res[i] = s.path
	}
	return res
}

// browseDirs handles the "/" mode. rest is the text after the leading slash,
// e.g. "cc-launch" or "work/api". It lists subdirectories of prefix[/work]
// whose name matches the final fragment: exact, then prefix, then substring.
func browseDirs(prefix, rest string) []string {
	if filepath.Separator == '\\' {
		rest = strings.ReplaceAll(rest, `\`, "/")
	}
	dir, frag := "", rest
	if i := strings.LastIndex(rest, "/"); i >= 0 {
		dir, frag = rest[:i], rest[i+1:]
	}
	base := filepath.Join(prefix, dir)
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	lf := strings.ToLower(frag)
	type scored struct {
		path  string
		score int
	}
	var out []scored
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(frag, ".") {
			continue
		}
		p := filepath.Join(base, name)
		if !isDir(p) { // also follows symlinks
			continue
		}
		ln := strings.ToLower(name)
		switch {
		case frag == "":
			out = append(out, scored{p, 2})
		case ln == lf:
			out = append(out, scored{p, 0})
		case strings.HasPrefix(ln, lf):
			out = append(out, scored{p, 1})
		case strings.Contains(ln, lf):
			out = append(out, scored{p, 2})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score < out[j].score })
	res := make([]string, len(out))
	for i, s := range out {
		res[i] = s.path
	}
	return res
}

type pickerModel struct {
	cfg    *Config
	prefix string // expanded default prefix
	input  textinput.Model
	spin   spinner.Model

	items  []string
	cursor int // len(items) is the trailing "other…" row

	direct bool // path given on the command line: no list, just upgrade and launch
	other  bool // free-form path entry mode
	path   textinput.Model

	upgrading bool
	upStatus  string // final upgrade line once finished
	pending   string // choice made while the upgrade is still running

	chosen    string
	cancelled bool
	notice    string
}

func newPickerModel(cfg *Config) pickerModel {
	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.Placeholder = "filter, or / to browse"
	ti.Focus()
	pi := textinput.New()
	pi.Prompt = "path: "
	pi.Placeholder = "~/some/dir"
	sp := spinner.New(spinner.WithSpinner(spinner.MiniDot))
	m := pickerModel{
		cfg:       cfg,
		prefix:    expandHome(cfg.DefaultPrefix),
		input:     ti,
		path:      pi,
		spin:      sp,
		upgrading: true,
	}
	m.refresh()
	return m
}

func (m pickerModel) browsing() bool { return strings.HasPrefix(m.input.Value(), "/") }

func (m *pickerModel) refresh() {
	v := m.input.Value()
	if m.browsing() {
		m.items = browseDirs(m.prefix, v[1:])
	} else {
		dirs := make([]string, len(m.cfg.Directories))
		for i, d := range m.cfg.Directories {
			dirs[i] = expandHome(d)
		}
		m.items = filterDirs(dirs, v)
	}
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items)
	}
}

func (m pickerModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spin.Tick, upgradeCmd())
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case upgradeDoneMsg:
		m.upgrading = false
		switch {
		case msg.timedOut:
			m.upStatus = warnStyle.Render("! update timed out")
		case msg.err != nil:
			m.upStatus = warnStyle.Render("! update failed")
		case msg.before != "" && msg.after != "" && msg.before != msg.after:
			m.upStatus = okStyle.Render("✓ updated to " + msg.after)
		default:
			m.upStatus = okStyle.Render("✓ up to date")
		}
		if m.pending != "" {
			m.chosen = m.pending
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		if !m.upgrading {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.pending != "" { // waiting on upgrade; only allow bailing out
			if k := msg.String(); k == "ctrl+c" || k == "esc" {
				m.cancelled = true
				return m, tea.Quit
			}
			return m, nil
		}
		if m.other {
			return m.updateOther(msg)
		}
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+n":
			if m.cursor < len(m.items) {
				m.cursor++
			}
			return m, nil
		case "tab":
			if m.browsing() && m.cursor < len(m.items) {
				rel, err := filepath.Rel(m.prefix, m.items[m.cursor])
				if err == nil {
					m.input.SetValue("/" + filepath.ToSlash(rel) + "/")
					m.input.CursorEnd()
					m.cursor = 0
					m.refresh()
				}
			}
			return m, nil
		case "enter":
			return m.choose()
		}
	}

	var cmd tea.Cmd
	before := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != before {
		m.notice = ""
		m.cursor = 0
		m.refresh()
	}
	return m, cmd
}

func (m pickerModel) choose() (tea.Model, tea.Cmd) {
	if m.cursor == len(m.items) { // "other…"
		m.other = true
		m.notice = ""
		m.path.SetValue("")
		return m, m.path.Focus()
	}
	return m.launchDir(m.items[m.cursor])
}

// launchDir quits with dir chosen, or parks it until the upgrade finishes.
func (m pickerModel) launchDir(dir string) (tea.Model, tea.Cmd) {
	if m.upgrading {
		m.pending = dir
		return m, nil
	}
	m.chosen = dir
	return m, tea.Quit
}

func (m pickerModel) updateOther(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	case "esc":
		m.other = false
		m.notice = ""
		m.path.Blur()
		return m, nil
	case "enter":
		v := strings.TrimSpace(m.path.Value())
		dir, err := filepath.Abs(expandHome(v))
		if v == "" || err != nil || !isDir(dir) {
			m.notice = "not a directory"
			return m, nil
		}
		return m.launchDir(dir)
	}
	var cmd tea.Cmd
	m.path, cmd = m.path.Update(msg)
	m.notice = ""
	return m, cmd
}

func (m pickerModel) View() string {
	if m.chosen != "" || m.cancelled {
		return ""
	}
	var b strings.Builder

	b.WriteString(titleStyle.Render("cc-launch") + dimStyle.Render(" | "))
	if m.upgrading {
		b.WriteString(m.spin.View() + dimStyle.Render(" upgrading"))
	} else {
		b.WriteString(m.upStatus)
	}
	if m.direct {
		b.WriteString("\n" + dimStyle.Render("launching "+tildify(m.pending)) + "\n")
		return b.String()
	}
	if m.other {
		b.WriteString("\n" + m.path.View() + "\n")
		if m.notice != "" {
			b.WriteString(errStyle.Render(m.notice) + "\n")
		}
		b.WriteString(dimStyle.Render("enter launch · esc back") + "\n")
		return b.String()
	}
	b.WriteString("\n" + m.input.View() + "\n")

	rows := make([]string, 0, len(m.items)+1)
	for _, it := range m.items {
		rows = append(rows, tildify(it))
	}
	rows = append(rows, "other…")
	start := 0
	if m.cursor >= maxRows {
		start = m.cursor - maxRows + 1
	}
	end := min(len(rows), start+maxRows)
	for i := start; i < end; i++ {
		switch {
		case i == m.cursor:
			b.WriteString(selStyle.Render("▸ "+rows[i]) + "\n")
		case i == len(m.items):
			b.WriteString(dimStyle.Render("  "+rows[i]) + "\n")
		default:
			b.WriteString("  " + rows[i] + "\n")
		}
	}
	if len(m.items) == 0 {
		switch {
		case m.input.Value() != "":
			b.WriteString(dimStyle.Render("  no matches") + "\n")
		case len(m.cfg.Directories) == 0:
			b.WriteString(dimStyle.Render("  no saved directories; type / to browse "+tildify(m.prefix)) + "\n")
		}
	} else if len(rows) > end {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  … %d more", len(rows)-end)) + "\n")
	}

	if m.notice != "" {
		b.WriteString(errStyle.Render(m.notice) + "\n")
	}
	b.WriteString(dimStyle.Render("↑/↓ select · tab complete · enter launch · esc quit") + "\n")
	return b.String()
}

// runPicker returns the chosen directory, or "" if cancelled.
func runPicker(cfg *Config) (string, error) {
	final, err := tea.NewProgram(newPickerModel(cfg)).Run()
	if err != nil {
		return "", err
	}
	return final.(pickerModel).chosen, nil
}

// runDirect upgrades, then returns dir without showing the picker.
func runDirect(dir string) (string, error) {
	m := newPickerModel(&Config{})
	m.direct, m.pending = true, dir
	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	return final.(pickerModel).chosen, nil
}
