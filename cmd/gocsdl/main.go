package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pboyd04/gocsdl/pkg/csdl"
	"github.com/pboyd04/gocsdl/pkg/odata"
)

func main() {
	ignoreCollections := flag.Bool("ignore-collections", false, "ignore collection resources")
	individualFiles := flag.Bool("individual-files", true, "generate individual files")
	packageName := flag.String("package-name", "standard", "package name for the generated file(s)")
	flag.Parse()
	leftOverArgs := flag.Args()
	if len(leftOverArgs) == 0 {
		fmt.Printf("No file specified\n")
	}
	if !*ignoreCollections || !*individualFiles {
		fmt.Printf("Unimplemented\n")
	}
	parser := csdl.NewParser()
	parser.IgnoreCollections = *ignoreCollections
	for _, fileName := range leftOverArgs {
		processFile(fileName, parser)
	}
	types, err := parser.Parse()
	if err != nil {
		fmt.Printf("Error parsing CSDL: %s\n", err)
		os.Exit(1)
	}
	parser.Fold(types)
	if *individualFiles {
		// Generate individual files
		// Create the basic types...
		err = odata.GenBoilerPlate(*packageName)
		if err != nil {
			fmt.Printf("Error generating boilerplate: %s\n", err)
			os.Exit(1)
		}
		files := map[string]*csdl.File{}
		for name, t := range types {
			prefix := splitNamespacePrefix(name)
			file, ok := files[prefix]
			if !ok {
				file = csdl.NewFile(*packageName)
				files[prefix] = file
			}
			err = file.AddType(t)
			if err != nil {
				fmt.Printf("Error adding type %s to file %s: %s\n", name, prefix, err)
				os.Exit(1)
			}
		}
		for prefix, file := range files {
			fileName := prefix + ".go"
			data, err := file.Flush(types)
			if err != nil {
				fmt.Printf("Error generating file %s: %s\n", fileName, err)
				os.Exit(1)
			}
			osFile, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error creating file %s: %s\n", fileName, err)
				os.Exit(1)
			}
			//nolint:errcheck // Ignore error on close, not sure what we can do about it
			defer osFile.Close()
			_, err = osFile.Write(data)
			if err != nil {
				fmt.Printf("Error writing file %s: %s\n", fileName, err)
				os.Exit(1)
			}
		}
	}
}

func splitNamespacePrefix(name string) string {
	index := strings.Index(name, ".")
	if index == -1 {
		return name
	}
	return name[:index]
}

func processFile(fileName string, parser *csdl.Parser) {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".xml":
		// Process CSDL File...
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Printf("Error opening CSDL file: %s\n", err)
			os.Exit(1)
		}
		parser.AddFile(filepath.Base(fileName), file)
	case ".zip":
		// Process ZIP File...
		processZipFile(fileName, parser)
	default:
		fmt.Printf("Unknown file type: %s\n", fileName)
		os.Exit(1)
	}
}

func processZipFile(fileName string, parser *csdl.Parser) {
	r, err := zip.OpenReader(fileName)
	if err != nil {
		fmt.Printf("Error opening ZIP file: %s\n", err)
		os.Exit(1)
	}
	// Ideally we should close this, but we shouldn't be long running...
	// defer r.Close()
	for _, f := range r.File {
		// Skip the pdfs, html, and directories...
		if filepath.Ext(f.Name) != ".xml" {
			continue
		}
		file, err := f.Open()
		if err != nil {
			fmt.Printf("Error opening CSDL file in zip: %s\n", err)
			os.Exit(1)
		}
		parser.AddFile(filepath.Base(f.Name), file)
	}
}
