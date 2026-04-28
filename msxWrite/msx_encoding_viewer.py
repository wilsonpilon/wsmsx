from __future__ import annotations

import json
import os
import tkinter as tk
from tkinter import messagebox, Canvas
from pathlib import Path

import customtkinter as ctk
from PIL import Image, ImageTk, ImageDraw, ImageFont


class MSXEncodingViewer(ctk.CTkFrame):
    def __init__(self, parent: ctk.CTk, insert_callback=None) -> None:
        super().__init__(parent)
        self.insert_callback = insert_callback
        self.font_data: bytes | None = None
        self.char_images: list[Image.Image] = []
        self.tk_table_img: ImageTk.PhotoImage | None = None
        self.charsets: dict[str, list[str]] | None = None
        self.current_table: list[str] | None = None
        
        # Layout principal
        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)

        # Header com controles
        header_frame = ctk.CTkFrame(self)
        header_frame.grid(row=0, column=0, sticky="ew", padx=10, pady=(10, 5))
        
        ctk.CTkLabel(header_frame, text="Versão MSX:").pack(side="left", padx=10)
        
        self.version_var = ctk.StringVar(value="International")
        self.version_combo = ctk.CTkComboBox(
            header_frame, 
            values=["International", "Japanese", "Brazilian", "Arabic", "Russian", "Korean"],
            variable=self.version_var,
            command=self._on_version_change
        )
        self.version_combo.pack(side="left", padx=10)

        # Área principal
        main_frame = ctk.CTkFrame(self)
        main_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=5)
        main_frame.grid_columnconfigure(0, weight=1)
        main_frame.grid_rowconfigure(0, weight=1)

        # Tabela de caracteres
        self.canvas_table = Canvas(main_frame, width=512, height=256, bg="black", highlightthickness=0)
        self.canvas_table.grid(row=0, column=0, padx=10, pady=10)
        self.canvas_table.bind("<Button-1>", self._on_table_click)

        # Buffer e controles inferiores
        bottom_frame = ctk.CTkFrame(self)
        bottom_frame.grid(row=2, column=0, sticky="ew", padx=10, pady=10)
        bottom_frame.grid_columnconfigure(1, weight=1)

        ctk.CTkLabel(bottom_frame, text="Buffer:").grid(row=0, column=0, padx=10, pady=5)
        
        self.buffer_text = ctk.CTkEntry(bottom_frame)
        self.buffer_text.grid(row=0, column=1, sticky="ew", padx=10, pady=5)
        
        btn_frame = ctk.CTkFrame(bottom_frame, fg_color="transparent")
        btn_frame.grid(row=0, column=2, padx=10)
        
        ctk.CTkButton(btn_frame, text="Limpar", command=self._clear_buffer, width=80).pack(side="left", padx=5)
        ctk.CTkButton(btn_frame, text="Inserir no Editor", command=self._insert_to_editor).pack(side="left", padx=5)

        # Garantir que os charsets estejam carregados
        self._ensure_charsets_loaded()
        # Carregar dados iniciais
        self._load_charset("International")

    def _on_version_change(self, version: str) -> None:
        self._load_charset(version)

    def _load_charset(self, version: str) -> None:
        # Seleciona a tabela correta a partir do JSON gerado do subprojeto msx-encoding
        if not self.charsets:
            self._ensure_charsets_loaded()
        # Tratar Korean como fallback para International (até termos tabela específica)
        key = version
        if version == "Korean":
            key = "International"
        table = (self.charsets or {}).get(key)
        if not table or len(table) != 256:
            # fallback robusto
            table = ["\uFFFD"] * 256
            for i in range(0x20, 0x7F):
                table[i] = chr(i)
        self.current_table = table

        self.char_images = []
        # Tenta carregar fonte TTF se disponível para desenhar
        try:
            font = ImageFont.truetype("MSX-Screen0.ttf", 16)
        except Exception:
            try:
                font = ImageFont.truetype("MSX-Screen1.ttf", 16)
            except Exception:
                font = ImageFont.load_default()

        # Gera imagens para os 256 caracteres com base na tabela
        for i in range(256):
            img = Image.new("RGB", (16, 16), color="black")
            draw = ImageDraw.Draw(img)
            mapped = table[i]
            # mapped pode ser uma string vazia ou especial; garantir caractere de fallback
            if not mapped or mapped == "\\uFFFD":
                glyph = "."
            else:
                glyph = mapped
            try:
                draw.text((2, -1), glyph, font=font, fill="white")
            except Exception:
                # fallback para evitar quebra com glifos não suportados
                draw.text((4, 0), ".", font=font, fill="white")
            self.char_images.append(img)

        self._draw_table()

    def _draw_table(self) -> None:
        if not self.char_images:
            return

        self.canvas_table.delete("all")

        # Tabela 32 colunas x 8 linhas para caber melhor (32*16=512px largura)
        cols = 32
        rows = 8
        
        full_table = Image.new("RGB", (cols * 16, rows * 16), "black")

        for idx, img in enumerate(self.char_images):
            if idx >= 256: break
            
            row = idx // cols
            col = idx % cols
            x = col * 16
            y = row * 16
            full_table.paste(img, (x, y))

        self.tk_table_img = ImageTk.PhotoImage(full_table)
        self.canvas_table.create_image(0, 0, anchor=tk.NW, image=self.tk_table_img)

    def _on_table_click(self, event: tk.Event) -> None:
        x = event.x
        y = event.y
        
        col = x // 16
        row = y // 16
        cols = 32
        
        char_index = (row * cols) + col
        
        if 0 <= char_index <= 255:
            current_text = self.buffer_text.get()
            # Usa o mapeamento real da tabela atual, caindo para \xHH se não mapeado
            mapped = None
            if self.current_table and 0 <= char_index < len(self.current_table):
                mapped = self.current_table[char_index]
            if not mapped or mapped == "\\uFFFD":
                char_to_add = f"\\x{char_index:02X}"
            else:
                char_to_add = mapped
            
            self.buffer_text.delete(0, tk.END)
            self.buffer_text.insert(0, current_text + char_to_add)

    def _clear_buffer(self) -> None:
        self.buffer_text.delete(0, tk.END)

    def _insert_to_editor(self) -> None:
        text = self.buffer_text.get()
        if not text:
            return
            
        if self.insert_callback:
            self.insert_callback(text)
            messagebox.showinfo("Sucesso", "Texto inserido no editor.")
        else:
            messagebox.showwarning("Aviso", "Editor não conectado.")

    def _ensure_charsets_loaded(self) -> None:
        if self.charsets is not None:
            return
        json_path = Path("msx_charsets.json")
        if not json_path.exists():
            # Tenta gerar automaticamente a partir do subprojeto msx-encoding
            try:
                from extract_charsets import extract_charsets
                extract_charsets()
            except Exception:
                pass
        try:
            with open(json_path, "r", encoding="utf-8") as f:
                data = json.load(f)
        except Exception:
            data = {}
        # Normaliza chaves esperadas
        # O extrator gera: International, Japanese, Brazilian, Russian, Arabic
        wanted_order = ["International", "Japanese", "Brazilian", "Arabic", "Russian"]
        self.charsets = {}
        for k in wanted_order:
            if k in data:
                self.charsets[k] = data[k]
        # Se faltar algo essencial, criar fallback básico
        if "International" not in self.charsets:
            intl = ["\uFFFD"] * 256
            for i in range(0x20, 0x7F):
                intl[i] = chr(i)
            self.charsets["International"] = intl
