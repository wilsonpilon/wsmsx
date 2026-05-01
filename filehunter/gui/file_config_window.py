import customtkinter as ctk
import os

class FileConfigWindow(ctk.CTkToplevel):
    def __init__(self, parent, db, file_relative_path):
        super().__init__(parent)
        self.title("Configurar Execução")
        self.geometry("450x650") # Aumentado para caber as 4 extensões
        self.db = db
        self.file_path = file_relative_path

        # Buscar caminho do openMSX para listar máquinas e extensões
        config = self.db.get_config()
        self.openmsx_path = os.path.dirname(config.get('openmsx_exe', '')) if config else ""

        machines = self._get_openmsx_items("machines")
        extensions = self._get_openmsx_items("extensions")

        # Carregar config atual se existir
        current = self.db.get_file_config(self.file_path)

        ctk.CTkLabel(self, text=f"Configurações para:\n{file_relative_path.split('/')[-1]}",
                     font=("Arial", 12, "bold"), wraplength=350).pack(pady=10)

        # Máquina
        ctk.CTkLabel(self, text="Máquina MSX:").pack(pady=(5, 0))
        self.machine_combo = ctk.CTkComboBox(self, values=machines, width=250)
        self.machine_combo.pack(pady=5)
        if current and current[0] in machines: self.machine_combo.set(current[0])

        # Tipo de Mídia
        ctk.CTkLabel(self, text="Executar como:").pack(pady=(5, 0))
        self.media_combo = ctk.CTkComboBox(self, values=["Auto", "ROM", "DSK", "CAS", "DirAsDisk"], width=250)
        self.media_combo.pack(pady=5)
        if current: self.media_combo.set(current[1])

        # Extensões 1 a 4
        self.ext_combos = []
        for i in range(4):
            ctk.CTkLabel(self, text=f"Extensão {i+1}:").pack(pady=(5, 0))
            combo = ctk.CTkComboBox(self, values=extensions, width=250)
            combo.pack(pady=5)
            # Se houver config salva (índices 2, 3 no seu DB atual, mas precisamos de 4)
            # Nota: Se o banco só tem 2 colunas de ext, ele pegará apenas as duas primeiras
            if current and len(current) > (i + 2) and current[i+2] in extensions:
                combo.set(current[i+2])
            else:
                combo.set("_nenhuma_")
            self.ext_combos.append(combo)

        btn_save = ctk.CTkButton(self, text="Salvar Configuração", fg_color="#2E7D32", command=self.save)
        btn_save.pack(pady=20)

    def _get_openmsx_items(self, folder):
        """Lista itens (máquinas ou extensões) do diretório do openMSX"""
        items = ["_nenhuma_"]
        if not self.openmsx_path:
            return items

        full_path = os.path.join(self.openmsx_path, "share", folder)
        if os.path.exists(full_path):
            for f in os.listdir(full_path):
                if f.endswith(".xml"):
                    items.append(f[:-4]) # Remove o .xml
        return sorted(items)

    def save(self):
        # Nota: Se o seu método save_file_config no manager.py suportar apenas 2 extensões,
        # você precisará atualizar o banco para suportar as 4.
        # Vou assumir que você passará as 4 agora.
        self.db.save_file_config(
            self.file_path,
            self.machine_combo.get(),
            self.media_combo.get(),
            self.ext_combos[0].get(),
            self.ext_combos[1].get(),
            self.ext_combos[2].get(), # Adicionado
            self.ext_combos[3].get()  # Adicionado
        )
        self.destroy()