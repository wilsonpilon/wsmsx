package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// cmdMakeDSKDialog opens a dual-pane file manager style dialog used to prepare
// a future DSK creation workflow.
func (e *editorUI) cmdMakeDSKDialog() {
	if e.window == nil {
		return
	}

	sourceDir := e.initialMakeDSKSourceDir()
	destDir := sourceDir
	if !dirExists(destDir) {
		destDir = configureInitialDirectory("", "")
	}
	if !dirExists(destDir) {
		destDir = "."
	}

	selected := map[string]bool{}
	entries := []fileEntry{}

	sourceDirLabel := widget.NewLabel("")
	sourceDirLabel.Wrapping = fyne.TextWrapWord
	selectedLabel := widget.NewLabel("Selecionados: 0")

	fileRows := container.NewVBox()
	fileScroll := container.NewVScroll(fileRows)
	fileScroll.SetMinSize(fyne.NewSize(460, 320))

	refreshSelectedLabel := func() {
		count := 0
		for _, ok := range selected {
			if ok {
				count++
			}
		}
		selectedLabel.SetText(fmt.Sprintf("Selecionados: %d", count))
	}

	var loadSourceDir func(string)
	refreshRows := func() {
		rows := make([]fyne.CanvasObject, 0, len(entries))
		for _, entry := range entries {
			entry := entry
			if entry.isDir || entry.isUp {
				label := "[" + entry.name + "]"
				if entry.isUp {
					label = "[..]"
				}
				btn := widget.NewButton(label, func() { loadSourceDir(entry.fullPath) })
				rows = append(rows, btn)
				continue
			}
			check := widget.NewCheck(entry.name, func(on bool) {
				selected[entry.fullPath] = on
				refreshSelectedLabel()
			})
			check.SetChecked(selected[entry.fullPath])
			rows = append(rows, check)
		}
		fileRows.Objects = rows
		fileRows.Refresh()
		refreshSelectedLabel()
	}

	loadSourceDir = func(dir string) {
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = dir
		}
		if !dirExists(abs) {
			return
		}
		sourceDir = abs
		sourceDirLabel.SetText("Origem: " + sourceDir)
		entries = dskPanelEntries(abs)
		selected = map[string]bool{}
		refreshRows()
	}

	destDirLabel := widget.NewLabel("Destino: " + destDir)
	destDirLabel.Wrapping = fyne.TextWrapWord

	dskName := widget.NewEntry()
	dskName.SetText("novo.dsk")

	pickSourceBtn := widget.NewButton("Procurar origem...", func() {
		initialDir := configureInitialDirectory(sourceDir, "")
		dlg := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			loadSourceDir(uri.Path())
		}, e.window)
		if initialDir != "" {
			if uri, err := storage.ListerForURI(storage.NewFileURI(initialDir)); err == nil {
				dlg.SetLocation(uri)
			}
		}
		dlg.Show()
	})

	pickDestBtn := widget.NewButton("Procurar destino...", func() {
		initialDir := configureInitialDirectory(destDir, "")
		dlg := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			destDir = uri.Path()
			destDirLabel.SetText("Destino: " + destDir)
		}, e.window)
		if initialDir != "" {
			if uri, err := storage.ListerForURI(storage.NewFileURI(initialDir)); err == nil {
				dlg.SetLocation(uri)
			}
		}
		dlg.Show()
	})

	selectAllBtn := widget.NewButton("Selecionar todos", func() {
		for _, entry := range entries {
			if !entry.isDir && !entry.isUp {
				selected[entry.fullPath] = true
			}
		}
		refreshRows()
	})
	clearSelectionBtn := widget.NewButton("Limpar selecao", func() {
		for k := range selected {
			selected[k] = false
		}
		refreshRows()
	})

	leftPanel := container.NewBorder(
		container.NewVBox(sourceDirLabel, pickSourceBtn, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewHBox(selectAllBtn, clearSelectionBtn), selectedLabel),
		nil,
		nil,
		fileScroll,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("Diretorio de destino:"),
		destDirLabel,
		pickDestBtn,
		widget.NewSeparator(),
		widget.NewLabel("Nome do arquivo DSK:"),
		dskName,
	)

	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.62

	loadSourceDir(sourceDir)

	d := dialog.NewCustomConfirm("Make DSK", "Criar", "Fechar", split, func(ok bool) {
		if !ok {
			return
		}
		selectedFiles := make([]string, 0, len(selected))
		for path, on := range selected {
			if on {
				selectedFiles = append(selectedFiles, path)
			}
		}
		sort.Strings(selectedFiles)
		if len(selectedFiles) == 0 {
			dialog.ShowInformation("Make DSK", "Selecione ao menos um arquivo no painel esquerdo.", e.window)
			return
		}

		name := strings.TrimSpace(dskName.Text)
		if name == "" {
			dialog.ShowInformation("Make DSK", "Informe o nome do arquivo DSK.", e.window)
			return
		}
		if !strings.HasSuffix(strings.ToLower(name), ".dsk") {
			name += ".dsk"
		}

		dskOut := filepath.Join(destDir, name)
		if err := dskCreateImage(dskOut, selectedFiles); err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		dialog.ShowInformation("Make DSK",
			fmt.Sprintf("Imagem criada com sucesso!\n\n%s\n\n%d arquivo(s) adicionado(s).", dskOut, len(selectedFiles)),
			e.window)
	}, e.window)
	d.Resize(fyne.NewSize(920, 520))
	d.Show()
}

func (e *editorUI) initialMakeDSKSourceDir() string {
	if e.filePath != "" {
		dir := filepath.Dir(e.filePath)
		if dirExists(dir) {
			return dir
		}
	}
	if e.browser != nil && dirExists(e.browser.dir) {
		return e.browser.dir
	}
	if cwd, err := os.Getwd(); err == nil && dirExists(cwd) {
		return cwd
	}
	return "."
}

func dskPanelEntries(dir string) []fileEntry {
	infos, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	entries := make([]fileEntry, 0, len(infos)+1)
	parent := filepath.Dir(dir)
	if parent != dir {
		entries = append(entries, fileEntry{name: "..", fullPath: parent, isDir: true, isUp: true})
	}

	dirs := make([]fileEntry, 0)
	files := make([]fileEntry, 0)
	for _, info := range infos {
		entry := fileEntry{name: info.Name(), fullPath: filepath.Join(dir, info.Name()), isDir: info.IsDir()}
		if info.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name) })
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].name) < strings.ToLower(files[j].name) })
	entries = append(entries, dirs...)
	entries = append(entries, files...)
	return entries
}
