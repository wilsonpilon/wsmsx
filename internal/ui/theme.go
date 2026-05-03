package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const defaultEditorThemeID = "dark"
const defaultSyntaxThemeID = "msx-dark"

const (
	defaultEditorFontFamily = "Source Code Pro"
	defaultEditorFontWeight = "Regular"
	defaultEditorFontSize   = float32(14)
)

const (
	editorThemeDarkID      = "dark"
	editorThemeLightID     = "light"
	editorThemeOneDarkID   = "one-dark"
	editorThemeMonokaiID   = "monokai"
	editorThemeSolarizedID = "solarized-dark"
	editorThemeGithubID    = "github-dark"
)

const (
	syntaxThemeDarkID        = "msx-dark"
	syntaxThemeLightID       = "msx-light"
	syntaxThemeGreenScreenID = "msx-green-screen"
	syntaxThemeCobaltID      = "msx-cobalt"
	syntaxThemeAmberID       = "msx-amber"
)

type editorPalette struct {
	Background      color.NRGBA
	Foreground      color.NRGBA
	MenuBackground  color.NRGBA
	InputBackground color.NRGBA
	Overlay         color.NRGBA
	Primary         color.NRGBA
}

type syntaxPalette struct {
	Plain       color.NRGBA
	Instruction color.NRGBA
	Jump        color.NRGBA
	Function    color.NRGBA
	Operator    color.NRGBA
	Number      color.NRGBA
	String      color.NRGBA
	Comment     color.NRGBA
	Identifier  color.NRGBA
}

type syntaxThemeDefinition struct {
	ID      string
	Name    string
	Palette syntaxPalette
	Builtin bool
}

type syntaxThemeOption struct {
	Label string
	ID    string
}

type syntaxThemeColorField struct {
	Key   string
	Label string
}

type syntaxThemeFile struct {
	Version int               `json:"version"`
	ID      string            `json:"id,omitempty"`
	Name    string            `json:"name"`
	Colors  map[string]string `json:"colors"`
}

type syntaxThemeCatalogFile struct {
	Version int               `json:"version"`
	Themes  []syntaxThemeFile `json:"themes"`
}

var syntaxThemeColorFields = []syntaxThemeColorField{
	{Key: "plain", Label: "Plain"},
	{Key: "instruction", Label: "Instruction"},
	{Key: "jump", Label: "Jump"},
	{Key: "function", Label: "Function"},
	{Key: "operator", Label: "Operator"},
	{Key: "number", Label: "Number"},
	{Key: "string", Label: "String"},
	{Key: "comment", Label: "Comment"},
	{Key: "identifier", Label: "Identifier"},
}

var editorPalettes = map[string]editorPalette{
	editorThemeDarkID: {
		Background:      color.NRGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x3D, G: 0x3D, B: 0x3D, A: 0xFF},
		Primary:         color.NRGBA{R: 0x00, G: 0x7A, B: 0xCC, A: 0xFF},
	},
	editorThemeLightID: {
		Background:      color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		Foreground:      color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0xF3, G: 0xF3, B: 0xF3, A: 0xFF},
		InputBackground: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		Overlay:         color.NRGBA{R: 0xE5, G: 0xE5, B: 0xE5, A: 0xFF},
		Primary:         color.NRGBA{R: 0x00, G: 0x54, B: 0x99, A: 0xFF},
	},
	editorThemeOneDarkID: {
		Background:      color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xAB, G: 0xB2, B: 0xBF, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x21, G: 0x25, B: 0x2B, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x1E, G: 0x22, B: 0x27, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x3B, G: 0x40, B: 0x48, A: 0xFF},
		Primary:         color.NRGBA{R: 0x61, G: 0xAF, B: 0xEF, A: 0xFF},
	},
	editorThemeMonokaiID: {
		Background:      color.NRGBA{R: 0x27, G: 0x28, B: 0x22, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xF8, G: 0xF8, B: 0xF2, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x19, G: 0x19, B: 0x19, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x3E, G: 0x3D, B: 0x32, A: 0xFF},
		Primary:         color.NRGBA{R: 0xA6, G: 0xE2, B: 0x2E, A: 0xFF},
	},
	editorThemeSolarizedID: {
		Background:      color.NRGBA{R: 0x00, G: 0x2B, B: 0x36, A: 0xFF},
		Foreground:      color.NRGBA{R: 0x83, G: 0x94, B: 0x96, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x07, G: 0x36, B: 0x42, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x00, G: 0x1E, B: 0x26, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x58, G: 0x6E, B: 0x75, A: 0xFF},
		Primary:         color.NRGBA{R: 0x26, G: 0x8B, B: 0xD2, A: 0xFF},
	},
	editorThemeGithubID: {
		Background:      color.NRGBA{R: 0x0D, G: 0x11, B: 0x17, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xC9, G: 0xD1, B: 0xD9, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x16, G: 0x1B, B: 0x22, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x01, G: 0x04, B: 0x09, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x30, G: 0x36, B: 0x3D, A: 0xFF},
		Primary:         color.NRGBA{R: 0x58, G: 0xA6, B: 0xFF, A: 0xFF},
	},
}

