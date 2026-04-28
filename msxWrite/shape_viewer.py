from __future__ import annotations

import struct
import tkinter as tk
from tkinter import messagebox, Canvas

import customtkinter as ctk
from PIL import Image, ImageTk

MSX_PALETTE = [
    (0, 0, 0),  # 0: transparent (black)
    (0, 0, 0),  # 1: black
    (32, 192, 32),  # 2: medium green
    (96, 224, 96),  # 3: light green
    (32, 32, 224),  # 4: dark blue
    (64, 96, 224),  # 5: light blue
    (160, 32, 32),  # 6: dark red
    (64, 192, 224),  # 7: cyan
    (224, 32, 32),  # 8: medium red
    (224, 96, 96),  # 9: light red
    (192, 192, 32),  # 10: dark yellow
    (192, 192, 128),  # 11: light yellow
    (32, 128, 32),  # 12: dark green
    (192, 64, 160),  # 13: magenta
    (160, 160, 160),  # 14: gray
    (224, 224, 224),  # 15: white
]


class ShapeViewerFrame(ctk.CTkFrame):
    def __init__(self, parent: ctk.CTk, file_path: str | None = None) -> None:
        super().__init__(parent)

        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)

        header_frame = ctk.CTkFrame(self)
        header_frame.grid(row=0, column=0, sticky="ew", padx=10, pady=(10, 5))
        ctk.CTkLabel(
            header_frame,
            text="Visualizador SHP",
            font=ctk.CTkFont(size=16, weight="bold"),
        ).pack(pady=6)

        self.main_frame = ctk.CTkFrame(self)
        self.main_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=5)

        self.canvas = Canvas(self.main_frame, bg="black", highlightthickness=0)
        self.canvas.pack(fill="both", expand=True, padx=5, pady=5)

        footer_frame = ctk.CTkFrame(self)
        footer_frame.grid(row=2, column=0, sticky="ew", padx=10, pady=(5, 10))

        nav_frame = ctk.CTkFrame(footer_frame, fg_color="transparent")
        nav_frame.pack(side="left", expand=True, padx=10, pady=10)

        self.btn_prev = ctk.CTkButton(
            nav_frame,
            text="< Anterior",
            width=80,
            command=self._prev_shape,
            state="disabled",
        )
        self.btn_prev.pack(side="left", padx=5)

        self.lbl_counter = ctk.CTkLabel(nav_frame, text="0 / 0", font=("Arial", 14, "bold"))
        self.lbl_counter.pack(side="left", padx=15)

        self.btn_next = ctk.CTkButton(
            nav_frame,
            text="Proximo >",
            width=80,
            command=self._next_shape,
            state="disabled",
        )
        self.btn_next.pack(side="left", padx=5)

        self.file_path: str | None = None
        self.shape_offsets: list[int] = []
        self.current_index = -1
        self.current_pil_image: Image.Image | None = None
        self.tk_img: ImageTk.PhotoImage | None = None

        if file_path:
            self.set_file(file_path)

    def set_file(self, path: str) -> None:
        self.file_path = path
        if self._scan_file_offsets(path):
            self.current_index = 0
            self._update_controls()
            self._load_shape_at_index(0)
        else:
            self._update_controls()
            messagebox.showinfo("Info", "Nenhum shape valido encontrado ou arquivo vazio.")

    def _scan_file_offsets(self, path: str) -> bool:
        self.shape_offsets = []
        try:
            with open(path, "rb") as handle:
                while True:
                    offset = handle.tell()
                    k_byte = handle.read(1)
                    if not k_byte:
                        break
                    k = struct.unpack("B", k_byte)[0]
                    if k == 0xFF:
                        break
                    header = handle.read(3)
                    if len(header) < 3:
                        break
                    t, s, h = struct.unpack("BBB", header)

                    self.shape_offsets.append(offset)

                    plane_size = s * h
                    if t == 1:
                        skip_bytes = plane_size
                    elif t == 2:
                        skip_bytes = plane_size * 2
                    elif t == 3:
                        skip_bytes = plane_size * 2
                    elif t == 4:
                        skip_bytes = plane_size * 3
                    else:
                        skip_bytes = 0
                    handle.seek(skip_bytes, 1)

            return len(self.shape_offsets) > 0
        except Exception as exc:
            messagebox.showerror("Erro", f"Erro ao indexar arquivo: {exc}")
            return False

    def _update_controls(self) -> None:
        total = len(self.shape_offsets)
        if total == 0:
            self.lbl_counter.configure(text="0 / 0")
            self.btn_prev.configure(state="disabled")
            self.btn_next.configure(state="disabled")
            return

        display_idx = self.current_index + 1
        self.lbl_counter.configure(text=f"{display_idx} / {total}")

        if self.current_index > 0:
            self.btn_prev.configure(state="normal")
        else:
            self.btn_prev.configure(state="disabled")

        if self.current_index < total - 1:
            self.btn_next.configure(state="normal")
        else:
            self.btn_next.configure(state="disabled")


    def _prev_shape(self) -> None:
        if self.current_index > 0:
            self.current_index -= 1
            self._load_shape_at_index(self.current_index)
            self._update_controls()

    def _next_shape(self) -> None:
        if self.current_index < len(self.shape_offsets) - 1:
            self.current_index += 1
            self._load_shape_at_index(self.current_index)
            self._update_controls()

    def _load_shape_at_index(self, index: int) -> None:
        if not self.file_path or index < 0 or index >= len(self.shape_offsets):
            return

        offset = self.shape_offsets[index]

        try:
            with open(self.file_path, "rb") as handle:
                handle.seek(offset)
                _k = struct.unpack("B", handle.read(1))[0]
                t, s, h = struct.unpack("BBB", handle.read(3))

                w_blocks = s // 8
                buffer_size = s * h

                buffer_cgp = bytearray(buffer_size)
                buffer_col = bytearray(buffer_size)
                buffer_msk = bytearray(buffer_size)

                if t == 1:
                    buffer_cgp = handle.read(buffer_size)
                    buffer_col = b"\xF0" * buffer_size
                elif t == 2:
                    buffer_cgp = handle.read(buffer_size)
                    buffer_col = handle.read(buffer_size)
                elif t == 3:
                    buffer_msk = handle.read(buffer_size)
                    buffer_cgp = handle.read(buffer_size)
                    buffer_col = b"\xF0" * buffer_size
                elif t == 4:
                    buffer_msk = handle.read(buffer_size)
                    buffer_cgp = handle.read(buffer_size)
                    buffer_col = handle.read(buffer_size)

                self._draw_shape(w_blocks, h, buffer_cgp, buffer_col, buffer_msk)
        except Exception as exc:
            print(f"Erro ao ler shape no index {index}: {exc}")

    def _draw_shape(self, w_tiles: int, h_tiles: int, buf_pattern, buf_color, _buf_mask) -> None:
        self.canvas.delete("all")

        px_width = w_tiles * 8
        px_height = h_tiles * 8

        img = Image.new("RGB", (px_width, px_height), "black")
        pixels = img.load()

        for ty in range(h_tiles):
            for tx in range(w_tiles):
                tile_index = (tx + ty * w_tiles) * 8
                for line in range(8):
                    byte_pos = tile_index + line
                    if byte_pos >= len(buf_pattern):
                        break

                    pattern_byte = buf_pattern[byte_pos]
                    color_byte = buf_color[byte_pos]

                    fg_rgb = MSX_PALETTE[color_byte // 16]
                    bg_rgb = MSX_PALETTE[color_byte % 16]

                    for bit in range(8):
                        is_set = (pattern_byte & (0x80 >> bit)) != 0
                        pixel_x = (tx * 8) + bit
                        pixel_y = (ty * 8) + line
                        pixels[pixel_x, pixel_y] = fg_rgb if is_set else bg_rgb

        self.current_pil_image = img

        viewer_zoom = 4
        w_zoom = px_width * viewer_zoom
        h_zoom = px_height * viewer_zoom

        img_zoomed = img.resize((w_zoom, h_zoom), Image.Resampling.NEAREST)
        self.tk_img = ImageTk.PhotoImage(img_zoomed)

        canvas_w = int(self.canvas.winfo_width())
        canvas_h = int(self.canvas.winfo_height())
        cx, cy = canvas_w // 2, canvas_h // 2

        self.canvas.create_image(cx, cy, image=self.tk_img, anchor="center")
        self.canvas.create_rectangle(
            cx - w_zoom // 2,
            cy - h_zoom // 2,
            cx + w_zoom // 2,
            cy + h_zoom // 2,
            outline="white",
            width=2,
        )
