// Package tui provides an interactive terminal form that builds the fscanx
// command-line arguments, replacing hand-typed flags with a guided UI.
//
// It does NOT reimplement the scanner: RunTUI collects user input and returns
// a constructed os.Args slice (e.g. []string{"fscanx","-h","1.1.1.1","-p","80"})
// that main.go feeds straight into the existing common.Flag + Plugins.Scan path.
// This keeps the scan engine untouched and preserves the -std pipe and Win7
// (non-TTY) fallback behavior.
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// field describes one form row.
type field struct {
	key     string // flag key, e.g. "h", "p", "socks5"
	label   string // human label shown in the UI
	value   string // current value
	help    string // one-line hint
	boolKey bool   // if true this is a boolean toggle (no text input)
}

// model is the bubbletea model for the form.
type model struct {
	inputs []textinput.Model
	fields []field
	focus  int
	submit bool
	err    error
}

func newModel() model {
	// Order matters: rendered top to bottom.
	defs := []field{
		{key: "h", label: "目标网段/主机", help: "如 192.168.1.0/24 或 192.168.1.1"},
		{key: "hf", label: "目标文件(-hf)", help: "每行一个 ip/url/域名/cidr"},
		{key: "p", label: "端口(-p)", help: "默认全端口,可指定 80,443,3306"},
		{key: "t", label: "端口扫描线程(-t)", help: "默认 512"},
		{key: "socks5", label: "SOCKS5 代理", help: "如 socks5://127.0.0.1:1080"},
		{key: "uf", label: "URL 文件(-uf)", help: "批量 url 扫描"},
		{key: "auto", label: "智能大网段探测(-auto)", boolKey: true},
		{key: "nmap", label: "协议指纹识别(-nmap)", boolKey: true},
		{key: "poc", label: "PoC 扫描(-poc)", boolKey: true},
		{key: "br", label: "弱口令爆破(-br)", boolKey: true},
		{key: "pd", label: "解析域名C段(-pd)", boolKey: true},
		{key: "silent", label: "静默扫描(-silent)", boolKey: true},
	}

	m := model{fields: defs}
	m.inputs = make([]textinput.Model, 0, len(defs))
	for _, f := range defs {
		ti := textinput.New()
		ti.Placeholder = f.help
		ti.CharLimit = 256
		if f.boolKey {
			// bool fields use y/n style; default empty = off
			ti.Placeholder = "y/n (默认 n)"
			ti.Width = 3
		}
		ti.Prompt = ""
		m.inputs = append(m.inputs, ti)
	}
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
		m.inputs[0].Prompt = "> "
	}
	return m
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.err = errCancelled
			return m, tea.Quit
		case "enter":
			if m.focus == len(m.inputs)-1 {
				m.submit = true
				return m, tea.Quit
			}
			fallthrough
		case "tab", "down", "shift+tab", "up":
			// move focus
			if msg.String() == "enter" {
				// handled above for last field
			}
			dir := 1
			if msg.String() == "shift+tab" || msg.String() == "up" {
				dir = -1
			}
			m.inputs[m.focus].Blur()
			m.inputs[m.focus].Prompt = ""
			m.focus += dir
			if m.focus < 0 {
				m.focus = len(m.inputs) - 1
			}
			if m.focus >= len(m.inputs) {
				m.focus = 0
			}
			m.inputs[m.focus].Focus()
			m.inputs[m.focus].Prompt = "> "
			return m, textinput.Blink
		}
	}

	// route text input to the focused field
	cmd := m.inputs[m.focus].Focus()
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)
	var inputCmd tea.Cmd
	m.inputs[m.focus], inputCmd = m.inputs[m.focus].Update(msg)
	cmds = append(cmds, inputCmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.submit || m.err != nil {
		return ""
	}
	s := "fscanx · 交互式参数\n\n"
	for i, f := range m.fields {
		cursor := "  "
		if i == m.focus {
			cursor = "> "
		}
		val := m.inputs[i].View()
		s += cursor + f.label + ":\n  " + val + "\n\n"
	}
	s += "↑/↓ 切换 · 输入值 · Enter 在末行提交 · Esc 取消\n"
	return s
}

// BuildArgs turns the collected form values into an os.Args-style slice
// (without the program name). Returns nil if cancelled.
func (m model) BuildArgs() []string {
	var args []string
	for i, f := range m.fields {
		v := m.inputs[i].Value()
		if f.boolKey {
			if v == "y" || v == "Y" || v == "yes" {
				args = append(args, "-"+f.key)
			}
			continue
		}
		if v != "" {
			args = append(args, "-"+f.key, v)
		}
	}
	return args
}