var syntaxPalettes = map[string]syntaxPalette{
	syntaxThemeDarkID: {
		Plain:       color.NRGBA{R: 0xD4, G: 0xD4, B: 0xD4, A: 0xFF},
		Instruction: color.NRGBA{R: 0x56, G: 0x9C, B: 0xD6, A: 0xFF},
		Jump:        color.NRGBA{R: 0xC5, G: 0x86, B: 0xC0, A: 0xFF},
		Function:    color.NRGBA{R: 0x4E, G: 0xC9, B: 0xB0, A: 0xFF},
		Operator:    color.NRGBA{R: 0xD4, G: 0xD4, B: 0xD4, A: 0xFF},
		Number:      color.NRGBA{R: 0xB5, G: 0xCE, B: 0xA8, A: 0xFF},
		String:      color.NRGBA{R: 0xCE, G: 0x91, B: 0x78, A: 0xFF},
		Comment:     color.NRGBA{R: 0x6A, G: 0x99, B: 0x55, A: 0xFF},
		Identifier:  color.NRGBA{R: 0x9C, G: 0xDC, B: 0xFE, A: 0xFF},
	},
	syntaxThemeLightID: {
		Plain:       color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF},
		Instruction: color.NRGBA{R: 0x00, G: 0x00, B: 0xCC, A: 0xFF},
		Jump:        color.NRGBA{R: 0x7A, G: 0x3E, B: 0x9D, A: 0xFF},
		Function:    color.NRGBA{R: 0x04, G: 0x5B, B: 0x67, A: 0xFF},
		Operator:    color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF},
		Number:      color.NRGBA{R: 0x09, G: 0x80, B: 0x5A, A: 0xFF},
		String:      color.NRGBA{R: 0xA3, G: 0x15, B: 0x15, A: 0xFF},
		Comment:     color.NRGBA{R: 0x00, G: 0x80, B: 0x00, A: 0xFF},
		Identifier:  color.NRGBA{R: 0x00, G: 0x11, B: 0x88, A: 0xFF},
	},
	syntaxThemeGreenScreenID: {
		Plain:       color.NRGBA{R: 0x7C, G: 0xFF, B: 0x7C, A: 0xFF},
		Instruction: color.NRGBA{R: 0x9C, G: 0xFF, B: 0x9C, A: 0xFF},
		Jump:        color.NRGBA{R: 0xC8, G: 0xFF, B: 0x78, A: 0xFF},
		Function:    color.NRGBA{R: 0x66, G: 0xFF, B: 0xB2, A: 0xFF},
		Operator:    color.NRGBA{R: 0x7C, G: 0xFF, B: 0x7C, A: 0xFF},
		Number:      color.NRGBA{R: 0xE4, G: 0xFF, B: 0x8A, A: 0xFF},
		String:      color.NRGBA{R: 0xB6, G: 0xFF, B: 0xB6, A: 0xFF},
		Comment:     color.NRGBA{R: 0x52, G: 0xB8, B: 0x52, A: 0xFF},
		Identifier:  color.NRGBA{R: 0x9C, G: 0xFF, B: 0xD0, A: 0xFF},
	},
	syntaxThemeCobaltID: {
		Plain:       color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		Instruction: color.NRGBA{R: 0x9E, G: 0xD6, B: 0xFF, A: 0xFF},
		Jump:        color.NRGBA{R: 0xFF, G: 0x9D, B: 0xDC, A: 0xFF},
		Function:    color.NRGBA{R: 0xFF, G: 0xE0, B: 0x6B, A: 0xFF},
		Operator:    color.NRGBA{R: 0xE1, G: 0xF2, B: 0xFF, A: 0xFF},
		Number:      color.NRGBA{R: 0xFF, G: 0xB0, B: 0x6B, A: 0xFF},
		String:      color.NRGBA{R: 0xA5, G: 0xFF, B: 0x90, A: 0xFF},
		Comment:     color.NRGBA{R: 0x7F, G: 0x9F, B: 0xBF, A: 0xFF},
		Identifier:  color.NRGBA{R: 0xD7, G: 0xF0, B: 0xFF, A: 0xFF},
	},
	syntaxThemeAmberID: {
		Plain:       color.NRGBA{R: 0xFF, G: 0xC8, B: 0x5A, A: 0xFF},
		Instruction: color.NRGBA{R: 0xFF, G: 0xD8, B: 0x76, A: 0xFF},
		Jump:        color.NRGBA{R: 0xFF, G: 0xA9, B: 0x3D, A: 0xFF},
		Function:    color.NRGBA{R: 0xFF, G: 0xE7, B: 0x9B, A: 0xFF},
		Operator:    color.NRGBA{R: 0xFF, G: 0xC8, B: 0x5A, A: 0xFF},
		Number:      color.NRGBA{R: 0xFF, G: 0xF1, B: 0xA8, A: 0xFF},
		String:      color.NRGBA{R: 0xFF, G: 0xB8, B: 0x52, A: 0xFF},
		Comment:     color.NRGBA{R: 0xC7, G: 0x86, B: 0x2A, A: 0xFF},
		Identifier:  color.NRGBA{R: 0xFF, G: 0xDB, B: 0x8B, A: 0xFF},
	},
}

