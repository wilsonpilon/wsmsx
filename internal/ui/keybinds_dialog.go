package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/input"
	"ws7/internal/store/sqlite"
)

type keybindImportChange struct {
	CommandID string
	Label     string
	Before    string
	After     string
}

type keybindImportPreview struct {
	Result  []sqlite.KeybindRecord
	Changes []keybindImportChange
	Skipped []string
	Errors  []string
}

func defaultKeybindRecords() []sqlite.KeybindRecord {
	defs := input.DefaultKeybindDefinitions()
	records := make([]sqlite.KeybindRecord, 0, len(defs))
	for _, def := range defs {
		records = append(records, sqlite.KeybindRecord{
			CommandID:    strings.TrimSpace(def.ID),
			Label:        strings.TrimSpace(def.Label),
			Shortcut:     strings.TrimSpace(def.Shortcut),
			Context:      strings.TrimSpace(def.Context),
			Implemented:  def.Implemented,
			Configurable: def.Configurable,
		})
	}
	return records
}

func recordsToDefinitions(records []sqlite.KeybindRecord) []input.KeybindDefinition {
	out := make([]input.KeybindDefinition, 0, len(records))
	for _, rec := range records {
		out = append(out, input.KeybindDefinition{
			ID:           strings.TrimSpace(rec.CommandID),
			Label:        strings.TrimSpace(rec.Label),
			Shortcut:     strings.TrimSpace(rec.Shortcut),
			Context:      strings.TrimSpace(rec.Context),
			Implemented:  rec.Implemented,
			Configurable: rec.Configurable,
		})
	}
	return out
}

