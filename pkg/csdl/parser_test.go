package csdl_test

import (
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockReadCloser struct {
	*strings.Reader

	data       string
	errOnClose error
}

func (m *mockReadCloser) Close() error {
	return m.errOnClose
}

func newMockReadCloser(s string) *mockReadCloser {
	return &mockReadCloser{strings.NewReader(s), s, nil}
}

func (m *mockReadCloser) Reset() {
	m.Reader.Reset(m.data)
}

const (
	computerSystemV1File      = "ComputerSystem_v1.xml"
	computerSystemV1TypeCount = 434

	externalRefXML = `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="External.v1_0_0">
			<ComplexType Name="ExternalType">
				<Property Name="ExternalProperty" Type="Edm.String"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`
	badCSDLXML = `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Bad.v1_0_0">
			<EntityType Name="BadEntity">
				<Property Name="BadProperty" Type="BadType"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`
)

func readComputerSystemV1(tb testing.TB) string {
	tb.Helper()
	data, err := os.ReadFile(computerSystemV1File)
	if err != nil {
		tb.Fatalf("reading %s: %v", computerSystemV1File, err)
	}
	return string(data)
}

func TestParserSimpleOperations(t *testing.T) {
	t.Parallel()
	t.Run("new", func(t *testing.T) {
		t.Parallel()
		parser := csdl.NewParser()
		require.NotNil(t, parser)
		assert.Empty(t, parser.Files)
	})

	t.Run("add file", func(t *testing.T) {
		t.Parallel()
		parser := csdl.NewParser()
		file := newMockReadCloser("<test>content</test>")

		parser.AddFile("test.xml", file)
		parser.AddFile("test2", file)
		assert.Len(t, parser.Files, 2)
	})

	t.Run("close", func(t *testing.T) {
		t.Parallel()
		parser := csdl.NewParser()
		file := newMockReadCloser("<test>content</test>")
		parser.AddFile("test", file)

		err := parser.Close()
		assert.NoError(t, err)
	})

	t.Run("close error", func(t *testing.T) {
		t.Parallel()
		parser := csdl.NewParser()
		file := newMockReadCloser("<test>content</test>")
		file.errOnClose = io.ErrClosedPipe
		parser.AddFile("test", file)

		err := parser.Close()
		assert.Error(t, err)
	})
}

type parserParseCase struct {
	name         string
	xmlData      string
	wantTypes    int
	wantError    bool
	wantTypeDefs map[string]parserExpectedType
	wantAbsent   []string
}

type parserExpectedType struct {
	name        string
	namespace   string
	complexType bool
	properties  map[string]string
	members     []string
}

func getParserParseCasesNoExternal() []parserParseCase {
	return []parserParseCase{
		{
			name: "simple entity type",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
				<Property Name="Name" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestEntity": {
					name:        "TestEntity",
					namespace:   "Test.v1_0_0",
					complexType: false,
					properties: map[string]string{
						"Id":   "Edm.String",
						"Name": "Edm.String",
					},
				},
			},
		},
		{
			name: "complex type",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="TestComplex">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestComplex": {
					name:        "TestComplex",
					namespace:   "Test.v1_0_0",
					complexType: true,
					properties: map[string]string{
						"Value": "Edm.String",
					},
				},
			},
		},
		{
			name: "enum type",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EnumType Name="TestEnum">
				<Member Name="Value1"/>
				<Member Name="Value2"/>
			</EnumType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestEnum": {
					name:      "TestEnum",
					namespace: "Test.v1_0_0",
					members:   []string{"Value1", "Value2"},
				},
			},
		},
		{
			name: "type definition",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<TypeDefinition Name="CustomString" UnderlyingType="Edm.String"/>
			<EntityType Name="TestEntity">
				<Property Name="Value" Type="Test.v1_0_0.CustomString"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestEntity": {
					name:        "TestEntity",
					namespace:   "Test.v1_0_0",
					complexType: false,
					properties: map[string]string{
						"Value": "Test.v1_0_0.CustomString",
					},
				},
			},
		},
		{
			name: "multiple types",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="Entity1">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
			<ComplexType Name="Complex1">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
			<EnumType Name="Enum1">
				<Member Name="Value1"/>
			</EnumType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 3,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.Entity1": {
					name:        "Entity1",
					namespace:   "Test.v1_0_0",
					complexType: false,
					properties: map[string]string{
						"Id": "Edm.String",
					},
				},
				"Test.v1_0_0.Complex1": {
					name:        "Complex1",
					namespace:   "Test.v1_0_0",
					complexType: true,
					properties: map[string]string{
						"Value": "Edm.String",
					},
				},
				"Test.v1_0_0.Enum1": {
					name:      "Enum1",
					namespace: "Test.v1_0_0",
					members:   []string{"Value1"},
				},
			},
		},
		{
			name: "reference ignored without lookup options",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:Reference Uri="https://example.com/schemas/ExternalTypes.xml"/>
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="LocalEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.LocalEntity": {
					name:        "LocalEntity",
					namespace:   "Test.v1_0_0",
					complexType: false,
					properties: map[string]string{
						"Id": "Edm.String",
					},
				},
			},
			wantAbsent: []string{"ExternalTypes.v1_0_0.ExternalEntity"},
		},
		{
			name:         "invalid XML",
			xmlData:      `<invalid>xml`,
			wantTypes:    0,
			wantError:    true,
			wantTypeDefs: map[string]parserExpectedType{},
		},
	}
}

