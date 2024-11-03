package fileutils

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// WriteToFile writes a string content to a file (creates the file if it does not exist)
func WriteToFile(filename string, content string) error {
	// Create the directory if it does not exist
	dir := path.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return err
}

// DeleteFile deletes a file / directory
func DeleteDirectory(path string) error {
	return os.RemoveAll(path)
}

// CreateDirectoryIfNotExists creates a directory if it does not exist
// Returns no error if the directory already exists
func CreateDirectoryIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func GetFileContent(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func ExtractTarToDestination(reader io.Reader, destination string) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar archive: %w", err)
		}

		// Compute the target path for the extracted file/directory
		target := filepath.Join(destination, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
		case tar.TypeReg:
			// Create the file and write its contents
			file, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return fmt.Errorf("error writing file: %w", err)
			}

			// Set file permissions
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("error setting file permissions: %w", err)
			}
		default:
			return fmt.Errorf("unsupported file type in tar archive: %d", header.Typeflag)
		}
	}

	return nil
}

// CompressFiles compresses a list of files into a single archive file at the specified path.
// The archive file will be created if it does not exist. If the archive file already exists,
// The new files will be appended to the archive file.
func CompressFiles(archive_path string, files []string) error {
	archive, err := os.Create(archive_path)
	if err != nil {
		return err
	}
	defer archive.Close()

	zip_writer := zip.NewWriter(archive)
	defer zip_writer.Close()

	for _, file := range files {
		if err := AddFileToZip(zip_writer, file); err != nil {
			return err
		}
	}

	return nil
}

// AddFileToZip adds a file to a zip.Writer.
func AddFileToZip(zip_writer *zip.Writer, file_path string) error {
	file, err := os.Open(file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(file_path)

	writer, err := zip_writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
