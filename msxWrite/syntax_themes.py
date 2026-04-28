from __future__ import annotations

# Nomes de chaves padrão para as cores no banco de dados
# Usaremos o prefixo "syntax_color_" para as cores de sintaxe
# e "syntax_theme" para o nome do tema selecionado.

SYNTAX_THEMES = {
    "Classico": {
        "command": "#2E6F9E",
        "function": "#2B7A5B",
        "string": "#B54D2B",
        "number": "#7A3E9D",
        "comment": "#3F6A3F",
        "line_number": "#6B6B6B",
        "bg": "#FFFFFF",
        "fg": "#000000",
    },
    "Neon": {
        "command": "#00B3FF",
        "function": "#00D18B",
        "string": "#FF8A00",
        "number": "#A15CFF",
        "comment": "#00A388",
        "line_number": "#8A8A8A",
        "bg": "#1A1A1A",
        "fg": "#FFFFFF",
    },
    "Papel": {
        "command": "#2C4E75",
        "function": "#2F6A52",
        "string": "#8B3A2A",
        "number": "#5E3D77",
        "comment": "#546E3B",
        "line_number": "#7C6F64",
        "bg": "#FDF6E3",
        "fg": "#657B83",
    },
    "Dark": {
        "command": "#569CD6",
        "function": "#DCDCAA",
        "string": "#CE9178",
        "number": "#B5CEA8",
        "comment": "#6A9955",
        "line_number": "#858585",
        "bg": "#2B2B2B",
        "fg": "#FFFFFF",
    }
}

DEFAULT_SYNTAX_THEME = "Classico"

def get_syntax_colors(db) -> dict[str, str]:
    """Recupera as cores de sintaxe do banco de dados ou usa o padrão."""
    theme_name = db.get_setting("syntax_theme", DEFAULT_SYNTAX_THEME)
    if theme_name not in SYNTAX_THEMES:
        theme_name = DEFAULT_SYNTAX_THEME
        
    theme = SYNTAX_THEMES[theme_name].copy()
    
    for key in theme.keys():
        saved = db.get_setting(f"syntax_color_{key}", "")
        if saved:
            theme[key] = saved
            
    return theme

def save_syntax_colors(db, theme_name: str, colors: dict[str, str]):
    """Salva as cores de sintaxe no banco de dados."""
    db.set_setting("syntax_theme", theme_name)
    for key, value in colors.items():
        db.set_setting(f"syntax_color_{key}", value)