var builtinSyntaxThemes = map[string]syntaxThemeDefinition{
	syntaxThemeDarkID:        {ID: syntaxThemeDarkID, Name: "MSX Dark", Palette: syntaxPalettes[syntaxThemeDarkID], Builtin: true},
	syntaxThemeLightID:       {ID: syntaxThemeLightID, Name: "MSX Light", Palette: syntaxPalettes[syntaxThemeLightID], Builtin: true},
	syntaxThemeGreenScreenID: {ID: syntaxThemeGreenScreenID, Name: "MSX Green Screen", Palette: syntaxPalettes[syntaxThemeGreenScreenID], Builtin: true},
	syntaxThemeCobaltID:      {ID: syntaxThemeCobaltID, Name: "Cobalt", Palette: syntaxPalettes[syntaxThemeCobaltID], Builtin: true},
	syntaxThemeAmberID:       {ID: syntaxThemeAmberID, Name: "Amber", Palette: syntaxPalettes[syntaxThemeAmberID], Builtin: true},
}

var customSyntaxThemes = map[string]syntaxThemeDefinition{}

func normalizeEditorThemeID(id string) string {
	id = strings.TrimSpace(strings.ToLower(id))
	if _, ok := editorPalettes[id]; ok {
		return id
	}
	return editorThemeDarkID
}

func normalizeSyntaxThemeID(id string) string {
	id = strings.TrimSpace(strings.ToLower(id))
	if _, ok := syntaxThemeDefinitionByID(id); ok {
		return id
	}
	return defaultSyntaxThemeID
}

func defaultSyntaxThemeForEditor(editorThemeID string) string {
	if normalizeEditorThemeID(editorThemeID) == editorThemeLightID {
		return syntaxThemeLightID
	}
	return syntaxThemeDarkID
}

func syntaxPaletteByID(id string) syntaxPalette {
	def, ok := syntaxThemeDefinitionByID(normalizeSyntaxThemeID(id))
	if ok {
		return def.Palette
	}
	return builtinSyntaxThemes[defaultSyntaxThemeID].Palette
}

func syntaxThemeDefinitionByID(id string) (syntaxThemeDefinition, bool) {
	return syntaxThemeDefinitionByIDWithCustom(id, customSyntaxThemes)
}

func syntaxThemeDefinitionByIDWithCustom(id string, custom map[string]syntaxThemeDefinition) (syntaxThemeDefinition, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	if def, ok := custom[id]; ok {
		return def, true
	}
	def, ok := builtinSyntaxThemes[id]
	return def, ok
}

func syntaxThemeOptions() []syntaxThemeOption {
	return syntaxThemeOptionsForCustom(customSyntaxThemes)
}

