import tkinter as tk
from tkinter import ttk
import customtkinter as ctk
from pathlib import Path

try:
    from tkinterweb import HtmlFrame
    TKINTERWEB_AVAILABLE = True
except ImportError:
    TKINTERWEB_AVAILABLE = False

from chm_parser import CHMParser

class CHMViewer(ctk.CTkToplevel):
    def __init__(self, master=None):
        super().__init__(master)
        self.title("Leitor de CHM Moderno")
        self.geometry("1100x800")
        
        self.grid_columnconfigure(0, weight=0)
        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(1, weight=1)
        
        self.current_parser = None
        self.chm_files = ["MSXBIOS.CHM", "MANUALS.CHM", "MSX.CHM", "SOFTWARE.CHM"]
        
        self._build_ui()
        self._load_chm(self.chm_files[0])

    def _build_ui(self):
        # Header with selector
        header = ctk.CTkFrame(self)
        header.grid(row=0, column=0, columnspan=2, sticky="ew", padx=10, pady=10)
        
        ctk.CTkLabel(header, text="Arquivo CHM:").pack(side="left", padx=10)
        
        self.chm_selector = ctk.CTkComboBox(header, values=self.chm_files, command=self._on_chm_change, width=200)
        self.chm_selector.pack(side="left", padx=10)
        self.chm_selector.set(self.chm_files[0])

        # Left panel (TOC)
        left_panel = ctk.CTkFrame(self, width=300)
        left_panel.grid(row=1, column=0, sticky="nsew", padx=(10, 5), pady=(0, 10))
        left_panel.grid_rowconfigure(0, weight=1)
        left_panel.grid_columnconfigure(0, weight=1)
        
        # We use a standard Treeview for TOC because CTK doesn't have one yet
        style = ttk.Style()
        theme = ctk.get_appearance_mode()
        
        if theme == "Dark":
            bg_color = "#2b2b2b"
            fg_color = "white"
            selected_color = "#1f538d"
        else:
            bg_color = "#dbdbdb"
            fg_color = "black"
            selected_color = "#3a7ebf"

        style.theme_use("default")
        style.configure("Treeview", 
                        font=("Segoe UI", 10), 
                        background=bg_color, 
                        foreground=fg_color, 
                        fieldbackground=bg_color,
                        borderwidth=0)
        style.map("Treeview", background=[('selected', selected_color)], foreground=[('selected', 'white')])
        
        self.tree = ttk.Treeview(left_panel, show="tree", selectmode="browse")
        self.tree.grid(row=0, column=0, sticky="nsew")
        self.tree.bind("<<TreeviewSelect>>", self._on_tree_select)
        self.tree.bind("<<TreeviewOpen>>", self._on_tree_open)
        self.tree.bind("<<TreeviewClose>>", self._on_tree_close)
        
        # Icons
        self.icon_folder = "📁"
        self.icon_folder_open = "📂"
        self.icon_doc = "📄"
        
        scrollbar = ttk.Scrollbar(left_panel, orient="vertical", command=self.tree.yview)
        scrollbar.grid(row=0, column=1, sticky="ns")
        self.tree.configure(yscrollcommand=scrollbar.set)

        # Right panel (HTML Content)
        right_panel = ctk.CTkFrame(self)
        right_panel.grid(row=1, column=1, sticky="nsew", padx=(5, 10), pady=(0, 10))
        right_panel.grid_rowconfigure(1, weight=1)
        right_panel.grid_columnconfigure(0, weight=1)
        
        # Navigation toolbar for HTML content
        nav_toolbar = ctk.CTkFrame(right_panel, height=40)
        nav_toolbar.grid(row=0, column=0, sticky="ew", padx=5, pady=5)
        
        self.btn_back = ctk.CTkButton(nav_toolbar, text="⬅ Voltar", width=80, command=self._on_back)
        self.btn_back.pack(side="left", padx=5)
        
        self.btn_forward = ctk.CTkButton(nav_toolbar, text="Avançar ➡", width=80, command=self._on_forward)
        self.btn_forward.pack(side="left", padx=5)
        
        self.btn_top = ctk.CTkButton(nav_toolbar, text="⬆ Topo", width=80, command=self._on_back_to_top)
        self.btn_top.pack(side="left", padx=5)

        if TKINTERWEB_AVAILABLE:
            self.html_view = HtmlFrame(right_panel, on_link_click=self._on_html_link_click)
            self.html_view.grid(row=1, column=0, sticky="nsew")
        else:
            self.html_view = None
            error_frame = ctk.CTkFrame(right_panel)
            error_frame.grid(row=1, column=0, sticky="nsew")
            
            error_msg = (
                "O módulo 'tkinterweb' não foi encontrado.\n\n"
                "Para visualizar o conteúdo CHM, instale-o usando:\n"
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

    def _on_back_to_top(self):
        if not self.html_view: return
        try:
            self.html_view.yview_moveto(0)
        except:
            pass

    def _on_html_link_click(self, url):
        # When a link is clicked, we might want to sync the tree.
        # normalize url to path
        if not url: return
        
        # tkinterweb might return file:///C:/...
        path_str = url
        if path_str.startswith("file:///"):
            path_str = path_str[8:]
        
        path_str = path_str.replace("/", "\\")
        
        # Try to find this path in the tree
        self._select_tree_by_path(path_str)

    def _select_tree_by_path(self, target_path):
        # Recursive search in tree
        def search_node(node):
            for child in self.tree.get_children(node):
                values = self.tree.item(child, "values")
                if values and values[0].lower() == target_path.lower():
                    self.tree.selection_set(child)
                    self.tree.see(child)
                    return True
                if search_node(child):
                    return True
            return False
        
        search_node("")

    def _load_chm(self, filename):
        path = Path(filename)
        if not path.exists():
            print(f"Arquivo não encontrado: {filename}")
            return
            
        self.current_parser = CHMParser(path)
        toc = self.current_parser.get_toc()
        
        # Clear tree
        for item in self.tree.get_children():
            self.tree.delete(item)
            
        self._populate_tree("", toc)
        
        # Fallback if tree is empty
        if not self.tree.get_children():
            self._build_tree_from_files()

    def _build_tree_from_files(self):
        """Fallback: build a simple tree from all .htm files if no TOC found"""
        if not self.current_parser: return
        
        temp_dir = self.current_parser.temp_dir
        htm_files = list(temp_dir.rglob("*.htm")) + list(temp_dir.rglob("*.html"))
        
        for htm in htm_files:
            rel_path = htm.relative_to(temp_dir)
            self.tree.insert("", "end", text=f"{self.icon_doc} {htm.name}", values=(str(htm),))

    def _populate_tree(self, parent, items):
        for item in items:
            icon = self.icon_doc
            if item['children']:
                icon = self.icon_folder
            
            node = self.tree.insert(parent, "end", text=f"{icon} {item['name']}", values=(item['local'],))
            if item['children']:
                self._populate_tree(node, item['children'])

    def _on_chm_change(self, choice):
        self._load_chm(choice)

    def _on_tree_open(self, event):
        item = self.tree.focus()
        if not item: return
        text = self.tree.item(item, "text")
        if text.startswith(self.icon_folder):
            new_text = text.replace(self.icon_folder, self.icon_folder_open, 1)
            self.tree.item(item, text=new_text)

    def _on_tree_close(self, event):
        item = self.tree.focus()
        if not item: return
        text = self.tree.item(item, "text")
        if text.startswith(self.icon_folder_open):
            new_text = text.replace(self.icon_folder_open, self.icon_folder, 1)
            self.tree.item(item, text=new_text)

    def _on_tree_select(self, event):
        selected_item = self.tree.selection()
        if not selected_item:
            return
            
        local_path = self.tree.item(selected_item[0], "values")[0]
        if local_path and Path(local_path).exists() and self.html_view:
            # Temporarily disable link click callback to avoid infinite loop or redundant tree selection
            self.html_view.on_link_click(None)
            self.html_view.load_file(local_path)
            self.html_view.on_link_click(self._on_html_link_click)

if __name__ == "__main__":
    app = ctk.CTk()
    viewer = CHMViewer(app)
    app.mainloop()
