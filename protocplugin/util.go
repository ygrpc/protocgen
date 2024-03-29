package protocplugin

import "path/filepath"

// ExtractFilename   filename.ext -> filename
func ExtractFilename(filenameWithExt string) string {
	var extension = filepath.Ext(filenameWithExt)
	return filenameWithExt[0 : len(filenameWithExt)-len(extension)]
}
