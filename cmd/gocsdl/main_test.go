package main

import (
	"archive/zip"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCSDL = `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EnumType Name="State">
				<Member Name="Ready" Value="0"/>
			</EnumType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

func TestSplitNamespacePrefix(t *testing.T) {
	tests := map[string]string{
		"Test.v1_0_0.Widget": "Test",
		"Widget":             "Widget",
		".Widget":            "",
	}

	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, want, splitNamespacePrefix(input))
		})
	}
}

func TestProcessFileAddsXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model.xml")
	require.NoError(t, os.WriteFile(path, []byte(testCSDL), 0o600))

	parser := csdl.NewParser()
	t.Cleanup(func() { require.NoError(t, parser.Close()) })
	processFile(path, parser)

	types, err := parser.Parse()
	require.NoError(t, err)
	assert.Contains(t, types, "Test.v1_0_0.State")
}

func TestProcessZipFileAddsOnlyXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.zip")
	zipFile, err := os.Create(path)
	require.NoError(t, err)
	writer := zip.NewWriter(zipFile)

	xmlFile, err := writer.Create("nested/model.xml")
	require.NoError(t, err)
	_, err = xmlFile.Write([]byte(testCSDL))
	require.NoError(t, err)
	ignoredFile, err := writer.Create("README.txt")
	require.NoError(t, err)
	_, err = ignoredFile.Write([]byte("not CSDL"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.NoError(t, zipFile.Close())

	parser := csdl.NewParser()
	t.Cleanup(func() { require.NoError(t, parser.Close()) })
	processFile(path, parser)

	assert.Len(t, parser.Files, 1)
	assert.Contains(t, parser.Files, "model")
	types, err := parser.Parse()
	require.NoError(t, err)
	assert.Contains(t, types, "Test.v1_0_0.State")
}

func TestMainGeneratesGoFiles(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "model.xml")
	require.NoError(t, os.WriteFile(inputPath, []byte(testCSDL), 0o600))

	oldArgs := os.Args
	oldCommandLine := flag.CommandLine
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldCommandLine
		require.NoError(t, os.Chdir(oldDir))
	})

	flag.CommandLine = flag.NewFlagSet("gocsdl", flag.ContinueOnError)
	os.Args = []string{"gocsdl", "-package-name", "generated", inputPath}
	require.NoError(t, os.Chdir(dir))

	main()

	boilerplate, err := os.ReadFile(filepath.Join(dir, "odata.go"))
	require.NoError(t, err)
	assert.Contains(t, string(boilerplate), "package generated")
	generated, err := os.ReadFile(filepath.Join(dir, "Test.go"))
	require.NoError(t, err)
	assert.Contains(t, string(generated), "type Test_State string")
}
