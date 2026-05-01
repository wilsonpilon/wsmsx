import customtkinter as ctk
import tkinter as tk
from tkinter import messagebox
from PIL import Image, ImageTk
import fitz  # PyMuPDF
import os


class PDFViewerWindow(ctk.CTkToplevel):
    def __init__(self, parent, file_path, title="Visualizador de PDF"):
        super().__init__(parent)
        self.title(f"FileHunter - {title}")
        self.geometry("1000x900")

        self.file_path = file_path
        self.current_page = 0
        self.zoom = 1.5  # Zoom inicial

        try:
            self.doc = fitz.open(file_path)
            self.total_pages = len(self.doc)
        except Exception as e:
            self.doc = None
            messagebox.showerror("Erro", f"Não foi possível abrir o PDF: {e}")
            self.destroy()
            return

        # --- Toolbar ---
        self.toolbar = ctk.CTkFrame(self, height=50)
        self.toolbar.pack(side="top", fill="x", padx=10, pady=5)

        self.btn_prev = ctk.CTkButton(self.toolbar, text="<< Anterior", width=80, command=self.prev_page)
        self.btn_prev.pack(side="left", padx=5)

        self.page_label = ctk.CTkLabel(self.toolbar, text=f"Página 1 de {self.total_pages}")
        self.page_label.pack(side="left", padx=10)

        self.btn_next = ctk.CTkButton(self.toolbar, text="Próxima >>", width=80, command=self.next_page)
        self.btn_next.pack(side="left", padx=5)

        # Opções de Zoom
        self.zoom_options = [
            "50%", "75%", "100%", "125%", "150%", "200%",
            "Ajustar Largura", "Ajustar Altura", "Ajustar Página"
        ]
        self.zoom_menu = ctk.CTkOptionMenu(
            self.toolbar,
            values=self.zoom_options,
            command=self.change_zoom,
            width=140
        )
        self.zoom_menu.set("150%")
        self.zoom_menu.pack(side="right", padx=10)

        ctk.CTkLabel(self.toolbar, text="Zoom:").pack(side="right", padx=2)

        # --- Área de Exibição ---
        self.scroll_canvas = ctk.CTkScrollableFrame(self, fg_color="#525659")
        self.scroll_canvas.pack(fill="both", expand=True, padx=10, pady=10)

        self.pdf_label = tk.Label(self.scroll_canvas, bg="#525659")
        self.pdf_label.pack(pady=20)

        self.render_page()
        self.after(200, self.focus_force)

    def change_zoom(self, choice):
        if not self.doc: return

        page = self.doc.load_page(self.current_page)
        rect = page.rect

        # Dimensões disponíveis (descontando margens aproximadas do frame)
        available_width = self.scroll_canvas.winfo_width() - 60
        available_height = self.scroll_canvas.winfo_height() - 60

        # Prevenção contra valores zerados antes da janela renderizar totalmente
        if available_width <= 0: available_width = 800
        if available_height <= 0: available_height = 600

        if choice == "Ajustar Largura":
            self.zoom = available_width / rect.width
        elif choice == "Ajustar Altura":
            self.zoom = available_height / rect.height
        elif choice == "Ajustar Página":
            zoom_w = available_width / rect.width
            zoom_h = available_height / rect.height
            self.zoom = min(zoom_w, zoom_h)
        else:
            # Porcentagens fixas (ex: "100%" -> 1.0)
            try:
                self.zoom = int(choice.replace("%", "")) / 100
            except ValueError:
                self.zoom = 1.0

        self.render_page()

    def render_page(self):
        if not self.doc: return

        # Renderiza a página como imagem usando a matriz de zoom
        page = self.doc.load_page(self.current_page)
        mat = fitz.Matrix(self.zoom, self.zoom)
        pix = page.get_pixmap(matrix=mat)

        # Converte para formato PIL e depois PhotoImage para o Tkinter
        img = Image.frombytes("RGB", [pix.width, pix.height], pix.samples)
        self.photo = ImageTk.PhotoImage(img)

        self.pdf_label.configure(image=self.photo)
        self.page_label.configure(text=f"Página {self.current_page + 1} de {self.total_pages}")

    def next_page(self):
        if self.current_page < self.total_pages - 1:
            self.current_page += 1
            self.render_page()

    def prev_page(self):
        if self.current_page > 0:
            self.current_page -= 1
            self.render_page()