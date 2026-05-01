import requests
import os
import hashlib
from datetime import datetime

class FileHunterSyncer:
    BASE_URL = "https://download.file-hunter.com/"
    HEADERS = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    }

    def __init__(self, db_manager, status_callback):
        self.db = db_manager
        self.log = status_callback

    def check_for_updates(self):
        config = self.db.get_config()
        base_url = config.get('filehunter_url') if config else self.BASE_URL
        if not base_url.endswith('/'): base_url += '/'

        self.log("Verificando atualizações no servidor...")
        try:
            response = requests.get(f"{base_url}allfiles.txt", headers=self.HEADERS, stream=True)
            remote_date = response.headers.get('Last-Modified')
            response.close()

            last_update = config.get('last_update') if config else None
            db_empty = self.db.is_database_empty()

            if db_empty or last_update != remote_date:
                self.log(f"Sincronização necessária: {'Banco vazio' if db_empty else 'Nova data detectada'}")
                self.sync_files()
                self.db.update_last_sync(remote_date)
                self.log("Sincronização concluída com sucesso!")
            else:
                self.log("O banco de dados já está atualizado.")
        except Exception as e:
            self.log(f"Erro na sincronização: {str(e)}")

    def sync_files(self):
        # AllFiles
        self.log("Baixando listagem de arquivos...")
        r = requests.get(f"{self.BASE_URL}allfiles.txt", headers=self.HEADERS)
        # Normaliza removendo './' e garante que vazios sejam descartados
        lines = [l.strip().lstrip("./") for l in r.text.splitlines() if l.strip()]
        self.db.clear_and_populate_files(lines)

        # SHA1
        self.log("Baixando hashes SHA1...")
        r = requests.get(f"{self.BASE_URL}sha1sums.txt", headers=self.HEADERS)
        sha_data = []
        for line in r.text.splitlines():
            parts = line.split(None, 1)
            if len(parts) == 2:
                clean_path = parts[1].strip().lstrip("./").replace("\\", "/")
                sha_data.append((parts[0].strip(), clean_path))

        self.db.clear_and_populate_hashes(sha_data)

    def download_file(self, remote_path):
        """Baixa o arquivo, cria estrutura local e valida integridade."""
        # Garante que o remote_path usado para busca no banco esteja limpo
        clean_remote_path = remote_path.lstrip("./").replace("\\", "/")

        local_dir = "downloads"
        full_local_path = os.path.join(local_dir, clean_remote_path.replace("/", os.sep))
        os.makedirs(os.path.dirname(full_local_path), exist_ok=True)

        url = f"{self.BASE_URL}{remote_path}"
        try:
            r = requests.get(url, headers=self.HEADERS, timeout=30)
            r.raise_for_status()

            with open(full_local_path, "wb") as f:
                f.write(r.content)

            sha1_local = hashlib.sha1(r.content).hexdigest()
            # Busca usando o caminho limpo
            sha1_remoto = self.db.get_sha1_for_file(clean_remote_path)

            if sha1_remoto:
                if sha1_local.lower() == sha1_remoto.lower():
                    self.log(f"✅ OK: {remote_path}")
                    return "success"
                else:
                    self.log(f"❌ ERRO HASH: {remote_path}")
                    return "danger"
            else:
                self.db.add_sha1(remote_path, sha1_local)
                self.log(f"ℹ️ NOVO HASH: {remote_path}")
                return "warning"
        except Exception as e:
            self.log(f"❌ ERRO DOWNLOAD: {str(e)}")
            return "error"