// Copyright 2020 Saferwall. All rights reserved.
// Use of this source code is governed by Apache v2 license
// license that can be found in the LICENSE file.

package pe

import (
	"io"
	"os"

	mmap "github.com/edsrzf/mmap-go"
)

// A File represents an open PE file.
type File struct {
	DosHeader    ImageDosHeader
	NtHeader     ImageNtHeader
	Sections     []ImageSectionHeader
	Imports      []Import
	Exports      []ExportFunction
	Debugs       []DebugEntry
	Relocations  []Relocation
	Resources    ResourceDirectory
	TLS          TLSDirectory
	LoadConfig   interface{}
	Exceptions   []Exception
	Certificates Certificate
	DelayImports []DelayImport
	BoundImports []BoundImportDescriptorData
	GlobalPtr    uint32
	RichHeader   RichHeader
	CLRHeader    ImageCOR20Header

	Header    []byte
	data      mmap.MMap
	closer    io.Closer
	Is64      bool
	Is32      bool
	Anomalies []string
	size      uint32
	f         *os.File
}

// Open opens the named file using os.Open and prepares it for use as a PE binary.
func Open(name string) (File, error) {

	// Init an File instance
	file := File{}

	f, err := os.Open(name)
	if err != nil {
		return file, err
	}

	// Memory map the file insead of using read/write.
	data, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		f.Close()
		return file, err
	}

	file.data = data
	file.size = uint32(len(file.data))
	file.f = f
	return file, nil
}

// Close closes the File.
func (pe *File) Close() error {
	var err error
	if pe.f != nil {
		err = f.Close()
	}
	return err
}

// Parse performs the file parsing for a PE binary.
func (pe *File) Parse() error {

	// check for the smallest PE size.
	if len(pe.data) < TinyPESize {
		return ErrInvalidPESize
	}

	// Parse the DOS header.
	err := pe.parseDosHeader()
	if err != nil {
		return err
	}

	// Parse the NT header.
	err = pe.parseNtHeader()
	if err != nil {
		return err
	}

	// Parse the Section Header.
	err = pe.parseSectionHeader()
	if err != nil {
		return err
	}

	// Parse the Data Directory entries.
	err = pe.parseDataDirectories()
	if err != nil {
		return err
	}

	return nil
}

func (pe *File) parseDataDirectories() (err error) {
	oh32 := ImageOptionalHeader32{}
	oh64 := ImageOptionalHeader64{}
	switch pe.Is64 {
	case true:
		oh64 = pe.NtHeader.OptionalHeader.(ImageOptionalHeader64)
	case false:
		oh32 = pe.NtHeader.OptionalHeader.(ImageOptionalHeader32)
	}

	// Maps data directory index to function which parses that directory.
	funcMaps := map[int](func(uint32, uint32) error){
		ImageDirectoryEntryExport:      pe.parseExportDirectory,
		ImageDirectoryEntryImport:      pe.parseImportDirectory,
		ImageDirectoryEntryResource:    pe.parseResourceDirectory,
		ImageDirectoryEntryException:   pe.parseExceptionDirectory,
		ImageDirectoryEntryCertificate: pe.parseSecurityDirectory,
		ImageDirectoryEntryBaseReloc:   pe.parseRelocDirectory,
		ImageDirectoryEntryDebug:       pe.parseDebugDirectory,
		ImageDirectoryEntryGlobalPtr:   pe.parseGlobalPtrDirectory,
		ImageDirectoryEntryTLS:         pe.parseTLSDirectory,
		ImageDirectoryEntryLoadConfig:  pe.parseLoadConfigDirectory,
		ImageDirectoryEntryBoundImport: pe.parseBoundImportDirectory,
		ImageDirectoryEntryIAT:         pe.parseIATDirectory,
		ImageDirectoryEntryDelayImport: pe.parseDelayImportDirectory,
		ImageDirectoryEntryCLR:         pe.parseCLRHeaderDirectory,
	}

	// Iterate over data directories and call the appropriate function.
	for entryIndex := 0; entryIndex < ImageNumberOfDirectoryEntries; entryIndex++ {

		var va, size uint32

		switch pe.Is64 {
		case true:
			dirEntry := oh64.DataDirectory[entryIndex]
			va = dirEntry.VirtualAddress
			size = dirEntry.Size
		case false:
			dirEntry := oh32.DataDirectory[entryIndex]
			va = dirEntry.VirtualAddress
			size = dirEntry.Size
		}

		if va != 0 {
			err := funcMaps[entryIndex](va, size)
			if err != nil {
				return err
			}
		}
	}

	return err
}
