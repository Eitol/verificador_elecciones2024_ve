package main

import (
	"fmt"
	"log"

	"github.com/Eitol/verificador_elecciones2024_ve/pkg/zipfiles"
)

func main() {
	inputDir := "assets/results"
	outputDir := "assets/results_zips"
	const maxZipSize int64 = 2_000_000_000 // 1 GB

	fmt.Println("creating zip files...")
	err := zipfiles.CreateZipFiles(inputDir, outputDir, maxZipSize)
	if err != nil {
		log.Panicf("error creating zip files: %v", err)
	}
	fmt.Println("creating zip files done")
}
