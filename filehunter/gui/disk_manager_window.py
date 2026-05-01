import customtkinter as ctk
from tkinter import filedialog, messagebox, ttk
from support.disk_utils import MSXDiskHandler
import os
import platform
import zipfile
import tempfile
import shutil


class DiskManagerWindow(ctk.CTkToplevel):
    def __init__(self, parent):
        super().__init__(parent)
        self.title("FileHunter MSX - Gerenciador de Discos Unificado")
        self.geometry("1100x750")

        # Estado do Sistema
        self.current_disk_path = None
        self.handler = None
        self.max_capacity = 720 * 1024
        self.temp_dir = None
        self.original_path_before_zip = None

        # Caminho inicial: Tenta downloads, senão usa a home do usuário
        self.current_local_path = os.path.abspath("downloads")
        if not os.path.exists(self.current_local_path):
            self.current_local_path = os.path.expanduser("~")

        self.setup_ui()
        self.load_local_files(self.current_local_path)

        # Garante foco na janela
        self.after(200, lambda: self.focus_force())

    def setup_ui(self):
        self.grid_columnconfigure(0, weight=1)
        self.grid_columnconfigure(1, weight=1)
        self.grid_rowconfigure(1, weight=1)

        # --- Topo: Operações de Disco ---
        self.top_bar = ctk.CTkFrame(self)
        self.top_bar.grid(row=0, column=0, columnspan=2, sticky="ew", padx=10, pady=5)

        ctk.CTkButton(self.top_bar, text="Novo Disco", width=100, command=self.new_disk).pack(side="left", padx=5)
        ctk.CTkButton(self.top_bar, text="Abrir Disco", width=100, command=self.open_disk).pack(side="left", padx=5)

        self.disk_label = ctk.CTkLabel(self.top_bar, text="Nenhum disco carregado", font=("Arial", 12, "italic"))
        self.disk_label.pack(side="left", padx=20)

        ctk.CTkButton(self.top_bar, text="Salvar Como...", width=120, fg_color="#1f538d",
                      command=self.save_disk_as).pack(side="right", padx=5)

        # --- Painel Esquerdo: PC / ZIP ---
        self.pc_frame = ctk.CTkFrame(self)
        self.pc_frame.grid(row=1, column=0, sticky="nsew", padx=10, pady=5)

        ctk.CTkLabel(self.pc_frame, text="Sistema de Arquivos (PC / ZIP)", font=("Arial", 13, "bold")).pack(pady=5)

        # Endereço PC
        addr_frame = ctk.CTkFrame(self.pc_frame, fg_color="transparent")
        addr_frame.pack(fill="x", padx=5, pady=2)
        self.pc_address = ctk.CTkEntry(addr_frame)
        self.pc_address.pack(side="left", fill="x", expand=True)
        self.pc_address.bind("<Return>", lambda e: self.load_local_files(self.pc_address.get()))
        ctk.CTkButton(addr_frame, text="Ir", width=40,
                      command=lambda: self.load_local_files(self.pc_address.get())).pack(side="left", padx=2)

        self.pc_tree = ttk.Treeview(self.pc_frame, columns=("size"), show="tree headings", selectmode="extended")
        self.pc_tree.heading("#0", text="Nome")
        self.pc_tree.heading("size", text="Tamanho")
        self.pc_tree.column("size", width=90, anchor="e")
        self.pc_tree.pack(fill="both", expand=True, padx=5, pady=5)
        self.pc_tree.bind("<Double-1>", self.on_pc_double_click)

        # --- Painel Direito: MSX Disk ---
        self.msx_frame = ctk.CTkFrame(self)
        self.msx_frame.grid(row=1, column=1, sticky="nsew", padx=10, pady=5)

        ctk.CTkLabel(self.msx_frame, text="Conteúdo do Disco (MSX)", font=("Arial", 13, "bold")).pack(pady=5)

        self.disk_info_label = ctk.CTkLabel(self.msx_frame, text="Capacidade: 720 KB", font=("Arial", 11))
        self.disk_info_label.pack(pady=2)

        self.msx_tree = ttk.Treeview(self.msx_frame, columns=("size", "date"), show="tree headings",
                                     selectmode="extended")
        self.msx_tree.heading("#0", text="Arquivo")
        self.msx_tree.heading("size", text="Tamanho")
        self.msx_tree.heading("date", text="Data")
        self.msx_tree.column("size", width=90, anchor="e")
        self.msx_tree.pack(fill="both", expand=True, padx=5, pady=5)

        # --- Barra Inferior: Ações ---
        self.bottom_bar = ctk.CTkFrame(self)
        self.bottom_bar.grid(row=2, column=0, columnspan=2, sticky="ew", padx=10, pady=10)

        ctk.CTkButton(self.bottom_bar, text="Inject no Disco >>", fg_color="#2E7D32", command=self.inject_to_disk).pack(
            side="left", padx=10)
        ctk.CTkButton(self.bottom_bar, text="<< Extrair p/ PC", fg_color="#2E7D32",
                      command=self.extract_from_disk).pack(side="left", padx=10)

        ctk.CTkButton(self.bottom_bar, text="Excluir no Disco", fg_color="#A13333", command=self.delete_from_msx).pack(
            side="right", padx=10)
        ctk.CTkButton(self.bottom_bar, text="Fechar", command=self.destroy).pack(side="right", padx=10)

    # --- Navegação PC / ZIP ---
    def load_local_files(self, path):
        try:
            path = os.path.abspath(path)

            # Se for um ZIP, abre como pasta
            if os.path.isfile(path) and path.lower().endswith('.zip'):
                self.open_zip_archive(path)
                return

            items = os.listdir(path)
            self.current_local_path = path
            self.pc_address.delete(0, "end")
            self.pc_address.insert(0, path)

            for i in self.pc_tree.get_children(): self.pc_tree.delete(i)

            # Botão de Voltar
            self.pc_tree.insert("", "end", text=".. [Voltar]", iid="UP")

            # Listar Diretórios e Arquivos
            dirs = sorted([d for d in items if os.path.isdir(os.path.join(path, d))], key=str.lower)
            files = sorted([f for f in items if os.path.isfile(os.path.join(path, f))], key=str.lower)

            for d in dirs:
                full_p = os.path.join(path, d)
                self.pc_tree.insert("", "end", text=f"📁 {d}", values=("DIR"), iid=full_p)

            for f in files:
                full_p = os.path.join(path, f)
                ext = os.path.splitext(f)[1].lower()
                is_arch = ext in ['.zip', '.rar', '.7z', '.tar', '.tgz']
                size = os.path.getsize(full_p)
                icon = "📦" if is_arch else "📄"
                self.pc_tree.insert("", "end", text=f"{icon} {f}", values=(f"{size}"), iid=full_p)

        except Exception as e:
            messagebox.showerror("Erro de Navegação", str(e))

    def open_zip_archive(self, zip_path):
        try:
            self.cleanup_temp()
            self.original_path_before_zip = os.path.dirname(zip_path)
            self.temp_dir = tempfile.mkdtemp(prefix="fh_msx_")

            with zipfile.ZipFile(zip_path, 'r') as z:
                z.extractall(self.temp_dir)

            self.load_local_files(self.temp_dir)
            self.disk_label.configure(text=f"ZIP: {os.path.basename(zip_path)}", text_color="#FFCC00")
        except Exception as e:
            messagebox.showerror("Erro ZIP", f"Falha ao abrir ZIP: {e}")

    def on_pc_double_click(self, event):
        sel = self.pc_tree.selection()
        if not sel: return
        path = sel[0]

        if path == "UP":
            if self.temp_dir and self.current_local_path.startswith(self.temp_dir):
                target = self.original_path_before_zip
                self.cleanup_temp()
                self.load_local_files(target)
                self.update_disk_title()
            else:
                self.load_local_files(os.path.dirname(self.current_local_path))
        elif os.path.isdir(path) or path.lower().endswith('.zip'):
            self.load_local_files(path)

    # --- Lógica de Disco MSX ---
    def new_disk(self):
        path = filedialog.asksaveasfilename(defaultextension=".dsk", filetypes=[("MSX Disk", "*.dsk")])
        if path:
            MSXDiskHandler.create_empty_disk(path)
            self.load_msx_disk(path)

    def open_disk(self):
        path = filedialog.askopenfilename(filetypes=[("MSX Disk", "*.dsk")])
        if path:
            self.load_msx_disk(path)

    def load_msx_disk(self, path):
        self.current_disk_path = path
        self.handler = MSXDiskHandler(path)
        self.update_disk_title()
        self.refresh_msx_view()

    def update_disk_title(self):
        if self.current_disk_path:
            name = os.path.basename(self.current_disk_path)
            self.disk_label.configure(text=f"Disco: {name}", text_color="#5cb85c")
        else:
            self.disk_label.configure(text="Nenhum disco carregado", text_color="gray")

    def refresh_msx_view(self):
        for i in self.msx_tree.get_children(): self.msx_tree.delete(i)
        if not self.handler: return

        try:
            self.handler.open_disk()
            files = self.handler.list_files()
            total = 0
            for f in files:
                self.msx_tree.insert("", "end", text=f['filename'], values=(f"{f['size']} b", f"{f['date']}"))
                total += f['size']

            percent = (total / self.max_capacity) * 100
            self.disk_info_label.configure(text=f"Usado: {total / 1024:.1f} KB / 720 KB ({percent:.1f}%)")
        except Exception as e:
            messagebox.showerror("Erro DSK", f"Erro ao ler disco: {e}")

    # --- Ações de Intercâmbio ---
    def inject_to_disk(self):
        if not self.handler:
            messagebox.showwarning("Aviso", "Selecione ou crie um disco primeiro.")
            return

        selected = self.pc_tree.selection()
        # Nota: O suporte a injeção real de arquivos requer manipulação de FAT12
        # Para este protótipo, simulamos a intenção de injeção.
        count = 0
        for path in selected:
            if os.path.isfile(path):
                # handler.add_file(path) - Seria implementado em disk_utils.py
                count += 1

        if count > 0:
            messagebox.showinfo("Injeção",
                                f"{count} arquivos selecionados para injeção.\n(Implementação de escrita FAT12 necessária em disk_utils.py)")
        self.refresh_msx_view()

    def extract_from_disk(self):
        if not self.handler: return
        selected = self.msx_tree.selection()
        if not selected: return

        for item_id in selected:
            fname = self.msx_tree.item(item_id, "text")
            dest = os.path.join(self.current_local_path, fname)
            self.handler.extract_file(fname, dest)

        self.load_local_files(self.current_local_path)
        messagebox.showinfo("Sucesso", "Arquivos extraídos do disco MSX.")

    def delete_from_msx(self):
        if not self.handler: return
        selected = self.msx_tree.selection()
        if not selected or not messagebox.askyesno("Confirmar", "Deseja realmente excluir os arquivos do DSK?"): return

        for item_id in selected:
            fname = self.msx_tree.item(item_id, "text")
            self.handler.delete_file(fname)
        self.refresh_msx_view()

    def save_disk_as(self):
        if not self.current_disk_path: return
        new_path = filedialog.asksaveasfilename(defaultextension=".dsk", filetypes=[("MSX Disk", "*.dsk")])
        if new_path:
            shutil.copy2(self.current_disk_path, new_path)
            self.load_msx_disk(new_path)

    def cleanup_temp(self):
        if self.temp_dir and os.path.exists(self.temp_dir):
            try:
                shutil.rmtree(self.temp_dir)
            except:
                pass
            self.temp_dir = None

    def destroy(self):
        self.cleanup_temp()
        super().destroy()