func keybindContextOptions(records []sqlite.KeybindRecord) []string {
	seen := map[string]struct{}{"All": {}}
	for _, rec := range records {
		ctx := strings.TrimSpace(rec.Context)
		if ctx != "" {
			seen[ctx] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for opt := range seen {
		out = append(out, opt)
	}
	sort.Strings(out)
	if len(out) > 0 && out[0] != "All" {
		out = append([]string{"All"}, out...)
	}
	return out
}

func applyKeybindFilters(records []sqlite.KeybindRecord, contextFilter, implementedFilter, configurableFilter string) []int {
	contextFilter = strings.TrimSpace(strings.ToLower(contextFilter))
	implementedFilter = strings.TrimSpace(strings.ToLower(implementedFilter))
	configurableFilter = strings.TrimSpace(strings.ToLower(configurableFilter))

	idx := make([]int, 0, len(records))
	for i, rec := range records {
		if contextFilter != "" && contextFilter != "all" && strings.ToLower(rec.Context) != contextFilter {
			continue
		}
		switch implementedFilter {
		case "implemented":
			if !rec.Implemented {
				continue
			}
		case "not implemented":
			if rec.Implemented {
				continue
			}
		}
		switch configurableFilter {
		case "configurable":
			if !rec.Configurable {
				continue
			}
		case "locked":
			if rec.Configurable {
				continue
			}
		}
		idx = append(idx, i)
	}
	return idx
}

func findShortcutConflict(records []sqlite.KeybindRecord, shortcut, currentID string) (sqlite.KeybindRecord, bool) {
	for _, rec := range records {
		if rec.CommandID == currentID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(rec.Shortcut), strings.TrimSpace(shortcut)) && strings.TrimSpace(shortcut) != "" {
			return rec, true
		}
	}
	return sqlite.KeybindRecord{}, false
}

func buildKeybindImportPreview(records []sqlite.KeybindRecord, data []byte) (keybindImportPreview, error) {
	rawRows, err := parseImportedKeybindRows(data)
	if err != nil {
		return keybindImportPreview{}, err
	}

	preview := keybindImportPreview{Result: make([]sqlite.KeybindRecord, len(records))}
	copy(preview.Result, records)
	indexByID := map[string]int{}
	for i, rec := range preview.Result {
		indexByID[strings.TrimSpace(rec.CommandID)] = i
	}

	for _, row := range rawRows {
		id := firstStringField(row, "command_id", "commandId", "CommandID", "id", "command")
		shortcut := firstStringField(row, "shortcut", "Shortcut")
		id = strings.TrimSpace(id)
		if id == "" {
			preview.Errors = append(preview.Errors, "entry without command_id")
			continue
		}
		idx, ok := indexByID[id]
		if !ok {
			preview.Errors = append(preview.Errors, "unknown command_id: "+id)
			continue
		}
		rec := preview.Result[idx]
		if !rec.Configurable {
			preview.Skipped = append(preview.Skipped, rec.Label+" (locked)")
			continue
		}

		normalized := ""
		if strings.TrimSpace(shortcut) != "" {
			norm, err := input.NormalizeShortcut(shortcut)
			if err != nil {
				preview.Errors = append(preview.Errors, fmt.Sprintf("%s: %v", rec.Label, err))
				continue
			}
			normalized = norm
		}
		if strings.EqualFold(strings.TrimSpace(rec.Shortcut), normalized) {
			continue
		}
		preview.Changes = append(preview.Changes, keybindImportChange{
			CommandID: rec.CommandID,
			Label:     rec.Label,
			Before:    rec.Shortcut,
			After:     normalized,
		})
		rec.Shortcut = normalized
		preview.Result[idx] = rec
	}

	seenShortcut := map[string]string{}
	for _, rec := range preview.Result {
		s := strings.TrimSpace(rec.Shortcut)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if otherID, exists := seenShortcut[key]; exists && otherID != rec.CommandID {
			preview.Errors = append(preview.Errors, fmt.Sprintf("conflict: %s is duplicated", rec.Shortcut))
			continue
		}
		seenShortcut[key] = rec.CommandID
	}

	r := input.NewResolver()
	if err := r.ApplyKeybinds(recordsToDefinitions(preview.Result)); err != nil {
		preview.Errors = append(preview.Errors, err.Error())
	}

	return preview, nil
}

func parseImportedKeybindRows(data []byte) ([]map[string]any, error) {
	var rawRows []map[string]any
	if err := json.Unmarshal(data, &rawRows); err == nil {
		return rawRows, nil
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}

	accepted := map[string]struct{}{
		"keybinds": {},
		"bindings": {},
		"items":    {},
	}
	for key, payload := range wrapped {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if _, ok := accepted[normalized]; !ok {
			continue
		}
		if err := json.Unmarshal(payload, &rawRows); err != nil {
			return nil, err
		}
		return rawRows, nil
	}

	return nil, fmt.Errorf("import JSON must be an array or object with keybinds/bindings/items")
}

func firstStringField(row map[string]any, keys ...string) string {
	for _, key := range keys {
		for actual, value := range row {
			if !strings.EqualFold(strings.TrimSpace(actual), key) {
				continue
			}
			if s, ok := value.(string); ok {
				return s
			}
		}
	}
	return ""
}

func summarizeKeybindImportPreview(preview keybindImportPreview) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Changes: %d\n", len(preview.Changes)))
	b.WriteString(fmt.Sprintf("Skipped: %d\n", len(preview.Skipped)))
	b.WriteString(fmt.Sprintf("Errors:  %d\n\n", len(preview.Errors)))
	if len(preview.Changes) > 0 {
		b.WriteString("Changed entries:\n")
		limit := len(preview.Changes)
		if limit > 20 {
			limit = 20
		}
		for i := 0; i < limit; i++ {
			item := preview.Changes[i]
			before := item.Before
			if strings.TrimSpace(before) == "" {
				before = "(none)"
			}
			after := item.After
			if strings.TrimSpace(after) == "" {
				after = "(none)"
			}
			b.WriteString(fmt.Sprintf("- %s: %s -> %s\n", item.Label, before, after))
		}
		if len(preview.Changes) > limit {
			b.WriteString(fmt.Sprintf("... and %d more\n", len(preview.Changes)-limit))
		}
	}
	if len(preview.Errors) > 0 {
		b.WriteString("\nValidation errors:\n")
		for _, errMsg := range preview.Errors {
			b.WriteString("- " + errMsg + "\n")
		}
	}
	if len(preview.Skipped) > 0 {
		b.WriteString("\nSkipped entries:\n")
		for _, item := range preview.Skipped {
			b.WriteString("- " + item + "\n")
		}
	}
	return strings.TrimSpace(b.String())
}

