package ui

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/version"
)

type openMSXBridgeSession struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	window   fyne.Window
	output   *widget.Entry
	command  *widget.Entry
	sendBtn  *widget.Button
	mu       sync.Mutex
	closed   bool
	outputSB strings.Builder
}

const (
	openMSXBootDelay    = 3 * time.Second
	openMSXBootInterval = 500 * time.Millisecond
)

func openMSXBootCommands() []string {
	return []string{"set renderer sdlgl-pp", "set power on"}
}

func buildOpenMSXBridgeCommand(resolvedPath, machine string, extensions []string) (name string, args []string, workDir string) {
	extra := make([]string, 0, 2+2*len(extensions)+2)
	if machine = normalizeOpenMSXResourceName(machine); machine != "" {
		extra = append(extra, "-machine", machine)
	}
	for _, extension := range extensions {
		if extension = normalizeOpenMSXResourceName(extension); extension != "" {
			extra = append(extra, "-ext", extension)
		}
	}
	extra = append(extra, "-control", "stdio")
	return buildConfiguredToolCommand(settingOpenMSXExeKey, resolvedPath, extra)
}

func (e *editorUI) loadOpenMSXBridgeLaunchOptions() (machine string, extensions []string) {
	if e == nil || e.store == nil {
		return "", nil
	}

	machine, _ = e.store.GetSetting(context.Background(), settingOpenMSXMachineKey)
	machine = normalizeOpenMSXResourceName(machine)

	for _, key := range openMSXExtensionSettingKeys() {
		raw, _ := e.store.GetSetting(context.Background(), key)
		if name := normalizeOpenMSXResourceName(raw); name != "" {
			extensions = append(extensions, name)
		}
	}

	return machine, extensions
}

func buildOpenMSXCommandXML(command string) string {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("<command>")
	_ = xml.EscapeText(&b, []byte(trimmed))
	b.WriteString("</command>")
	return b.String()
}

func (s *openMSXBridgeSession) isAlive() bool {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return false
	}
	if s.cmd.ProcessState != nil && s.cmd.ProcessState.Exited() {
		return false
	}
	return true
}

func (s *openMSXBridgeSession) focusWindow() {
	if s == nil || s.window == nil {
		return
	}
	fyne.Do(func() {
		s.window.Show()
	})
}

func (s *openMSXBridgeSession) setInputEnabled(enabled bool) {
	if s == nil {
		return
	}
	fyne.Do(func() {
		if s.command != nil {
			if enabled {
				s.command.Enable()
			} else {
				s.command.Disable()
			}
		}
		if s.sendBtn != nil {
			if enabled {
				s.sendBtn.Enable()
			} else {
				s.sendBtn.Disable()
			}
		}
	})
}

func (s *openMSXBridgeSession) appendOutputLine(line string) {
	if s == nil {
		return
	}
	line = strings.TrimRight(line, "\r\n")
	s.mu.Lock()
	if s.outputSB.Len() > 0 {
		s.outputSB.WriteByte('\n')
	}
	s.outputSB.WriteString(line)
	text := s.outputSB.String()
	s.mu.Unlock()

	fyne.Do(func() {
		if s.output != nil {
			s.output.SetText(text)
		}
	})
}

func (s *openMSXBridgeSession) sendCommand(raw string) error {
	xmlCommand := buildOpenMSXCommandXML(raw)
	if xmlCommand == "" {
		return nil
	}

	s.mu.Lock()
	if s.closed || s.stdin == nil {
		s.mu.Unlock()
		return fmt.Errorf("openMSX bridge is closed")
	}
	stdin := s.stdin
	s.mu.Unlock()

	if _, err := io.WriteString(stdin, xmlCommand+"\n"); err != nil {
		return err
	}
	s.appendOutputLine("> " + xmlCommand)
	return nil
}

