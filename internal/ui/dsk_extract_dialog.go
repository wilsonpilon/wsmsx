package ui

import (
	"encoding/binary"
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

// cmdExtractDSKDialog opens a dual-pane dialog to pick a DSK file (source) and
// a destination directory, then select which files to extract from the image.
func (e *editorUI) cmdExtractDSKDialog() {
	if e.window == nil {
		return
	}

	dskPath := ""
	destDir := e.initialMakeDSKSourceDir()
	if !dirExists(destDir) {
		destDir = configureInitialDirectory("", "")
	}
	if !dirExists(destDir) {
		destDir = "."
	}

	selectedFiles := map[string]bool{}
	dskEntries := []string{} // filenames inside the DSK

	dskFileLabel := widget.NewLabel("Origem DSK: (nenhum arquivo selecionado)")
	dskFileLabel.Wrapping = fyne.TextWrapWord
	selectedLabel := widget.NewLabel("Selecionados: 0")

	fileRows := container.NewVBox()
	fileScroll := container.NewVScroll(fileRows)
	fileScroll.SetMinSize(fyne.NewSize(460, 320))

	refreshSelectedLabel := func() {
		n := 0
		for _, on := range selectedFiles {
			if on {
				n++
			}
		}
		selectedLabel.SetText(fmt.Sprintf("Selecionados: %d", n))
	}

	refreshRows := func() {
		rows := make([]fyne.CanvasObject, 0, len(dskEntries))
		if len(dskEntries) == 0 {
			rows = append(rows, widget.NewLabel("(selecione um arquivo DSK para listar o conteudo)"))
		}
		for _, name := range dskEntries {
			name := name
			check := widget.NewCheck(name, func(on bool) {
				selectedFiles[name] = on
				refreshSelectedLabel()
			})
			check.SetChecked(selectedFiles[name])
			rows = append(rows, check)
		}
		fileRows.Objects = rows
		fileRows.Refresh()
		refreshSelectedLabel()
	}

	loadDSKFile := func(path string) {
		abs, err := filepath.Abs(path)
		if err != nil {
			abs = path
		}
		dskPath = abs
		dskFileLabel.SetText("Origem DSK: " + filepath.Base(dskPath))
		names, readErr := dskListFiles(abs)
		if readErr != nil {
			dskEntries = nil
			selectedFiles = map[string]bool{}
			fileRows.Objects = []fyne.CanvasObject{
				widget.NewLabel("Erro ao ler o arquivo DSK: " + readErr.Error()),
			}
			fileRows.Refresh()
			refreshSelectedLabel()
			return
		}
		dskEntries = names
		selectedFiles = map[string]bool{}
		refreshRows()
	}

	destDirLabel := widget.NewLabel("Destino: " + destDir)
	destDirLabel.Wrapping = fyne.TextWrapWord

	pickDSKBtn := widget.NewButton("Procurar arquivo DSK...", func() {
		initialDir := configureInitialDirectory(dskPath, "")
		if initialDir == "" {
			initialDir = configureInitialDirectory(destDir, "")
		}
		dlg := dialog.NewFileOpen(func(uri fyne.URIReadCloser, err error) {
			if err != nil || uri == nil {
				return
			}
			uri.Close()
			loadDSKFile(uri.URI().Path())
		}, e.window)
		dlg.SetFilter(storage.NewExtensionFileFilter([]string{".dsk", ".DSK"}))
		if initialDir != "" {
			if lister, lerr := storage.ListerForURI(storage.NewFileURI(initialDir)); lerr == nil {
				dlg.SetLocation(lister)
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
			if lister, lerr := storage.ListerForURI(storage.NewFileURI(initialDir)); lerr == nil {
				dlg.SetLocation(lister)
			}
		}
		dlg.Show()
	})

	selectAllBtn := widget.NewButton("Selecionar todos", func() {
		for _, name := range dskEntries {
			selectedFiles[name] = true
		}
		refreshRows()
	})
	clearSelectionBtn := widget.NewButton("Limpar selecao", func() {
		for k := range selectedFiles {
			selectedFiles[k] = false
		}
		refreshRows()
	})

	leftPanel := container.NewBorder(
		container.NewVBox(dskFileLabel, pickDSKBtn, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewHBox(selectAllBtn, clearSelectionBtn), selectedLabel),
		nil,
		nil,
		fileScroll,
	)

	rightPanel := container.NewVBox(
		widget.NewLabel("Diretorio de destino:"),
		destDirLabel,
		pickDestBtn,
	)

	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.62

	refreshRows() // show placeholder before any DSK is loaded

	d := dialog.NewCustomConfirm("Extract DSK", "Extrair", "Fechar", split, func(ok bool) {
		if !ok {
			return
		}
		if dskPath == "" {
			dialog.ShowInformation("Extract DSK", "Selecione um arquivo DSK no painel esquerdo.", e.window)
			return
		}
		chosen := make([]string, 0, len(selectedFiles))
		for name, on := range selectedFiles {
			if on {
				chosen = append(chosen, name)
			}
		}
		sort.Strings(chosen)
		if len(chosen) == 0 {
			dialog.ShowInformation("Extract DSK", "Selecione ao menos um arquivo para extrair.", e.window)
			return
		}
		if err := dskExtractFiles(dskPath, destDir, chosen); err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		dialog.ShowInformation("Extract DSK",
			fmt.Sprintf("Extraido com sucesso!\n\n%d arquivo(s) extraido(s) para:\n%s", len(chosen), destDir),
			e.window)
	}, e.window)
	d.Resize(fyne.NewSize(920, 520))
	d.Show()
}

// dskListFiles reads a standard MSX-DOS FAT12 disk image and returns the
// filenames stored in its root directory.
func dskListFiles(dskPath string) ([]string, error) {
	const sectorSize = 512

	f, err := os.Open(dskPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	boot := make([]byte, sectorSize)
	if _, err := f.ReadAt(boot, 0); err != nil {
		return nil, fmt.Errorf("leitura do boot sector: %w", err)
	}

	reservedSec := int(binary.LittleEndian.Uint16(boot[0x0E:0x10]))
	numFATs := int(boot[0x10])
	rootEntries := int(binary.LittleEndian.Uint16(boot[0x11:0x13]))
	secPerFAT := int(binary.LittleEndian.Uint16(boot[0x16:0x18]))

	dirOffset := int64(sectorSize) * int64(reservedSec+numFATs*secPerFAT)

	dirData := make([]byte, rootEntries*32)
	if _, err := f.ReadAt(dirData, dirOffset); err != nil {
		return nil, fmt.Errorf("leitura do diretorio raiz: %w", err)
	}

	var names []string
	for i := 0; i < rootEntries; i++ {
		entry := dirData[i*32 : i*32+32]
		first := entry[0]
		if first == 0x00 {
			break // fim das entradas
		}
		if first == 0xE5 {
			continue // entrada deletada
		}
		attr := entry[0x0B]
		if attr&0x08 != 0 || attr&0x10 != 0 {
			continue // volume label ou subdiretório
		}

		rawName := strings.TrimRight(string(entry[0:8]), " ")
		rawExt := strings.TrimRight(string(entry[8:11]), " ")
		if rawName == "" {
			continue
		}
		name := rawName
		if rawExt != "" {
			name = rawName + "." + rawExt
		}
		// Skip entries with non-printable ASCII
		valid := true
		for _, c := range name {
			if c < 0x20 || c > 0x7E {
				valid = false
				break
			}
		}
		if valid {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}