func (e *editorUI) ensureKeybindCache() []sqlite.KeybindRecord {
	if len(e.keybindCatalog) == 0 {
		e.keybindCatalog = defaultKeybindRecords()
	}
	if e.store == nil {
		return e.keybindCatalog
	}
	if list, err := e.store.ListKeybinds(context.Background()); err == nil && len(list) > 0 {
		e.keybindCatalog = list
	}
	return e.keybindCatalog
}

func (e *editorUI) applyKeybindResolverMappings() {
	if e == nil || e.resolver == nil {
		return
	}
	_ = e.resolver.ApplyKeybinds(recordsToDefinitions(e.keybindCatalog))
}

func (e *editorUI) shortcutLabelForCommand(cmd input.Command) string {
	if e == nil {
		return ""
	}
	for _, rec := range e.keybindCatalog {
		if strings.TrimSpace(rec.CommandID) == string(cmd) {
			return strings.TrimSpace(rec.Shortcut)
		}
	}
	if e.resolver != nil {
		return strings.TrimSpace(e.resolver.ShortcutForCommand(cmd))
	}
	return ""
}

func (e *editorUI) cmdKeybinds() {
	records := e.ensureKeybindCache()
	if len(records) == 0 {
		dialog.ShowInformation("Keybinds", "No keybind entries were found.", e.window)
		return
	}

	contextOptions := keybindContextOptions(records)
	contextFilter := widget.NewSelect(contextOptions, nil)
	contextFilter.SetSelected("All")
	implementedFilter := widget.NewSelect([]string{"All", "Implemented", "Not Implemented"}, nil)
	implementedFilter.SetSelected("All")
	configurableFilter := widget.NewSelect([]string{"All", "Configurable", "Locked"}, nil)
	configurableFilter.SetSelected("All")

	filtered := applyKeybindFilters(records, contextFilter.Selected, implementedFilter.Selected, configurableFilter.Selected)
	selectedRecord := -1

	defaultShortcuts := map[string]string{}
	for _, def := range input.DefaultKeybindDefinitions() {
		if norm, err := input.NormalizeShortcut(def.Shortcut); err == nil {
			defaultShortcuts[def.ID] = norm
		}
	}

	selectedLabel := widget.NewLabel("Select a command row to edit the shortcut.")
	shortcutEntry := widget.NewEntry()
	shortcutEntry.SetPlaceHolder("Ctrl+K,S")
	shortcutEntry.Disable()
	saveBtn := widget.NewButton("Apply Shortcut", nil)
	saveBtn.Disable()
	resetBtn := widget.NewButton("Reset to Default", nil)
	resetBtn.Disable()

	headers := []string{"Command", "Shortcut", "Context", "Implemented", "Configurable"}
	table := widget.NewTable(
		func() (int, int) {
			return len(filtered) + 1, len(headers)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.SetText(headers[id.Col])
				return
			}
			label.TextStyle = fyne.TextStyle{}
			rec := records[filtered[id.Row-1]]
			switch id.Col {
			case 0:
				label.SetText(rec.Label)
			case 1:
				if strings.TrimSpace(rec.Shortcut) == "" {
					label.SetText("(none)")
				} else {
					label.SetText(rec.Shortcut)
				}
			case 2:
				label.SetText(rec.Context)
			case 3:
				if rec.Implemented {
					label.SetText("Yes")
				} else {
					label.SetText("No")
				}
			case 4:
				if rec.Configurable {
					label.SetText("Yes")
				} else {
					label.SetText("No")
				}
			}
		},
	)
	table.SetColumnWidth(0, 270)
	table.SetColumnWidth(1, 170)
	table.SetColumnWidth(2, 110)
	table.SetColumnWidth(3, 110)
	table.SetColumnWidth(4, 110)

	refreshSelection := func(idx int) {
		selectedRecord = idx
		if idx < 0 || idx >= len(records) {
			selectedLabel.SetText("Select a command row to edit the shortcut.")
			shortcutEntry.SetText("")
			shortcutEntry.Disable()
			saveBtn.Disable()
			resetBtn.Disable()
			return
		}
		rec := records[idx]
		selectedLabel.SetText(fmt.Sprintf("%s (%s)", rec.Label, rec.CommandID))
		shortcutEntry.SetText(rec.Shortcut)
		if rec.Configurable {
			shortcutEntry.Enable()
			saveBtn.Enable()
			if _, ok := defaultShortcuts[rec.CommandID]; ok {
				resetBtn.Enable()
			} else {
				resetBtn.Disable()
			}
		} else {
			shortcutEntry.Disable()
			saveBtn.Disable()
			resetBtn.Disable()
		}
	}

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		if id.Row-1 >= 0 && id.Row-1 < len(filtered) {
			refreshSelection(filtered[id.Row-1])
		}
	}

	refreshTable := func() {
		filtered = applyKeybindFilters(records, contextFilter.Selected, implementedFilter.Selected, configurableFilter.Selected)
		table.Refresh()
		if selectedRecord >= 0 {
			visible := false
			for _, idx := range filtered {
				if idx == selectedRecord {
					visible = true
					break
				}
			}
			if !visible {
				refreshSelection(-1)
			}
		}
	}

	persistRecord := func(rec sqlite.KeybindRecord) error {
		if e.store != nil {
			if err := e.store.UpsertKeybind(context.Background(), rec); err != nil {
				return err
			}
		}
		e.keybindCatalog = records
		e.applyKeybindResolverMappings()
		return nil
	}

	saveBtn.OnTapped = func() {
		if selectedRecord < 0 || selectedRecord >= len(records) {
			return
		}
		rec := records[selectedRecord]
		if !rec.Configurable {
			return
		}
		raw := strings.TrimSpace(shortcutEntry.Text)
		normalized := ""
		if raw != "" {
			norm, err := input.NormalizeShortcut(raw)
			if err != nil {
				dialog.ShowInformation("Keybinds", "Invalid shortcut: "+err.Error(), e.window)
				return
			}
			normalized = norm
		}
		if conflict, ok := findShortcutConflict(records, normalized, rec.CommandID); ok {
			dialog.ShowInformation("Keybind Conflict", fmt.Sprintf("%s already uses %s", conflict.Label, normalized), e.window)
			return
		}
		rec.Shortcut = normalized
		records[selectedRecord] = rec
		if err := persistRecord(rec); err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		refreshTable()
		refreshSelection(selectedRecord)
		if e.status != nil {
			e.status.SetText("Keybind updated: " + rec.Label)
		}
	}

	resetBtn.OnTapped = func() {
		if selectedRecord < 0 || selectedRecord >= len(records) {
			return
		}
		rec := records[selectedRecord]
		defaultShortcut, ok := defaultShortcuts[rec.CommandID]
		if !ok {
			return
		}
		if conflict, hasConflict := findShortcutConflict(records, defaultShortcut, rec.CommandID); hasConflict {
			dialog.ShowInformation("Keybind Conflict", fmt.Sprintf("%s already uses %s", conflict.Label, defaultShortcut), e.window)
			return
		}
		rec.Shortcut = defaultShortcut
		records[selectedRecord] = rec
		if err := persistRecord(rec); err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		refreshTable()
		refreshSelection(selectedRecord)
	}

	exportBtn := widget.NewButton("Export current keybinds (JSON)", func() {
		payload := make([]sqlite.KeybindRecord, 0, len(filtered))
		for _, idx := range filtered {
			payload = append(payload, records[idx])
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		save := dialog.NewFileSave(func(writer fyne.URIWriteCloser, writeErr error) {
			if writeErr != nil {
				dialog.ShowError(writeErr, e.window)
				return
			}
			if writer == nil {
				return
			}
			if _, err := writer.Write(data); err != nil {
				_ = writer.Close()
				dialog.ShowError(err, e.window)
				return
			}
			if err := writer.Close(); err != nil {
				dialog.ShowError(err, e.window)
				return
			}
			if e.status != nil {
				e.status.SetText("Keybinds exported: " + writer.URI().Path())
			}
		}, e.window)
		save.SetFileName("keybinds.json")
		save.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		if e.store != nil {
			if lastDir, _ := e.store.GetSetting(context.Background(), "last_dir"); strings.TrimSpace(lastDir) != "" {
				u, err := storage.ParseURI("file://" + filepath.ToSlash(lastDir))
				if err == nil {
					if lister, lErr := storage.ListerForURI(u); lErr == nil {
						save.SetLocation(lister)
					}
				}
			}
		}
		save.Show()
	})

	importBtn := widget.NewButton("Import keybinds (JSON)", func() {
		open := dialog.NewFileOpen(func(reader fyne.URIReadCloser, openErr error) {
			if openErr != nil {
				dialog.ShowError(openErr, e.window)
				return
			}
			if reader == nil {
				return
			}
			defer func() { _ = reader.Close() }()

			data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, e.window)
				return
			}

			preview, err := buildKeybindImportPreview(records, data)
			if err != nil {
				dialog.ShowError(err, e.window)
				return
			}

			previewText := widget.NewMultiLineEntry()
			previewText.SetText(summarizeKeybindImportPreview(preview))
			previewText.Disable()
			previewText.Wrapping = fyne.TextWrapWord
			previewText.SetMinRowsVisible(18)

			var previewDlg *dialog.CustomDialog
			applyBtn := widget.NewButton("Apply Import", func() {
				if len(preview.Errors) > 0 {
					return
				}
				for i := range preview.Result {
					before := records[i]
					after := preview.Result[i]
					if before.Shortcut == after.Shortcut {
						continue
					}
					if e.store != nil {
						if upsertErr := e.store.UpsertKeybind(context.Background(), after); upsertErr != nil {
							dialog.ShowError(upsertErr, e.window)
							return
						}
					}
					records[i] = after
				}
				e.keybindCatalog = records
				e.applyKeybindResolverMappings()
				refreshTable()
				refreshSelection(-1)
				if previewDlg != nil {
					previewDlg.Hide()
				}
				if e.status != nil {
					e.status.SetText(fmt.Sprintf("Keybind import applied (%d change(s))", len(preview.Changes)))
				}
			})
			if len(preview.Errors) > 0 {
				applyBtn.Disable()
			}
			cancelBtn := widget.NewButton("Cancel", func() {
				if previewDlg != nil {
					previewDlg.Hide()
				}
			})
			content := container.NewBorder(nil, container.NewHBox(widget.NewLabel("Preview"), applyBtn, cancelBtn), nil, nil, container.NewVScroll(previewText))
			previewDlg = dialog.NewCustomWithoutButtons("Import Keybinds Preview", content, e.window)
			previewDlg.Resize(fyne.NewSize(860, 520))
			previewDlg.Show()
		}, e.window)
		open.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		if e.store != nil {
			if lastDir, _ := e.store.GetSetting(context.Background(), "last_dir"); strings.TrimSpace(lastDir) != "" {
				u, err := storage.ParseURI("file://" + filepath.ToSlash(lastDir))
				if err == nil {
					if lister, lErr := storage.ListerForURI(u); lErr == nil {
						open.SetLocation(lister)
					}
				}
			}
		}
		open.Show()
	})

	contextFilter.OnChanged = func(string) { refreshTable() }
	implementedFilter.OnChanged = func(string) { refreshTable() }
	configurableFilter.OnChanged = func(string) { refreshTable() }

	filters := container.NewGridWithColumns(6,
		widget.NewLabel("Context"), contextFilter,
		widget.NewLabel("Implemented"), implementedFilter,
		widget.NewLabel("Configurable"), configurableFilter,
	)
	editorBox := container.NewVBox(
		selectedLabel,
		shortcutEntry,
		container.NewHBox(saveBtn, resetBtn),
	)
	content := container.NewBorder(
		container.NewVBox(filters, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewHBox(importBtn, exportBtn), editorBox),
		nil,
		nil,
		table,
	)

	dlg := dialog.NewCustom("Keybinds", "Close", content, e.window)
	dlg.Resize(fyne.NewSize(980, 620))
	dlg.Show()
	if e.status != nil {
		e.status.SetText("Keybinds: manage shortcuts")
	}
}
