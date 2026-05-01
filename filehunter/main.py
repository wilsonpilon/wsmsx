import customtkinter as ctk
from PIL import Image
import os
from database.manager import DatabaseManager
from database.syncer import FileHunterSyncer
from gui.settings_window import SettingsWindow
from gui.file_list_window import AllFilesWindow
from tkinter import messagebox
from datetime import datetime
import time


class SplashScreen(ctk.CTkToplevel):
    def __init__(self):
        super().__init__()

        # Configuração da Imagem
        img_path = os.path.join(os.path.dirname(__file__), "splashscreen.png")
        try:
            pil_img = Image.open(img_path)
            # Reduz as dimensões para 1/2 (metade)
            img_width = int(pil_img.size[0] * 0.5)
            img_height = int(pil_img.size[1] * 0.5)
        except Exception as e:
            print(f"Erro ao carregar splash: {e}")
            img_width, img_height = 400, 300  # Fallback
            pil_img = None

        self.overrideredirect(True)

        # Centralizar com base no novo tamanho reduzido
        screen_width = self.winfo_screenwidth()
        screen_height = self.winfo_screenheight()
        x = (screen_width // 2) - (img_width // 2)
        y = (screen_height // 2) - (img_height // 2)
        self.geometry(f"{img_width}x{img_height}+{x}+{y}")

        # Exibe a imagem redimensionada
        if pil_img:
            # O CTkImage aceita o tamanho que deve ser renderizado
            self.splash_img = ctk.CTkImage(light_image=pil_img,
                                           dark_image=pil_img,
                                           size=(img_width, img_height))
            self.label = ctk.CTkLabel(self, image=self.splash_img, text="")
            self.label.pack()
        else:
            self.label = ctk.CTkLabel(self, text="FileHunter MSX\n(Imagem não encontrada)", font=("Arial", 20))
            self.label.pack(expand=True)

        self.attributes("-alpha", 1.0)

    def fade_out(self):
        try:
            if not self.winfo_exists():
                return

            # Forçamos a conversão para float para evitar erros de comparação
            current_alpha = float(self.attributes("-alpha"))

            if current_alpha > 0:
                new_alpha = max(0, current_alpha - 0.1)
                self.attributes("-alpha", new_alpha)
                self.after(30, self.fade_out)
            else:
                self.destroy()
        except Exception:
            # Se a janela for fechada durante o processo, encerramos graciosamente
            try:
                self.destroy()
            except:
                pass


class FileHunterApp(ctk.CTk):
    def __init__(self):
        super().__init__()
        self.withdraw()

        self.title("FileHunter MSX Manager")
        self.geometry("600x450")

        self.db = DatabaseManager()
        self.syncer = FileHunterSyncer(self.db, self.update_status)

        # Inicializa a Splash mas garante que a UI principal está pronta
        self.splash = SplashScreen()

        self.setup_ui()
        self.apply_initial_config()

        # O uso do after aqui é correto para não bloquear o init
        self.after(3000, self.show_main_window)

    def show_main_window(self):
        if self.splash and self.splash.winfo_exists():
            self.splash.fade_out()

        # Pequeno delay para garantir que o fade iniciou antes de mostrar a principal
        self.after(100, self.deiconify)

    def setup_ui(self):
        # Criamos um container para evitar conflitos de recursão de eventos no root
        self.main_container = ctk.CTkFrame(self)
        self.main_container.pack(fill="both", expand=True)
        self.all_files_ui = AllFilesWindow(self.main_container, self.db, self.syncer, embed=True)

    def apply_initial_config(self):
        config = self.db.get_config()
        if config:
            ctk.set_appearance_mode(config.get("appearance_mode", "Dark"))
            ctk.set_default_color_theme(config.get("color_theme", "blue"))

    def update_status(self, message):
        # Redireciona para a caixa de status que agora vive dentro da AllFilesWindow (embed)
        if hasattr(self, 'all_files_ui') and self.all_files_ui:
            try:
                # Verifica se o método existe para evitar recursão se AllFilesWindow chamar de volta
                self.all_files_ui.update_status(message)
            except Exception:
                print(f"Status: {message}")

    def open_settings(self):
        SettingsWindow(self, self.db, self.apply_initial_config)

    def open_all_files(self):
        if self.db.is_database_empty():
            messagebox.showwarning("Aviso", "O banco está vazio. Sincronize primeiro!")
            return
        AllFilesWindow(self, self.db, self.syncer)


if __name__ == "__main__":
    app = FileHunterApp()
    app.mainloop()