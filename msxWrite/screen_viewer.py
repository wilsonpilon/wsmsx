from __future__ import annotations

import tkinter as tk
from tkinter import messagebox

import customtkinter as ctk
from PIL import Image, ImageTk

MSX_PALETTE = [
    (0, 0, 0),  # 0: transparent (black)
    (0, 0, 0),  # 1: black
    (35, 178, 53),  # 2: medium green
    (109, 231, 116),  # 3: light green
    (54, 59, 236),  # 4: dark blue
    (115, 119, 246),  # 5: light blue
    (171, 53, 49),  # 6: dark red
    (74, 213, 247),  # 7: cyan
    (229, 62, 54),  # 8: medium red
    (241, 123, 117),  # 9: light red
    (201, 196, 56),  # 10: dark yellow
    (218, 215, 125),  # 11: light yellow
    (31, 138, 56),  # 12: dark green
    (176, 87, 182),  # 13: magenta
    (176, 176, 176),  # 14: gray
    (255, 255, 255),  # 15: white
]


class ScreenViewerFrame(ctk.CTkFrame):
    def __init__(self, parent: ctk.CTk, file_path: str | None = None) -> None:
        super().__init__(parent)

        self.raw_data: bytes | None = None
        self.original_image: Image.Image | None = None
        self.current_zoom = 4
        self.tk_image: ImageTk.PhotoImage | None = None

        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(0, weight=1)

        sidebar = ctk.CTkFrame(self, width=200)
        sidebar.grid(row=0, column=0, sticky="ns", padx=(10, 5), pady=10)

        ctk.CTkLabel(sidebar, text="Visualizador Screen 2", font=ctk.CTkFont(size=16, weight="bold")).pack(pady=10)

        ctk.CTkLabel(sidebar, text="Modo de Visualizacao:", anchor="w").pack(padx=10, pady=(10, 5), fill="x")
        self.mode_var = ctk.StringVar(value="normal")

        ctk.CTkRadioButton(
            sidebar,
            text="Normal (Pixels + Cor)",
            variable=self.mode_var,
            value="normal",
            command=self._update_display,
        ).pack(padx=10, pady=4, anchor="w")

        ctk.CTkRadioButton(
            sidebar,
            text="Preto e Branco",
            variable=self.mode_var,
            value="bw",
            command=self._update_display,
        ).pack(padx=10, pady=4, anchor="w")

        ctk.CTkRadioButton(
            sidebar,
            text="Apenas Cores",
            variable=self.mode_var,
            value="color",
            command=self._update_display,
        ).pack(padx=10, pady=4, anchor="w")

        ctk.CTkLabel(sidebar, text="Zoom:", anchor="w").pack(padx=10, pady=(12, 5), fill="x")
        self.zoom_combo = ctk.CTkComboBox(
            sidebar,
            values=["1x", "2x", "3x", "4x", "5x"],
            command=self._on_zoom_change,
        )
        self.zoom_combo.set("4x")
        self.zoom_combo.pack(padx=10, pady=(0, 10))

        display_area = ctk.CTkScrollableFrame(self, label_text="Visualizacao")
        display_area.grid(row=0, column=1, sticky="nsew", padx=(5, 10), pady=10)
        display_area.grid_columnconfigure(0, weight=1)

        self.image_label = ctk.CTkLabel(
            display_area,
            text="Nenhuma imagem carregada.\nSelecione um arquivo .SCR (Graphos III).",
        )
        self.image_label.pack(expand=True, pady=50)

        if file_path:
            self.set_file(file_path)

    def set_file(self, filepath: str) -> None:
        try:
            with open(filepath, "rb") as handle:
                handle.read(128)
                content = handle.read(12288)

            if len(content) < 12288:
                content = content + b"\x00" * (12288 - len(content))

            self.raw_data = content
            self._update_display()
        except Exception as exc:
            messagebox.showerror("Erro", f"Falha ao ler arquivo: {exc}")

    def _process_msx_screen2(self, data: bytes, mode: str) -> Image.Image:
        width, height = 256, 192
        img = Image.new("RGB", (width, height), "black")
        pixels = img.load()

        base_pattern = 0
        base_color = 0x1800

        for t in range(3):
            for a in range(256):
                col_char = a % 32
                row_char = a // 32

                block_offset = (a * 8) + (t * 0x800)
                screen_x_base = col_char * 8
                screen_y_base = (t * 64) + (row_char * 8)

                for d in range(8):
                    pattern_byte = data[base_pattern + block_offset + d]
                    color_byte = data[base_color + block_offset + d]

                    fg_idx = (color_byte >> 4) & 0x0F
                    bg_idx = color_byte & 0x0F

                    if mode == "bw":
                        fg_rgb = MSX_PALETTE[15]
                        bg_rgb = MSX_PALETTE[1]
                    elif mode == "color":
                        pattern_byte = 0xF0
                        fg_rgb = MSX_PALETTE[fg_idx]
                        bg_rgb = MSX_PALETTE[bg_idx]
                    else:
                        fg_rgb = MSX_PALETTE[fg_idx]
                        bg_rgb = MSX_PALETTE[bg_idx]

                    for e in range(8):
                        bit = (pattern_byte >> (7 - e)) & 1
                        if bit == 1:
                            pixels[screen_x_base + e, screen_y_base + d] = fg_rgb
                        else:
                            pixels[screen_x_base + e, screen_y_base + d] = bg_rgb

        return img

    def _update_display(self) -> None:
        if self.raw_data is None:
            return

        mode = self.mode_var.get()
        self.original_image = self._process_msx_screen2(self.raw_data, mode)

        w, h = self.original_image.size
        new_w = w * self.current_zoom
        new_h = h * self.current_zoom

        zoomed_img = self.original_image.resize((new_w, new_h), Image.NEAREST)
        self.tk_image = ImageTk.PhotoImage(zoomed_img)
        self.image_label.configure(image=self.tk_image, text="")

    def _on_zoom_change(self, value: str) -> None:
        self.current_zoom = int(value.replace("x", ""))
        self._update_display()
