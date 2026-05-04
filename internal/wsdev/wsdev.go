package wsdev

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/config"
	"ws7/internal/store/sqlite"
	"ws7/internal/version"
)

const wsdevWSMSXPathKey = "wsdev_wsmsx_exe"

func Run() error {
	a := app.NewWithID("ws7.wsdev")
	w := a.NewWindow(version.Full() + " - wsdev")

	dbPath, err := config.DBPath()
	if err != nil {
		return err
	}
	store, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()
	currentPath, _ := store.GetSetting(ctx, wsdevWSMSXPathKey)

	pathLabel := widget.NewLabel("")
	pathLabel.Wrapping = fyne.TextWrapWord
	refreshPath := func() {
		if strings.TrimSpace(currentPath) == "" {
			pathLabel.SetText("WSMSX path: (not configured)")
			return
		}
		pathLabel.SetText("WSMSX path: " + currentPath)
	}
	refreshPath()

	programAction := func() {
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, openErr error) {
			if openErr != nil {
				dialog.ShowError(openErr, w)
				return
			}
			if reader == nil {
				return
			}
			defer func() { _ = reader.Close() }()

			picked := filepath.Clean(reader.URI().Path())
			if strings.TrimSpace(picked) == "" {
				return
			}
			if err := store.SetSetting(ctx, wsdevWSMSXPathKey, picked); err != nil {
				dialog.ShowError(err, w)
				return
			}
			currentPath = picked
			refreshPath()
			dialog.ShowInformation("Configure Program", "Path saved successfully.", w)
		}, w)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".exe", ".EXE"}))
		d.Show()
	}

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Exit", func() { w.Close() }),
	)
	configureMenu := fyne.NewMenu("Configure")
	helpItem := fyne.NewMenuItem("Help", func() {
		dialog.ShowInformation("Configure > Help", "Not implemented yet.", w)
	})
	programItem := fyne.NewMenuItem("Program", programAction)
	configureMenu.Items = []*fyne.MenuItem{helpItem, programItem}

	w.SetMainMenu(fyne.NewMainMenu(fileMenu, configureMenu))
	w.SetContent(container.NewVBox(
		widget.NewLabel("wsdev utility"),
		widget.NewSeparator(),
		pathLabel,
		widget.NewLabel(fmt.Sprintf("Shared DB: %s", dbPath)),
	))
	w.Resize(fyne.NewSize(760, 320))
	w.ShowAndRun()
	return nil
}