func (s *openMSXBridgeSession) close() {
	if s == nil {
		return
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	stdin := s.stdin
	s.stdin = nil
	cmd := s.cmd
	win := s.window
	s.mu.Unlock()

	if stdin != nil {
		_ = stdin.Close()
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	if win != nil {
		fyne.Do(func() {
			win.SetCloseIntercept(nil)
			win.Close()
		})
	}
}

func (s *openMSXBridgeSession) sendBootSequence(commands []string, interval time.Duration) {
	for _, command := range commands {
		if !s.isAlive() {
			return
		}
		if err := s.sendCommand(command); err != nil {
			s.appendOutputLine("[bridge] bootstrap failed: " + err.Error())
			return
		}
		if interval > 0 {
			time.Sleep(interval)
		}
	}
}

func (s *openMSXBridgeSession) scheduleBootSequence(delay time.Duration, commands []string, interval time.Duration) {
	if s == nil || len(commands) == 0 {
		return
	}
	go func() {
		if delay > 0 {
			time.Sleep(delay)
		}
		s.sendBootSequence(commands, interval)
	}()
}

func scanOpenMSXOutput(r io.Reader, prefix string, onLine func(string)) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		onLine(prefix + scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		onLine(prefix + "[reader error] " + err.Error())
	}
}

func (e *editorUI) clearOpenMSXBridge(target *openMSXBridgeSession) {
	if e.openMSXBridge == target {
		e.openMSXBridge = nil
	}
}

func (e *editorUI) closeOpenMSXBridge() {
	if e.openMSXBridge == nil {
		return
	}
	active := e.openMSXBridge
	e.openMSXBridge = nil
	active.close()
}

func (e *editorUI) openOpenMSXBridge() {
	if e.openMSXBridge != nil && e.openMSXBridge.isAlive() {
		e.openMSXBridge.focusWindow()
		if e.status != nil {
			e.status.SetText("openMSX bridge: already active")
		}
		return
	}
	if e.openMSXBridge != nil {
		e.closeOpenMSXBridge()
	}

	resolved, err := e.resolveConfiguredToolPathFromSettings(settingOpenMSXExeKey)
	if err != nil {
		msg := err.Error() + "\n\nUse Utilities -> Configure... to set the tool path."
		dialog.ShowInformation("openMSX", msg, e.window)
		if e.status != nil {
			e.status.SetText("openMSX: not configured")
		}
		return
	}

	machine, extensions := e.loadOpenMSXBridgeLaunchOptions()
	name, args, workDir := buildOpenMSXBridgeCommand(resolved, machine, extensions)
	cmd := exec.Command(name, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		dialog.ShowError(err, e.window)
		if e.status != nil {
			e.status.SetText("openMSX bridge: failed to create stdin pipe")
		}
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		dialog.ShowError(err, e.window)
		if e.status != nil {
			e.status.SetText("openMSX bridge: failed to create stdout pipe")
		}
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		dialog.ShowError(err, e.window)
		if e.status != nil {
			e.status.SetText("openMSX bridge: failed to create stderr pipe")
		}
		return
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		dialog.ShowError(err, e.window)
		if e.status != nil {
			e.status.SetText("openMSX bridge: failed to start")
		}
		return
	}

	bridgeWindow := e.fyneApp.NewWindow(version.Full() + " - openMSX Bridge")
	bridgeWindow.Resize(fyne.NewSize(900, 520))

	output := widget.NewMultiLineEntry()
	output.Wrapping = fyne.TextWrapWord
	output.SetMinRowsVisible(20)
	output.Disable()
	output.SetPlaceHolder("openMSX output will appear here")

	command := widget.NewEntry()
	command.SetPlaceHolder("Type MSX command (example: set pause on)")

	session := &openMSXBridgeSession{
		cmd:     cmd,
		stdin:   stdin,
		window:  bridgeWindow,
		output:  output,
		command: command,
	}
	e.openMSXBridge = session

	send := widget.NewButton("SEND", func() {
		if err := session.sendCommand(command.Text); err != nil {
			session.appendOutputLine("[error] " + err.Error())
			return
		}
		command.SetText("")
	})
	session.sendBtn = send
	command.OnSubmitted = func(string) {
		send.OnTapped()
	}

	header := widget.NewLabel("openMSX Remote Bridge (-control stdio)")
	content := container.NewBorder(
		header,
		nil,
		nil,
		nil,
		container.NewVBox(
			container.NewBorder(nil, nil, nil, send, command),
			output,
		),
	)
	bridgeWindow.SetContent(content)
	bridgeWindow.SetCloseIntercept(func() {
		e.closeOpenMSXBridge()
	})
	bridgeWindow.Show()

	session.appendOutputLine("[bridge] openMSX process started")
	session.appendOutputLine("[bridge] waiting boot sequence")
	session.scheduleBootSequence(openMSXBootDelay, openMSXBootCommands(), openMSXBootInterval)
	session.appendOutputLine("[bridge] ready for commands")

	go scanOpenMSXOutput(stdout, "", session.appendOutputLine)
	go scanOpenMSXOutput(stderr, "[stderr] ", session.appendOutputLine)
	go func(local *openMSXBridgeSession) {
		waitErr := cmd.Wait()
		if waitErr != nil {
			local.appendOutputLine("[bridge] openMSX exited: " + waitErr.Error())
		} else {
			local.appendOutputLine("[bridge] openMSX exited")
		}
		local.setInputEnabled(false)
		fyne.Do(func() {
			e.clearOpenMSXBridge(local)
		})
	}(session)

	if e.status != nil {
		e.status.SetText("openMSX bridge: started")
	}
}