func TestParserParseUsesCachedPrimaryFile(t *testing.T) {
	parser := csdl.NewParser()
	parser.AddFile("test", newMockReadCloser("this is not xml"))

	cache := &mockCache{data: map[string]*csdl.CacheEntry{
		"test": {
			Edmx: &csdl.Edmx{
				DataServices: csdl.DataServices{
					Schema: []*csdl.Schema{
						{Namespace: "CachedMain.v1_0_0"},
					},
				},
			},
			File: newMockReadCloser("cached file"),
		},
	}}

	types, err := parser.Parse(csdl.WithExternalParseCache(cache))
	assert.NoError(t, err)
	assert.Empty(t, types)
	assert.Contains(t, parser.Parsed, "test")
}

func TestParserParseReturnsErrorForCachedPrimaryWithBadReference(t *testing.T) {
	tempDir := t.TempDir()
	helperCreateXMLFiles(t, tempDir, map[string]string{
		"BadRef.xml": "not xml",
	})

	parser := csdl.NewParser()
	parser.AddFile("test", newMockReadCloser("this is not xml"))
	cache := &mockCache{data: map[string]*csdl.CacheEntry{
		"test": {
			Edmx: &csdl.Edmx{
				Reference: []*csdl.Reference{
					{Uri: "https://example.com/schemas/BadRef.xml"},
				},
				DataServices: csdl.DataServices{
					Schema: []*csdl.Schema{
						{Namespace: "CachedMain.v1_0_0"},
					},
				},
			},
			File: newMockReadCloser("cached file"),
		},
	}}

	_, err := parser.Parse(csdl.WithExternalParseCache(cache), csdl.WithLocalReferenceLookup(tempDir))
	assert.Error(t, err)
}

