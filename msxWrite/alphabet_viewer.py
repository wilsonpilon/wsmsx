from __future__ import annotations

import tkinter as tk
from tkinter import messagebox, Canvas

import customtkinter as ctk
from PIL import Image, ImageTk


class AlphabetViewerFrame(ctk.CTkFrame):
    def __init__(self, parent: ctk.CTk, file_path: str | None = None) -> None:
        super().__init__(parent)

        self.font_data: bytes | None = None
        self.char_images: list[Image.Image] = []
        self.tk_table_img: ImageTk.PhotoImage | None = None
        self.tk_detail_img: ImageTk.PhotoImage | None = None

        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)

        header_frame = ctk.CTkFrame(self)
        header_frame.grid(row=0, column=0, sticky="ew", padx=10, pady=(10, 5))
        ctk.CTkLabel(
            header_frame,
            text="Visualizador de Alfabeto (.ALF)",
            font=ctk.CTkFont(size=16, weight="bold"),
        ).pack(pady=6)

        main_frame = ctk.CTkFrame(self)
        main_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=5)
        main_frame.grid_columnconfigure(0, weight=0)
        main_frame.grid_columnconfigure(1, weight=1)
        main_frame.grid_rowconfigure(0, weight=1)

        left_panel = ctk.CTkFrame(main_frame)
        left_panel.grid(row=0, column=0, sticky="ns", padx=(0, 10), pady=10)

        ctk.CTkLabel(left_panel, text="Tabela 16x16 (Zoom 2x)").pack(pady=5)
        self.canvas_table = Canvas(left_panel, width=256, height=256, bg="black", highlightthickness=0)
        self.canvas_table.pack(padx=10, pady=10)
        self.canvas_table.bind("<Button-1>", self._on_table_click)

        right_panel = ctk.CTkFrame(main_frame)
        right_panel.grid(row=0, column=1, sticky="nsew", padx=(0, 10), pady=10)
        right_panel.grid_columnconfigure(0, weight=1)

        ctk.CTkLabel(right_panel, text="Caractere Selecionado (Zoom 16x)").pack()

        self.canvas_detail = Canvas(
            right_panel,
            width=128,
            height=128,
            bg="black",
            highlightthickness=1,
            highlightbackground="gray",
        )
        self.canvas_detail.pack(pady=10)

        self.lbl_char_info = ctk.CTkLabel(right_panel, text="Codigo: ---")
        self.lbl_char_info.pack()

        if file_path:
            self.set_file(file_path)

    def set_file(self, file_path: str) -> None:
        try:
            with open(file_path, "rb") as handle:
                _header = handle.read(7)
                data = handle.read(2048)
            if len(data) != 2048:
                raise ValueError(f"Arquivo incompleto. Esperado 2048 bytes de dados, lido {len(data)}.")

            self.font_data = data
            self._process_data()
            self._draw_table()
            self._select_char(65)
        except Exception as exc:
            messagebox.showerror("Erro", f"Falha ao ler arquivo:\n{exc}")

    def _process_data(self) -> None:
        self.char_images = []
        if not self.font_data:
            return

        for i in range(256):
            char_bytes = self.font_data[i * 8 : (i + 1) * 8]
            img = Image.new("RGB", (8, 8), color="black")
            pixels = img.load()

            for row in range(8):
                byte = char_bytes[row]
                for col in range(8):
                    if (byte >> (7 - col)) & 1:
                        pixels[col, row] = (255, 255, 255)

            self.char_images.append(img)

    def _draw_table(self) -> None:
        if not self.char_images:
            return

        self.canvas_table.delete("all")

        base_width = 16 * 8
        base_height = 16 * 8
        full_table = Image.new("RGB", (base_width, base_height), "black")

        for idx, img in enumerate(self.char_images):
            row = idx // 16
            col = idx % 16
            x = col * 8
            y = row * 8
            full_table.paste(img, (x, y))

        zoomed_table = full_table.resize((256, 256), resample=Image.NEAREST)
        self.tk_table_img = ImageTk.PhotoImage(zoomed_table)
        self.canvas_table.create_image(0, 0, anchor=tk.NW, image=self.tk_table_img)

    def _on_table_click(self, event: tk.Event) -> None:
        if not self.font_data:
            return

        x_zoom = event.x
        y_zoom = event.y
        x_orig = x_zoom // 2
        y_orig = y_zoom // 2

        col = x_orig // 8
        row = y_orig // 8
        char_index = (row * 16) + col

        if 0 <= char_index <= 255:
            self._select_char(char_index)

    def _select_char(self, index: int) -> None:
        if not self.char_images:
            return

        char_img = self.char_images[index]
        detail_img = char_img.resize((128, 128), resample=Image.NEAREST)

        self.tk_detail_img = ImageTk.PhotoImage(detail_img)
        self.canvas_detail.create_image(64, 64, anchor=tk.CENTER, image=self.tk_detail_img)

        char_display = chr(index) if 32 <= index <= 126 else "."
        self.lbl_char_info.configure(text=f"Char: {char_display} | Dec: {index} | Hex: ${index:02X}")
