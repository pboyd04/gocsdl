package odata

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})
}

func TestGenBoilerPlateWritesExpectedContent(t *testing.T) {
	tempDir := t.TempDir()
	withWorkingDir(t, tempDir)

	err := GenBoilerPlate("odata")
	require.NoError(t, err)

	generatedPath := filepath.Join(tempDir, Filename)
	content, err := os.ReadFile(generatedPath)
	require.NoError(t, err)

	text := string(content)
	assert.Contains(t, text, "package odata")
	assert.Contains(t, text, "Action struct")
	assert.Contains(t, text, "OdataID struct")
	assert.Contains(t, text, "DateTimeOffset struct")
	assert.Contains(t, text, "Duration struct")
	assert.Contains(t, text, "UUID struct")
	assert.Contains(t, text, "func (d *DateTimeOffset) MarshalJSON()")
	assert.Contains(t, text, "func (d *DateTimeOffset) UnmarshalJSON")
	assert.Contains(t, text, "func (d *Duration) MarshalJSON()")
	assert.Contains(t, text, "func (d *Duration) UnmarshalJSON")
	assert.NotContains(t, text, strings.TrimSpace(UUIDMarshalJSONText))

	fileSet := token.NewFileSet()
	_, err = parser.ParseFile(fileSet, generatedPath, content, parser.AllErrors)
	assert.NoError(t, err)
}

func TestGenBoilerPlateReturnsErrorWhenOutputPathIsDirectory(t *testing.T) {
	tempDir := t.TempDir()
	withWorkingDir(t, tempDir)

	require.NoError(t, os.Mkdir(Filename, 0o755))
	err := GenBoilerPlate("odata")
	require.Error(t, err)
}

func TestGenBoilerPlateReturnsErrorForInvalidPackageName(t *testing.T) {
	tempDir := t.TempDir()
	withWorkingDir(t, tempDir)

	err := GenBoilerPlate("123invalid")
	require.Error(t, err)
}

func TestGenBoilerPlateReturnsErrorWhenWriteFails(t *testing.T) {
	if _, err := os.Stat("/dev/full"); err != nil {
		t.Skip("/dev/full not available")
	}

	tempDir := t.TempDir()
	withWorkingDir(t, tempDir)
	require.NoError(t, os.Symlink("/dev/full", Filename))

	err := GenBoilerPlate("odata")
	require.Error(t, err)
}
