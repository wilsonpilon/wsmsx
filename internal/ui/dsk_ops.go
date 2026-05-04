package ui

// dsk_ops.go – MSX-DOS FAT12 DSK image operations.
//
// Ported from the C programs rddsk.c, wrdsk.c, DiskUtil.c and Boot.h
// by Arnold Metselaar (1996-1998) and Marat Fayzullin (1994-1995).
//
// Layout of a standard 720 KB MSX-DOS disk image (from the boot sector):
//   sector 0            : boot sector  (512 bytes)
//   sectors 1..6        : FAT  (2 copies × 3 sectors = 6 sectors)
//   sectors 7..13       : root directory (112 entries × 32 bytes = 3584 bytes = 7 sectors)
//   sectors 14..1439    : data area (clusters 2..714)

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	dskSectorSize = 512
	dskEOF12      = 0xFFF // FAT12 EOF marker
)

// dskBootBlock is the 512-byte MSX-DOS FAT12 boot sector used when creating
// a fresh disk image.  Sourced from fMSX Boot.h by Marat Fayzullin.
var dskBootBlock = [dskSectorSize]byte{
	0xEB, 0xFE, 0x90, 0x56, 0x46, 0x42, 0x2D, 0x31, 0x39, 0x38, 0x39, 0x00, 0x02, 0x02, 0x01, 0x00,
	0x02, 0x70, 0x00, 0xA0, 0x05, 0xF9, 0x03, 0x00, 0x09, 0x00, 0x02, 0x00, 0x00, 0x00, 0xD0, 0xED,
	0x53, 0x58, 0xC0, 0x32, 0xC2, 0xC0, 0x36, 0x55, 0x23, 0x36, 0xC0, 0x31, 0x1F, 0xF5, 0x11, 0x9D,
	0xC0, 0x0E, 0x0F, 0xCD, 0x7D, 0xF3, 0x3C, 0x28, 0x28, 0x11, 0x00, 0x01, 0x0E, 0x1A, 0xCD, 0x7D,
	0xF3, 0x21, 0x01, 0x00, 0x22, 0xAB, 0xC0, 0x21, 0x00, 0x3F, 0x11, 0x9D, 0xC0, 0x0E, 0x27, 0xCD,
	0x7D, 0xF3, 0xC3, 0x00, 0x01, 0x57, 0xC0, 0xCD, 0x00, 0x00, 0x79, 0xE6, 0xFE, 0xFE, 0x02, 0x20,
	0x07, 0x3A, 0xC2, 0xC0, 0xA7, 0xCA, 0x22, 0x40, 0x11, 0x77, 0xC0, 0x0E, 0x09, 0xCD, 0x7D, 0xF3,
	0x0E, 0x07, 0xCD, 0x7D, 0xF3, 0x18, 0xB4, 0x42, 0x6F, 0x6F, 0x74, 0x20, 0x65, 0x72, 0x72, 0x6F,
	0x72, 0x0D, 0x0A, 0x50, 0x72, 0x65, 0x73, 0x73, 0x20, 0x61, 0x6E, 0x79, 0x20, 0x6B, 0x65, 0x79,
	0x20, 0x66, 0x6F, 0x72, 0x20, 0x72, 0x65, 0x74, 0x72, 0x79, 0x0D, 0x0A, 0x24, 0x00, 0x4D, 0x53,
	0x58, 0x44, 0x4F, 0x53, 0x20, 0x20, 0x53, 0x59, 0x53, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF3, 0x2A,
	0x51, 0xF3, 0x11, 0x00, 0x01, 0x19, 0x01, 0x00, 0x01, 0x11, 0x00, 0xC1, 0xED, 0xB0, 0x3A, 0xEE,
	0xC0, 0x47, 0x11, 0xEF, 0xC0, 0x21, 0x00, 0x00, 0xCD, 0x51, 0x52, 0xF3, 0x76, 0xC9, 0x18, 0x64,
	0x3A, 0xAF, 0x80, 0xF9, 0xCA, 0x6D, 0x48, 0xD3, 0xA5, 0x0C, 0x8C, 0x2F, 0x9C, 0xCB, 0xE9, 0x89,
	0xD2, 0x00, 0x32, 0x26, 0x40, 0x94, 0x61, 0x19, 0x20, 0xE6, 0x80, 0x6D, 0x8A, 0x00, 0x00, 0x00,
	// remaining 240 bytes are zero
}

// ---------------------------------------------------------------------------
// FAT12 helpers  (equivalent of ReadFAT / WriteFAT in DiskUtil.c)
// ---------------------------------------------------------------------------

