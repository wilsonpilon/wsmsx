from __future__ import annotations

import tkinter as tk
import tkinter.font as tkfont
import re
from pathlib import Path
from tkinter import filedialog, messagebox

import customtkinter as ctk

from msx_basic_decoder import decode_msx_basic_segments
from help_viewer import HelpViewer
from msx_encoding_viewer import MSXEncodingViewer
from syntax_themes import SYNTAX_THEMES, DEFAULT_SYNTAX_THEME, get_syntax_colors, save_syntax_colors

# Paleta de cores do MSX1 (0-15)
# Fonte: https://paulwratt.github.io/programmers-palettes/HW-MSX/HW-MSX-palettes.html
# Cada item: (índice, nome, cor_hex)
MSX1_PALETTE: list[tuple[int, str, str]] = [
    (0,  "transparent",  "#000000"),
    (1,  "black",        "#010101"),
    (2,  "medium green", "#3EB849"),
    (3,  "light green",  "#74D07D"),
    (4,  "dark blue",    "#5955E0"),
    (5,  "light blue",   "#8076F1"),
    (6,  "dark red",     "#B95E51"),
    (7,  "cyan",         "#65DBEF"),
    (8,  "medium red",   "#DB6559"),
    (9,  "light red",    "#FF897D"),
    (10, "dark yellow",  "#CCC35E"),
    (11, "light yellow", "#DED087"),
    (12, "dark green",   "#3AA241"),
    (13, "magenta",      "#B766B5"),
    (14, "gray",         "#CCCCCC"),
    (15, "white",        "#FFFFFF"),
]


class LineNumbers(tk.Canvas):
    def __init__(self, master, font, editor, **kwargs):
        super().__init__(master, width=50, highlightthickness=0, bg="#2b2b2b", **kwargs)
        self.font = font
        self.editor = editor
        self.textbox = None

    def set_textbox(self, textbox):
        self.textbox = textbox

    def redraw(self):
        self.delete("all")
        if not self.textbox:
            return

        # Update background color from settings
        bg_color = self.editor.settings.get("color_bg", "#2b2b2b")
        self.configure(bg=bg_color)

        i = self.textbox.index("@0,0")
        while True:
            dline = self.textbox._textbox.dlineinfo(i)
            if dline is None:
                break
            y = dline[1]
            linenum = str(i).split(".")[0]
            
            # Highlight current line
            cursor_pos = self.textbox.index(tk.INSERT)
            current_linenum = cursor_pos.split(".")[0]
            
            fg = self.editor.settings.get("color_line_number", "#858585")
            if linenum == current_linenum:
                fg = self.editor.settings.get("color_fg", "#ffffff")
                
            self.create_text(45, y, anchor="ne", text=linenum, font=self.font, fill=fg)
            i = self.textbox.index(f"{i} + 1line")


class Ruler(tk.Canvas):
    def __init__(self, master, font, editor, **kwargs):
        super().__init__(master, height=25, highlightthickness=0, bg="#2b2b2b", **kwargs)
        self.font = font
        self.editor = editor
        self.textbox = None

    def set_textbox(self, textbox):
        self.textbox = textbox

    def redraw(self):
        self.delete("all")
        if not self.textbox:
            return

        # Update background color from settings
        bg_color = self.editor.settings.get("color_bg", "#2b2b2b")
        self.configure(bg=bg_color)

        # Get character width from internal textbox font
        font_val = self.textbox._textbox.cget("font")
        f = tkfont.Font(font=font_val)
        char_width = f.measure("0")
        
        # Current cursor column for highlighting
        cursor_pos = self.textbox.index(tk.INSERT)
        _, current_col = map(int, cursor_pos.split("."))
        
        # Numbers to highlight: 0, 32, 40, 80
        highlights = {0, 32, 40, 80}
        
        # Horizontal scroll position
        x_offset = self.textbox._textbox.xview()[0]
        # We need the total width of the text area to convert x_offset to pixels
        # But bbox is easier if we have text.
        
        # Use bbox to find the x position of column 0
        bbox = self.textbox._textbox.bbox("1.0")
        if bbox:
            start_x = bbox[0]
        else:
            # Fallback if text is empty or bbox fails
            # If bbox is None, it might be because the text is scrolled out of view
            # or the textbox is empty.
            # If it's scrolled out of view, we can't easily use bbox.
            # But we can try bbox for a visible line.
            
            # Get the first visible index
            first_visible = self.textbox.index("@0,0")
            line_start = first_visible.split(".")[0] + ".0"
            bbox = self.textbox._textbox.bbox(line_start)
            if bbox:
                start_x = bbox[0]
            else:
                # Absolute fallback - usually CTkTextbox has 2px padding
                # But it's better to be dynamic. 
                # If we are at scroll 0, it's likely 2.
                if x_offset == 0:
                    start_x = 2
                else:
                    # If we are scrolled, we have to estimate or use a different method.
                    # Since we want it to align with the text, if we can't find the text,
                    # we might as well not draw or draw at a best guess.
                    start_x = 2 - (x_offset * self.textbox._textbox.winfo_width()) # This is often not accurate
        
        # Adjust start_x for the fact that Ruler is grid-aligned with Textbox
        # Both are in column 1 of editor_frame.
        
        # Draw marks and numbers
        for col in range(120): # Draw up to 120 columns
            x = start_x + (col * char_width)
            
            if col in highlights:
                fg = self.editor.settings.get("color_command", "#569CD6") # Highlight color
                weight = "bold"
                # Small line mark
                self.create_line(x, 15, x, 25, fill=fg, width=2)
                # Number
                # For fixed-width fonts, we want the number to be centered or aligned exactly.
                # anchor="n" centers the text horizontally at x.
                self.create_text(x, 5, text=str(col), font=(self.font[0], 10, weight), fill=fg, anchor="n")
            elif col % 10 == 0:
                self.create_line(x, 20, x, 25, fill="#555555")
                if col not in highlights:
                    self.create_text(x, 8, text=str(col), font=(self.font[0], 8), fill="#858585", anchor="n")
            elif col % 5 == 0:
                self.create_line(x, 22, x, 25, fill="#555555")

            # Cursor highlight on ruler
            if col == current_col:
                cursor_fg = self.editor.settings.get("color_string", "#CE9178")
                self.create_line(x, 0, x, 25, fill=cursor_fg, width=1, dash=(2, 2))


