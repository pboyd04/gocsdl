package csdl

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
)

type File struct {
	w       *bytes.Buffer
	fileSet *token.FileSet
	types   map[string]*Type
}

func NewFile(packageName string) *File {
	ret := &File{
		w:       bytes.NewBuffer(nil),
		fileSet: token.NewFileSet(),
		types:   make(map[string]*Type),
	}
	fileToken := &ast.File{
		Name: ast.NewIdent(packageName),
	}
	err := format.Node(ret.w, ret.fileSet, fileToken)
	if err != nil {
		panic(err)
	}
	return ret
}

func (f *File) AddType(t *Type) error {
	if strings.HasPrefix(t.Namespace, "MessageRegistry") && t.Name == "Message" {
		// This is a special case for the Message type, which is replicated
		return nil
	}
	existing, ok := f.types[t.Name]
	if ok {
		if len(existing.Properties) > len(t.Properties) || len(existing.Members) > len(t.Members) {
			// We already have a more complete type, so just skip this one
			return nil
		}
	}
	f.types[t.Name] = t
	return nil
}

func (f *File) Flush(allTypes map[string]*Type) ([]byte, error) {
	for _, typeData := range f.types {
		typeTokens := typeData.Node(allTypes)
		for _, typeToken := range typeTokens {
			err := format.Node(f.w, f.fileSet, typeToken)
			if err != nil {
				return nil, err
			}
			_, err = f.w.Write([]byte("\n"))
			if err != nil {
				return nil, err
			}
		}
	}
	// We should be good, but run it through the formatter one more time to be sure...
	return format.Source(f.w.Bytes())
}
