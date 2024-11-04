package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipFiles takes a list of file paths and a destination path for the zip file.
func ZipFiles(files []string, dest string) (int64, error) {
	// Create the zip file
	zipFile, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer zipFile.Close()

	// Initialize the zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add each file to the zip archive
	for _, file := range files {
		// Open the file for reading
		f, err := os.Open(file)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		// Get file information
		fileInfo, err := f.Stat()
		if err != nil {
			return 0, err
		}

		// Create a zip header based on the file information
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return 0, err
		}

		// Set the header name to match the file path
		header.Name = filepath.Base(file)
		header.Method = zip.Deflate // Compression method

		// Create the writer for the file entry
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return 0, err
		}

		// Copy the file data to the zip writer
		_, err = io.Copy(writer, f)
		if err != nil {
			return 0, err
		}
	}

	zipFileInfo, _ := zipFile.Stat()
	zipFileSize := zipFileInfo.Size()

	return zipFileSize, nil
}
