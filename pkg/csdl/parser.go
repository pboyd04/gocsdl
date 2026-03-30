package csdl

import (
	"cmp"
	"encoding/xml"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Parser struct {
	IgnoreCollections bool
	Files             map[string]io.ReadCloser
	Replacements      map[string]string
	Parsed            map[string]*ParsedData

	localReferenceSearch []string
	externalCache        ExternalCache
	skipFileRead         bool
}

type ParsedData struct {
	Edmx        *Edmx
	IsReference bool
}

type CacheEntry struct {
	Edmx *Edmx
	File io.ReadCloser
}

type ExternalCache interface {
	GetAll() map[string]*CacheEntry
	Get(name string) (*Edmx, io.ReadCloser)
	Set(name string, edmx *Edmx, file io.ReadCloser)
}

// ParseOption is a function that configures a Parser
type ParseOption func(*Parser)

// WithLocalReferenceLookup adds a directory to search for referenced files
func WithLocalReferenceLookup(directory string) ParseOption {
	return func(p *Parser) {
		p.localReferenceSearch = append(p.localReferenceSearch, directory)
	}
}

func WithExternalParseCache(cache ExternalCache) ParseOption {
	return func(p *Parser) {
		p.externalCache = cache
	}
}

func WithSkipFileRead() ParseOption {
	return func(p *Parser) {
		p.skipFileRead = true
	}
}

func NewParser() *Parser {
	return &Parser{
		Files:        make(map[string]io.ReadCloser),
		Replacements: make(map[string]string),
		Parsed:       make(map[string]*ParsedData),
	}
}

func NewParserFromCache(cache ExternalCache) *Parser {
	ret := NewParser()
	entries := cache.GetAll()
	for name, entry := range entries {
		ret.Files[name] = entry.File
		ret.Parsed[name] = &ParsedData{
			Edmx:        entry.Edmx,
			IsReference: false,
		}
	}
	return ret
}

func (p *Parser) Close() error {
	var ret error
	for _, file := range p.Files {
		if err := file.Close(); err != nil {
			ret = err
		}
	}
	return ret
}

func (p *Parser) AddFile(name string, file io.ReadCloser) {
	name = strings.TrimSuffix(name, ".xml")
	p.Files[name] = file
}

func (p *Parser) Parse(opts ...ParseOption) (map[string]*Type, error) {
	for _, opt := range opts {
		opt(p)
	}
	types := map[string]*Type{}
	if p.skipFileRead {
		for _, parsed := range p.Parsed {
			if parsed.IsReference {
				continue
			}
			for _, schema := range parsed.Edmx.DataServices.Schema {
				p.makeTypesFromSchema(schema, types)
			}
		}
		return types, nil
	}
	// We keep a list to add at the end so we don't end up in a parsing loop.
	externalFiles := map[string]io.ReadCloser{}
	for name, reader := range p.Files {
		if p.externalCache != nil {
			edmx, file := p.externalCache.Get(name)
			if edmx != nil {
				p.Parsed[name] = &ParsedData{
					Edmx:        edmx,
					IsReference: false,
				}
				if _, err := p.DoReferenceLookup(edmx); err != nil {
					return nil, err
				}
				externalFiles[name] = file
				continue
			}
		}
		dec := xml.NewDecoder(reader)
		edmx := Edmx{}
		err := dec.Decode(&edmx)
		if err != nil {
			return nil, err
		}
		files, err := p.DoReferenceLookup(&edmx)
		if err != nil {
			return nil, err
		}
		for refName, refFile := range files {
			externalFiles[refName] = refFile
		}
		for _, schema := range edmx.DataServices.Schema {
			p.makeTypesFromSchema(schema, types)
		}
		p.Parsed[name] = &ParsedData{
			Edmx:        &edmx,
			IsReference: false,
		}
		if p.externalCache != nil {
			p.externalCache.Set(name, &edmx, reader)
		}
	}
	for externalFileName, externalFile := range externalFiles {
		p.Files[externalFileName] = externalFile
	}
	return types, nil
}

func (p *Parser) makeTypesFromSchema(schema *Schema, types map[string]*Type) {
	for i, entityType := range schema.EntityType {
		schema.EntityType[i].SetSchema(schema)
		types[schema.Namespace+"."+entityType.Name] = NewTypeFromEntityType(entityType, schema.Namespace)
	}
	for _, enumType := range schema.EnumType {
		types[schema.Namespace+"."+enumType.Name] = NewTypeFromEnumType(enumType, schema.Namespace)
	}
	for i, complexType := range schema.ComplexType {
		schema.ComplexType[i].SetSchema(schema)
		types[schema.Namespace+"."+complexType.Name] = NewTypeFromComplexType(complexType, schema.Namespace)
	}
	for _, typeDefinition := range schema.TypeDefinition {
		p.Replacements[schema.Namespace+"."+typeDefinition.Name] = typeDefinition.UnderlyingType
	}
}

// Fold will consolidate types, adding properties from base types to derived types
func (p *Parser) Fold(types map[string]*Type) {
	keys := maps.Keys(types)
	sortedKeys := slices.SortedFunc(keys, SortNamespace)
	slices.Reverse(sortedKeys)
	for _, name := range sortedKeys {
		t, ok := types[name]
		if !ok {
			// we removed this type in the fold
			continue
		}
		types[name] = t.Fold(types, p.Replacements)
		if types[name] == nil {
			delete(types, name)
		}
	}
}

// DoReferenceLookup will resolve references to other files if the parser is configured
// to do so.
func (p *Parser) DoReferenceLookup(edmx *Edmx) (map[string]io.ReadCloser, error) {
	// This eventually should allow for network lookup as well.
	if len(p.localReferenceSearch) == 0 {
		return nil, nil
	}
	files := make(map[string]io.ReadCloser)
	for _, reference := range edmx.Reference {
		fileName := filepath.Base(reference.Uri)
		file, err := p.addReferenceFile(fileName)
		if err != nil {
			return nil, err
		}
		if file != nil {
			files[strings.TrimSuffix(fileName, ".xml")] = file
		}
	}
	return files, nil
}

func (p *Parser) addReferenceFile(fileName string) (io.ReadCloser, error) {
	name := strings.TrimSuffix(fileName, ".xml")
	_, ok := p.Parsed[name]
	if ok {
		// Already parsed. Skip it.
		return nil, nil
	}
	if p.externalCache != nil {
		if parsed, file := p.externalCache.Get(name); parsed != nil {
			p.Parsed[name] = &ParsedData{
				Edmx:        parsed,
				IsReference: true,
			}
			return file, nil
		}
	}
	for _, dir := range p.localReferenceSearch {
		filePath := filepath.Join(dir, fileName)
		file, err := os.Open(filePath)
		if err == nil {
			dec := xml.NewDecoder(file)
			edmx := Edmx{}
			err = dec.Decode(&edmx)
			if err != nil {
				return nil, err
			}
			for _, schema := range edmx.DataServices.Schema {
				for i := range schema.EntityType {
					schema.EntityType[i].SetSchema(schema)
				}
				for i := range schema.ComplexType {
					schema.ComplexType[i].SetSchema(schema)
				}
			}
			p.Parsed[name] = &ParsedData{
				Edmx:        &edmx,
				IsReference: true,
			}
			if p.externalCache != nil {
				p.externalCache.Set(name, &edmx, file)
			}
			return file, nil
		}
	}
	return nil, nil
}

// need to do this because simple string sort doesn't work...
func SortNamespace(a, b string) int {
	a1, a2, a3 := splitNamespace(a)
	b1, b2, b3 := splitNamespace(b)
	if a1 != b1 {
		return strings.Compare(a1, b1)
	}
	if a3 == "" {
		return strings.Compare(a2, b2)
	}
	// The second component is where we have issues...
	aMajor, aMinor, aRev := splitVersion(a2)
	bMajor, bMinor, bRev := splitVersion(b2)
	if aMajor != bMajor {
		return cmp.Compare(aMajor, bMajor)
	}
	if aMinor != bMinor {
		return cmp.Compare(aMinor, bMinor)
	}
	if aRev != bRev {
		return cmp.Compare(aRev, bRev)
	}
	return strings.Compare(a3, b3)
}

func splitNamespace(namespace string) (string, string, string) {
	index := strings.Index(namespace, ".")
	if index == -1 {
		return namespace, "", ""
	}
	comp1 := namespace[:index]
	comp2 := namespace[index+1:]
	index = strings.Index(comp2, ".")
	if index == -1 {
		return comp1, comp2, ""
	}
	return comp1, comp2[:index], comp2[index+1:]
}

func splitVersion(version string) (int, int, int) {
	version = strings.TrimPrefix(version, "v")
	index := strings.Index(version, "_")
	if index == -1 {
		intVer, err := strconv.Atoi(version)
		if err != nil {
			return 0, 0, 0
		}
		return intVer, 0, 0
	}
	comp1 := version[:index]
	major, err := strconv.Atoi(comp1)
	if err != nil {
		return 0, 0, 0
	}
	comp2 := version[index+1:]
	index = strings.Index(comp2, "_")
	if index == -1 {
		minor, err := strconv.Atoi(comp2)
		if err != nil {
			return major, 0, 0
		}
		return major, minor, 0
	}
	minor, err := strconv.Atoi(comp2[:index])
	if err != nil {
		return major, 0, 0
	}
	rev, err := strconv.Atoi(comp2[index+1:])
	if err != nil {
		return major, minor, 0
	}
	return major, minor, rev
}