func builtinSyntaxThemeOptions() []syntaxThemeOption {
	return []syntaxThemeOption{
		{Label: builtinSyntaxThemes[syntaxThemeDarkID].Name, ID: syntaxThemeDarkID},
		{Label: builtinSyntaxThemes[syntaxThemeLightID].Name, ID: syntaxThemeLightID},
		{Label: builtinSyntaxThemes[syntaxThemeGreenScreenID].Name, ID: syntaxThemeGreenScreenID},
		{Label: builtinSyntaxThemes[syntaxThemeCobaltID].Name, ID: syntaxThemeCobaltID},
		{Label: builtinSyntaxThemes[syntaxThemeAmberID].Name, ID: syntaxThemeAmberID},
	}
}

func syntaxThemeOptionsForCustom(custom map[string]syntaxThemeDefinition) []syntaxThemeOption {
	builtins := builtinSyntaxThemeOptions()
	customOpts := make([]syntaxThemeOption, 0, len(custom))
	for _, def := range custom {
		customOpts = append(customOpts, syntaxThemeOption{Label: def.Name, ID: def.ID})
	}
	sort.Slice(customOpts, func(i, j int) bool {
		return strings.ToLower(customOpts[i].Label) < strings.ToLower(customOpts[j].Label)
	})
	return append(builtins, customOpts...)
}

func syntaxThemeLabelForID(id string, options []syntaxThemeOption) string {
	id = normalizeSyntaxThemeID(id)
	for _, opt := range options {
		if opt.ID == id {
			return opt.Label
		}
	}
	return builtinSyntaxThemes[defaultSyntaxThemeID].Name
}

func syntaxThemeIDForLabel(label string, options []syntaxThemeOption) string {
	for _, opt := range options {
		if opt.Label == label {
			return opt.ID
		}
	}
	return defaultSyntaxThemeID
}

func fallbackSyntaxThemeSelection(currentID, editorThemeID string, custom map[string]syntaxThemeDefinition) string {
	if _, ok := syntaxThemeDefinitionByIDWithCustom(currentID, custom); ok {
		return currentID
	}
	preferred := defaultSyntaxThemeForEditor(editorThemeID)
	if _, ok := syntaxThemeDefinitionByIDWithCustom(preferred, custom); ok {
		return preferred
	}
	return defaultSyntaxThemeID
}

func deleteCustomSyntaxTheme(currentID, editorThemeID string, custom map[string]syntaxThemeDefinition) (string, error) {
	def, ok := syntaxThemeDefinitionByIDWithCustom(currentID, custom)
	if !ok {
		return fallbackSyntaxThemeSelection("", editorThemeID, custom), fmt.Errorf("theme %q not found", currentID)
	}
	if def.Builtin {
		return currentID, fmt.Errorf("builtin themes cannot be deleted")
	}
	delete(custom, def.ID)
	return fallbackSyntaxThemeSelection("", editorThemeID, custom), nil
}

func resetCustomSyntaxThemeToBuiltin(currentID, builtinID string, custom map[string]syntaxThemeDefinition) (syntaxThemeDefinition, error) {
	def, ok := syntaxThemeDefinitionByIDWithCustom(currentID, custom)
	if !ok {
		return syntaxThemeDefinition{}, fmt.Errorf("theme %q not found", currentID)
	}
	if def.Builtin {
		return syntaxThemeDefinition{}, fmt.Errorf("builtin themes cannot be reset")
	}
	builtinDef, ok := builtinSyntaxThemes[normalizeSyntaxThemeID(builtinID)]
	if !ok {
		return syntaxThemeDefinition{}, fmt.Errorf("builtin preset %q not found", builtinID)
	}
	def.Palette = builtinDef.Palette
	custom[def.ID] = def
	return def, nil
}

func syntaxThemeColorHexMap(p syntaxPalette) map[string]string {
	return map[string]string{
		"plain":       formatHexColor(p.Plain),
		"instruction": formatHexColor(p.Instruction),
		"jump":        formatHexColor(p.Jump),
		"function":    formatHexColor(p.Function),
		"operator":    formatHexColor(p.Operator),
		"number":      formatHexColor(p.Number),
		"string":      formatHexColor(p.String),
		"comment":     formatHexColor(p.Comment),
		"identifier":  formatHexColor(p.Identifier),
	}
}