func getParserParseCasesExternal() []parserParseCase {
	return []parserParseCase{
		{
			name: "valid unused reference",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
    <edmx:Reference Uri="https://example.com/schemas/external1.xml"/>
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
				<Property Name="Name" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestEntity": {
					name:        "TestEntity",
					namespace:   "Test.v1_0_0",
					complexType: false,
					properties: map[string]string{
						"Id":   "Edm.String",
						"Name": "Edm.String",
					},
				},
			},
		},
		{
			name: "invalid unused reference",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:Reference Uri="https://example.com/schemas/external2.xml"/>
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="TestComplex">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantError: true,
		},
		{
			name: "used reference",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:Reference Uri="https://example.com/schemas/external1.xml">
		<edmx:Include Namespace="External.v1_0_0"/>
	</edmx:Reference>
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="TestComplex">
				<Property Name="Value" Type="External.v1_0_0.ExternalType"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestComplex": {
					name:        "TestComplex",
					namespace:   "Test.v1_0_0",
					complexType: true,
					properties: map[string]string{
						"Value": "External.v1_0_0.ExternalType",
					},
				},
			},
		},
		{
			name: "missing reference",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:Reference Uri="https://example.com/schemas/notfound.xml">
		<edmx:Include Namespace="External.v1_0_0"/>
	</edmx:Reference>
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="TestComplex">
				<Property Name="Value" Type="External.v1_0_0.ExternalType"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`,
			wantTypes: 1,
			wantError: false,
			wantTypeDefs: map[string]parserExpectedType{
				"Test.v1_0_0.TestComplex": {
					name:        "TestComplex",
					namespace:   "Test.v1_0_0",
					complexType: true,
					properties: map[string]string{
						"Value": "External.v1_0_0.ExternalType",
					},
				},
			},
		},
	}
}

func TestParserParseNoExternal(t *testing.T) {
	t.Parallel()
	for _, tt := range getParserParseCasesNoExternal() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := csdl.NewParser()
			file := newMockReadCloser(tt.xmlData)
			parser.AddFile("test", file)

			types, err := parser.Parse()
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, types, tt.wantTypes)
			assert.Len(t, tt.wantTypeDefs, tt.wantTypes)

			helperCompareTypesToExpected(t, types, tt.wantTypeDefs, tt.wantAbsent)
		})
	}
}

func TestParserParseComputerSystemV1(t *testing.T) {
	t.Parallel()
	parser := csdl.NewParser()
	parser.AddFile(computerSystemV1File, newMockReadCloser(readComputerSystemV1(t)))

	types, err := parser.Parse()
	require.NoError(t, err)
	require.Len(t, types, computerSystemV1TypeCount)

	computerSystem := types["ComputerSystem.v1_28_0.ComputerSystem"]
	require.NotNil(t, computerSystem)
	assert.Equal(t, "ComputerSystem.v1_27_1.ComputerSystem", computerSystem.BaseType)
	assert.Equal(t, "ComputerSystem.v1_28_0.Virtualization", computerSystem.Properties["Virtualization"].Type)

	assert.Contains(t, types, "ComputerSystem.v1_28_0.BootImageParameters")
	assert.Contains(t, types, "ComputerSystem.v1_28_0.ImageChecksumAlgorithm")
}

func TestParserParseLocalRefs(t *testing.T) {
	t.Parallel()
	for _, tt := range getParserParseCasesExternal() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := csdl.NewParser()
			file := newMockReadCloser(tt.xmlData)
			parser.AddFile("test", file)

			tempDir := t.TempDir()
			helperCreateXMLFiles(t, tempDir, map[string]string{
				"external1.xml": externalRefXML,
				"external2.xml": badCSDLXML,
			})

			types, err := parser.Parse(csdl.WithLocalReferenceLookup(tempDir))
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, types, tt.wantTypes)
			assert.Len(t, tt.wantTypeDefs, tt.wantTypes)

			helperCompareTypesToExpected(t, types, tt.wantTypeDefs, tt.wantAbsent)
		})
	}
}

func TestParserParseWithCache(t *testing.T) {
	t.Parallel()
	for _, tt := range getParserParseCasesExternal() {
		if tt.name == "invalid unused reference" {
			// Skip it since it would already have errored in a real use case.
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parser := csdl.NewParser()
			file := newMockReadCloser(tt.xmlData)
			parser.AddFile("test", file)

			cache := &mockCache{
				data: map[string]*csdl.CacheEntry{
					"external1.xml": {
						Edmx: &csdl.Edmx{
							DataServices: csdl.DataServices{
								Schema: []*csdl.Schema{
									{
										Namespace: "External.v1_0_0",
										ComplexType: []*csdl.ComplexType{
											{
												Name: "ExternalType",
												Property: []*csdl.Property{
													{
														Name: "ExternalProperty",
														Type: "Edm.String",
													},
												},
											},
										},
									},
								},
							},
						},
						File: newMockReadCloser(externalRefXML),
					},
					"external2.xml": {
						Edmx: nil,
						File: newMockReadCloser(badCSDLXML),
					},
				},
			}

			types, err := parser.Parse(csdl.WithExternalParseCache(cache))
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, types, tt.wantTypes)
			assert.Len(t, tt.wantTypeDefs, tt.wantTypes)

			helperCompareTypesToExpected(t, types, tt.wantTypeDefs, tt.wantAbsent)
		})
	}
}

func helperCompareTypesToExpected(t *testing.T, types map[string]*csdl.Type, expected map[string]parserExpectedType, wantAbsent []string) {
	t.Helper()
	for typeName, expected := range expected {
		parsedType, ok := types[typeName]
		require.True(t, ok, "expected parsed type %q not found", typeName)
		assert.Equal(t, expected.name, parsedType.Name)
		assert.Equal(t, expected.namespace, parsedType.Namespace)
		assert.Equal(t, expected.complexType, parsedType.ComplexType)

		if expected.properties != nil {
			assert.Len(t, parsedType.Properties, len(expected.properties))
			for propName, propType := range expected.properties {
				parsedProp, ok := parsedType.Properties[propName]
				require.True(t, ok, "expected property %q on type %q", propName, typeName)
				assert.Equal(t, propType, parsedProp.Type)
			}
		}
		if expected.members != nil {
			assert.Len(t, parsedType.Members, len(expected.members))
			for _, member := range expected.members {
				_, ok := parsedType.Members[member]
				require.True(t, ok, "expected member %q on type %q", member, typeName)
			}
		}
	}
	for _, absentTypeName := range wantAbsent {
		_, ok := types[absentTypeName]
		assert.False(t, ok, "unexpected referenced type %q parsed without lookup/cache options", absentTypeName)
	}
}

func helperCreateXMLFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for filename, content := range files {
		err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}
}

func TestParserFold(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="Base">
				<Property Name="BaseProp" Type="Edm.String"/>
			</EntityType>
			<EntityType Name="Derived" BaseType="Test.v1_0_0.Base">
				<Property Name="DerivedProp" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := csdl.NewParser()
	file := newMockReadCloser(xmlData)
	parser.AddFile("test", file)

	types, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	parser.Fold(types)

	derivedType := types["Test.v1_0_0.Derived"]
	if derivedType == nil {
		t.Fatal("Derived type not found after fold")
	}

	if len(derivedType.Properties) != 2 {
		t.Errorf("Derived type has %d properties, want 2", len(derivedType.Properties))
	}

	if _, ok := derivedType.Properties["BaseProp"]; !ok {
		t.Error("Derived type missing BaseProp after fold")
	}
	if _, ok := derivedType.Properties["DerivedProp"]; !ok {
		t.Error("Derived type missing DerivedProp after fold")
	}
}

func TestParserFoldDeletesTypeWhenBaseCannotBeResolved(t *testing.T) {
	parser := csdl.NewParser()
	types := map[string]*csdl.Type{
		"Test.v1_0_0.DeleteMe": {
			Name:      "DeleteMe",
			Namespace: "Test.v1_0_0",
			BaseType:  "Unknown.v1_0_0.Base",
		},
	}

	parser.Fold(types)
	assert.NotContains(t, types, "Test.v1_0_0.DeleteMe")
}

func TestParserFoldContinuesWhenKeyDeletedEarlier(t *testing.T) {
	parser := csdl.NewParser()
	types := map[string]*csdl.Type{
		// This key folds first (higher sort key) and deletes Test.v1_0_0.Victim
		"Z.v1_0_0.Alias": {
			Name:      "Victim",
			Namespace: "Test.v1_0_0",
			BaseType:  "Unknown.v1_0_0.Base",
		},
		"Test.v1_0_0.Victim": {
			Name:      "Victim",
			Namespace: "Test.v1_0_0",
			BaseType:  "",
		},
	}

	parser.Fold(types)
	assert.NotContains(t, types, "Z.v1_0_0.Alias")
	assert.NotContains(t, types, "Test.v1_0_0.Victim")
}

func TestDoReferenceLookupLocalReferenceMissing(t *testing.T) {
	// This error isn't propogated up...
	t.Parallel()
	parser := csdl.NewParser()
	csdl.WithLocalReferenceLookup(t.TempDir())(parser)
	edmx := &csdl.Edmx{
		Reference: []*csdl.Reference{
			{Uri: "https://example.com/schemas/NotFound.xml"},
		},
	}

	files, err := parser.DoReferenceLookup(edmx)
	assert.NoError(t, err)
	assert.Empty(t, files)
}

func TestDoReferenceLookupParsesReferenceOnlyOnce(t *testing.T) {
	tempDir := t.TempDir()
	refFileName := "OnlyOnce.xml"

	refXML := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="OnlyOnce.v1_0_0">
			<EntityType Name="RefEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`
	helperCreateXMLFiles(t, tempDir, map[string]string{refFileName: refXML})

	cache := &mockCache{data: make(map[string]*csdl.CacheEntry)}
	parser := csdl.NewParser()
	csdl.WithLocalReferenceLookup(tempDir)(parser)
	csdl.WithExternalParseCache(cache)(parser)

	uri := "https://example.com/schemas/" + refFileName
	edmx := &csdl.Edmx{
		Reference: []*csdl.Reference{
			{Uri: uri},
			{Uri: uri},
		},
	}

	files, err := parser.DoReferenceLookup(edmx)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.NotNil(t, parser.Parsed[strings.TrimSuffix(refFileName, ".xml")])
	assert.Equal(t, 1, cache.setCalls)

	// Re-running lookup with same references should not trigger another parse.
	_, err = parser.DoReferenceLookup(edmx)
	assert.NoError(t, err)
	assert.Equal(t, 1, cache.setCalls)
}

