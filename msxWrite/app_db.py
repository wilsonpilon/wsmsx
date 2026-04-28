from __future__ import annotations

import sqlite3
from pathlib import Path


class AppDatabase:
    def __init__(self, db_path: Path) -> None:
        self.db_path = db_path
        self._ensure_schema()

    def _connect(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn

    def _ensure_schema(self) -> None:
        with self._connect() as conn:
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS settings (
                    key TEXT PRIMARY KEY,
                    value TEXT NOT NULL
                )
                """
            )
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS recent_files (
                    path TEXT PRIMARY KEY,
                    last_opened INTEGER NOT NULL
                )
                """
            )
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS renum_map (
                    old_ln INTEGER PRIMARY KEY,
                    new_ln INTEGER NOT NULL
                )
                """
            )

    def get_setting(self, key: str, default: str | None = None) -> str | None:
        with self._connect() as conn:
            row = conn.execute(
                "SELECT value FROM settings WHERE key = ?",
                (key,),
            ).fetchone()
            if row:
                return row["value"]
        return default

    def set_setting(self, key: str, value: str) -> None:
        with self._connect() as conn:
            conn.execute(
                """
                INSERT INTO settings (key, value)
                VALUES (?, ?)
                ON CONFLICT(key) DO UPDATE SET value = excluded.value
                """,
                (key, value),
            )

    def touch_recent_file(self, path: str, timestamp: int) -> None:
        with self._connect() as conn:
            conn.execute(
                """
                INSERT INTO recent_files (path, last_opened)
                VALUES (?, ?)
                ON CONFLICT(path) DO UPDATE SET last_opened = excluded.last_opened
                """,
                (path, timestamp),
            )
