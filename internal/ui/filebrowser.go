package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type fileEntry struct {
	name     string
	fullPath string
	isDir    bool
	isUp     bool
}

// fileBrowser is the startup/opening-menu view, showing a navigable directory listing.
type fileBrowser struct {
	dir         string
	entries     []fileEntry
	list        *widget.List
	dirLabel    *widget.Label
	onOpenFile  func(string) // called when a file is activated
	onDirChange func(string) // called whenever the browsed directory changes (optional)
	selectedIdx int          // last highlighted / selected item index

	// Content is the root widget to embed in the window.
	Content fyne.CanvasObject
}

func newFileBrowser(startDir string, onOpenFile func(string)) *fileBrowser {
	fb := &fileBrowser{
		dirLabel:   widget.NewLabel(""),
		onOpenFile: onOpenFile,
	}

	fb.list = widget.NewList(
		func() int { return len(fb.entries) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("")
			lbl.TextStyle = fyne.TextStyle{Monospace: true}
			return lbl
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(fb.entries) {
				return
			}
			lbl := o.(*widget.Label)
			e := fb.entries[i]
			switch {
			case e.isUp:
				lbl.SetText("  [..]")
			case e.isDir:
				lbl.SetText("  [" + e.name + "]")
			default:
				lbl.SetText("     " + e.name)
			}
		},
	)

	fb.list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(fb.entries) {
			return
		}
		fb.selectedIdx = int(id)
		e := fb.entries[id]
		if e.isDir || e.isUp {
			fb.loadDir(e.fullPath)
		} else {
			fb.onOpenFile(e.fullPath)
		}
	}

	title := widget.NewLabel("  WS7 - Opening Menu")
	title.TextStyle = fyne.TextStyle{Bold: true}

	helpBar := widget.NewLabel(
		"  \u2191\u2193 Move   Enter Open   [DIR] Folder   [..] Parent")

	fb.Content = container.NewBorder(
		container.NewVBox(title, fb.dirLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), helpBar),
		nil, nil,
		fb.list,
	)

	fb.loadDir(startDir)
	return fb
}

// loadDir reloads the listing for the given directory.
func (fb *fileBrowser) loadDir(dir string) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	fb.dir = abs
	fb.dirLabel.SetText("  Directory: " + abs)

	infos, _ := os.ReadDir(abs)
	fb.entries = nil

	// ".." entry unless we are at a filesystem root
	parent := filepath.Dir(abs)
	if parent != abs {
		fb.entries = append(fb.entries, fileEntry{
			name:     "..",
			fullPath: parent,
			isDir:    true,
			isUp:     true,
		})
	}

	var dirs, files []fileEntry
	for _, info := range infos {
		e := fileEntry{
			name:     info.Name(),
			fullPath: filepath.Join(abs, info.Name()),
			isDir:    info.IsDir(),
		}
		if info.IsDir() {
			dirs = append(dirs, e)
		} else {
			files = append(files, e)
		}
	}
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})
	fb.entries = append(fb.entries, dirs...)
	fb.entries = append(fb.entries, files...)

	fb.list.Refresh()
	if len(fb.entries) > 0 {
		fb.list.ScrollTo(0)
	}

	if fb.onDirChange != nil {
		fb.onDirChange(fb.dir)
	}
}

// Refresh reloads the current directory (useful after file operations).
func (fb *fileBrowser) Refresh() {
	fb.loadDir(fb.dir)
}