func TestDoReferenceLookupUsesExternalCacheWhenPresent(t *testing.T) {
	cache := &mockCache{data: make(map[string]*csdl.CacheEntry)}
	refName := "CachedRef"
	refKey := refName
	refFile := newMockReadCloser("cached file")
	refEdmx := &csdl.Edmx{
		DataServices: csdl.DataServices{
			Schema: []*csdl.Schema{
				{Namespace: "CachedRef.v1_0_0"},
			},
		},
	}
	cache.data[refKey] = &csdl.CacheEntry{Edmx: refEdmx, File: refFile}

	parser := csdl.NewParser()
	// DoReferenceLookup short-circuits without local lookup paths, so configure one.
	csdl.WithLocalReferenceLookup(t.TempDir())(parser)
	csdl.WithExternalParseCache(cache)(parser)

	edmx := &csdl.Edmx{
		Reference: []*csdl.Reference{
			{Uri: "https://example.com/schemas/" + refName + ".xml"},
		},
	}

	files, err := parser.DoReferenceLookup(edmx)
	assert.NoError(t, err)
	assert.Len(t, files, 1)

	file, ok := files[refKey]
	assert.True(t, ok, "DoReferenceLookup() missing key %q, got keys: %v", refKey, slices.Sorted(maps.Keys(files)))
	assert.Equal(t, refFile, file, "DoReferenceLookup() returned unexpected file handle from cache")

	parsed, ok := parser.Parsed[refKey]
	assert.True(t, ok, "parser.Parsed missing %q", refKey)
	assert.Equal(t, refEdmx, parsed.Edmx, "parser.Parsed[%q].Edmx is not the cached Edmx instance", refKey)
	assert.True(t, parsed.IsReference, "parser.Parsed[%q].IsReference = false, want true", refKey)
	assert.Equal(t, 0, cache.setCalls, "cache.Set called %d times, want 0 when cache already has reference", cache.setCalls)
}

