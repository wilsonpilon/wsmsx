import os
import subprocess
import shutil
import tempfile
from pathlib import Path
from bs4 import BeautifulSoup

class CHMParser:
    def __init__(self, chm_path):
        self.chm_path = Path(chm_path)
        self.temp_dir = Path(tempfile.gettempdir()) / "msxwrite_chm" / self.chm_path.stem
        self._decompile()

    def _decompile(self):
        if not self.temp_dir.exists():
            self.temp_dir.mkdir(parents=True, exist_ok=True)
            # Use hh.exe to decompile CHM
            subprocess.run(['hh.exe', '-decompile', str(self.temp_dir), str(self.chm_path)], shell=True)

    def get_toc(self):
        # Look for .hhc or .hhk
        hhc_files = list(self.temp_dir.glob("*.hhc"))
        hhk_files = list(self.temp_dir.glob("*.hhk"))
        
        toc_file = hhc_files[0] if hhc_files else (hhk_files[0] if hhk_files else None)
        
        if not toc_file:
            return []

        with open(toc_file, 'r', encoding='latin-1', errors='replace') as f:
            soup = BeautifulSoup(f.read(), 'html.parser')
            
        return self._parse_ul(soup.find('ul'))

    def _parse_ul(self, ul):
        if not ul:
            return []
        
        items = []
        if not ul:
            return items

        # Find all LI and UL elements that are direct children (conceptually)
        # Note: Beautiful Soup find_all(recursive=False) might miss elements if the HTML is messy
        # Many CHM files have flat-ish structures where UL follows LI
        
        current_li = None
        for child in ul.find_all(['li', 'ul'], recursive=False):
            if child.name == 'li':
                obj = child.find('object', type="text/sitemap")
                if obj:
                    name = ""
                    local = ""
                    for param in obj.find_all('param'):
                        p_name = param.get('name', '').lower()
                        if p_name == 'name':
                            name = param.get('value', '')
                        elif p_name == 'local':
                            local = param.get('value', '')
                    
                    # If name is empty, try 'Keyword' as fallback (common in HHK)
                    if not name:
                        for param in obj.find_all('param'):
                            if param.get('name', '').lower() == 'keyword':
                                name = param.get('value', '')

                    item = {
                        'name': name,
                        'local': str(self.temp_dir / local.replace('\\', '/')) if local else "",
                        'children': []
                    }
                    items.append(item)
                    current_li = item
                
                # Check if there's a nested UL inside this LI
                nested_ul = child.find('ul')
                if nested_ul and current_li:
                    current_li['children'] = self._parse_ul(nested_ul)

            elif child.name == 'ul':
                # This UL follows a LI (common in some CHM generators)
                if current_li:
                    current_li['children'] = self._parse_ul(child)
        
        return items

    def cleanup(self):
        # Optional: cleanup temp files
        # shutil.rmtree(self.temp_dir, ignore_errors=True)
        pass
