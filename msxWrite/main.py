from __future__ import annotations

import os
import time
from pathlib import Path
import tkinter as tk
from tkinter import colorchooser, filedialog, messagebox

import customtkinter as ctk

from app_db import AppDatabase
from msx_basic_decoder import decode_msx_basic_segments
from alphabet_viewer import AlphabetViewerFrame
from layout_viewer import LayoutViewerFrame
from screen_viewer import ScreenViewerFrame
from shape_viewer import ShapeViewerFrame
from syntax_themes import SYNTAX_THEMES, DEFAULT_SYNTAX_THEME, get_syntax_colors, save_syntax_colors

try:
    from help_viewer import HelpViewer
    HELP_VIEWER_AVAILABLE = True
except ImportError:
    HELP_VIEWER_AVAILABLE = False


APP_TITLE = "MSX-Write"
DB_NAME = "msxread.db"
TEXT_ENCODINGS = ("utf-8", "cp1252", "latin-1")
HEX_PREVIEW_BYTES = 4096


class MSXViewer(ctk.CTkToplevel):
    def __init__(self, master: ctk.CTk | None = None) -> None:
        super().__init__(master)

        if master and hasattr(master, "db"):
            self.db = master.db
        else:
            self.db = AppDatabase(Path(DB_NAME))

        self.appearance_mode = self.db.get_setting("appearance_mode", "System")
        self.color_theme = self.db.get_setting("color_theme", "blue")
        ctk.set_appearance_mode(self.appearance_mode)
        ctk.set_default_color_theme(self.color_theme)

        self.title(APP_TITLE)
        geometry = self.db.get_setting("window_geometry", "1100x700")
        if geometry:
            self.geometry(geometry)

        self.base_dir = self.db.get_setting("last_dir", str(Path.cwd()))
        self.current_file: str | None = None
        self.current_file_kind: str | None = None
        self.current_msx_segments: list[tuple[str, str]] | None = None
        self.shape_viewer: ShapeViewerFrame | None = None
        self.alphabet_viewer: AlphabetViewerFrame | None = None
        self.layout_viewer: LayoutViewerFrame | None = None
        self.screen_viewer: ScreenViewerFrame | None = None

        self.syntax_theme_name = self.db.get_setting("syntax_theme", DEFAULT_SYNTAX_THEME)
        self.syntax_colors = get_syntax_colors(self.db)
        self.viewer_text_bg = self.syntax_colors.get("bg", "")
        self.viewer_text_fg = self.syntax_colors.get("fg", "")

        self._load_fonts()
        self._build_layout()
        self._apply_viewer_colors()
        self._refresh_file_list()
        self._load_last_file()

        self.protocol("WM_DELETE_WINDOW", self._on_close)

    def _load_fonts(self) -> None:
        self.text_font = ("Consolas", 12)
        ttf_path = Path("MSX-Screen0.ttf")
        if ttf_path.exists():
            try:
                ctk.FontManager.load_font(str(ttf_path))
                self.text_font = ("MSX Screen 0", 14)
            except Exception:
                self.text_font = ("Consolas", 12)

    def _build_layout(self) -> None:
        self.grid_columnconfigure(0, weight=0)
        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(1, weight=1)

        header = ctk.CTkFrame(self)
        header.grid(row=0, column=0, columnspan=2, sticky="ew", padx=10, pady=10)
        header.grid_columnconfigure(1, weight=1)

        self.dir_label = ctk.CTkLabel(header, text="Diretorio:")
        self.dir_label.grid(row=0, column=0, sticky="w", padx=(10, 5), pady=10)

        self.dir_value = ctk.CTkLabel(header, text=self.base_dir, anchor="w")
        self.dir_value.grid(row=0, column=1, sticky="ew", padx=5, pady=10)

        choose_button = ctk.CTkButton(header, text="Selecionar", command=self._choose_directory)
        choose_button.grid(row=0, column=2, padx=5, pady=10)

        refresh_button = ctk.CTkButton(header, text="Atualizar", command=self._refresh_file_list)
        refresh_button.grid(row=0, column=3, padx=(5, 10), pady=10)

        settings_button = ctk.CTkButton(header, text="Configuracao", command=self._open_settings)
        settings_button.grid(row=0, column=4, padx=(0, 10), pady=10)

        left = ctk.CTkFrame(self)
        left.grid(row=1, column=0, sticky="nsew", padx=(10, 5), pady=(0, 10))
        left.grid_rowconfigure(0, weight=1)
        left.grid_columnconfigure(0, weight=1)

        list_frame = ctk.CTkFrame(left)
        list_frame.grid(row=0, column=0, sticky="nsew", padx=10, pady=10)
        list_frame.grid_rowconfigure(0, weight=1)
        list_frame.grid_columnconfigure(0, weight=1)

        self.file_listbox = tk.Listbox(list_frame, activestyle="none")
        self.file_listbox.grid(row=0, column=0, sticky="nsew")
        self.file_listbox.bind("<<ListboxSelect>>", self._on_file_select)

        scrollbar = tk.Scrollbar(list_frame, command=self.file_listbox.yview)
        scrollbar.grid(row=0, column=1, sticky="ns")
        self.file_listbox.config(yscrollcommand=scrollbar.set)

        right = ctk.CTkFrame(self)
        right.grid(row=1, column=1, sticky="nsew", padx=(5, 10), pady=(0, 10))
        right.grid_rowconfigure(0, weight=1)
        right.grid_columnconfigure(0, weight=1)

        self.right_tabs = ctk.CTkTabview(right)
        self.right_tabs.grid(row=0, column=0, sticky="nsew", padx=10, pady=10)

        content_tab = self.right_tabs.add("Conteudo")
        content_tab.grid_rowconfigure(1, weight=1)
        content_tab.grid_columnconfigure(0, weight=1)

        self.file_label = ctk.CTkLabel(content_tab, text="Selecione um arquivo")
        self.file_label.grid(row=0, column=0, sticky="w", padx=10, pady=(10, 5))

        self.textbox = ctk.CTkTextbox(content_tab, wrap="none")
        self.textbox.grid(row=1, column=0, sticky="nsew", padx=10, pady=(0, 10))
        self.textbox.configure(font=self.text_font, state="disabled")
        self.text_widget = self.textbox._textbox if hasattr(self.textbox, "_textbox") else self.textbox
        self.default_text_bg = self.textbox.cget("fg_color")
        self.default_text_fg = self.textbox.cget("text_color")
        self._configure_syntax_tags()

        self.status_label = ctk.CTkLabel(content_tab, text="", anchor="w")
        self.status_label.grid(row=2, column=0, sticky="ew", padx=10, pady=(0, 10))

        shape_tab = self.right_tabs.add("Shape")
        shape_tab.grid_rowconfigure(0, weight=1)
        shape_tab.grid_columnconfigure(0, weight=1)
        self.shape_viewer = ShapeViewerFrame(shape_tab)
        self.shape_viewer.grid(row=0, column=0, sticky="nsew")

        alphabet_tab = self.right_tabs.add("Alfabeto")
        alphabet_tab.grid_rowconfigure(0, weight=1)
        alphabet_tab.grid_columnconfigure(0, weight=1)
        self.alphabet_viewer = AlphabetViewerFrame(alphabet_tab)
        self.alphabet_viewer.grid(row=0, column=0, sticky="nsew")

        layout_tab = self.right_tabs.add("Layout")
        layout_tab.grid_rowconfigure(0, weight=1)
        layout_tab.grid_columnconfigure(0, weight=1)
        self.layout_viewer = LayoutViewerFrame(layout_tab)
        self.layout_viewer.grid(row=0, column=0, sticky="nsew")

        screen_tab = self.right_tabs.add("Screen")
        screen_tab.grid_rowconfigure(0, weight=1)
        screen_tab.grid_columnconfigure(0, weight=1)
        self.screen_viewer = ScreenViewerFrame(screen_tab)
        self.screen_viewer.grid(row=0, column=0, sticky="nsew")

        editor_button = ctk.CTkButton(header, text="Editor BASIC", command=self._open_basic_editor)
        editor_button.grid(row=0, column=5, padx=(0, 10), pady=10)

        help_button = ctk.CTkButton(header, text="Ajuda MSX", command=self._open_help_viewer)
        help_button.grid(row=0, column=6, padx=(0, 10), pady=10)



    def _choose_directory(self) -> None:
        path = filedialog.askdirectory(initialdir=self.base_dir, title="Selecione o diretorio")
        if not path:
            return
        self.base_dir = path
        self.dir_value.configure(text=path)
        self.db.set_setting("last_dir", path)
        self._refresh_file_list()

    def _refresh_file_list(self) -> None:
        self.file_listbox.delete(0, tk.END)
        base_path = Path(self.base_dir)
        files = []
        if base_path.exists():
            for entry in base_path.iterdir():
                if entry.is_file():
                    files.append(entry.name)
        for name in sorted(files, key=str.lower):
            self.file_listbox.insert(tk.END, name)
        self.status_label.configure(text=f"{len(files)} arquivos encontrados")

    def _looks_like_msx_basic(self, path: Path) -> bool:
        try:
            with path.open("rb") as handle:
                first = handle.read(1)
            return first == b"\xFF"
        except OSError:
            return False

    def _looks_like_msx_basic_data(self, data: bytes) -> bool:
        return data[:1] == b"\xFF"

    def _decode_text(self, data: bytes) -> str | None:
        if b"\x00" in data:
            return None
        for encoding in TEXT_ENCODINGS:
            try:
                return data.decode(encoding)
            except UnicodeDecodeError:
                continue
        return data.decode("latin-1", errors="replace")

    def _hex_dump(self, data: bytes) -> str:
        preview = data[:HEX_PREVIEW_BYTES]
        lines = []
        for offset in range(0, len(preview), 16):
            chunk = preview[offset : offset + 16]
            hex_part = " ".join(f"{byte:02X}" for byte in chunk)
            text_part = "".join(chr(byte) if 32 <= byte <= 126 else "." for byte in chunk)
            lines.append(f"{offset:08X}  {hex_part:<47}  {text_part}")
        if len(data) > HEX_PREVIEW_BYTES:
            lines.append("")
            lines.append(f"... {len(data) - HEX_PREVIEW_BYTES} bytes nao exibidos")
        return "\n".join(lines)

    def _on_file_select(self, _event: tk.Event) -> None:
        selection = self.file_listbox.curselection()
        if not selection:
            return
        name = self.file_listbox.get(selection[0])
        file_path = str(Path(self.base_dir) / name)
        self._open_file(file_path)

    def _open_file(self, file_path: str) -> None:
        segments: list[tuple[str, str]] | None = None
        try:
            data = Path(file_path).read_bytes()
            if Path(file_path).suffix.lower() == ".shp":
                self._open_shape_viewer(file_path)
                decoded = "Arquivo SHP aberto no visualizador."
                file_kind = "Graphos Shape"
            elif Path(file_path).suffix.lower() == ".alf":
                self._open_alphabet_viewer(file_path)
                decoded = "Arquivo ALF aberto no visualizador."
                file_kind = "Graphos Alphabet"
            elif Path(file_path).suffix.lower() == ".lay":
                self._open_layout_viewer(file_path)
                decoded = "Arquivo LAY aberto no visualizador."
                file_kind = "Graphos Layout"
            elif Path(file_path).suffix.lower() == ".scr":
                self._open_screen_viewer(file_path)
                decoded = "Arquivo SCR aberto no visualizador."
                file_kind = "Graphos Screen 2"
            elif self._looks_like_msx_basic_data(data):
                segments = decode_msx_basic_segments(data)
                decoded = "".join(text for _kind, text in segments)
                file_kind = "MSX BASIC"
                self.right_tabs.set("Conteudo")
            else:
                text = self._decode_text(data)
                if text is not None:
                    decoded = text
                    file_kind = "Texto"
                    self.right_tabs.set("Conteudo")
                else:
                    decoded = self._hex_dump(data)
                    file_kind = "Binario"
                    self.right_tabs.set("Conteudo")
        except Exception as exc:
            messagebox.showerror("Erro ao abrir", str(exc))
            return

        self.current_file = file_path
        self.current_file_kind = file_kind
        self.current_msx_segments = segments if file_kind == "MSX BASIC" else None
        self.db.set_setting("last_file", file_path)
        self.db.touch_recent_file(file_path, int(time.time()))

        self.file_label.configure(text=f"{Path(file_path).name} ({file_kind})")
        if file_kind == "MSX BASIC" and self.current_msx_segments:
            self._set_msx_text(self.current_msx_segments)
        else:
            self._set_text(decoded)

    def _open_shape_viewer(self, file_path: str) -> None:
        if self.shape_viewer:
            self.shape_viewer.set_file(file_path)
        self.right_tabs.set("Shape")

    def _open_alphabet_viewer(self, file_path: str) -> None:
        if self.alphabet_viewer:
            self.alphabet_viewer.set_file(file_path)
        self.right_tabs.set("Alfabeto")

    def _open_layout_viewer(self, file_path: str) -> None:
        if self.layout_viewer:
            self.layout_viewer.set_file(file_path)
        self.right_tabs.set("Layout")

    def _open_screen_viewer(self, file_path: str) -> None:
        if self.screen_viewer:
            self.screen_viewer.set_file(file_path)
        self.right_tabs.set("Screen")

    def _set_text(self, text: str) -> None:
        self.textbox.configure(state="normal")
        self.textbox.delete("1.0", tk.END)
        self.textbox.insert("1.0", text)
        self._clear_highlighting()
        self.textbox.configure(state="disabled")

    def _load_last_file(self) -> None:
        last_file = self.db.get_setting("last_file")
        if not last_file:
            return
        path = Path(last_file)
        if path.exists():
            self.base_dir = str(path.parent)
            self.dir_value.configure(text=self.base_dir)
            self._refresh_file_list()
            self._open_file(str(path))

    def _on_close(self) -> None:
        self.db.set_setting("window_geometry", self.geometry())
        self.destroy()

    def _set_msx_text(self, segments: list[tuple[str, str]]) -> None:
        self.textbox.configure(state="normal")
        self.textbox.delete("1.0", tk.END)
        for kind, text in segments:
            tag = self._map_kind_to_tag(kind)
            if tag:
                self.text_widget.insert(tk.END, text, (tag,))
            else:
                self.text_widget.insert(tk.END, text)
        self.textbox.configure(state="disabled")

    def _map_kind_to_tag(self, kind: str) -> str | None:
        mapping = {
            "command": "msx_command",
            "function": "msx_function",
            "string": "msx_string",
            "number": "msx_number",
            "comment": "msx_comment",
            "line_number": "msx_line_number",
        }
        return mapping.get(kind)

    def _clear_highlighting(self) -> None:
        for tag in (
            "msx_command",
            "msx_function",
            "msx_string",
            "msx_number",
            "msx_comment",
            "msx_line_number",
        ):
            self.text_widget.tag_remove(tag, "1.0", tk.END)

    def _configure_syntax_tags(self) -> None:
        colors = self.syntax_colors
        self.text_widget.tag_config("msx_command", foreground=colors.get("command", "#2E6F9E"))
        self.text_widget.tag_config("msx_function", foreground=colors.get("function", "#2B7A5B"))
        self.text_widget.tag_config("msx_string", foreground=colors.get("string", "#B54D2B"))
        self.text_widget.tag_config("msx_number", foreground=colors.get("number", "#7A3E9D"))
        self.text_widget.tag_config("msx_comment", foreground=colors.get("comment", "#3F6A3F"))
        self.text_widget.tag_config("msx_line_number", foreground=colors.get("line_number", "#6B6B6B"))

    def _apply_viewer_colors(self) -> None:
        bg = self.viewer_text_bg or self.default_text_bg
        fg = self.viewer_text_fg or self.default_text_fg
        self.textbox.configure(fg_color=bg, text_color=fg)

    def _open_settings(self) -> None:
        dialog = ctk.CTkToplevel(self)
        dialog.title("Configuracao")
        dialog.resizable(False, False)
        dialog.grab_set()

        container = ctk.CTkFrame(dialog)
        container.grid(row=0, column=0, sticky="nsew", padx=15, pady=15)

        appearance_var = tk.StringVar(value=self.appearance_mode)
        color_theme_var = tk.StringVar(value=self.color_theme)
        viewer_bg_var = tk.StringVar(value=self.viewer_text_bg)
        viewer_fg_var = tk.StringVar(value=self.viewer_text_fg)

        syntax_theme_var = tk.StringVar(value=self.syntax_theme_name)
        syntax_vars = {
            "command": tk.StringVar(value=self.syntax_colors.get("command", "")),
            "function": tk.StringVar(value=self.syntax_colors.get("function", "")),
            "string": tk.StringVar(value=self.syntax_colors.get("string", "")),
            "number": tk.StringVar(value=self.syntax_colors.get("number", "")),
            "comment": tk.StringVar(value=self.syntax_colors.get("comment", "")),
            "line_number": tk.StringVar(value=self.syntax_colors.get("line_number", "")),
            "bg": tk.StringVar(value=self.syntax_colors.get("bg", "")),
            "fg": tk.StringVar(value=self.syntax_colors.get("fg", "")),
        }

        row = 0
        ctk.CTkLabel(container, text="Tema do aplicativo").grid(row=row, column=0, sticky="w", pady=(0, 4))
        ctk.CTkOptionMenu(container, values=["System", "Light", "Dark"], variable=appearance_var).grid(
            row=row, column=1, sticky="ew", pady=(0, 4)
        )
        row += 1

        ctk.CTkLabel(container, text="Paleta do aplicativo").grid(row=row, column=0, sticky="w", pady=(0, 12))
        ctk.CTkOptionMenu(container, values=["blue", "green", "dark-blue"], variable=color_theme_var).grid(
            row=row, column=1, sticky="ew", pady=(0, 12)
        )
        row += 1

        ctk.CTkLabel(container, text="Cor de fundo do texto").grid(row=row, column=0, sticky="w", pady=(0, 4))
        ctk.CTkButton(
            container,
            text="Escolher",
            command=lambda: self._pick_color(dialog, syntax_vars["bg"]),
        ).grid(row=row, column=1, sticky="w", pady=(0, 4))
        ctk.CTkLabel(container, textvariable=syntax_vars["bg"], width=90).grid(row=row, column=2, sticky="w", pady=(0, 4))
        row += 1

        ctk.CTkLabel(container, text="Cor do texto").grid(row=row, column=0, sticky="w", pady=(0, 12))
        ctk.CTkButton(
            container,
            text="Escolher",
            command=lambda: self._pick_color(dialog, syntax_vars["fg"]),
        ).grid(row=row, column=1, sticky="w", pady=(0, 12))
        ctk.CTkLabel(container, textvariable=syntax_vars["fg"], width=90).grid(row=row, column=2, sticky="w", pady=(0, 12))
        row += 1

        ctk.CTkLabel(container, text="Tema do MSX BASIC").grid(row=row, column=0, sticky="w", pady=(0, 4))
        ctk.CTkOptionMenu(container, values=list(SYNTAX_THEMES.keys()), variable=syntax_theme_var).grid(
            row=row, column=1, sticky="ew", pady=(0, 4)
        )
        ctk.CTkButton(
            container,
            text="Aplicar tema",
            command=lambda: self._apply_theme_to_vars(syntax_theme_var.get(), syntax_vars),
        ).grid(row=row, column=2, sticky="w", pady=(0, 4))
        row += 1

        row = self._add_syntax_color_row(container, row, "Comandos", "command", syntax_vars)
        row = self._add_syntax_color_row(container, row, "Funcoes", "function", syntax_vars)
        row = self._add_syntax_color_row(container, row, "Strings", "string", syntax_vars)
        row = self._add_syntax_color_row(container, row, "Numeros", "number", syntax_vars)
        row = self._add_syntax_color_row(container, row, "Comentarios", "comment", syntax_vars)
        row = self._add_syntax_color_row(container, row, "Numeracao", "line_number", syntax_vars)

        button_frame = ctk.CTkFrame(container)
        button_frame.grid(row=row, column=0, columnspan=3, sticky="ew", pady=(12, 0))
        button_frame.grid_columnconfigure(0, weight=1)
        button_frame.grid_columnconfigure(1, weight=1)

        ctk.CTkButton(
            button_frame,
            text="Cancelar",
            command=dialog.destroy,
        ).grid(row=0, column=0, sticky="ew", padx=(0, 6))

        ctk.CTkButton(
            button_frame,
            text="Salvar",
            command=lambda: self._save_settings(
                dialog,
                appearance_var,
                color_theme_var,
                syntax_theme_var,
                syntax_vars,
            ),
        ).grid(row=0, column=1, sticky="ew", padx=(6, 0))

        dialog.columnconfigure(0, weight=1)
        container.columnconfigure(1, weight=1)

    def _add_syntax_color_row(
        self,
        parent: ctk.CTkFrame,
        row: int,
        label: str,
        key: str,
        vars_map: dict[str, tk.StringVar],
    ) -> int:
        ctk.CTkLabel(parent, text=label).grid(row=row, column=0, sticky="w", pady=(0, 4))
        ctk.CTkButton(
            parent,
            text="Escolher",
            command=lambda: self._pick_color(parent, vars_map[key]),
        ).grid(row=row, column=1, sticky="w", pady=(0, 4))
        ctk.CTkLabel(parent, textvariable=vars_map[key], width=90).grid(row=row, column=2, sticky="w", pady=(0, 4))
        return row + 1

    def _apply_theme_to_vars(self, theme_name: str, vars_map: dict[str, tk.StringVar]) -> None:
        theme = SYNTAX_THEMES.get(theme_name, SYNTAX_THEMES[DEFAULT_SYNTAX_THEME])
        for key, value in theme.items():
            vars_map[key].set(value)

    def _pick_color(self, parent: tk.Misc, target_var: tk.StringVar) -> None:
        result = colorchooser.askcolor(parent=parent, color=target_var.get() or None)
        if result and result[1]:
            target_var.set(result[1])

    def _save_settings(
        self,
        dialog: ctk.CTkToplevel,
        appearance_var: tk.StringVar,
        color_theme_var: tk.StringVar,
        syntax_theme_var: tk.StringVar,
        syntax_vars: dict[str, tk.StringVar],
    ) -> None:
        self.appearance_mode = appearance_var.get()
        self.color_theme = color_theme_var.get()
        self.syntax_theme_name = syntax_theme_var.get()
        self.syntax_colors = {key: var.get() for key, var in syntax_vars.items()}
        self.viewer_text_bg = self.syntax_colors.get("bg", "").strip()
        self.viewer_text_fg = self.syntax_colors.get("fg", "").strip()

        self.db.set_setting("appearance_mode", self.appearance_mode)
        self.db.set_setting("color_theme", self.color_theme)
        
        save_syntax_colors(self.db, self.syntax_theme_name, self.syntax_colors)

        ctk.set_appearance_mode(self.appearance_mode)
        ctk.set_default_color_theme(self.color_theme)
        self._apply_viewer_colors()
        self._configure_syntax_tags()
        if self.current_file_kind == "MSX BASIC":
            if self.current_msx_segments:
                self._set_msx_text(self.current_msx_segments)

        dialog.destroy()

    def _open_basic_editor(self) -> None:
        self.focus()

    def _open_help_viewer(self) -> None:
        viewer = HelpViewer(self)
        viewer.focus()


def main() -> None:
    from msx_basic_editor import MSXBasicEditor
    app = MSXBasicEditor()
    app.mainloop()


if __name__ == "__main__":
    main()
