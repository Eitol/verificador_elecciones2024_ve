package zipfiles

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Eitol/verificador_elecciones2024_ve/pkg/iocloser"
)

func CreateZipFiles(sourceDir, destDir string, maxZipSize int64) error {
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creando el directorio de destino: %v", err)
	}

	files, err := getValidFiles(sourceDir)
	if err != nil {
		return err
	}

	return processFiles(files, sourceDir, destDir, maxZipSize)
}

func getValidFiles(dir string) ([]os.DirEntry, error) {
	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error leyendo el directorio de origen: %v", err)
	}

	var vf []os.DirEntry
	for _, file := range allFiles {
		if !file.IsDir() {
			vf = append(vf, file)
		}
	}
	return vf, nil
}

func processFiles(files []os.DirEntry, sourceDir, destDir string, maxZipSize int64) error {
	var currentZipFile *os.File
	var zipWriter *zip.Writer
	var currentZipSize int64
	zipIndex := 1

	for _, file := range files {
		filePath := filepath.Join(sourceDir, file.Name())
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("error obteniendo información del archivo %s: %v", file.Name(), err)
		}

		if needNewZipFile(currentZipFile, currentZipSize, fileInfo.Size(), maxZipSize) {
			currentZipFile, zipWriter, err = createNewZipFile(destDir, zipIndex, currentZipFile, zipWriter)
			if err != nil {
				return err
			}
			currentZipSize = 0
			zipIndex++
		}

		if err := addFileToZip(zipWriter, filePath, file.Name()); err != nil {
			return fmt.Errorf("error añadiendo %s al ZIP: %v", file.Name(), err)
		}

		currentZipSize += fileInfo.Size()
	}

	return closeZipFile(zipWriter, currentZipFile)
}

func needNewZipFile(currentZipFile *os.File, currentZipSize, fileSize, maxZipSize int64) bool {
	return currentZipFile == nil || currentZipSize+fileSize > maxZipSize
}

func createNewZipFile(destDir string, zipIndex int, currentZipFile *os.File, zipWriter *zip.Writer) (*os.File, *zip.Writer, error) {
	if currentZipFile != nil {
		iocloser.CLose(zipWriter)
		iocloser.CLose(currentZipFile)
	}

	zipFileName := filepath.Join(destDir, fmt.Sprintf("archive_%d.zip", zipIndex))
	newZipFile, err := os.Create(zipFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("error creando el archivo ZIP %s: %v", zipFileName, err)
	}

	return newZipFile, zip.NewWriter(newZipFile), nil
}

func closeZipFile(zipWriter *zip.Writer, zipFile *os.File) error {
	if zipWriter != nil {
		iocloser.CLose(zipWriter)
	}
	if zipFile != nil {
		iocloser.CLose(zipFile)
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath, fileName string) error {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer iocloser.CLose(fileToZip)

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = fileName
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