type sortNamespaceCase struct {
	name     string
	a        string
	b        string
	wantSign int
}

func getSortNamespaceCases() []sortNamespaceCase {
	return []sortNamespaceCase{
		{
			name:     "empty namespace compared to namespaced value",
			a:        "",
			b:        "Test.v1_0_0",
			wantSign: -1,
		},
		{
			name:     "namespace with no dot compared to versioned namespace",
			a:        "Test",
			b:        "Test.v1_0_0",
			wantSign: -1,
		},
		{
			name:     "same namespace",
			a:        "Test.v1_0_0",
			b:        "Test.v1_0_0",
			wantSign: 0,
		},
		{
			name:     "different namespace root",
			a:        "Alpha.v1_0_0",
			b:        "Beta.v1_0_0",
			wantSign: -1,
		},
		{
			name:     "same root compares raw second component when no type component",
			a:        "Test.v10_0_0",
			b:        "Test.v2_0_0",
			wantSign: -1,
		},
		{
			name:     "same root with type component compares major version numerically",
			a:        "Test.v2_0_0.Entity",
			b:        "Test.v10_0_0.Entity",
			wantSign: -1,
		},
		{
			name:     "type component with versions lacking underscores",
			a:        "Test.v2.Entity",
			b:        "Test.v10.Entity",
			wantSign: -1,
		},
		{
			name:     "type component with one underscore and numeric suffix",
			a:        "Test.v1_2.Entity",
			b:        "Test.v1_10.Entity",
			wantSign: -1,
		},
		{
			name:     "type component with one underscore and non-numeric suffix",
			a:        "Test.v1_x.Entity",
			b:        "Test.v1_2.Entity",
			wantSign: -1,
		},
		{
			name:     "type component with underscore version and non-numeric first component",
			a:        "Test.vx_1_0.Entity",
			b:        "Test.v2_0_0.Entity",
			wantSign: -1,
		},
		{
			name:     "type component with non-numeric versions lacking underscores",
			a:        "Test.vabc.Entity",
			b:        "Test.vdef.Entity",
			wantSign: 0, // We can't compare these correctly, so leave the order in tact.
		},
		{
			name:     "same major compares minor",
			a:        "Test.v1_2_0.Entity",
			b:        "Test.v1_10_0.Entity",
			wantSign: -1,
		},
		{
			name:     "non-numeric minor treated as zero",
			a:        "Test.v1_x_0.Entity",
			b:        "Test.v1_2_0.Entity",
			wantSign: -1,
		},
		{
			name:     "non-numeric revision treated as zero",
			a:        "Test.v1_2_x.Entity",
			b:        "Test.v1_2_1.Entity",
			wantSign: -1,
		},
		{
			name:     "same major and minor compares revision",
			a:        "Test.v1_0_1.Entity",
			b:        "Test.v1_0_2.Entity",
			wantSign: -1,
		},
		{
			name:     "same version compares type name",
			a:        "Test.v1_0_0.Alpha",
			b:        "Test.v1_0_0.Beta",
			wantSign: -1,
		},
	}
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

func assertSortNamespaceCase(tb testing.TB, tc sortNamespaceCase) {
	tb.Helper()

	got := csdl.SortNamespace(tc.a, tc.b)
	gotSign := sign(got)
	if gotSign != tc.wantSign {
		tb.Fatalf("SortNamespace(%q, %q) sign = %d, want %d (raw value: %d)", tc.a, tc.b, gotSign, tc.wantSign, got)
	}

	reverse := csdl.SortNamespace(tc.b, tc.a)
	reverseSign := sign(reverse)
	if tc.wantSign == 0 {
		if reverseSign != 0 {
			tb.Fatalf("reverse SortNamespace(%q, %q) sign = %d, want 0", tc.b, tc.a, reverseSign)
		}
		return
	}
	if reverseSign != -tc.wantSign {
		tb.Fatalf("reverse SortNamespace(%q, %q) sign = %d, want %d", tc.b, tc.a, reverseSign, -tc.wantSign)
	}
}

func TestSortNamespace(t *testing.T) {
	for _, tc := range getSortNamespaceCases() {
		t.Run(tc.name, func(t *testing.T) {
			assertSortNamespaceCase(t, tc)
		})
	}
}

func TestWithSkipFileRead(t *testing.T) {
	parser := csdl.NewParser()

	// Add invalid XML to ensure Parse would fail if file reads are not skipped.
	parser.AddFile("test", newMockReadCloser("this is not xml"))

	// Pre-populate Parsed data so Parse() can build types without reading files.
	parser.Parsed["test"] = &csdl.ParsedData{
		Edmx: &csdl.Edmx{
			DataServices: csdl.DataServices{
				Schema: []*csdl.Schema{
					{
						Namespace: "Test.v1_0_0",
						EntityType: []*csdl.EntityType{
							{
								Name: "TestEntity",
								Property: []*csdl.Property{
									{Name: "Id", Type: "Edm.String"},
								},
							},
						},
					},
				},
			},
		},
		IsReference: false,
	}

	types, err := parser.Parse(csdl.WithSkipFileRead())
	assert.NoError(t, err)
	assert.NotEmpty(t, types, "Parse() with WithSkipFileRead returned no types")
	assert.Contains(t, types, "Test.v1_0_0.TestEntity", "expected type Test.v1_0_0.TestEntity not found; got keys: %v", slices.Sorted(maps.Keys(types)))
}

func TestWithSkipFileReadSkipsReferenceParsedEntries(t *testing.T) {
	parser := csdl.NewParser()
	parser.AddFile("test", newMockReadCloser("this is not xml"))

	parser.Parsed["nonref"] = &csdl.ParsedData{
		Edmx: &csdl.Edmx{
			DataServices: csdl.DataServices{Schema: []*csdl.Schema{{
				Namespace: "Test.v1_0_0",
				EntityType: []*csdl.EntityType{{
					Name:     "NonRefType",
					Property: []*csdl.Property{{Name: "Id", Type: "Edm.String"}},
				}},
			}}},
		},
		IsReference: false,
	}
	parser.Parsed["ref"] = &csdl.ParsedData{
		Edmx: &csdl.Edmx{
			DataServices: csdl.DataServices{Schema: []*csdl.Schema{{
				Namespace: "Ref.v1_0_0",
				EntityType: []*csdl.EntityType{{
					Name:     "RefType",
					Property: []*csdl.Property{{Name: "Id", Type: "Edm.String"}},
				}},
			}}},
		},
		IsReference: true,
	}

	types, err := parser.Parse(csdl.WithSkipFileRead())
	assert.NoError(t, err)
	assert.Contains(t, types, "Test.v1_0_0.NonRefType")
	assert.NotContains(t, types, "Ref.v1_0_0.RefType")
}

type mockCache struct {
	data     map[string]*csdl.CacheEntry
	setCalls int
}

func (m *mockCache) GetAll() map[string]*csdl.CacheEntry {
	return m.data
}

func (m *mockCache) Get(name string) (*csdl.Edmx, io.ReadCloser) {
	entry, ok := m.data[name]
	if !ok {
		return nil, nil
	}
	return entry.Edmx, entry.File
}

func (m *mockCache) Set(name string, edmx *csdl.Edmx, file io.ReadCloser) {
	m.setCalls++
	m.data[name] = &csdl.CacheEntry{
		Edmx: edmx,
		File: file,
	}
}

func TestNewParserFromCache(t *testing.T) {
	cache := &mockCache{data: make(map[string]*csdl.CacheEntry)}
	entry := &csdl.Edmx{}
	cache.Set("test", entry, nil)

	parser := csdl.NewParserFromCache(cache)
	require.NotNil(t, parser)
	assert.Contains(t, parser.Files, "test")
}

func BenchmarkParserParse(b *testing.B) {
	for _, tc := range getParserParseCasesNoExternal() {
		tc := tc
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			file := newMockReadCloser(tc.xmlData)
			for b.Loop() {
				parser := csdl.NewParser()
				parser.AddFile("test", file)
				types, err := parser.Parse()
				if tc.wantError {
					if err == nil {
						b.Fatal("Expected error, got nil")
					}
				} else {
					if err != nil {
						b.Fatalf("Parse() returned unexpected error: %v", err)
					}
					if len(types) != tc.wantTypes {
						b.Fatalf("Parse() returned %d types, want %d", len(types), tc.wantTypes)
					}
				}
				file.Reset()
			}
		})
	}
}