func syntaxPaletteFromHexMap(colors map[string]string) (syntaxPalette, error) {
	get := func(key string) (color.NRGBA, error) {
		value, ok := colors[key]
		if !ok {
			return color.NRGBA{}, fmt.Errorf("missing color category %q", key)
		}
		return parseHexColor(value)
	}
	plain, err := get("plain")
	if err != nil {
		return syntaxPalette{}, err
	}
	instruction, err := get("instruction")
	if err != nil {
		return syntaxPalette{}, err
	}
	jump, err := get("jump")
	if err != nil {
		return syntaxPalette{}, err
	}
	function, err := get("function")
	if err != nil {
		return syntaxPalette{}, err
	}
	operator, err := get("operator")
	if err != nil {
		return syntaxPalette{}, err
	}
	number, err := get("number")
	if err != nil {
		return syntaxPalette{}, err
	}
	stringColor, err := get("string")
	if err != nil {
		return syntaxPalette{}, err
	}
	comment, err := get("comment")
	if err != nil {
		return syntaxPalette{}, err
	}
	identifier, err := get("identifier")
	if err != nil {
		return syntaxPalette{}, err
	}
	return syntaxPalette{
		Plain:       plain,
		Instruction: instruction,
		Jump:        jump,
		Function:    function,
		Operator:    operator,
		Number:      number,
		String:      stringColor,
		Comment:     comment,
		Identifier:  identifier,
	}, nil
}

func cloneCustomSyntaxThemes() map[string]syntaxThemeDefinition {
	return cloneSyntaxThemeDefinitions(customSyntaxThemes)
}

func cloneSyntaxThemeDefinitions(src map[string]syntaxThemeDefinition) map[string]syntaxThemeDefinition {
	if len(src) == 0 {
		return map[string]syntaxThemeDefinition{}
	}
	out := make(map[string]syntaxThemeDefinition, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func setCustomSyntaxThemes(themes map[string]syntaxThemeDefinition) {
	customSyntaxThemes = cloneSyntaxThemeDefinitions(themes)
}

func validateSyntaxThemeName(name, currentID string, custom map[string]syntaxThemeDefinition) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	for _, def := range builtinSyntaxThemes {
		if strings.EqualFold(def.Name, name) && def.ID != currentID {
			return fmt.Errorf("theme name %q already exists", name)
		}
	}
	for _, def := range custom {
		if strings.EqualFold(def.Name, name) && def.ID != currentID {
			return fmt.Errorf("theme name %q already exists", name)
		}
	}
	return nil
}

func makeCustomSyntaxTheme(name string, base syntaxThemeDefinition, custom map[string]syntaxThemeDefinition) (syntaxThemeDefinition, error) {
	name = uniqueSyntaxThemeName(strings.TrimSpace(name), custom, "")
	if err := validateSyntaxThemeName(name, "", custom); err != nil {
		return syntaxThemeDefinition{}, err
	}
	return syntaxThemeDefinition{
		ID:      makeSyntaxThemeID(name, custom, ""),
		Name:    name,
		Palette: base.Palette,
		Builtin: false,
	}, nil
}

func uniqueSyntaxThemeName(base string, custom map[string]syntaxThemeDefinition, excludeID string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "Custom Theme"
	}
	name := base
	for i := 2; ; i++ {
		if validateSyntaxThemeName(name, excludeID, custom) == nil {
			return name
		}
		name = fmt.Sprintf("%s %d", base, i)
	}
}

func makeSyntaxThemeID(name string, custom map[string]syntaxThemeDefinition, excludeID string) string {
	base := slugifySyntaxThemeName(name)
	if base == "" {
		base = "custom-theme"
	}
	id := base
	for i := 2; ; i++ {
		if excludeID != "" && id == excludeID {
			return id
		}
		if _, ok := builtinSyntaxThemes[id]; ok {
			id = fmt.Sprintf("%s-%d", base, i)
			continue
		}
		if _, ok := custom[id]; ok {
			id = fmt.Sprintf("%s-%d", base, i)
			continue
		}
		return id
	}
}

func slugifySyntaxThemeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func serializeCustomSyntaxThemes(themes map[string]syntaxThemeDefinition) (string, error) {
	items := make([]syntaxThemeFile, 0, len(themes))
	for _, def := range themes {
		if def.Builtin {
			continue
		}
		items = append(items, syntaxThemeFile{
			Version: 1,
			ID:      def.ID,
			Name:    def.Name,
			Colors:  syntaxThemeColorHexMap(def.Palette),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	b, err := json.Marshal(syntaxThemeCatalogFile{Version: 1, Themes: items})
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func parseCustomSyntaxThemesJSON(raw string) (map[string]syntaxThemeDefinition, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]syntaxThemeDefinition{}, nil
	}
	var catalog syntaxThemeCatalogFile
	if err := json.Unmarshal([]byte(raw), &catalog); err != nil {
		return nil, err
	}
	custom := make(map[string]syntaxThemeDefinition, len(catalog.Themes))
	for _, item := range catalog.Themes {
		palette, err := syntaxPaletteFromHexMap(item.Colors)
		if err != nil {
			return nil, err
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			return nil, fmt.Errorf("custom theme name cannot be empty")
		}
		def := syntaxThemeDefinition{
			ID:      makeSyntaxThemeID(coalesceSyntaxThemeID(item.ID, name), custom, ""),
			Name:    uniqueSyntaxThemeName(name, custom, ""),
			Palette: palette,
			Builtin: false,
		}
		custom[def.ID] = def
	}
	return custom, nil
}

func exportSyntaxThemeJSON(def syntaxThemeDefinition) ([]byte, error) {
	return json.MarshalIndent(syntaxThemeFile{
		Version: 1,
		ID:      def.ID,
		Name:    def.Name,
		Colors:  syntaxThemeColorHexMap(def.Palette),
	}, "", "  ")
}

func importSyntaxThemeJSON(data []byte, custom map[string]syntaxThemeDefinition) (syntaxThemeDefinition, error) {
	var item syntaxThemeFile
	if err := json.Unmarshal(data, &item); err != nil {
		return syntaxThemeDefinition{}, err
	}
	palette, err := syntaxPaletteFromHexMap(item.Colors)
	if err != nil {
		return syntaxThemeDefinition{}, err
	}
	name := uniqueSyntaxThemeName(strings.TrimSpace(item.Name), custom, "")
	if err := validateSyntaxThemeName(name, "", custom); err != nil {
		return syntaxThemeDefinition{}, err
	}
	baseID := item.ID
	if strings.TrimSpace(baseID) == "" || strings.HasPrefix(baseID, "msx-") {
		baseID = name
	}
	return syntaxThemeDefinition{
		ID:      makeSyntaxThemeID(baseID, custom, ""),
		Name:    name,
		Palette: palette,
		Builtin: false,
	}, nil
}

func coalesceSyntaxThemeID(id, fallbackName string) string {
	id = slugifySyntaxThemeName(id)
	if id != "" {
		return id
	}
	return fallbackName
}

func formatHexColor(c color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func parseHexColor(value string) (color.NRGBA, error) {
	value = strings.TrimSpace(strings.TrimPrefix(value, "#"))
	if len(value) != 6 {
		return color.NRGBA{}, fmt.Errorf("invalid color %q: expected RRGGBB", value)
	}
	var c color.NRGBA
	if _, err := fmt.Sscanf(strings.ToUpper(value), "%02X%02X%02X", &c.R, &c.G, &c.B); err != nil {
		return color.NRGBA{}, fmt.Errorf("invalid color %q", value)
	}
	c.A = 0xFF
	return c, nil
}

type sourceCodeProTheme struct {
	fyne.Theme
	font         fyne.Resource // regular weight
	fontBold     fyne.Resource // bold weight
	fontItalic   fyne.Resource // italic style
	fontBoldItal fyne.Resource // bold+italic style
	palette      editorPalette
	textSize     float32
}

func availableEditorFontFamilies() []string {
	families := []string{defaultEditorFontFamily, "MSX Screen 0", "MSX Screen 1"}
	sort.Strings(families)
	return families
}

func editorFontWeightsForFamily(family string) []string {
	switch normalizeEditorFontFamily(family) {
	case "MSX Screen 0", "MSX Screen 1":
		return []string{"Regular"}
	default:
		return []string{"ExtraLight", "Light", "Regular", "Medium", "SemiBold", "Bold", "ExtraBold", "Black"}
	}
}

func editorFontFamilySupportsItalic(family string) bool {
	return normalizeEditorFontFamily(family) == defaultEditorFontFamily
}

func normalizeEditorFontFamily(family string) string {
	f := strings.TrimSpace(family)
	for _, known := range availableEditorFontFamilies() {
		if strings.EqualFold(known, f) {
			return known
		}
	}
	return defaultEditorFontFamily
}

func normalizeEditorFontWeight(family, weight string) string {
	family = normalizeEditorFontFamily(family)
	w := strings.TrimSpace(weight)
	for _, known := range editorFontWeightsForFamily(family) {
		if strings.EqualFold(known, w) {
			return known
		}
	}
	return defaultEditorFontWeight
}

func normalizeEditorFontSize(size float32) float32 {
	if size < 8 || size > 48 {
		return defaultEditorFontSize
	}
	return size
}

func nextHeavierWeight(weight string) string {
	order := []string{"ExtraLight", "Light", "Regular", "Medium", "SemiBold", "Bold", "ExtraBold", "Black"}
	for i, w := range order {
		if strings.EqualFold(w, weight) {
			if i+1 < len(order) {
				return order[i+1]
			}
			return order[i]
		}
	}
	return "Bold"
}

func fontFileName(family, weight string, italic bool) string {
	family = normalizeEditorFontFamily(family)
	weight = normalizeEditorFontWeight(family, weight)
	suffix := ""
	if italic {
		suffix = "Italic"
	}
	switch family {
	case "MSX Screen 0":
		return "MSX-Screen0.ttf"
	case "MSX Screen 1":
		return "MSX-Screen1.ttf"
	default:
		return "SourceCodePro-" + weight + suffix + ".ttf"
	}
}

func loadFontResource(path string) fyne.Resource {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(filepath.Base(path), b)
}

func newConfiguredEditorTheme(resDir, editorThemeID, family, weight string, textSize float32) (fyne.Theme, error) {
	editorThemeID = normalizeEditorThemeID(editorThemeID)
	palette := editorPalettes[editorThemeID]

	base := theme.DarkTheme()
	if editorThemeID == editorThemeLightID {
		base = theme.LightTheme()
	}

	family = normalizeEditorFontFamily(family)
	weight = normalizeEditorFontWeight(family, weight)
	textSize = normalizeEditorFontSize(textSize)

	regular := loadFontResource(filepath.Join(resDir, fontFileName(family, weight, false)))
	italic := loadFontResource(filepath.Join(resDir, fontFileName(family, weight, true)))
	bold := loadFontResource(filepath.Join(resDir, fontFileName(family, nextHeavierWeight(weight), false)))
	boldItalic := loadFontResource(filepath.Join(resDir, fontFileName(family, nextHeavierWeight(weight), true)))

	if bold == nil {
		bold = regular
	}
	if italic == nil {
		italic = regular
	}
	if boldItalic == nil {
		if italic != nil {
			boldItalic = italic
		} else {
			boldItalic = bold
		}
	}

	if regular == nil && bold == nil && italic == nil && boldItalic == nil {
		return &sourceCodeProTheme{Theme: base, palette: palette, textSize: textSize}, nil
	}
	return &sourceCodeProTheme{
		Theme:        base,
		font:         regular,
		fontBold:     bold,
		fontItalic:   italic,
		fontBoldItal: boldItalic,
		palette:      palette,
		textSize:     textSize,
	}, nil
}

func newSourceCodeProTheme(fontPath, editorThemeID string) (fyne.Theme, error) {
	return newConfiguredEditorTheme(filepath.Dir(fontPath), editorThemeID, defaultEditorFontFamily, defaultEditorFontWeight, defaultEditorFontSize)
}

func (t *sourceCodeProTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold && style.Italic && t.fontBoldItal != nil {
		return t.fontBoldItal
	}
	if style.Bold && t.fontBold != nil {
		return t.fontBold
	}
	if style.Italic && t.fontItalic != nil {
		return t.fontItalic
	}
	if t.font != nil {
		return t.font
	}
	return t.Theme.Font(style)
}

func (t *sourceCodeProTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText && t.textSize > 0 {
		return t.textSize
	}
	return t.Theme.Size(name)
}

func (t *sourceCodeProTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.palette.Background
	case theme.ColorNameInputBackground:
		return t.palette.InputBackground
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return t.palette.MenuBackground
	case theme.ColorNameForeground:
		return t.palette.Foreground
	case theme.ColorNamePrimary:
		return t.palette.Primary
	case theme.ColorNameSelection, theme.ColorNameFocus, theme.ColorNameHover:
		c := t.palette.Primary
		return color.NRGBA{R: c.R, G: c.G, B: c.B, A: 0x44}
	}
	return t.Theme.Color(name, variant)
}
