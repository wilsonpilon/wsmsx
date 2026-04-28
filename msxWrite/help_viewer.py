import tkinter as tk
from tkinter import ttk
import customtkinter as ctk
from pathlib import Path

try:
    from tkinterweb import HtmlFrame
    TKINTERWEB_AVAILABLE = True
except ImportError:
    TKINTERWEB_AVAILABLE = False

class HelpViewer(ctk.CTkToplevel):
    def __init__(self, master=None, initial_file="MANUALS.html"):
        super().__init__(master)
        self.title("Documentação MSX")
        self.geometry("1100x800")
        
        self.grid_columnconfigure(0, weight=1)
        self.grid_rowconfigure(1, weight=1)
        
        self.doc_files = {
            "Manual Geral": "MANUALS.html",
            "Referência MSX": "MSX.html",
            "MSX BIOS": "MSXBIOS.html",
            "Softwares": "SOFTWARE.html"
        }
        
        self._build_ui()
        
        # Carregar arquivo inicial
        if initial_file in self.doc_files.values():
            self.file_selector.set(next(k for k, v in self.doc_files.items() if v == initial_file))
            self._load_html(initial_file)
        else:
            self._load_html("MANUALS.html")

    def _build_ui(self):
        # Header com seletor
        header = ctk.CTkFrame(self)
        header.grid(row=0, column=0, sticky="ew", padx=10, pady=10)
        
        ctk.CTkLabel(header, text="Documento:").pack(side="left", padx=10)
        
        self.file_selector = ctk.CTkComboBox(
            header, 
            values=list(self.doc_files.keys()), 
            command=self._on_file_change, 
            width=200
        )
        self.file_selector.pack(side="left", padx=10)
        self.file_selector.set("Manual Geral")

        # Navigation buttons
        self.btn_back = ctk.CTkButton(header, text="⬅ Voltar", width=80, command=self._on_back)
        self.btn_back.pack(side="left", padx=5)
        
        self.btn_forward = ctk.CTkButton(header, text="Avançar ➡", width=80, command=self._on_forward)
        self.btn_forward.pack(side="left", padx=5)

        # HTML Viewer
        if TKINTERWEB_AVAILABLE:
            self.html_view = HtmlFrame(self)
            self.html_view.grid(row=1, column=0, sticky="nsew", padx=10, pady=(0, 10))
        else:
            self.html_view = None
            error_frame = ctk.CTkFrame(self)
            error_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=(0, 10))
            
            error_msg = (
                "O módulo 'tkinterweb' não foi encontrado.\n\n"
                "Para visualizar a documentação, instale-o usando:\n"
                "pip install tkinterweb"
            )
            label = ctk.CTkLabel(error_frame, text=error_msg, text_color="red", font=("Consolas", 14))
            label.pack(expand=True)

    def _on_back(self):
        if not self.html_view: return
        try:
            self.html_view.go_back()
        except:
            pass

    def _on_forward(self):
        if not self.html_view: return
        try:
            self.html_view.go_forward()
        except:
            pass

    def _load_html(self, filename):
        if not self.html_view: return
        path = Path(filename).absolute()
        if not path.exists():
            # Tentar no diretório do script se não encontrado no root
            path = Path(__file__).parent / filename
            
        if path.exists():
            self.html_view.load_file(str(path))
        else:
            self.html_view.load_html(f"<h1>Erro</h1><p>Arquivo não encontrado: {filename}</p>")

    def _on_file_change(self, choice):
        filename = self.doc_files.get(choice)
        if filename:
            self._load_html(filename)

if __name__ == "__main__":
    app = ctk.CTk()
    viewer = HelpViewer(app)
    app.mainloop()