// dskReadFAT12 reads a 12-bit FAT entry for cluster clnr.
func dskReadFAT12(clnr int, fat []byte) int {
	offset := (clnr * 3) / 2
	if offset+1 >= len(fat) {
		return dskEOF12
	}
	val := int(fat[offset]) | (int(fat[offset+1]) << 8)
	if clnr&1 != 0 {
		return (val >> 4) & 0x0FFF
	}
	return val & 0x0FFF
}

// dskWriteFAT12 writes a 12-bit FAT entry for cluster clnr.
func dskWriteFAT12(clnr int, val int, fat []byte) {
	offset := (clnr * 3) / 2
	if offset+1 >= len(fat) {
		return
	}
	if clnr&1 != 0 {
		fat[offset] = (fat[offset] & 0x0F) | byte(val<<4)
		fat[offset+1] = byte(val >> 4)
	} else {
		fat[offset] = byte(val)
		fat[offset+1] = (fat[offset+1] & 0xF0) | byte((val>>8)&0x0F)
	}
}

// ---------------------------------------------------------------------------
// dskGeometry holds the layout parameters derived from the boot sector.
// ---------------------------------------------------------------------------
type dskGeometry struct {
	secPerCluster int
	numFATs       int
	secPerFAT     int
	rootEntries   int
	totalSectors  int
	mediaDesc     byte
	clusterLen    int64 // bytes per cluster
	dirOfs        int64 // byte offset of root directory
	dataOfs       int64 // byte offset of cluster 2
	maxCl         int   // highest valid cluster number
}

func dskParseGeometry(boot []byte) dskGeometry {
	secPerCluster := int(boot[0x0D])
	reservedSec := int(binary.LittleEndian.Uint16(boot[0x0E:0x10]))
	numFATs := int(boot[0x10])
	rootEntries := int(binary.LittleEndian.Uint16(boot[0x11:0x13]))
	totalSectors := int(binary.LittleEndian.Uint16(boot[0x13:0x15]))
	mediaDesc := boot[0x15]
	secPerFAT := int(binary.LittleEndian.Uint16(boot[0x16:0x18]))

	clusterLen := int64(dskSectorSize) * int64(secPerCluster)
	dirOfs := int64(dskSectorSize) * int64(reservedSec+numFATs*secPerFAT)
	dataOfs := dirOfs + int64(rootEntries)*32
	maxCl := (totalSectors-int(dataOfs/dskSectorSize))/secPerCluster + 1

	return dskGeometry{
		secPerCluster: secPerCluster,
		numFATs:       numFATs,
		secPerFAT:     secPerFAT,
		rootEntries:   rootEntries,
		totalSectors:  totalSectors,
		mediaDesc:     mediaDesc,
		clusterLen:    clusterLen,
		dirOfs:        dirOfs,
		dataOfs:       dataOfs,
		maxCl:         maxCl,
	}
}

