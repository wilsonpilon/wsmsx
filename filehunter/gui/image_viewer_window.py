import customtkinter as ctk
from PIL import Image
import os


class ImageViewerWindow(ctk.CTkToplevel):
    def __init__(self, parent, image_paths, title="Slideshow"):
        super().__init__(parent)
        self.title(f"Screenshots - {title}")
        self.geometry("800x600")

        self.image_paths = image_paths
        self.current_index = 0

        self.setup_ui()
        self.show_image()

        self.bind("<Left>", lambda e: self.prev_image())
        self.bind("<Right>", lambda e: self.next_image())
        self.after(200, self.focus_force)

    def setup_ui(self):
        self.grid_rowconfigure(0, weight=1)
        self.grid_columnconfigure(0, weight=1)

        self.canvas_label = ctk.CTkLabel(self, text="", fg_color="black")
        self.canvas_label.grid(row=0, column=0, sticky="nsew", padx=10, pady=10)

        self.nav_frame = ctk.CTkFrame(self, height=50)
        self.nav_frame.grid(row=1, column=0, sticky="ew", padx=10, pady=5)

        self.btn_prev = ctk.CTkButton(self.nav_frame, text="<< Anterior", command=self.prev_image, width=100)
        self.btn_prev.pack(side="left", padx=20, pady=5)

        self.info_label = ctk.CTkLabel(self.nav_frame, text="0 / 0")
        self.info_label.pack(side="left", expand=True)

        self.btn_next = ctk.CTkButton(self.nav_frame, text="Próxima >>", command=self.next_image, width=100)
        self.btn_next.pack(side="right", padx=20, pady=5)

    def show_image(self):
        if not self.image_paths:
            return

        img_path = self.image_paths[self.current_index]
        try:
            pil_img = Image.open(img_path)

            # Redimensionamento mantendo proporção
            canvas_w = self.canvas_label.winfo_width()
            canvas_h = self.canvas_label.winfo_height()
            if canvas_w < 10: canvas_w, canvas_h = 780, 500  # Fallback inicial

            img_w, img_h = pil_img.size
            ratio = min(canvas_w / img_w, canvas_h / img_h)
            new_w, new_h = int(img_w * ratio), int(img_h * ratio)

            ctk_img = ctk.CTkImage(light_image=pil_img, dark_image=pil_img, size=(new_w, new_h))
            self.canvas_label.configure(image=ctk_img)
            self.canvas_label.image = ctk_img  # Referência

            self.info_label.configure(
                text=f"Imagem {self.current_index + 1} de {len(self.image_paths)}\n{os.path.basename(img_path)}")
        except Exception as e:
            self.canvas_label.configure(text=f"Erro ao carregar imagem:\n{e}")

    def next_image(self):
        self.current_index = (self.current_index + 1) % len(self.image_paths)
        self.show_image()

    def prev_image(self):
        self.current_index = (self.current_index - 1) % len(self.image_paths)
        self.show_image()