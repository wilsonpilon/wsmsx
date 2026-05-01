import customtkinter as ctk
import tkinter as tk
import ctypes
import os


class TextViewerWindow(ctk.CTkToplevel):
    def __init__(self, parent, file_path, title="Visualizador de Texto"):
        super().__init__(parent)
        self.title(f"FileHunter - {title}")
        self.geometry("1100x850")
        self.configure(fg_color="#f0f0f0")

        # Estado inicial
        self.current_font_family = "Dotrice"
        self.current_font_size = 12
        self.chars_per_line = 80

        self.load_all_fonts()

        try:
            with open(file_path, 'r', encoding='utf-8', errors='replace') as f:
                self.content = f.read()
        except Exception as e:
            self.content = f"Erro ao carregar arquivo: {e}"

        # --- BARRA DE FERRAMENTAS ---
        self.toolbar = ctk.CTkFrame(self, height=50, fg_color="#dbdbdb")
        self.toolbar.pack(side="top", fill="x", padx=20, pady=(10, 0))

        # Escolha da Fonte
        ctk.CTkLabel(self.toolbar, text="Fonte:", text_color="black").pack(side="left", padx=(10, 2))
        self.combo_font = ctk.CTkComboBox(self.toolbar, values=["Dotrice", "Consolas", "Courier", "Monospace"],
                                          command=self.update_view, width=140)
        self.combo_font.set("Dotrice")
        self.combo_font.pack(side="left", padx=5)

        # Tamanho da Fonte
        ctk.CTkLabel(self.toolbar, text="Tamanho:", text_color="black").pack(side="left", padx=(10, 2))
        self.combo_size = ctk.CTkComboBox(self.toolbar, values=[str(x) for x in range(8, 26, 2)],
                                          command=self.update_view, width=70)
        self.combo_size.set("12")
        self.combo_size.pack(side="left", padx=5)

        # Largura em Caracteres
        ctk.CTkLabel(self.toolbar, text="Colunas:", text_color="black").pack(side="left", padx=(10, 2))
        self.combo_cols = ctk.CTkComboBox(self.toolbar, values=["40", "64", "80", "120", "132"],
                                          command=self.update_view, width=80)
        self.combo_cols.set("80")
        self.combo_cols.pack(side="left", padx=5)

        # --- ÁREA DE TEXTO (PAGINADA) ---
        self.main_container = ctk.CTkFrame(self, fg_color="#e0e0e0", corner_radius=0)
        self.main_container.pack(fill="both", expand=True, padx=20, pady=10)

        # Scrollbars
        self.v_scroll = ctk.CTkScrollbar(self.main_container, orientation="vertical")
        self.v_scroll.pack(side="right", fill="y")

        self.h_scroll = ctk.CTkScrollbar(self.main_container, orientation="horizontal")
        self.h_scroll.pack(side="bottom", fill="x")

        # O widget de texto principal
        self.text_area = tk.Text(
            self.main_container,
            wrap="word",  # Word wrap ativado conforme solicitado
            bg="white",
            fg="#1a1a1a",
            padx=20,
            pady=20,
            borderwidth=0,
            highlightthickness=0,
            yscrollcommand=self.v_scroll.set,
            xscrollcommand=self.h_scroll.set
        )
        self.text_area.pack(side="left", fill="both", expand=True)

        self.v_scroll.configure(command=self.text_area.yview)
        self.h_scroll.configure(command=self.text_area.xview)

        # Botão Sair
        self.btn_close = ctk.CTkButton(self, text="Fechar Impressão", fg_color="#A13333",
                                       hover_color="#7A2626", command=self.destroy)
        self.btn_close.pack(pady=10)

        # Configurar tags de cores para o zebrado
        self.text_area.tag_configure("zebra_green", background="#e8f5e9")
        self.text_area.tag_configure("zebra_white", background="white")

        self.update_view()
        self.after(200, self.focus_force)

    def load_all_fonts(self):
        support_path = os.path.abspath("support")
        if os.path.exists(support_path):
            FR_PRIVATE = 0x10
            for file in os.listdir(support_path):
                if file.lower().endswith((".otf", ".ttf")):
                    ctypes.windll.gdi32.AddFontResourceExW(os.path.join(support_path, file), FR_PRIVATE, 0)

    def update_view(self, _=None):
        self.current_font_family = self.combo_font.get()
        self.current_font_size = int(self.combo_size.get())
        self.chars_per_line = int(self.combo_cols.get())

        # Configurar Fonte
        weight = "bold" if "Dotrice" in self.current_font_family else "normal"
        f_obj = (self.current_font_family, self.current_font_size, weight)
        self.text_area.configure(font=f_obj)

        # Ajustar a largura do formulário (em caracteres)
        self.text_area.configure(width=self.chars_per_line)

        # Atualizar conteúdo e aplicar zebrado
        self.apply_content_and_zebra()

    def apply_content_and_zebra(self):
        self.text_area.configure(state="normal")
        self.text_area.delete("1.0", "end")
        self.text_area.insert("1.0", self.content)

        # Aplica o zebrado linha por linha
        lines = self.text_area.get("1.0", "end-1c").split("\n")
        for i in range(len(lines)):
            line_num = i + 1
            tag = "zebra_green" if i % 2 == 0 else "zebra_white"
            self.text_area.tag_add(tag, f"{line_num}.0", f"{line_num}.end+1c")

        self.text_area.configure(state="disabled")
