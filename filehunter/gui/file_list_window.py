import customtkinter as ctk
from tkinter import ttk, messagebox
import re
import os
import threading
import time
import zipfile
import glob
import shutil
from datetime import datetime
from gui.disk_manager_window import DiskManagerWindow
from gui.image_viewer_window import ImageViewerWindow
from gui.file_config_window import FileConfigWindow
from gui.text_viewer_window import TextViewerWindow
from gui.pdf_viewer_window import PDFViewerWindow
from support.msx_bridge import OpenMSXBridge


class AllFilesWindow(ctk.CTkToplevel):
    def __init__(self, parent, db, syncer, embed=False):
        if embed:
            self.master = parent
        else:
            super().__init__(parent)
            self.title("FileHunter - Gerenciador de Arquivos (Modo Explorer)")
            self.geometry("1200x800")

        self.db = db
        self.syncer = syncer

        # Estado
        self.selected_category_id = None
        self.all_data = []  # Cache de arquivos da categoria atual
        self.filtered_data = []
        self.sort_asc = True
        self.current_page = 0
        self.items_per_page = 50
        self.selected_row_frame = None  # Armazena a linha selecionada atualmente

        # Inicializa a bridge com o executável configurado
        config = self.db.get_config()
        openmsx_exe = config.get('openmsx_exe', 'openmsx.exe') if config else "openmsx.exe"

        self.msx_bridge = OpenMSXBridge(executable=openmsx_exe)
        self.msx_bridge.on_output_received = self.update_status

        # Backup do callback original do syncer
        self.original_status_callback = self.syncer.log
        self.syncer.log = self.update_status

        self.setup_ui(embed=embed)
        self.load_root_categories()
        self.apply_search()

        if not embed:
            self.protocol("WM_DELETE_WINDOW", self.on_close)

    def update_status(self, message):
        if hasattr(self, "status_box") and self.status_box.winfo_exists():
            try:
                prefix = "openMSX >> " if "openMSX:" not in message and "> Enviado" not in message else ""
                self.status_box.insert("end", f"[{datetime.now().strftime('%H:%M:%S')}] {prefix}{message}\n")
                self.status_box.see("end")
            except Exception:
                pass

    def setup_ui(self, embed=False):
        container = self.master if embed else self

        if embed:
            root_window = container.winfo_toplevel()
            root_window.geometry("1200x800")
            root_window.title("FileHunter MSX Manager - Explorer")

        # 1. Barra Superior
        top_frame = ctk.CTkFrame(container)
        top_frame.pack(side="top", fill="x", padx=10, pady=10)

        ctk.CTkButton(top_frame, text="Sair", width=80, fg_color="#A13333", hover_color="#7A2626",
                      command=self.quit_application).pack(side="left", padx=5)

        def get_root():
            return container.winfo_toplevel()

        ctk.CTkButton(top_frame, text="Configurações", width=120,
                      command=lambda: get_root().open_settings()).pack(side="left", padx=5)

        ctk.CTkButton(top_frame, text="Discos", width=100, fg_color="#1f538d",
                      command=self.open_disk_manager).pack(side="left", padx=5)

        self.btn_sync = ctk.CTkButton(top_frame, text="Sincronizar Banco", width=140, fg_color="#2E7D32",
                                      hover_color="#1B5E20", command=self.start_sync_thread)
        self.btn_sync.pack(side="left", padx=5)

        self.search_entry = ctk.CTkEntry(top_frame, placeholder_text="Filtrar nesta pasta (Regex)...")
        self.search_entry.pack(side="left", fill="x", expand=True, padx=5)
        self.search_entry.bind("<Return>", lambda e: self.apply_search())

        ctk.CTkButton(top_frame, text="Buscar", width=80, command=self.apply_search).pack(side="left", padx=2)
        ctk.CTkButton(top_frame, text="Limpar", width=80, fg_color="#A13333", command=self.clear_search).pack(
            side="left", padx=2)

        # --- ÁREA DE COMANDO MSX ---
        self.msx_cmd_frame = ctk.CTkFrame(container)
        self.msx_cmd_frame.pack(side="bottom", fill="x", padx=10, pady=(10, 0))

        self.msx_entry = ctk.CTkEntry(
            self.msx_cmd_frame,
            placeholder_text="Comando openMSX (ex: set pause on, screenshot, reset)..."
        )
        self.msx_entry.pack(side="left", fill="x", expand=True, padx=(0, 5))
        self.msx_entry.bind("<Return>", lambda e: self.send_msx_command())

        self.btn_send_msx = ctk.CTkButton(
            self.msx_cmd_frame,
            text="Enviar",
            width=80,
            command=self.send_msx_command
        )
        self.btn_send_msx.pack(side="right")

        # 2. Console de Status
        self.status_box = ctk.CTkTextbox(container, height=100)
        self.status_box.pack(side="bottom", fill="x", padx=10, pady=(0, 10))

        # 3. Container Principal
        self.main_container = ctk.CTkFrame(container, fg_color="transparent")
        self.main_container.pack(fill="both", expand=True, padx=10, pady=0)

        self.left_panel = ctk.CTkFrame(self.main_container, width=300)
        self.left_panel.pack(side="left", fill="y", padx=(0, 5))

        ctk.CTkLabel(self.left_panel, text="Diretórios", font=("Arial", 14, "bold")).pack(pady=5)

        style = ttk.Style()
        style.configure("Treeview", rowheight=25)
        self.tree = ttk.Treeview(self.left_panel, show="tree")
        self.tree.pack(fill="both", expand=True, padx=5, pady=5)
        self.tree.bind("<<TreeviewOpen>>", self.on_tree_expand)
        self.tree.bind("<<TreeviewSelect>>", self.on_tree_select)

        self.right_panel = ctk.CTkFrame(self.main_container)
        self.right_panel.pack(side="right", fill="both", expand=True)

        self.pagination_frame = ctk.CTkFrame(self.right_panel)
        self.pagination_frame.pack(side="bottom", fill="x", padx=5, pady=5)

        self.btn_prev = ctk.CTkButton(self.pagination_frame, text="<<", width=40, command=self.prev_page)
        self.btn_prev.pack(side="left", padx=5)

        self.page_label = ctk.CTkLabel(self.pagination_frame, text="Página 1")
        self.page_label.pack(side="left", expand=True)

        self.btn_download_all = ctk.CTkButton(
            self.pagination_frame,
            text="Baixar Todos",
            fg_color="#2E7D32",
            hover_color="#1B5E20",
            command=self.download_all_current
        )

        self.btn_next = ctk.CTkButton(self.pagination_frame, text=">>", width=40, command=self.next_page)
        self.btn_next.pack(side="right", padx=5)

        self.scroll_frame = ctk.CTkScrollableFrame(self.right_panel)
        self.scroll_frame.pack(fill="both", expand=True, padx=5, pady=5)

    def execute_file(self, local_path, relative_path=None):
        try:
            screenshot_base_dir = "screenshots"
            if relative_path:
                relative_dir = os.path.dirname(relative_path)
                target_screenshot_dir = os.path.join(screenshot_base_dir, relative_dir)
                os.makedirs(target_screenshot_dir, exist_ok=True)
            else:
                target_screenshot_dir = screenshot_base_dir

            config = self.db.get_config()
            if not config or not config.get('openmsx_exe'):
                messagebox.showwarning("Configuração", "Configure o executável do openMSX.")
                return

            file_cfg = self.db.get_file_config(relative_path) if relative_path else None

            if file_cfg:
                machine, media_type, ext1, ext2 = file_cfg[0], file_cfg[1], file_cfg[2], file_cfg[3]
                exts = [ext1, ext2]
            else:
                machine = config.get('default_msx_machine')
                media_type = "Auto"
                exts = [config.get(f'ext{i}') for i in range(1, 5)]

            abs_local_path = os.path.abspath(local_path)
            path_upper = local_path.upper()
            media_args = []

            if media_type == "ROM" or (
                    media_type == "Auto" and any(path_upper.endswith(e) for e in [".ROM", ".MX1", ".MX2"])):
                media_args.extend(["-carta", abs_local_path])
            elif media_type == "DSK" or (media_type == "Auto" and path_upper.endswith(".DSK")):
                media_args.extend(["-diska", abs_local_path])
            elif media_type == "CAS" or (media_type == "Auto" and path_upper.endswith(".CAS")):
                media_args.extend(["-cassetteplayer", abs_local_path])
            else:
                media_args.append(abs_local_path)

            extra_args = []
            if machine and machine != "_nenhuma_":
                extra_args.extend(["-machine", machine])
            for ext in exts:
                if ext and ext != "_nenhuma_":
                    extra_args.extend(["-ext", ext])
            extra_args.extend(media_args)

            self.update_status("Iniciando openMSX...")
            self.msx_bridge.start(extra_args=extra_args)

            if relative_path:
                threading.Thread(
                    target=self.monitor_and_collect_screenshots,
                    args=(relative_path, target_screenshot_dir),
                    daemon=True
                ).start()

        except Exception as e:
            self.update_status(f"Erro: {e}")
            messagebox.showerror("Erro", str(e))

    def monitor_and_collect_screenshots(self, relative_path, target_dir):
        if not self.msx_bridge:
            return

        base_filename = os.path.basename(relative_path)
        while '.' in base_filename:
            base_filename = os.path.splitext(base_filename)[0]

        docs_dir = os.path.join(os.path.expanduser("~"), "Documents", "openMSX", "screenshots")
        abs_target = os.path.abspath(target_dir)
        mask = f"{base_filename} [0-9][0-9][0-9][0-9].png"

        self.update_status(f"Monitoramento Iniciado:")
        self.update_status(f"  > Origem: {docs_dir}")
        self.update_status(f"  > Destino: {abs_target}")
        self.update_status(f"  > Máscara: '{mask}'")

        initial_check = glob.glob(os.path.join(docs_dir, mask))
        if initial_check:
            self.update_status(f"  > Já existem {len(initial_check)} imagens na origem.")

        wait_count = 0
        while True:
            if not self.msx_bridge.is_running():
                self.update_status("Detectado encerramento do openMSX.")
                break
            time.sleep(1)
            wait_count += 1
            if wait_count % 10 == 0:
                self.update_status(f"Aguardando fechamento do emulador... ({wait_count}s)")

        try:
            self.update_status(f"--- Fim da Execução: Coletando Imagens ---")
            if not os.path.exists(docs_dir):
                self.update_status("Erro: Pasta de origem não encontrada.")
                return

            found_screenshots = glob.glob(os.path.join(docs_dir, mask))
            if found_screenshots:
                self.update_status(f"Processando {len(found_screenshots)} imagens...")
                os.makedirs(abs_target, exist_ok=True)
                for src_path in found_screenshots:
                    filename = os.path.basename(src_path)
                    dest_path = os.path.join(abs_target, filename)
                    self.update_status(f"Movendo: {filename}")
                    shutil.move(src_path, dest_path)
                self.update_status("Concluído: Screenshots movidas.")
            else:
                self.update_status(f"Nenhuma imagem encontrada com a máscara '{mask}'.")
                files = os.listdir(docs_dir)
                if files:
                    self.update_status(f"Dica: Existem outros arquivos na pasta: {files[:3]}...")
            self.update_status("---------------------------------------")
        except Exception as e:
            self.update_status(f"Erro no coletor: {e}")

    def send_msx_command(self):
        if not hasattr(self, 'msx_entry'): return
        command = self.msx_entry.get().strip()
        if not command: return
        if self.msx_bridge and self.msx_bridge.is_running():
            self.msx_bridge.send_command(command)
            self.msx_entry.delete(0, "end")
        else:
            self.update_status("Erro: openMSX não está em execução.")

    def start_sync_thread(self):
        self.btn_sync.configure(state="disabled", text="Sincronizando...")
        thread = threading.Thread(target=self.run_sync, daemon=True)
        thread.start()

    def run_sync(self):
        try:
            self.syncer.check_for_updates()
            self.master.after(0, self.finalize_sync)
        except Exception as e:
            self.master.after(0, lambda: self.update_status(f"Erro: {e}"))
            self.master.after(0, lambda: self.btn_sync.configure(state="normal", text="Sincronizar Banco"))

    def finalize_sync(self):
        self.btn_sync.configure(state="normal", text="Sincronizar Banco")
        self.load_root_categories()
        messagebox.showinfo("Sucesso", "Banco atualizado!")

    def on_close(self):
        if self.msx_bridge:
            self.msx_bridge.stop()
        self.syncer.log = self.original_status_callback

        if hasattr(self, "master") and isinstance(self, ctk.CTkToplevel):
            try:
                self.destroy()
            except Exception:
                pass

    def load_root_categories(self):
        for i in self.tree.get_children(): self.tree.delete(i)
        for cat_id, name in self.db.get_categories(None):
            node = self.tree.insert("", "end", text=name, iid=f"cat_{cat_id}", open=False)
            self.tree.insert(node, "end", text="_dummy")

    def on_tree_expand(self, event):
        node_id = self.tree.focus()
        if not node_id.startswith("cat_"): return
        children = self.tree.get_children(node_id)
        if len(children) == 1 and self.tree.item(children[0], "text") == "_dummy":
            self.tree.delete(children[0])
            cat_id = int(node_id.split("_")[1])
            for sid, sname in self.db.get_categories(cat_id):
                snode = self.tree.insert(node_id, "end", text=sname, iid=f"cat_{sid}", open=False)
                self.tree.insert(snode, "end", text="_dummy")

    def on_tree_select(self, event):
        selected = self.tree.selection()
        if not selected or not selected[0].startswith("cat_"): return
        cat_id = int(selected[0].split("_")[1])
        if not self.db.get_categories(cat_id):
            self.btn_download_all.pack(side="right", padx=10)
        else:
            self.btn_download_all.pack_forget()
        self.selected_category_id = cat_id
        self.all_data = self.db.get_all_files(category_id=cat_id)
        self.apply_search()

    def apply_search(self):
        pattern = self.search_entry.get()
        source = self.all_data if self.selected_category_id else self.db.get_all_files()
        if not pattern:
            self.filtered_data = list(source)
        else:
            try:
                regex = re.compile(pattern, re.IGNORECASE)
                self.filtered_data = [f for f in source if regex.search(f)]
            except:
                self.filtered_data = []
        self.current_page = 0
        self.refresh_list()

    def clear_search(self):
        self.search_entry.delete(0, "end")
        self.apply_search()

    def select_row(self, row_frame):
        # Desmarca a linha anterior
        if self.selected_row_frame and self.selected_row_frame.winfo_exists():
            self.selected_row_frame.configure(fg_color="transparent")

        # Marca a nova linha
        self.selected_row_frame = row_frame
        row_frame.configure(fg_color="#3B3B3B")  # Cor de destaque (cinza escuro)

    def refresh_list(self):
        for widget in self.scroll_frame.winfo_children(): widget.destroy()
        self.selected_row_frame = None

        start = self.current_page * self.items_per_page
        page_items = self.filtered_data[start:start + self.items_per_page]
        total_pages = max(1, (len(self.filtered_data) + self.items_per_page - 1) // self.items_per_page)
        self.page_label.configure(text=f"Pag {self.current_page + 1}/{total_pages} ({len(self.filtered_data)} arq)")

        if hasattr(self, "master") and self.master:
            actual_parent = self.master.winfo_toplevel()
        else:
            actual_parent = self.winfo_toplevel() if hasattr(self, "tk") else self

        for path in page_items:
            # Row Frame: fg_color transparente por padrão, cursor de mão
            row = ctk.CTkFrame(self.scroll_frame, fg_color="transparent")
            row.pack(fill="x", pady=0, padx=2)  # pady=0 para ficar bem colado

            filename = path.split('/')[-1]
            local_path = os.path.join("downloads", path.replace("/", os.sep))
            is_txt = filename.lower().endswith('.txt')
            is_pdf = filename.lower().endswith('.pdf')

            # Busca screenshots correspondentes de forma simplificada
            # path: "Games/MSX1/DSK/alcatraz-bra.zip"
            # alvo: "screenshots/Games/MSX1/DSK/alcatraz-bra*.png"

            # Remove a extensão do path original (ex: .zip) e troca as barras
            rel_path_no_ext = os.path.splitext(path)[0].replace("/", os.sep)

            # Junta com o prefixo screenshots e o wildcard para as imagens
            pattern = os.path.join("screenshots", rel_path_no_ext + "*.png")

            # Busca as imagens usando o caminho absoluto para garantir precisão
            screenshots = sorted(glob.glob(os.path.abspath(pattern)))

            is_real_container = False
            display_icon = ""
            if filename.lower().endswith('.zip') and os.path.exists(local_path):
                if self.is_nested_archive(local_path):
                    is_real_container = True
                    display_icon = "📦 "

            # Label de Nome: bind de clique para seleção visual
            lbl = ctk.CTkLabel(row, text=f"{display_icon}{filename}", anchor="w", cursor="hand2")
            lbl.pack(side="left", fill="x", expand=False, padx=5)
            lbl.bind("<Button-1>", lambda e, r=row: self.select_row(r))

            actions_frame = ctk.CTkFrame(row, fg_color="transparent")
            actions_frame.pack(side="right")

            if os.path.exists(local_path):
                if is_real_container:
                    ctk.CTkButton(actions_frame, text="Descompactar", width=100, fg_color="#E67E22",
                                  command=lambda lp=local_path, rp=path: self.handle_unzip_container(lp, rp)).pack(
                        side="right", padx=2, pady=1)
                elif is_txt:
                    ctk.CTkButton(actions_frame, text="View", width=60, fg_color="#2E7D32",
                                  hover_color="#1B5E20",
                                  command=lambda lp=local_path: TextViewerWindow(actual_parent, lp, filename)).pack(
                        side="right", padx=2, pady=1)
                elif is_pdf:
                    ctk.CTkButton(actions_frame, text="View", width=60, fg_color="#2E7D32",
                                  hover_color="#1B5E20",
                                  command=lambda lp=local_path: PDFViewerWindow(actual_parent, lp, filename)).pack(
                        side="right", padx=2, pady=1)
                else:
                    ctk.CTkButton(actions_frame, text="Exec", width=60, fg_color="#2E7D32",
                                  command=lambda lp=local_path, rp=path: self.execute_file(lp, rp)).pack(side="right",
                                                                                                         padx=2, pady=1)

                if screenshots:
                    ctk.CTkButton(actions_frame, text="View", width=60, fg_color="#1f538d",
                                  command=lambda s=screenshots, f=filename: ImageViewerWindow(actual_parent, s, f)).pack(
                        side="right", padx=2, pady=1)

                if not is_txt and not is_pdf and not is_real_container:
                    ctk.CTkButton(actions_frame, text="Config", width=60,
                                  command=lambda p=path: self.open_file_config(p)).pack(side="right", padx=2, pady=1)
            else:
                ctk.CTkButton(actions_frame, text="Baixar", width=60,
                              command=lambda p=path: self.handle_download(p)).pack(side="right", padx=2, pady=1)

    def is_nested_archive(self, zip_path):
        try:
            if not zipfile.is_zipfile(zip_path):
                return False
            with zipfile.ZipFile(zip_path, 'r') as z:
                infolist = z.infolist()
                if len(infolist) > 1:
                    return True
                if len(infolist) == 1:
                    name = infolist[0].filename.lower()
                    if name.endswith(('.zip', '.rar', '.7z')):
                        return True
                    if name.endswith(('.dsk', '.rom', '.mx1', '.mx2', '.cas')):
                        return False
            return True
        except Exception:
            pass
        return False

    def handle_unzip_container(self, local_path, relative_path):
        try:
            base_dir = os.path.dirname(local_path)
            folder_name = os.path.splitext(os.path.basename(local_path))[0]
            extract_path = os.path.join(base_dir, folder_name)
            os.makedirs(extract_path, exist_ok=True)
            with zipfile.ZipFile(local_path, 'r') as z:
                z.extractall(extract_path)

            new_files_to_register = []
            rel_base_path = os.path.dirname(relative_path)
            for root, dirs, files in os.walk(extract_path):
                for file in files:
                    abs_f_path = os.path.join(root, file)
                    rel_suffix = os.path.relpath(abs_f_path, os.path.dirname(local_path)).replace(os.sep, '/')
                    db_path = f"{rel_base_path}/{rel_suffix}".strip('/')
                    new_files_to_register.append(db_path)

            if new_files_to_register:
                self.register_extracted_files(new_files_to_register)

            messagebox.showinfo("Sucesso",
                                f"Container extraído em:\n{folder_name}\n\n{len(new_files_to_register)} novos itens adicionados.")
            self.load_root_categories()
            self.refresh_list()
        except Exception as e:
            messagebox.showerror("Erro na Extração", str(e))

    def register_extracted_files(self, file_paths):
        with self.db.get_connection() as conn:
            cursor = conn.cursor()
            for filepath in file_paths:
                parts = filepath.split('/')
                filename = parts[-1]
                directories = parts[:-1]
                current_parent_id = None
                for dir_name in directories:
                    cursor.execute(
                        "SELECT id FROM categories WHERE name = ? AND (parent_id = ? OR (parent_id IS NULL AND ? IS NULL))",
                        (dir_name, current_parent_id, current_parent_id))
                    row = cursor.fetchone()
                    if row:
                        current_parent_id = row[0]
                    else:
                        cursor.execute("INSERT INTO categories (name, parent_id) VALUES (?, ?)",
                                       (dir_name, current_parent_id))
                        current_parent_id = cursor.lastrowid
                cursor.execute("SELECT 1 FROM allfiles WHERE filepath = ?", (filepath,))
                if not cursor.fetchone():
                    cursor.execute("INSERT INTO allfiles (filepath, filename, category_id) VALUES (?, ?, ?)",
                                   (filepath, filename, current_parent_id))
            conn.commit()

    def handle_download(self, path, silent=False):
        status = self.syncer.download_file(path)
        if status in ["success", "warning"]:
            if not silent: self.refresh_list()
            return True
        return False

    def download_all_current(self):
        if not self.filtered_data: return
        if messagebox.askyesno("Confirmar", f"Baixar {len(self.filtered_data)} arquivos?"):
            self.btn_download_all.configure(state="disabled", text="Baixando...")
            for path in self.filtered_data:
                local_path = os.path.join("downloads", path.replace("/", os.sep))
                if not os.path.exists(local_path):
                    self.handle_download(path, silent=True)
                    self.update_idletasks()
            self.btn_download_all.configure(state="normal", text="Baixar Todos")
            self.refresh_list()

    def open_file_config(self, relative_path):
        if hasattr(self, "master") and self.master:
            actual_parent = self.master.winfo_toplevel()
        else:
            actual_parent = self.winfo_toplevel() if hasattr(self, "tk") else self

        FileConfigWindow(actual_parent, self.db, relative_path)

    def open_disk_manager(self):
        DiskManagerWindow(self)

    def next_page(self):
        if (self.current_page + 1) * self.items_per_page < len(self.filtered_data):
            self.current_page += 1
            self.refresh_list()

    def prev_page(self):
        if self.current_page > 0:
            self.current_page -= 1
            self.refresh_list()

    def quit_application(self):
        self.on_close()
        try:
            root = self.winfo_toplevel()
            root.destroy()
        except Exception:
            import sys
            sys.exit(0)