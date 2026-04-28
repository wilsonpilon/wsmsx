from __future__ import annotations
import re

class MSXBasicAnalyzer:
    def __init__(self, content: str):
        self.content = content
        self.lines = content.splitlines()
        self.variables = {} # name: {type: str, lines: set, count: int}
        self.flow = []      # list of {from: int, to: int, type: 'GOTO'|'GOSUB'}
        self.subroutines = set() # target line numbers of GOSUB
        
    def analyze(self):
        # Regex for line numbers: ^\s*(\d+)
        # Regex for GOTO/GOSUB: \b(GOTO|GOSUB)\b\s*(\d+)
        # Regex for variables: 
        #   MSX BASIC variables: [A-Z][A-Z0-9]?[$%!#]?
        #   Must avoid keywords.
        
        from msx_basic_decoder import TOKEN_MAP, TOKEN_MAP_FF
        keywords = set(TOKEN_MAP) | set(TOKEN_MAP_FF)
        
        # Simple parser to skip strings and comments
        for line in self.lines:
            line_match = re.match(r"^\s*(\d+)", line)
            if not line_match:
                continue
            
            line_num = int(line_match.group(1))
            code_part = line[line_match.end():]
            
            # Remove strings
            code_part = re.sub(r'"[^"]*"', '""', code_part)
            
            # Remove comments (REM or ')
            code_part = re.split(r"REM|'", code_part, flags=re.IGNORECASE)[0]
            
            # Extract GOTO / GOSUB
            # Note: MSX BASIC allows GOTO 100,200 but that's ON GOTO. 
            # Simple GOTO/GOSUB:
            flow_matches = re.finditer(r"\b(GOTO|GOSUB)\b\s*(\d+)", code_part, re.IGNORECASE)
            for m in flow_matches:
                ftype = m.group(1).upper()
                target = int(m.group(2))
                self.flow.append({"from": line_num, "to": target, "type": ftype})
                if ftype == "GOSUB":
                    self.subroutines.add(target)

            # Extract variables
            # [A-Z][A-Z0-9]*[$%!#]?
            # We capture the full word first, then check if it's a variable.
            # MSX BASIC only uses the first two characters for variable names, 
            # but in modern dialects or even for readability we might have longer names.
            # However, the classic MSX BASIC tokens often start with these.
            
            # Using a more greedy match for characters and then filtering.
            words = re.finditer(r"\b([A-Z][A-Z0-9]*)([$%!#]?)(?!\()", code_part, re.IGNORECASE)
            for m in words:
                full_word = m.group(1).upper()
                suffix = m.group(2)
                
                # If the word itself is a keyword, skip it.
                if full_word in keywords:
                    continue
                
                # If it's a long word, MSX BASIC (classic) would only see the first 2 chars.
                # But if it's a keyword like PRINT, the regex [A-Z][A-Z0-9]* matches "PRINT".
                # If it was [A-Z][A-Z0-9]?, it would match "PR" and "INT" would be left over? 
                # No, \bPR would match PR and then INT would be another word.
                
                # Let's truncate to 2 chars for the variable name identification if we want to be classic,
                # but let's keep what the user wrote for the map.
                var_name = full_word
                if len(var_name) > 2:
                    # If the first 2 chars or the whole word is a keyword, it's likely not a variable
                    # unless it's a dialect that allows it.
                    # But the issue is specifically about "pedacos (2 letras) de palavras chave".
                    # Note: MSX BASIC keywords can be used if they have a suffix like % or $ in some cases,
                    # but usually they are reserved. If len > 2 and it starts with a keyword, skip.
                    is_kw_start = False
                    for kw in keywords:
                        if len(kw) >= 2 and var_name.startswith(kw):
                            is_kw_start = True
                            break
                    if is_kw_start:
                        continue
                
                full_name = var_name + suffix
                if full_name in keywords:
                    continue

                vsize = 4
                vtype = "SNG"
                if suffix == "$":
                    vtype = "STR"
                    vsize = 3
                elif suffix == "%":
                    vtype = "INT"
                    vsize = 2
                elif suffix == "!":
                    vtype = "SNG"
                    vsize = 4
                elif suffix == "#":
                    vtype = "DBL"
                    vsize = 8
                
                if full_name not in self.variables:
                    self.variables[full_name] = {
                        "type": vtype,
                        "lines": {line_num},
                        "count": 1,
                        "size": vsize
                    }
                else:
                    self.variables[full_name]["lines"].add(line_num)
                    self.variables[full_name]["count"] += 1

    def get_summary(self):
        sorted_vars = dict(sorted(self.variables.items()))
        total_memory = 0
        # Convert lines set to sorted list for display
        for v in sorted_vars.values():
            v["lines"] = sorted(list(v["lines"]))
            total_memory += v["size"]
        
        return {
            "variables": sorted_vars,
            "flow": self.flow,
            "subroutines": sorted(list(self.subroutines)),
            "total_memory_est": total_memory
        }
