package editor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	tea "github.com/charmbracelet/bubbletea"
)

// OpenMsg is sent when the editor process exits.
type OpenMsg struct {
	Code string
	Err  error
}

// Open opens the given code in an external editor and returns the edited content.
func Open(editor, lang, code, titleSlug string) tea.Cmd {
	path := TempFilePath(lang, titleSlug)

	// Prepare the file before launching the editor
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return func() tea.Msg { return OpenMsg{Err: fmt.Errorf("create temp dir: %w", err)} }
	}
	if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
		return func() tea.Msg { return OpenMsg{Err: fmt.Errorf("write temp file: %w", err)} }
	}

	cmd := editorCmd(editor, path)
	return tea.Exec(&editorExecCmd{cmd: cmd}, func(err error) tea.Msg {
		if err != nil {
			return OpenMsg{Err: err}
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return OpenMsg{Err: readErr}
		}
		return OpenMsg{Code: string(data)}
	})
}

// editorExecCmd wraps exec.Cmd to run the editor in its own process group.
// After the editor exits, it restores the parent as the foreground process group
// and kills any lingering child processes (e.g., gopls spawned by nvim).
type editorExecCmd struct {
	cmd *exec.Cmd
}

func (e *editorExecCmd) Run() error {
	err := e.cmd.Run()

	// Temporarily ignore SIGTTOU so we can call tcsetpgrp from a background group.
	signal.Ignore(syscall.SIGTTOU)
	// Restore our process group as the foreground process group.
	pgrp := syscall.Getpgrp()
	tcsetpgrp(os.Stdin.Fd(), pgrp)
	signal.Reset(syscall.SIGTTOU)

	// Kill any lingering child processes in the editor's process group (e.g., gopls).
	if e.cmd.Process != nil {
		_ = syscall.Kill(-e.cmd.Process.Pid, syscall.SIGTERM)
	}

	return err
}

func (e *editorExecCmd) SetStdin(r io.Reader)  { e.cmd.Stdin = r }
func (e *editorExecCmd) SetStdout(w io.Writer) { e.cmd.Stdout = w }
func (e *editorExecCmd) SetStderr(w io.Writer) { e.cmd.Stderr = w }

func tcsetpgrp(fd uintptr, pgrp int) {
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCSPGRP, uintptr(unsafe.Pointer(&pgrp)))
}

func editorCmd(editor, path string) *exec.Cmd {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		parts = []string{"vim"}
	}

	args := append(parts[1:], path)
	cmd := exec.Command(parts[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: true,
		Ctty:       int(os.Stdin.Fd()),
	}
	return cmd
}

// TempFilePath returns the path to the temporary file for a given language and slug.
func TempFilePath(lang, titleSlug string) string {
	dir := filepath.Join(os.TempDir(), "leet-tui")
	ext := LangExtension(lang)
	return filepath.Join(dir, fmt.Sprintf("%s.%s", titleSlug, ext))
}

// LangExtension returns the file extension for a given language.
func LangExtension(lang string) string {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return "go"
	case "python", "python3":
		return "py"
	case "cpp", "c++":
		return "cpp"
	case "java":
		return "java"
	case "javascript":
		return "js"
	case "typescript":
		return "ts"
	case "rust":
		return "rs"
	case "c":
		return "c"
	case "ruby":
		return "rb"
	case "swift":
		return "swift"
	case "kotlin":
		return "kt"
	default:
		return "txt"
	}
}