class MSXBasicEditor(ctk.CTk):
    def __init__(self) -> None:
        super().__init__()

        from app_db import AppDatabase
        self.db = AppDatabase(Path("msxread.db"))

        # Default settings
        self.settings = {
            "dialect": "MSX-BASIC",
            "start_line": "10",
            "increment": "10",
            "keep_case": "False",
            "openmsx_path": "",
            "fmsx_path": "",
            "extra_configs": ""
        }
        self._load_settings()
        
        # Load syntax colors from unified theme
        self.syntax_colors = get_syntax_colors(self.db)
        for key, value in self.syntax_colors.items():
            self.settings[f"color_{key}"] = value

        self.title("MSX-Write - Editor MSX BASIC")
        self.geometry("1000x700")
        
        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)

        self._build_ui()
        self._setup_syntax_highlighting()

    def _build_ui(self) -> None:
        # Menu
        self.option_add("*tearOff", tk.FALSE)
        self.menubar = tk.Menu(self)
        self.config(menu=self.menubar)

        # File Menu
        self.file_menu = tk.Menu(self.menubar)
        self.menubar.add_cascade(label="Arquivo", menu=self.file_menu)
        self.file_menu.add_command(label="Novo", command=self._clear_editor)
        self.file_menu.add_command(label="Abrir...", command=self._open_file)
        self.file_menu.add_command(label="Salvar", command=self._save_file)
        self.file_menu.add_separator()
        self.file_menu.add_command(label="Sair", command=self.destroy)

        # Edit Menu
        self.edit_menu = tk.Menu(self.menubar)
        self.menubar.add_cascade(label="Editar", menu=self.edit_menu)
        self.edit_menu.add_command(label="Desfazer", command=lambda: self.textbox.event_generate("<<Undo>>"))
        self.edit_menu.add_command(label="Refazer", command=lambda: self.textbox.event_generate("<<Redo>>"))
        self.edit_menu.add_separator()
        self.edit_menu.add_command(label="Localizar...", command=self._on_find)
        self.edit_menu.add_command(label="Substituir...", command=self._on_replace)
        self.edit_menu.add_separator()
        self.edit_menu.add_command(label="Renumerar (RENUM)", command=self._on_renum)

        # Tools Menu
        self.tools_menu = tk.Menu(self.menubar)
        self.menubar.add_cascade(label="Ferramentas", menu=self.tools_menu)
        self.tools_menu.add_command(label="Remover Números de Linha", command=self._remove_line_numbers)
        self.tools_menu.add_command(label="Adicionar Números de Linha", command=self._add_line_numbers)
        self.tools_menu.add_command(label="Formatar Código (Beautify)", command=self._on_beautify_all)
        self.tools_menu.add_command(label="Mapa do Programa", command=self._on_program_map)
        
        # Submenu: Cores MSX1
        self.msx_colors_menu = tk.Menu(self.tools_menu)
        self.tools_menu.add_cascade(label="Cores MSX1", menu=self.msx_colors_menu)
        for idx, name, hex_color in MSX1_PALETTE:
            self.msx_colors_menu.add_command(
                label=f"{idx} - {name} ({hex_color})",
                command=lambda i=idx: self._show_msx1_palette(i)
            )
        self.msx_colors_menu.add_separator()
        self.msx_colors_menu.add_command(label="Mostrar Paleta...", command=self._show_msx1_palette)

        self.tools_menu.add_separator()
        self.tools_menu.add_command(label="Configurações", command=self._on_settings)

        # Help Menu
        self.help_menu = tk.Menu(self.menubar)
        self.menubar.add_cascade(label="Ajuda", menu=self.help_menu)
        self.help_menu.add_command(label="Manuais MSX", command=self._open_help_viewer)
        self.help_menu.add_separator()
        self.help_menu.add_command(label="Sobre", command=lambda: messagebox.showinfo("Sobre", "MSX-Write Editor"))

        # Toolbar
        toolbar = ctk.CTkFrame(self)
        toolbar.grid(row=0, column=0, sticky="ew", padx=10, pady=5)

        btn_open = ctk.CTkButton(toolbar, text="Abrir", width=80, command=self._open_file)
        btn_open.grid(row=0, column=0, padx=2, pady=2)

        btn_save = ctk.CTkButton(toolbar, text="Salvar", width=80, command=self._save_file)
        btn_save.grid(row=0, column=1, padx=2, pady=2)

        btn_renum = ctk.CTkButton(toolbar, text="RENUM", width=80, command=self._on_renum)
        btn_renum.grid(row=0, column=2, padx=2, pady=2)

        btn_settings = ctk.CTkButton(toolbar, text="Config", width=80, command=self._on_settings)
        btn_settings.grid(row=0, column=3, padx=2, pady=2)

        btn_viewer = ctk.CTkButton(toolbar, text="msxRead (Viewer)", width=120, command=self._open_viewer)
        btn_viewer.grid(row=0, column=4, padx=2, pady=2)

        btn_help = ctk.CTkButton(toolbar, text="Ajuda", width=80, command=self._open_help_viewer)
        btn_help.grid(row=0, column=5, padx=2, pady=2)
        
        btn_encoding = ctk.CTkButton(toolbar, text="Encoding", width=80, command=self._open_encoding_viewer)
        btn_encoding.grid(row=0, column=6, padx=2, pady=2)


        # Editor Area
        editor_frame = ctk.CTkFrame(self)
        editor_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=(0, 5))
        editor_frame.grid_columnconfigure(1, weight=1)
        editor_frame.grid_rowconfigure(1, weight=1)

        # Ruler (at the top of the textbox)
        self.ruler = Ruler(editor_frame, font=("Consolas", 14), editor=self)
        self.ruler.grid(row=0, column=1, sticky="ew")

        # Line Numbers (at the left of the textbox)
        self.line_numbers = LineNumbers(editor_frame, font=("Consolas", 14), editor=self)
        self.line_numbers.grid(row=1, column=0, sticky="ns")
        
        self.line_numbers.configure(bg=self.settings.get("color_bg", "#2b2b2b"))
        self.ruler.configure(bg=self.settings.get("color_bg", "#2b2b2b"))

        self.textbox = ctk.CTkTextbox(editor_frame, wrap="none", font=("Consolas", 14), undo=True)
        self.textbox.grid(row=1, column=1, sticky="nsew")
        
        # Link Line Numbers and Ruler to Textbox
        self.line_numbers.set_textbox(self.textbox)
        self.ruler.set_textbox(self.textbox)

        # Status Bar
        self.status_bar = ctk.CTkLabel(self, text="Linha: 1, Coluna: 0", anchor="w")
        self.status_bar.grid(row=2, column=0, sticky="ew", padx=10, pady=(0, 5))

        self.textbox.bind("<<Modified>>", self._on_text_modified)
        self.textbox.bind("<KeyRelease-space>", self._on_key_beautify)
        self.textbox.bind("<KeyRelease-Return>", self._on_key_beautify)
        
        # Cursor position update
        self.textbox.bind("<KeyRelease>", self._update_status_bar)
        self.textbox.bind("<ButtonRelease-1>", self._update_status_bar)

        # Keybinding: SHIFT+HOME -> ir para início absoluto e listar do topo
        self.textbox.bind("<Shift-Home>", self._on_shift_home)
        
        # Keybindings: CONTROL+B e CONTROL+F -> navegação por palavras
        self.textbox.bind("<Control-b>", self._on_ctrl_b)
        self.textbox.bind("<Control-f>", self._on_ctrl_f)

        # Keybinding: CONTROL+E -> apagar até o fim da linha
        self.textbox.bind("<Control-e>", self._on_ctrl_e)
        
        # Keybinding: CONTROL+H -> backspace
        self.textbox.bind("<Control-h>", self._on_ctrl_h)

        # Keybinding: CONTROL+I -> ir para próxima tabulação (8 colunas)
        self.textbox.bind("<Control-i>", self._on_ctrl_i)
        
        # Keybinding: CONTROL+J -> ir para próxima linha
        self.textbox.bind("<Control-j>", self._on_ctrl_j)

        # Keybinding: CONTROL+K -> mover cursor para o topo da tela visível
        self.textbox.bind("<Control-k>", self._on_ctrl_k)
        self.textbox._textbox.bind("<Control-k>", self._on_ctrl_k)

        # Keybinding: CONTROL+L -> idêntico a SHIFT+HOME (início absoluto)
        self.textbox.bind("<Control-l>", self._on_shift_home)
        self.textbox._textbox.bind("<Control-l>", self._on_shift_home)

        # Keybinding: CONTROL+M -> Enter
        self.textbox.bind("<Control-m>", self._on_ctrl_m)
        self.textbox._textbox.bind("<Control-m>", self._on_ctrl_m)
        
        # Keybinding: CONTROL+N -> ir para o final da linha
        self.textbox.bind("<Control-n>", self._on_ctrl_n)
        self.textbox._textbox.bind("<Control-n>", self._on_ctrl_n)
        
        # Sync scrolling for line numbers and ruler
        self.textbox._textbox.configure(yscrollcommand=self._on_textbox_scroll_y, xscrollcommand=self._on_textbox_scroll_x)

    def _on_textbox_scroll_y(self, *args) -> None:
        # Standard yscrollcommand handling
        self.textbox._y_scrollbar.set(*args)
        self.line_numbers.redraw()

    def _on_textbox_scroll_x(self, *args) -> None:
        self.textbox._x_scrollbar.set(*args)
        self.ruler.redraw()

    def _on_shift_home(self, event=None) -> str | None:
        """
        SHIFT+HOME: mover o cursor para a coluna 1, linha 1 e mostrar o início do programa na tela.
        """
        try:
            # Posiciona o cursor no início absoluto
            self.textbox.mark_set(tk.INSERT, "1.0")
            # Garante que o topo fique visível (lista a partir do início)
            try:
                # Preferir API direta do CTkTextbox, se disponível
                self.textbox.see("1.0")
            except Exception:
                # Fallback para o widget interno do tkinter.Text
                self.textbox._textbox.see("1.0")
            # Atualiza status/régua/linhas
            self._update_status_bar()
            self.textbox.focus_set()
            return "break"  # impede o comportamento padrão do Tk
        except Exception:
            # Em caso de qualquer falha, não bloqueia o comportamento padrão
            return None

    def _on_ctrl_b(self, event=None) -> str:
        """
        CONTROL+B: move o cursor para o início da palavra anterior.
        """
        self.textbox.mark_set(tk.INSERT, self.textbox.index("insert -1c wordstart"))
        self.textbox.see(tk.INSERT)
        self._update_status_bar()
        return "break"

    def _on_ctrl_f(self, event=None) -> str:
        """
        CONTROL+F: move o cursor para o fim da palavra seguinte.
        """
        self.textbox.mark_set(tk.INSERT, self.textbox.index("insert +1c wordend"))
        self.textbox.see(tk.INSERT)
        self._update_status_bar()
        return "break"

    def _on_ctrl_e(self, event=None) -> str:
        """
        CONTROL+E: apaga da posição do cursor até o final da linha.
        """
        self.textbox.delete(tk.INSERT, "insert lineend")
        self._update_status_bar()
        return "break"

    def _on_ctrl_h(self, event=None) -> str:
        """
        CONTROL+H: volta o cursor apagando como se fosse backspace.
        """
        try:
            # Se houver seleção, apaga a seleção
            if self.textbox.tag_ranges(tk.SEL):
                self.textbox.delete(tk.SEL_FIRST, tk.SEL_LAST)
            else:
                # Caso contrário, apaga o caractere anterior
                # "insert -1c" refere-se ao caractere imediatamente antes do cursor
                if self.textbox.index(tk.INSERT) != "1.0":
                    self.textbox.delete("insert -1c", tk.INSERT)
            
            self.textbox.see(tk.INSERT)
            self._update_status_bar()
        except Exception:
            pass
        return "break"

    def _on_ctrl_i(self, event=None) -> str:
        """
        CONTROL+I: move o cursor para a próxima tabulação (8 colunas).
        Apenas move o cursor, não insere caractere.
        """
        try:
            # Obtém a linha e coluna atual
            line, col = map(int, self.textbox.index(tk.INSERT).split('.'))
            
            # Calcula a próxima tabulação (múltiplos de 8)
            # No Tkinter a coluna 0 é a primeira. 
            # Tabulações MSX padrão: 0, 8, 16, 24, 32, 40, 48, 56, 64, 72...
            next_col = ((col // 8) + 1) * 8
            
            # Define a nova posição
            new_index = f"{line}.{next_col}"
            
            # Move o cursor
            self.textbox.mark_set(tk.INSERT, new_index)
            self.textbox.see(tk.INSERT)
            self._update_status_bar()
        except Exception:
            pass
        return "break"

    def _on_ctrl_j(self, event=None) -> str:
        """
        CONTROL+J: move o cursor para a próxima linha.
        """
        try:
            # Move o cursor para a linha de baixo, mantendo a coluna se possível
            # "+1line" é o modificador do Tkinter para ir para a próxima linha
            new_index = self.textbox.index("insert +1line")
            self.textbox.mark_set(tk.INSERT, new_index)
            self.textbox.see(tk.INSERT)
            self._update_status_bar()
        except Exception:
            pass
        return "break"

    def _on_ctrl_k(self, event=None) -> str:
        """
        CONTROL+K: move o cursor para o topo da tela visível (primeira linha visível),
        mantendo a coluna atual.
        """
        try:
            # Obtém a linha visível no topo (@0,0 retorna o índice no canto superior esquerdo)
            top_index = self.textbox._textbox.index("@0,0")
            top_line = int(top_index.split('.')[0])
            
            # Obtém a coluna atual do cursor
            _, current_col = map(int, self.textbox.index(tk.INSERT).split('.'))
            
            # Define a nova posição: mesma coluna, mas na linha do topo
            new_index = f"{top_line}.{current_col}"
            
            # Move o cursor
            self.textbox.mark_set(tk.INSERT, new_index)
            # Não chamamos see() aqui porque queremos "manter o display do texto"
            # como solicitado ("mantendo a posicao e o display do texto")
            self._update_status_bar()
        except Exception:
            pass
        return "break"

    def _on_ctrl_m(self, event=None) -> str:
        """
        CONTROL+M: funciona igual ao ENTER.
        """
        try:
            # Insere uma nova linha na posição do cursor
            self.textbox.insert(tk.INSERT, "\n")
            self.textbox.see(tk.INSERT)
            
            # Simula o evento de soltar a tecla Return para acionar o beautify se necessário
            # Criamos um evento fake para o _on_key_beautify
            class FakeEvent:
                def __init__(self):
                    self.keysym = "Return"
            
            self._on_key_beautify(FakeEvent())
            self._update_status_bar()
        except Exception:
            pass
        return "break"

    def _on_ctrl_n(self, event=None) -> str:
        """
        CONTROL+N: move o cursor para o final da linha.
        """
        try:
            # "insert lineend" move o cursor para o final da linha atual no Tkinter
            new_index = self.textbox.index("insert lineend")
            self.textbox.mark_set(tk.INSERT, new_index)
            self.textbox.see(tk.INSERT)
            self._update_status_bar()
        except Exception:
            pass
        return "break"

    def _update_status_bar(self, event=None) -> None:
        cursor_pos = self.textbox.index(tk.INSERT)
        line, col = cursor_pos.split(".")
        self.status_bar.configure(text=f"Linha: {line}, Coluna: {col}")
        self.line_numbers.redraw()
        self.ruler.redraw()

    def _setup_syntax_highlighting(self) -> None:
        self.textbox.tag_config("keyword", foreground=self.settings.get("color_command", "#569CD6"))
        self.textbox.tag_config("comment", foreground=self.settings.get("color_comment", "#6A9955"))
        self.textbox.tag_config("string", foreground=self.settings.get("color_string", "#CE9178"))
        self.textbox.tag_config("number", foreground=self.settings.get("color_number", "#B5CEA8"))
        self.textbox.tag_config("linenumber", foreground=self.settings.get("color_line_number", "#858585"))
        self.textbox.tag_config("function", foreground=self.settings.get("color_function", "#DCDCAA"))
        
        # Apply BG/FG to textbox
        bg = self.settings.get("color_bg", "#2b2b2b")
        fg = self.settings.get("color_fg", "#ffffff")
        self.textbox.configure(fg_color=bg, text_color=fg)

    def _on_text_modified(self, event=None) -> None:
        if self.textbox.edit_modified():
            self._apply_syntax_highlighting()
            self.textbox.edit_modified(False)

    def _apply_syntax_highlighting(self) -> None:
        if hasattr(self, "line_numbers"):
            self.line_numbers.redraw()
        content = self.textbox.get("1.0", tk.END)
        # Clear existing tags
        for tag in ["keyword", "comment", "string", "number", "linenumber", "function"]:
            self.textbox.tag_remove(tag, "1.0", tk.END)

        from msx_basic_decoder import TOKEN_MAP, TOKEN_MAP_FF

        keywords = set(TOKEN_MAP)
        functions = set(TOKEN_MAP_FF)
        
        lines = content.split("\n")
        for i, line in enumerate(lines):
            line_idx = i + 1
            
            # Line numbers
            match_ln = re.match(r"^\s*(\d+)", line)
            if match_ln:
                start = f"{line_idx}.{line.find(match_ln.group(1))}"
                end = f"{line_idx}.{len(match_ln.group(1)) + line.find(match_ln.group(1))}"
                self.textbox.tag_add("linenumber", start, end)

            # Comments (REM and ')
            for match in re.finditer(r"(REM.*|'.*)", line, re.IGNORECASE):
                self.textbox.tag_add("comment", f"{line_idx}.{match.start()}", f"{line_idx}.{match.end()}")
            
            # Strings
            for match in re.finditer(r'("[^"]*")', line):
                self.textbox.tag_add("string", f"{line_idx}.{match.start()}", f"{line_idx}.{match.end()}")

            # Keywords and Functions
            for word_match in re.finditer(r"\b[A-Z$]+\b", line, re.IGNORECASE):
                word = word_match.group(0).upper()
                if word in keywords:
                    self.textbox.tag_add("keyword", f"{line_idx}.{word_match.start()}", f"{line_idx}.{word_match.end()}")
                elif word in functions:
                    self.textbox.tag_add("function", f"{line_idx}.{word_match.start()}", f"{line_idx}.{word_match.end()}")

            # Numbers
            for num_match in re.finditer(r"\b\d+\b", line):
                # Avoid tagging if it's already a line number (at start of line)
                if not (match_ln and num_match.start() == line.find(match_ln.group(1))):
                    self.textbox.tag_add("number", f"{line_idx}.{num_match.start()}", f"{line_idx}.{num_match.end()}")

    def _on_find(self) -> None:
        dialog = ctk.CTkToplevel(self)
        dialog.title("Localizar")
        dialog.geometry("300x150")
        dialog.attributes("-topmost", True)
        
        ctk.CTkLabel(dialog, text="Localizar:").pack(pady=5)
        entry = ctk.CTkEntry(dialog, width=200)
        entry.pack(pady=5)
        entry.focus_set()

        def do_find():
            search_text = entry.get()
            self.textbox.tag_remove("search", "1.0", tk.END)
            if search_text:
                idx = "1.0"
                while True:
                    idx = self.textbox.search(search_text, idx, nocase=True, stopindex=tk.END)
                    if not idx:
                        break
                    lastidx = f"{idx}+{len(search_text)}c"
                    self.textbox.tag_add("search", idx, lastidx)
                    idx = lastidx
                self.textbox.tag_config("search", background="yellow", foreground="black")

        ctk.CTkButton(dialog, text="Localizar Todos", command=do_find).pack(pady=5)

    def _on_replace(self) -> None:
        dialog = ctk.CTkToplevel(self)
        dialog.title("Substituir")
        dialog.geometry("300x200")
        dialog.attributes("-topmost", True)
        
        ctk.CTkLabel(dialog, text="Localizar:").pack(pady=2)
        find_entry = ctk.CTkEntry(dialog, width=200)
        find_entry.pack(pady=2)
        
        ctk.CTkLabel(dialog, text="Substituir por:").pack(pady=2)
        replace_entry = ctk.CTkEntry(dialog, width=200)
        replace_entry.pack(pady=2)

        def do_replace():
            search_text = find_entry.get()
            replace_text = replace_entry.get()
            if search_text:
                content = self.textbox.get("1.0", tk.END)
                new_content = content.replace(search_text, replace_text)
                self.textbox.delete("1.0", tk.END)
                self.textbox.insert("1.0", new_content)
                self._apply_syntax_highlighting()

        ctk.CTkButton(dialog, text="Substituir Tudo", command=do_replace).pack(pady=10)

    def _on_renum(self) -> None:
        content = self.textbox.get("1.0", tk.END).strip()
        if not content:
            return
        
        lines = content.split("\n")
        
        # 1. Mapear linhas antigas para novas
        mapping = []
        try:
            start_line = int(self.settings.get("start_line", 10))
            increment = int(self.settings.get("increment", 10))
        except ValueError:
            start_line = 10
            increment = 10
            
        new_ln = start_line
        import re
        
        pattern_ln = re.compile(r"^\s*(\d+)")
        
        valid_lines = []
        for line in lines:
            if not line.strip():
                continue
            match = pattern_ln.match(line)
            if match:
                old_ln = int(match.group(1))
                pure_content = line[match.end():].strip()
                mapping.append((old_ln, new_ln, pure_content))
                new_ln += increment
                valid_lines.append(line)
            else:
                # Se não tem número de linha, vamos atribuir um
                mapping.append((None, new_ln, line.strip()))
                new_ln += increment
                valid_lines.append(line)

        # 2. Salvar no SQLite se disponível
        if self.db:
            with self.db._connect() as conn:
                conn.execute("DELETE FROM renum_map")
                for old, new, _ in mapping:
                    if old is not None:
                        conn.execute("INSERT INTO renum_map (old_ln, new_ln) VALUES (?, ?)", (old, new))
                conn.commit()

        # 3. Atualizar referências (GOTO, GOSUB, etc)
        # Comandos que podem ter números de linha: GOTO, GOSUB, THEN, ELSE, RESTORE, RUN, ON...GOTO/GOSUB
        # Regex simplificada para capturar números após esses comandos
        ref_keywords = ["GOTO", "GOSUB", "THEN", "ELSE", "RESTORE", "RUN"]
        
        def update_refs(text: str) -> str:
            # Regex para encontrar números de linha que não estão dentro de aspas ou comentários
            # Esta é uma aproximação. Para ser perfeito precisaria de um parser real.
            # Vamos focar nos padrões comuns: KEYWORD <número>
            for kw in ref_keywords:
                # Procura KEYWORD seguido de espaços e um ou mais números separados por vírgula (para ON GOTO)
                pattern = re.compile(rf"({kw}\s+)(\d+(?:\s*,\s*\d+)*)", re.IGNORECASE)
                
                def replace_func(match):
                    prefix = match.group(1)
                    nums_str = match.group(2)
                    nums = [n.strip() for n in nums_str.split(",")]
                    new_nums = []
                    for n in nums:
                        if self.db:
                            with self.db._connect() as conn:
                                row = conn.execute("SELECT new_ln FROM renum_map WHERE old_ln = ?", (n,)).fetchone()
                                if row:
                                    new_nums.append(str(row["new_ln"]))
                                else:
                                    new_nums.append(n) # Mantém se não encontrar
                        else:
                            # Fallback se não tiver DB (mesmo que a task peça DB)
                            # Poderia usar um dict aqui
                            found = False
                            for m_old, m_new, _ in mapping:
                                if m_old == int(n):
                                    new_nums.append(str(m_new))
                                    found = True
                                    break
                            if not found:
                                new_nums.append(n)
                    return prefix + ", ".join(new_nums)
                
                text = pattern.sub(replace_func, text)
            return text

        new_lines = []
        for old, new, pure in mapping:
            updated_pure = update_refs(pure)
            new_lines.append(f"{new} {updated_pure}")
        
        self.textbox.delete("1.0", tk.END)
        self.textbox.insert("1.0", "\n".join(new_lines))
        self._apply_syntax_highlighting()

    def _remove_line_numbers(self) -> None:
        content = self.textbox.get("1.0", tk.END).strip()
        if not content:
            return
        lines = content.split("\n")
        new_lines = []
        for line in lines:
            import re
            new_lines.append(re.sub(r"^\s*\d+\s*", "", line))
        self.textbox.delete("1.0", tk.END)
        self.textbox.insert("1.0", "\n".join(new_lines))

    def _add_line_numbers(self) -> None:
        self._on_renum()

    def _open_file(self) -> None:
        file_path = filedialog.askopenfilename(
            title="Abrir arquivo MSX BASIC",
            filetypes=[("Arquivos MSX", "*.bas *.asc *.txt"), ("Todos os arquivos", "*.*")]
        )
        if not file_path:
            return

        try:
            path = Path(file_path)
            data = path.read_bytes()
            
            # If it's tokenized MSX BASIC (starts with 0xFF)
            if data.startswith(b"\xFF"):
                segments = decode_msx_basic_segments(data)
                text = "".join(seg[1] for seg in segments)
            else:
                # Try common encodings
                try:
                    text = data.decode("utf-8")
                except UnicodeDecodeError:
                    text = data.decode("latin-1")
            
            self.textbox.delete("1.0", tk.END)
            self.textbox.insert("1.0", text)
            self._apply_syntax_highlighting()
        except Exception as e:
            messagebox.showerror("Erro", f"Nao foi possivel abrir o arquivo:\n{e}")

    def _save_file(self) -> None:
        content = self.textbox.get("1.0", tk.END).strip()
        
        # Dialect restrictions
        if self.settings.get("dialect") == "MSX-BASIC":
            lines = content.split("\n")
            for i, line in enumerate(lines):
                if line.strip() and not re.match(r"^\d+", line.strip()):
                    messagebox.showerror("Erro de Dialeto", f"No MSX-BASIC clássico, todas as linhas devem ser numeradas.\nErro na linha {i+1}: {line[:30]}...")
                    return

        file_path = filedialog.asksaveasfilename(
            title="Salvar como",
            defaultextension=".bas",
            filetypes=[("MSX BASIC", "*.bas"), ("Texto", "*.txt"), ("Todos os arquivos", "*.*")]
        )
        if not file_path:
            return

        try:
            content = self.textbox.get("1.0", tk.END)
            # Saving as plain text (ASCII) which MSX can LOAD "filename.bas",A
            Path(file_path).write_text(content, encoding="latin-1", errors="replace")
            messagebox.showinfo("Sucesso", "Arquivo salvo com sucesso (formato ASCII).")
        except Exception as e:
            messagebox.showerror("Erro", f"Erro ao salvar:\n{e}")

    def _clear_editor(self) -> None:
        if messagebox.askyesno("Limpar", "Deseja limpar todo o conteudo?"):
            self.textbox.delete("1.0", tk.END)

    def _beautify_line(self, line: str) -> str:
        from msx_basic_decoder import TOKEN_MAP, TOKEN_MAP_FF

        if not line.strip():
            return line

        # 1. Separar número da linha
        match_ln = re.match(r"^(\s*\d+)\s*(.*)", line)
        if match_ln:
            ln_part = match_ln.group(1).strip()
            code_part = match_ln.group(2)
        else:
            ln_part = ""
            code_part = line

        # 2. Tokenizar mantendo strings e comentários intactos
        # Ordem de importância: strings, comentários, palavras-chave, outros
        keywords = sorted(list(set(TOKEN_MAP) | set(TOKEN_MAP_FF)), key=len, reverse=True)
        # Escapar keywords que são operadores
        escaped_keywords = [re.escape(k) for k in keywords]
        
        # Regex para strings: "[^"]*"
        # Regex para comentários: (REM|').*
        # Regex para keywords
        combined_pattern = f"(\"[^\"]*\")|(REM.*|'.*)|({'|'.join(escaped_keywords)})|([^\"R'\\s]+)"
        
        tokens = []
        # re.finditer para pegar tudo inclusive espaços se necessário, mas vamos processar code_part
        # Melhor: iterar pelo texto e identificar componentes
        
        pos = 0
        result_parts = []
        if ln_part:
            result_parts.append(ln_part)
            result_parts.append(" ")

        while pos < len(code_part):
            char = code_part[pos]
            
            # Pular espaços já existentes (serão recalculados)
            if char.isspace():
                pos += 1
                continue

            # String
            if char == '"':
                end = code_part.find('"', pos + 1)
                if end == -1:
                    result_parts.append(code_part[pos:])
                    pos = len(code_part)
                else:
                    result_parts.append(code_part[pos:end+1])
                    pos = end + 1
                continue

            # Comentário
            if code_part[pos:].upper().startswith("REM"):
                rem_text = code_part[pos:pos+3]
                if self.settings.get("keep_case") == "False":
                    rem_text = rem_text.upper()
                result_parts.append(rem_text + code_part[pos+3:])
                pos = len(code_part)
                continue
            if char == "'":
                result_parts.append(code_part[pos:])
                pos = len(code_part)
                continue

            # Keyword?
            found_kw = False
            for kw in keywords:
                if code_part[pos:].upper().startswith(kw):
                    # Se for uma keyword que termina com (, não adicionar espaço depois (ex: TAB()
                    # Mas o MSX BASIC as vezes tem "TAB (10)" ? Geralmente é colado.
                    kw_text = code_part[pos:pos+len(kw)]
                    if self.settings.get("keep_case") == "False":
                        kw_text = kw_text.upper()
                    result_parts.append(kw_text)
                    pos += len(kw)
                    found_kw = True
                    break
            
            if found_kw:
                continue

            # Outros caracteres (variáveis, operadores não-keyword, etc)
            # Pegar até o próximo delimitador ou espaço ou keyword
            start_other = pos
            while pos < len(code_part):
                curr = code_part[pos]
                if curr.isspace() or curr == '"' or curr == "'":
                    break
                if code_part[pos:].upper().startswith("REM"):
                    break
                
                # Tratar operadores comuns como tokens separados para forçar espaços
                if curr in "=+-*/^\\<>:":
                    if pos == start_other:
                        pos += 1
                    break

                any_kw = False
                for kw in keywords:
                    if code_part[pos:].upper().startswith(kw):
                        any_kw = True
                        break
                if any_kw:
                    break
                pos += 1
            
            other_text = code_part[start_other:pos]
            if other_text:
                result_parts.append(other_text)

        # Montar a linha final com espaços
        # Regra: espaço entre quase tudo, exceto:
        # - Após número da linha (já adicionado)
        # - Antes de ":" se for separador de comandos? MSX permite "PRINT:PRINT". QB usa "PRINT : PRINT"
        # O usuário pediu: 10 PRINT "TESTE" : IF A$="S" THEN 10 ELSE 40
        
        final_line = ""
        for i, part in enumerate(result_parts):
            if i == 0:
                final_line += part
                continue
            
            prev_part = result_parts[i-1]
            
            # Não adicionar espaço se:
            # - part for ":" e prev_part for algo que não precisa de espaço?
            # Na verdade, o usuário quer " : ", então adicionamos espaço.
            
            # Exceções onde NÃO colocar espaço:
            # - Antes de "(" (ex: TAB()
            # - Entre nome de variável e "$" ou "%" ou "!" ou "#" (já virão juntos no 'other')
            # - Dentro de expressões compactas? O usuário quer DESCOLAR.
            
            # Se a parte atual é uma string ou comentário, e a anterior não é espaço, bota espaço.
            # Se a parte anterior é uma keyword, bota espaço.
            
            # Vamos ser agressivos na inserção de espaços como solicitado.
            if not final_line.endswith(" ") and not part.startswith(" "):
                # Não colocar espaço antes de ( se prev for function?
                # TOKEN_MAP_FF são funções.
                is_prev_func = prev_part.upper() in TOKEN_MAP_FF or prev_part.upper().endswith("(")
                if is_prev_func and part.startswith("("):
                    pass # Sem espaço
                else:
                    final_line += " "
            
            final_line += part

        return final_line.rstrip()

    def _on_key_beautify(self, event) -> None:
        # Pega a linha atual
        cursor_pos = self.textbox.index(tk.INSERT)
        line_num = cursor_pos.split(".")[0]
        col_idx = int(cursor_pos.split(".")[1])
        
        line_content = self.textbox.get(f"{line_num}.0", f"{line_num}.end")

        # Verificar se o cursor está dentro de uma string
        # Contamos as aspas antes do cursor na linha atual
        quotes_before = line_content[:col_idx].count('"')
        if quotes_before % 2 != 0:
            # Se ímpar, o cursor está dentro de uma string (aberta)
            return
        
        # Se for Enter, formatar a linha ANTERIOR
        if event.keysym == "Return":
            prev_line = str(int(line_num) - 1)
            if int(prev_line) >= 1:
                prev_content = self.textbox.get(f"{prev_line}.0", f"{prev_line}.end")
                beautified_prev = self._beautify_line(prev_content)
                if beautified_prev != prev_content:
                    self.textbox.delete(f"{prev_line}.0", f"{prev_line}.end")
                    self.textbox.insert(f"{prev_line}.0", beautified_prev)
            return

        # Para Espaço, formatar a linha atual
        beautified = self._beautify_line(line_content)
        if beautified != line_content:
            # Salvar posição do cursor (tentar manter a lógica)
            # Ao adicionar espaços antes do cursor, precisamos compensar
            
            # Vamos ver quantos espaços foram adicionados antes da posição do cursor
            orig_before = line_content[:col_idx]
            # Simplificação: se o texto mudou, re-inserimos e tentamos achar a nova posição
            # Mas para o usuário não "sentir" o salto, o ideal seria apenas formatar
            # quando ele termina uma palavra.
            
            self.textbox.delete(f"{line_num}.0", f"{line_num}.end")
            self.textbox.insert(f"{line_num}.0", beautified)
            
            # Reposicionar o cursor: se adicionamos espaços, ele deve ir mais para a frente
            # Contagem simples de diferença de tamanho (imperfeito mas ajuda)
            diff = len(beautified) - len(line_content)
            new_col = col_idx + diff
            self.textbox.mark_set(tk.INSERT, f"{line_num}.{new_col}")

    def _on_beautify_all(self) -> None:
        content = self.textbox.get("1.0", tk.END).strip()
        if not content:
            return
        
        lines = content.split("\n")
        new_lines = [self._beautify_line(line) for line in lines]
        
        self.textbox.delete("1.0", tk.END)
        self.textbox.insert("1.0", "\n".join(new_lines))
        self._apply_syntax_highlighting()
        messagebox.showinfo("Beautify", "Código formatado com sucesso!")

    def _on_program_map(self) -> None:
        from msx_basic_analyzer import MSXBasicAnalyzer
        
        # O usuário sugeriu que o mapa seja feito após a formatação para evitar erros.
        # Vamos obter o conteúdo formatado sem necessariamente alterar o texto no editor,
        # ou apenas formatar as linhas internamente para a análise.
        content = self.textbox.get("1.0", tk.END).strip()
        if not content:
            return
            
        lines = content.split("\n")
        beautified_lines = [self._beautify_line(line) for line in lines]
        beautified_content = "\n".join(beautified_lines)
        
        analyzer = MSXBasicAnalyzer(beautified_content)
        analyzer.analyze()
        summary = analyzer.get_summary()
        
        self._show_program_map_window(summary)

    def _show_program_map_window(self, summary: dict) -> None:
        window = ctk.CTkToplevel(self)
        window.title("Mapa do Programa")
        window.geometry("800x600")
        window.grab_set()
        
        tabview = ctk.CTkTabview(window)
        tabview.pack(fill="both", expand=True, padx=10, pady=10)
        
        tab_vars = tabview.add("Variáveis")
        tab_flow = tabview.add("Fluxo e Subrotinas")
        
        # --- Aba Variáveis ---
        # Usar um CTkTextbox para mostrar como tabela ou lista formatada
        vars_text = ctk.CTkTextbox(tab_vars, font=("Consolas", 12))
        vars_text.pack(fill="both", expand=True, padx=5, pady=5)
        
        header = f"{'Nome':<10} | {'Tipo':<6} | {'Usos':<6} | {'Mem(est)':<8} | {'Linhas'}\n"
        header += "-" * 75 + "\n"
        vars_text.insert(tk.END, header)
        
        for name, info in summary["variables"].items():
            lines_str = ", ".join(map(str, info["lines"]))
            row = f"{name:<10} | {info['type']:<6} | {info['count']:<6} | {info['size']:<8} | {lines_str}\n"
            vars_text.insert(tk.END, row)
        
        vars_text.insert(tk.END, "\n" + "-" * 75 + "\n")
        vars_text.insert(tk.END, f"Memória total estimada para variáveis simples: {summary['total_memory_est']} bytes\n")
        vars_text.insert(tk.END, "(Nota: Strings usam 3 bytes para descritor + conteúdo; Arrays não contabilizados)\n")
        
        vars_text.configure(state="disabled")
        
        # --- Aba Fluxo ---
        flow_text = ctk.CTkTextbox(tab_flow, font=("Consolas", 12))
        flow_text.pack(fill="both", expand=True, padx=5, pady=5)
        
        flow_text.insert(tk.END, "--- Subrotinas Identificadas (GOSUB targets) ---\n")
        if summary["subroutines"]:
            for sub in summary["subroutines"]:
                flow_text.insert(tk.END, f"Linha {sub}\n")
        else:
            flow_text.insert(tk.END, "Nenhuma subrotina identificada.\n")
            
        flow_text.insert(tk.END, "\n--- Fluxo de Execução (GOTO/GOSUB) ---\n")
        header_flow = f"{'Origem':<10} | {'Destino':<10} | {'Tipo'}\n"
        header_flow += "-" * 40 + "\n"
        flow_text.insert(tk.END, header_flow)
        
        for f in summary["flow"]:
            row = f"{f['from']:<10} | {f['to']:<10} | {f['type']}\n"
            flow_text.insert(tk.END, row)
        
        flow_text.configure(state="disabled")

    def _open_viewer(self) -> None:
        from main import MSXViewer
        viewer = MSXViewer(self)
        viewer.focus()

    def _open_help_viewer(self) -> None:
        viewer = HelpViewer(self)
        viewer.focus()

    def _open_encoding_viewer(self) -> None:
        window = ctk.CTkToplevel(self)
        window.title("Visualizador de Encodings MSX")
        window.geometry("600x500")
        
        def insert_text(text: str):
            self.textbox.insert(tk.INSERT, text)
            self.textbox.focus_set()

        viewer = MSXEncodingViewer(window, insert_callback=insert_text)
        viewer.pack(fill="both", expand=True)
        window.focus()

    def _show_msx1_palette(self, highlight_index: int | None = None) -> None:
        dialog = ctk.CTkToplevel(self)
        dialog.title("Cores MSX1 (0-15)")
        dialog.geometry("420x520")
        dialog.grab_set()

        header = ctk.CTkFrame(dialog)
        header.pack(fill="x", padx=10, pady=(10, 0))
        ctk.CTkLabel(header, text="#", width=30, anchor="w").grid(row=0, column=0, sticky="w")
        ctk.CTkLabel(header, text="Nome", width=180, anchor="w").grid(row=0, column=1, sticky="w")
        ctk.CTkLabel(header, text="Hex", width=90, anchor="w").grid(row=0, column=2, sticky="w")
        ctk.CTkLabel(header, text="Amostra", width=90, anchor="w").grid(row=0, column=3, sticky="w")

        list_frame = ctk.CTkScrollableFrame(dialog)
        list_frame.pack(fill="both", expand=True, padx=10, pady=10)

        row_idx_to_widget = {}
        for r, (idx, name, hex_color) in enumerate(MSX1_PALETTE):
            row = ctk.CTkFrame(list_frame)
            row.grid(row=r, column=0, sticky="ew", pady=2)
            row.grid_columnconfigure(3, weight=1)

            ctk.CTkLabel(row, text=str(idx), width=30, anchor="w").grid(row=0, column=0, sticky="w", padx=(2, 6))
            ctk.CTkLabel(row, text=name, width=180, anchor="w").grid(row=0, column=1, sticky="w")
            ctk.CTkLabel(row, text=hex_color.upper(), width=90, anchor="w").grid(row=0, column=2, sticky="w", padx=(10, 10))

            # Amostra visual (quadrado)
            swatch = tk.Canvas(row, width=48, height=20, highlightthickness=1, highlightbackground="#444444")
            swatch.grid(row=0, column=3, sticky="w")
            swatch.create_rectangle(2, 2, 46, 18, fill=hex_color, outline="black")

            row_idx_to_widget[idx] = row

        # Se solicitado, destacar uma cor específica
        if highlight_index is not None and highlight_index in row_idx_to_widget:
            target = row_idx_to_widget[highlight_index]
            try:
                target.configure(fg_color=("#2B2B2B", "#D9D9D9"))
            except Exception:
                pass

    def _load_settings(self) -> None:
        if not self.db:
            return
        for key in self.settings:
            val = self.db.get_setting(f"editor_{key}")
            if val is not None:
                self.settings[key] = val

    def _save_settings(self) -> None:
        if not self.db:
            return
        for key, val in self.settings.items():
            self.db.set_setting(f"editor_{key}", val)

    def _on_settings(self) -> None:
        dialog = ctk.CTkToplevel(self)
        dialog.title("Configurações")
        dialog.geometry("500x700")
        dialog.grab_set()

        tabview = ctk.CTkTabview(dialog)
        tabview.pack(fill="both", expand=True, padx=10, pady=10)

        tab_main = tabview.add("Principal")
        tab_msx_basic = tabview.add("MSX-BASIC")
        tab_dignified = tabview.add("Dignified")
        tab_bas2rom = tabview.add("Bas2Rom")
        tab_emulator = tabview.add("Emulador")
        tab_extras = tabview.add("Extras")

        # Helper to create entry with label
        def create_entry(parent, label_text, default_val):
            frame = ctk.CTkFrame(parent)
            frame.pack(fill="x", padx=20, pady=5)
            ctk.CTkLabel(frame, text=label_text, width=150, anchor="w").pack(side="left")
            entry = ctk.CTkEntry(frame)
            entry.insert(0, str(default_val) if default_val is not None else "")
            entry.pack(side="right", expand=True, fill="x")
            return entry

        # --- Aba Principal ---
        syntax_theme_var = tk.StringVar(value=self.db.get_setting("syntax_theme", DEFAULT_SYNTAX_THEME))
        theme_frame = ctk.CTkFrame(tab_main)
        theme_frame.pack(fill="x", padx=20, pady=5)
        ctk.CTkLabel(theme_frame, text="Tema de Sintaxe:", width=150, anchor="w").pack(side="left")
        
        def on_theme_change(new_theme):
            theme_colors = SYNTAX_THEMES.get(new_theme, SYNTAX_THEMES[DEFAULT_SYNTAX_THEME])
            for k, v in theme_colors.items():
                if k in color_entries:
                    color_entries[k].delete(0, tk.END)
                    color_entries[k].insert(0, v)

        theme_menu = ctk.CTkOptionMenu(theme_frame, values=list(SYNTAX_THEMES.keys()), variable=syntax_theme_var, command=on_theme_change)
        theme_menu.pack(side="right", expand=True, fill="x")

        dialect_var = tk.StringVar(value=self.settings.get("dialect", "MSX-BASIC"))
        dialect_frame = ctk.CTkFrame(tab_main)
        dialect_frame.pack(fill="x", padx=20, pady=5)
        ctk.CTkLabel(dialect_frame, text="Dialeto Atual:", width=150, anchor="w").pack(side="left")
        dialects = ["MSX-BASIC", "MSX Basic Dignified", "MSX-Bas2Rom"]
        dialect_menu = ctk.CTkOptionMenu(dialect_frame, values=dialects, variable=dialect_var)
        dialect_menu.pack(side="right", expand=True, fill="x")

        start_line_entry = create_entry(tab_main, "Linha Inicial:", self.settings["start_line"])
        increment_entry = create_entry(tab_main, "Incremento:", self.settings["increment"])
        
        color_entries = {}
        color_labels = {
            "command": "Cor Comandos:",
            "function": "Cor Funções:",
            "string": "Cor Strings:",
            "number": "Cor Números:",
            "line_number": "Cor Nº Linha:",
            "comment": "Cor Comentários:",
            "bg": "Cor Fundo:",
            "fg": "Cor Texto:"
        }
        for key, label in color_labels.items():
            color_entries[key] = create_entry(tab_main, label, self.settings.get(f"color_{key}"))

        keep_case_var = tk.BooleanVar(value=self.settings.get("keep_case") == "True")
        keep_case_check = ctk.CTkCheckBox(tab_main, text="Manter palavras-chave como escritas", variable=keep_case_var)
        keep_case_check.pack(pady=10, padx=20, anchor="w")

        # --- Abas de Dialeto (Placeholder por enquanto) ---
        ctk.CTkLabel(tab_msx_basic, text="Configurações específicas do MSX-BASIC clássico").pack(pady=20)
        ctk.CTkLabel(tab_dignified, text="Configurações específicas do MSX Basic Dignified").pack(pady=20)
        ctk.CTkLabel(tab_bas2rom, text="Configurações específicas do MSX-Bas2Rom").pack(pady=20)

        # --- Aba Emulador ---
        openmsx_entry = create_entry(tab_emulator, "Caminho openMSX:", self.settings.get("openmsx_path", ""))
        fmsx_entry = create_entry(tab_emulator, "Caminho fMSX:", self.settings.get("fmsx_path", ""))

        # --- Aba Extras ---
        extra_entry = create_entry(tab_extras, "Configurações Extras:", self.settings.get("extra_configs", ""))

        def save():
            self.settings["dialect"] = dialect_var.get()
            self.settings["start_line"] = start_line_entry.get()
            self.settings["increment"] = increment_entry.get()
            
            new_colors = {}
            for key in color_entries:
                val = color_entries[key].get()
                self.settings[f"color_{key}"] = val
                new_colors[key] = val
            
            save_syntax_colors(self.db, syntax_theme_var.get(), new_colors)
            
            self.settings["keep_case"] = str(keep_case_var.get())
            self.settings["openmsx_path"] = openmsx_entry.get()
            self.settings["fmsx_path"] = fmsx_entry.get()
            self.settings["extra_configs"] = extra_entry.get()
            
            self._save_settings()
            
            # Update UI colors
            self.line_numbers.configure(bg=self.settings.get("color_bg", "#2b2b2b"))
            self.ruler.configure(bg=self.settings.get("color_bg", "#2b2b2b"))
            
            self._setup_syntax_highlighting()
            self._apply_syntax_highlighting()
            dialog.destroy()
            messagebox.showinfo("Configurações", "Configurações salvas com sucesso!")

        ctk.CTkButton(dialog, text="Salvar", command=save).pack(pady=10)
