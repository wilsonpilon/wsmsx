from __future__ import annotations

import os
import tkinter as tk
from tkinter import messagebox, Canvas

import customtkinter as ctk
from PIL import Image, ImageTk


class LayoutViewerFrame(ctk.CTkFrame):
    def __init__(self, parent: ctk.CTk, file_path: str | None = None) -> None:
        super().__init__(parent)

        self.current_pil_image: Image.Image | None = None
        self.current_filename: str | None = None
        self.tk_img: ImageTk.PhotoImage | None = None

        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)

        header_frame = ctk.CTkFrame(self)
        header_frame.grid(row=0, column=0, sticky="ew", padx=10, pady=(10, 5))
        ctk.CTkLabel(
            header_frame,
            text="Visualizador de Layout (.LAY)",
            font=ctk.CTkFont(size=16, weight="bold"),
        ).pack(pady=6)

        main_frame = ctk.CTkFrame(self)
        main_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=5)
        main_frame.grid_rowconfigure(0, weight=1)
        main_frame.grid_columnconfigure(0, weight=1)

        self.canvas = Canvas(main_frame, bg="black", highlightthickness=0)
        self.canvas.grid(row=0, column=0, sticky="nsew", padx=5, pady=5)

        footer_frame = ctk.CTkFrame(self)
        footer_frame.grid(row=2, column=0, sticky="ew", padx=10, pady=(5, 10))

        self.lbl_info = ctk.CTkLabel(footer_frame, text="Carregue um arquivo .LAY")
        self.lbl_info.pack(side="left", padx=10, pady=10)

        if file_path:
            self.set_file(file_path)

    def set_file(self, path: str) -> None:
        try:
            buffer = self._decode_graphos_lay(path)
            self.current_filename = os.path.basename(path)
            self._render_buffer_to_screen(buffer)
            self.lbl_info.configure(
                text=f"Arquivo: {self.current_filename} | Tamanho Decodificado: {len(buffer)} bytes"
            )
        except Exception as exc:
            messagebox.showerror("Erro de Leitura", f"Falha ao ler o arquivo .LAY:\n{exc}")

    def _decode_graphos_lay(self, filepath: str) -> bytearray:
        decoded_buffer = bytearray()

        with open(filepath, "rb") as handle:
            handle.read(3)
            byte_e = handle.read(1)
            byte_f = handle.read(1)

            if not byte_e or not byte_f:
                raise ValueError("Arquivo muito curto ou cabecalho invalido.")

            e_val = byte_e[0]
            f_val = byte_f[0]
            counter = (f_val * 256) + e_val + 1 - 0x9200

            handle.read(2)

            max_size = 0x1800
            while counter > 0 and len(decoded_buffer) < max_size:
                char = handle.read(1)
                if not char:
                    break

                raw_val = char[0]
                counter -= 1

                if raw_val >= 0x99:
                    val = raw_val - 0x99
                else:
                    val = raw_val + 0x67

                if val == 0x00 or val == 0xFF:
                    count_char = handle.read(1)
                    if count_char:
                        count = count_char[0]
                        for _ in range(count):
                            if len(decoded_buffer) < max_size:
                                decoded_buffer.append(val)
                else:
                    decoded_buffer.append(val)

        if len(decoded_buffer) < 6144:
            decoded_buffer.extend(b"\x00" * (6144 - len(decoded_buffer)))

        return decoded_buffer

    def _render_buffer_to_screen(self, buffer: bytearray) -> None:
        width, height = 256, 192
        img = Image.new("RGB", (width, height), (0, 0, 0))
        pixels = img.load()

        for t in range(3):
            for a in range(256):
                base_offset = (a * 8) + (t * 0x800)
                tile_x = (a % 32) * 8
                tile_y = (a // 32) * 8 + (t * 64)

                for d in range(8):
                    if base_offset + d >= len(buffer):
                        break

                    p = buffer[base_offset + d]
                    for bit in range(8):
                        if (p & (0x80 >> bit)) != 0:
                            px = tile_x + bit
                            py = tile_y + d
                            pixels[px, py] = (255, 255, 255)

        self.current_pil_image = img
        self._update_canvas_image(2)

    def _update_canvas_image(self, zoom_factor: int) -> None:
        if not self.current_pil_image:
            return

        w, h = self.current_pil_image.size
        new_w = w * zoom_factor
        new_h = h * zoom_factor
        img_zoomed = self.current_pil_image.resize((new_w, new_h), Image.Resampling.NEAREST)
        self.tk_img = ImageTk.PhotoImage(img_zoomed)

        self.canvas.delete("all")
        c_width = self.canvas.winfo_width()
        c_height = self.canvas.winfo_height()
        cx = c_width // 2
        cy = c_height // 2

        self.canvas.create_image(cx, cy, image=self.tk_img, anchor="center")
        self.canvas.create_rectangle(
            cx - new_w // 2,
            cy - new_h // 2,
            cx + new_w // 2,
            cy + new_h // 2,
            outline="white",
        )
