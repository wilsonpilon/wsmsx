"""Leitor de discos MSX-DOS FAT12"""
import struct


SECTOR_SIZE = 512


class MSXDiskReader:
    """Classe para ler imagens de disco MSX (FAT12)"""

    def __init__(self, disk_path: str) -> None:
        self.disk_path = disk_path
        self.boot_sector: bytes | None = None
        self.fat: bytearray | None = None
        self.dir_entries: list[bytes] = []
        self.params: dict[str, int] = {}

    def open_disk(self) -> None:
        """Lê os parâmetros do disco e a FAT"""
        with open(self.disk_path, "rb") as f:
            self.boot_sector = f.read(SECTOR_SIZE)

            # BPB (BIOS Parameter Block)
            self.params["sec_per_clus"] = self.boot_sector[0x0D]
            self.params["reserved_sec"] = struct.unpack("<H", self.boot_sector[0x0E:0x10])[0]
            self.params["num_fats"] = self.boot_sector[0x10]
            self.params["root_entries"] = struct.unpack("<H", self.boot_sector[0x11:0x13])[0]
            self.params["total_sectors"] = struct.unpack("<H", self.boot_sector[0x13:0x15])[0]
            self.params["sec_per_fat"] = struct.unpack("<H", self.boot_sector[0x16:0x18])[0]

            # Cálculos de Offset
            self.params["dir_ofs"] = SECTOR_SIZE * (
                self.params["reserved_sec"] + (self.params["num_fats"] * self.params["sec_per_fat"])
            )
            self.params["data_ofs"] = self.params["dir_ofs"] + (self.params["root_entries"] * 32)
            self.params["clus_len"] = SECTOR_SIZE * self.params["sec_per_clus"]

            # Carregar FAT
            f.seek(SECTOR_SIZE * self.params["reserved_sec"])
            fat_size = self.params["sec_per_fat"] * SECTOR_SIZE
            self.fat = bytearray(f.read(fat_size))

            # Carregar Diretório Raiz
            f.seek(self.params["dir_ofs"])
            self.dir_entries = []
            for _ in range(self.params["root_entries"]):
                entry_data = f.read(32)
                if entry_data[0] == 0:
                    break
                self.dir_entries.append(entry_data)

    def read_fat_entry(self, clnr: int) -> int:
        """Lê uma entrada de 12 bits da FAT"""
        if self.fat is None:
            return 0
        idx = (clnr * 3) // 2
        val = struct.unpack("<H", self.fat[idx : idx + 2])[0]
        if clnr & 1:
            return val >> 4
        else:
            return val & 0x0FFF

    def list_files(self) -> list[dict[str, str | int]]:
        """Retorna lista de arquivos do disco"""
        files = []
        for entry in self.dir_entries:
            if entry[0] == 0xE5 or entry[0] == 0x00:
                continue

            # Decodificar nome e extensão
            try:
                name = entry[0:8].decode("ascii").strip()
                ext = entry[8:11].decode("ascii").strip()
            except UnicodeDecodeError:
                continue

            size = struct.unpack("<I", entry[28:32])[0]

            # Decodificar Data/Hora MSX
            raw_time = struct.unpack("<H", entry[22:24])[0]
            raw_date = struct.unpack("<H", entry[24:26])[0]

            hour = (raw_time >> 11) & 0x1F
            minute = (raw_time >> 5) & 0x3F
            second = (raw_time & 0x1F) * 2

            year = ((raw_date >> 9) & 0x7F) + 1980
            month = (raw_date >> 5) & 0x0F
            day = raw_date & 0x1F

            files.append(
                {
                    "filename": f"{name}.{ext}" if ext else name,
                    "size": size,
                    "date": f"{day:02d}/{month:02d}/{year}",
                    "time": f"{hour:02d}:{minute:02d}:{second:02d}",
                }
            )
        return files

    def extract_file(self, filename: str, dest_path: str) -> bool:
        """Extrai um arquivo do DSK para o PC"""
        target_entry = None
        for entry in self.dir_entries:
            try:
                name = entry[0:8].decode("ascii").strip()
                ext = entry[8:11].decode("ascii").strip()
            except UnicodeDecodeError:
                continue

            full_name = f"{name}.{ext}" if ext else name
            if full_name.upper() == filename.upper():
                target_entry = entry
                break

        if not target_entry:
            return False

        first_cluster = struct.unpack("<H", target_entry[26:28])[0]
        file_size = struct.unpack("<I", target_entry[28:32])[0]

        with open(self.disk_path, "rb") as f_in, open(dest_path, "wb") as f_out:
            cur_cl = first_cluster
            bytes_left = file_size

            while bytes_left > 0 and 0x002 <= cur_cl <= 0xFF6:
                offset = self.params["data_ofs"] + (cur_cl - 2) * self.params["clus_len"]
                f_in.seek(offset)

                to_read = min(bytes_left, self.params["clus_len"])
                f_out.write(f_in.read(to_read))

                bytes_left -= to_read
                cur_cl = self.read_fat_entry(cur_cl)

        return True
