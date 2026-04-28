import os

path = "E:/msxWrite/msx-encoding"
try:
    files = os.listdir(path)
    print(f"Files in {path}:")
    for f in files:
        print(f)
except Exception as e:
    print(f"Error: {e}")