func BenchmarkParserParseComputerSystemV1(b *testing.B) {
	file := newMockReadCloser(readComputerSystemV1(b))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		parser := csdl.NewParser()
		parser.AddFile(computerSystemV1File, file)
		types, err := parser.Parse()
		if err != nil {
			b.Fatalf("Parse() returned unexpected error: %v", err)
		}
		if len(types) != computerSystemV1TypeCount {
			b.Fatalf("Parse() returned %d types, want %d", len(types), computerSystemV1TypeCount)
		}
		file.Reset()
	}
}

func BenchmarkSortNamespace(b *testing.B) {
	cases := getSortNamespaceCases()
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		for _, tc := range cases {
			assertSortNamespaceCase(b, tc)
		}
	}
}

func BenchmarkParserFold(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="Base">
				<Property Name="BaseProp1" Type="Edm.String"/>
				<Property Name="BaseProp2" Type="Edm.String"/>
			</EntityType>
			<EntityType Name="Derived" BaseType="Test.v1_0_0.Base">
				<Property Name="DerivedProp1" Type="Edm.String"/>
				<Property Name="DerivedProp2" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := csdl.NewParser()
	file := newMockReadCloser(xmlData)
	parser.AddFile("test", file)
	types, _ := parser.Parse()

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		typesCopy := make(map[string]*csdl.Type)
		for k, v := range types {
			typesCopy[k] = v
		}
		parser.Fold(typesCopy)
	}
}