// ---------------------------------------------------------------------------
// dskExtractFiles extracts the named files from a DSK image into destDir.
// If names is empty all files in the root directory are extracted.
// Equivalent to ReadFile loop in rddsk.c.
// ---------------------------------------------------------------------------
func dskExtractFiles(dskPath, destDir string, names []string) error {
	f, err := os.Open(dskPath)
	if err != nil {
		return fmt.Errorf("abrir imagem: %w", err)
	}
	defer f.Close()

	boot := make([]byte, dskSectorSize)
	if _, err := f.ReadAt(boot, 0); err != nil {
		return fmt.Errorf("ler boot sector: %w", err)
	}

	g := dskParseGeometry(boot)

	// Build lookup set (uppercase, as stored in the DSK).
	nameSet := map[string]bool{}
	for _, n := range names {
		nameSet[strings.ToUpper(n)] = true
	}

	// Read FAT.
	fatLen := (g.maxCl*3+1)/2 + 1
	fat := make([]byte, fatLen)
	if _, err := f.ReadAt(fat, int64(dskSectorSize)); err != nil {
		return fmt.Errorf("ler FAT: %w", err)
	}

	// Read directory.
	dirData := make([]byte, g.rootEntries*32)
	if _, err := f.ReadAt(dirData, g.dirOfs); err != nil {
		return fmt.Errorf("ler diretorio: %w", err)
	}

	secBuf := make([]byte, dskSectorSize)

	for i := 0; i < g.rootEntries; i++ {
		de := dirData[i*32 : i*32+32]
		first := de[0]
		if first == 0x00 {
			break // end of directory
		}
		if first == 0xE5 {
			continue // deleted entry
		}
		attr := de[0x0B]
		if attr&0x08 != 0 || attr&0x10 != 0 {
			continue // volume label or subdirectory
		}

		rawName := strings.TrimRight(string(de[0:8]), " ")
		rawExt := strings.TrimRight(string(de[8:11]), " ")
		if rawName == "" {
			continue
		}
		name83 := rawName
		if rawExt != "" {
			name83 = rawName + "." + rawExt
		}

		// Filter by requested names.
		if len(nameSet) > 0 && !nameSet[strings.ToUpper(name83)] {
			continue
		}

		// Build output filename (lowercase, matching rddsk.c Tolower).
		outName := strings.ToLower(rawName)
		if rawExt != "" {
			outName = outName + "." + strings.ToLower(rawExt)
		}
		outPath := filepath.Join(destDir, outName)

		fileSize := int64(binary.LittleEndian.Uint32(de[0x1C:0x20]))
		firstCluster := int(binary.LittleEndian.Uint16(de[0x1A:0x1C]))

		if err := os.MkdirAll(destDir, 0o755); err != nil {
			return fmt.Errorf("criar diretorio destino: %w", err)
		}

		out, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("criar %s: %w", outPath, err)
		}

		remaining := fileSize
		curCl := firstCluster

		for remaining > 0 && curCl >= 2 && curCl <= g.maxCl {
			clOfs := g.dataOfs + g.clusterLen*int64(curCl-2)
			for sec := 0; sec < g.secPerCluster && remaining > 0; sec++ {
				secOfs := clOfs + int64(sec)*dskSectorSize
				if _, err := f.ReadAt(secBuf, secOfs); err != nil {
					out.Close()
					return fmt.Errorf("ler setor de %s: %w", name83, err)
				}
				toWrite := int64(dskSectorSize)
				if remaining < toWrite {
					toWrite = remaining
				}
				if _, err := out.Write(secBuf[:toWrite]); err != nil {
					out.Close()
					return fmt.Errorf("escrever %s: %w", outPath, err)
				}
				remaining -= toWrite
			}
			curCl = dskReadFAT12(curCl, fat)
		}

		out.Close()
		if remaining != 0 {
			return fmt.Errorf("imagem corrompida: %s truncado", name83)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// dskCreateImage creates a fresh 720 KB MSX-DOS FAT12 disk image at dskPath
// and stores the files listed in srcFiles.
// Equivalent to main() + WriteFile() + CloseDisk() in wrdsk.c.
// ---------------------------------------------------------------------------
func dskCreateImage(dskPath string, srcFiles []string) error {
	boot := dskBootBlock // value copy
	g := dskParseGeometry(boot[:])

	// Initialise FAT.
	fatBufLen := dskSectorSize * g.secPerFAT
	fat := make([]byte, fatBufLen)
	dskWriteFAT12(0, int(g.mediaDesc)+0xF00, fat) // media descriptor
	dskWriteFAT12(1, dskEOF12, fat)

	// Initialise directory.
	dir := make([]byte, g.rootEntries*32)

	// Create image file.
	f, err := os.Create(dskPath)
	if err != nil {
		return fmt.Errorf("criar imagem: %w", err)
	}
	defer f.Close()

	// Write boot sector.
	if _, err := f.Write(boot[:]); err != nil {
		return fmt.Errorf("escrever boot sector: %w", err)
	}

	// Expand file to full image size (fills with zeros).
	totalBytes := int64(g.totalSectors) * dskSectorSize
	if err := f.Truncate(totalBytes); err != nil {
		// Fallback: seek to last byte and write a zero.
		if _, err := f.Seek(totalBytes-1, 0); err != nil {
			return err
		}
		if _, err := f.Write([]byte{0}); err != nil {
			return err
		}
	}

	curCl := 2  // first free cluster
	dirIdx := 0 // next free directory slot
	secBuf := make([]byte, dskSectorSize)

	for _, srcPath := range srcFiles {
		if dirIdx >= g.rootEntries {
			return fmt.Errorf("diretorio cheio: %s nao gravado", srcPath)
		}

		info, err := os.Stat(srcPath)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}

		// Build 8.3 directory entry -------------------------------------------
		var de [32]byte

		baseName := filepath.Base(srcPath)
		// Strip leading dots.
		for len(baseName) > 0 && baseName[0] == '.' {
			baseName = baseName[1:]
		}

		dotIdx := strings.LastIndex(baseName, ".")
		var rawName, rawExt string
		if dotIdx >= 0 {
			rawName = baseName[:dotIdx]
			rawExt = baseName[dotIdx+1:]
		} else {
			rawName = baseName
		}
		// Truncate to 8/3 and uppercase.
		fname := []byte(strings.ToUpper(rawName))
		if len(fname) > 8 {
			fname = fname[:8]
		}
		ext := []byte(strings.ToUpper(rawExt))
		if len(ext) > 3 {
			ext = ext[:3]
		}

		for i := 0; i < 8; i++ {
			if i < len(fname) {
				de[i] = fname[i]
			} else {
				de[i] = ' '
			}
		}
		for i := 0; i < 3; i++ {
			if i < len(ext) {
				de[8+i] = ext[i]
			} else {
				de[8+i] = ' '
			}
		}

		// MSX-DOS time/date stamp (same encoding as wrdsk.c).
		mt := info.ModTime()
		tval := uint16((mt.Second()>>1)&0x1F) |
			uint16(mt.Minute()&0x3F)<<5 |
			uint16(mt.Hour()&0x1F)<<11
		binary.LittleEndian.PutUint16(de[0x16:0x18], tval)
		yr := mt.Year() - 1980
		if yr < 0 {
			yr = 0
		}
		dval := uint16(mt.Day()&0x1F) |
			uint16(mt.Month()&0x0F)<<5 |
			uint16(yr&0x7F)<<9
		binary.LittleEndian.PutUint16(de[0x18:0x1A], dval)

		// Find first free cluster (same as wrdsk.c).
		for curCl <= g.maxCl && dskReadFAT12(curCl, fat) != 0 {
			curCl++
		}
		if curCl > g.maxCl {
			return fmt.Errorf("imagem cheia: %s nao gravado", srcPath)
		}
		firstCl := curCl
		binary.LittleEndian.PutUint16(de[0x1A:0x1C], uint16(firstCl))

		// Write file data into clusters ----------------------------------------
		src, err := os.Open(srcPath)
		if err != nil {
			continue
		}

		var fileSize int64
		prevCl := 0

	outerLoop:
		for {
			if curCl > g.maxCl {
				src.Close()
				return fmt.Errorf("imagem cheia durante gravacao de %s", srcPath)
			}
			clOfs := g.dataOfs + g.clusterLen*int64(curCl-2)

			for sec := 0; sec < g.secPerCluster; sec++ {
				n, readErr := src.Read(secBuf[:dskSectorSize])
				if n > 0 {
					fileSize += int64(n)
					// Zero-pad remainder of sector.
					for i := n; i < dskSectorSize; i++ {
						secBuf[i] = 0
					}
					secOfs := clOfs + int64(sec)*dskSectorSize
					if _, werr := f.WriteAt(secBuf[:dskSectorSize], secOfs); werr != nil {
						src.Close()
						return fmt.Errorf("escrever setor: %w", werr)
					}
				}
				if readErr != nil || n == 0 {
					// EOF or error – cluster write done.
					if prevCl != 0 {
						dskWriteFAT12(prevCl, curCl, fat)
					}
					prevCl = curCl
					break outerLoop
				}
			}

			// Link previous cluster to current.
			if prevCl != 0 {
				dskWriteFAT12(prevCl, curCl, fat)
			}
			prevCl = curCl

			// Advance to next free cluster.
			curCl++
			for curCl <= g.maxCl && dskReadFAT12(curCl, fat) != 0 {
				curCl++
			}
		}
		src.Close()

		if fileSize > 0 {
			dskWriteFAT12(prevCl, dskEOF12, fat)
		}
		binary.LittleEndian.PutUint32(de[0x1C:0x20], uint32(fileSize))

		// Advance cluster pointer past used cluster(s).
		curCl++
		for curCl <= g.maxCl && dskReadFAT12(curCl, fat) != 0 {
			curCl++
		}

		// Store directory entry.
		copy(dir[dirIdx*32:], de[:])
		dirIdx++
	}

	// Write FAT copies (numFATs copies, each secPerFAT sectors).
	for i := 0; i < g.numFATs; i++ {
		fatOfs := int64(dskSectorSize) + int64(i)*int64(g.secPerFAT)*dskSectorSize
		if _, err := f.WriteAt(fat, fatOfs); err != nil {
			return fmt.Errorf("escrever FAT copia %d: %w", i, err)
		}
	}

	// Write root directory.
	if _, err := f.WriteAt(dir, g.dirOfs); err != nil {
		return fmt.Errorf("escrever diretorio: %w", err)
	}

	return nil
}
