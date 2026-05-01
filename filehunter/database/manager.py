import sqlite3
import os

class DatabaseManager:
    def __init__(self, db_path="database/filehunter.db"):
        self.db_path = db_path
        os.makedirs(os.path.dirname(self.db_path), exist_ok=True)
        self.init_db()

    def get_connection(self):
        return sqlite3.connect(self.db_path)

    def init_db(self):
        with self.get_connection() as conn:
            # Tabela de Configuração atualizada
            conn.execute("""
                         CREATE TABLE IF NOT EXISTS config
                         (
                             id
                             INTEGER
                             PRIMARY
                             KEY
                             CHECK
                         (
                             id =
                             1
                         ),
                             filehunter_url TEXT,
                             openmsx_exe TEXT,
                             default_msx_machine TEXT,
                             ext1 TEXT, ext2 TEXT, ext3 TEXT, ext4 TEXT,
                             appearance_mode TEXT,
                             color_theme TEXT,
                             last_update TEXT
                             )
                         """)

            # Garante que existe exatamente um registro
            cursor = conn.cursor()
            cursor.execute("SELECT COUNT(*) FROM config")
            if cursor.fetchone()[0] == 0:
                conn.execute("""
                             INSERT INTO config (id, filehunter_url, appearance_mode, color_theme)
                             VALUES (1, 'https://download.file-hunter.com/', 'Dark', 'blue')
                             """)

            # Nova Tabela de Categorias (Diretórios)
            conn.execute("""
                         CREATE TABLE IF NOT EXISTS categories
                         (
                             id
                             INTEGER
                             PRIMARY
                             KEY
                             AUTOINCREMENT,
                             name
                             TEXT
                             NOT
                             NULL,
                             parent_id
                             INTEGER,
                             FOREIGN
                             KEY
                         (
                             parent_id
                         ) REFERENCES categories
                         (
                             id
                         )
                             )
                         """)

            # Tabelas de dados atualizada
            conn.execute("""
                         CREATE TABLE IF NOT EXISTS allfiles
                         (
                             id
                             INTEGER
                             PRIMARY
                             KEY
                             AUTOINCREMENT,
                             filepath
                             TEXT,
                             filename
                             TEXT,
                             category_id
                             INTEGER,
                             FOREIGN
                             KEY
                         (
                             category_id
                         ) REFERENCES categories
                         (
                             id
                         )
                             )
                         """)
            conn.execute("CREATE TABLE IF NOT EXISTS sha1sums (hash TEXT, filepath TEXT)")
            conn.commit()

            conn.execute('''
                         CREATE TABLE IF NOT EXISTS file_configs
                         (
                             file_path
                             TEXT
                             PRIMARY
                             KEY,
                             machine
                             TEXT,
                             media_type
                             TEXT,
                             ext1
                             TEXT,
                             ext2
                             TEXT,
                             ext3
                             TEXT,
                             ext4
                             TEXT
                         )
                         ''')
            conn.commit()

    def clear_and_populate_files(self, data_list):
        """Limpa e insere arquivos criando a árvore de categorias."""
        with self.get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("DELETE FROM allfiles")
            cursor.execute("DELETE FROM categories")
            cursor.execute("DELETE FROM sqlite_sequence WHERE name IN ('allfiles', 'categories')")

            category_cache = {"": None}  # Cache de categorias { "caminho": id }

            for filepath in data_list:
                # Normaliza para usar barras invertidas para o processamento de diretórios
                normalized_path = filepath.replace('/', '\\')
                parts = normalized_path.split('\\')
                filename = parts[-1]
                directories = parts[:-1]

                current_parent_id = None
                full_path_acc = ""

                for dir_name in directories:
                    full_path_acc = os.path.join(full_path_acc, dir_name) if full_path_acc else dir_name

                    if full_path_acc not in category_cache:
                        cursor.execute(
                            "INSERT INTO categories (name, parent_id) VALUES (?, ?)",
                            (dir_name, current_parent_id)
                        )
                        category_cache[full_path_acc] = cursor.lastrowid

                    current_parent_id = category_cache[full_path_acc]

                cursor.execute(
                    "INSERT INTO allfiles (filepath, filename, category_id) VALUES (?, ?, ?)",
                    (filepath, filename, current_parent_id)
                )
            conn.commit()

    def clear_and_populate_hashes(self, hash_data):
        """Limpa e insere os hashes SHA1."""
        with self.get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("DELETE FROM sha1sums")
            cursor.executemany("INSERT INTO sha1sums (hash, filepath) VALUES (?, ?)", hash_data)
            conn.commit()

    def update_last_sync(self, date_str):
        with self.get_connection() as conn:
            conn.execute("UPDATE config SET last_update = ? WHERE id = 1", (date_str,))
            conn.commit()

    def get_config(self):
        with self.get_connection() as conn:
            conn.execute("SELECT * FROM config WHERE id = 1")
            description = conn.execute("SELECT * FROM config LIMIT 1").description
            columns = [column[0] for column in description]
            row = conn.execute("SELECT * FROM config WHERE id = 1").fetchone()
            if row:
                return dict(zip(columns, row))
            return None

    def save_config(self, config_dict):
        with self.get_connection() as conn:
            keys = ", ".join(config_dict.keys())
            placeholders = ", ".join(["?"] * len(config_dict))
            values = list(config_dict.values())
            conn.execute(f"UPDATE config SET ({keys}) = ({placeholders}) WHERE id = 1", values)
            conn.commit()

    def is_database_empty(self):
        with self.get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT count(*) FROM allfiles")
            return cursor.fetchone()[0] == 0

    def get_all_files(self, category_id=None, search_pattern=None):
        with self.get_connection() as conn:
            cursor = conn.cursor()
            query = "SELECT filepath FROM allfiles WHERE 1=1"
            params = []

            if category_id:
                query += " AND category_id = ?"
                params.append(category_id)

            cursor.execute(query, params)
            return [row[0] for row in cursor.fetchall()]

    def get_categories(self, parent_id=None):
        """Busca subcategorias de um pai específico."""
        with self.get_connection() as conn:
            cursor = conn.cursor()
            if parent_id is None:
                cursor.execute("SELECT id, name FROM categories WHERE parent_id IS NULL")
            else:
                cursor.execute("SELECT id, name FROM categories WHERE parent_id = ?", (parent_id,))
            return cursor.fetchall()

    def get_sha1_for_file(self, filepath):
        with self.get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT hash FROM sha1sums WHERE filepath = ?", (filepath,))
            result = cursor.fetchone()
            return result[0] if result else None

    def add_sha1(self, filepath, sha1_hash):
        with self.get_connection() as conn:
            conn.execute("INSERT INTO sha1sums (hash, filepath) VALUES (?, ?)", (sha1_hash, filepath))
            conn.commit()

    def get_all_files(self, category_id=None):
        with self.get_connection() as conn:
            cursor = conn.cursor()
            if category_id:
                # Busca recursiva: arquivos da categoria e de todas as subcategorias
                query = """
                        WITH RECURSIVE subcats(id) AS (SELECT ? \
                                                       UNION ALL \
                                                       SELECT c.id \
                                                       FROM categories c \
                                                                JOIN subcats s ON c.parent_id = s.id)
                        SELECT filepath \
                        FROM allfiles \
                        WHERE category_id IN (SELECT id FROM subcats) \
                        """
                cursor.execute(query, (category_id,))
            else:
                cursor.execute("SELECT filepath FROM allfiles")

            return [row[0] for row in cursor.fetchall()]

    def has_files_in_category(self, category_id):
        """Verifica se existem arquivos especificamente nesta pasta (sem recursividade)."""
        with self.get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT 1 FROM allfiles WHERE category_id = ? LIMIT 1", (category_id,))
            return cursor.fetchone() is not None

    def save_file_config(self, file_path, machine, media_type, ext1, ext2, ext3, ext4):
        with self.get_connection() as conn:
            conn.execute('''
                INSERT OR REPLACE INTO file_configs (file_path, machine, media_type, ext1, ext2, ext3, ext4)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            ''', (file_path, machine, media_type, ext1, ext2, ext3, ext4))
            conn.commit()

    def get_file_config(self, file_path):
        with self.get_connection() as conn:
            cursor = conn.cursor()
            cursor.execute('SELECT machine, media_type, ext1, ext2, ext3, ext4 FROM file_configs WHERE file_path = ?',
                           (file_path,))
            return cursor.fetchone()
