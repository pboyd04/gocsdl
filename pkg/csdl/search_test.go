package csdl_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
)

func createTestParser(xmlData string) *csdl.Parser {
	parser := csdl.NewParser()
	file := &mockReadCloser{strings.NewReader(xmlData), xmlData, nil}
	parser.AddFile("test", file)
	_, _ = parser.Parse()
	return parser
}

func TestNewSearcher(t *testing.T) {
	parser := csdl.NewParser()
	searcher := csdl.NewSearcher(parser)
	if searcher == nil {
		t.Fatal("NewSearcher returned nil")
	}
}

func TestSearcherGetParser(t *testing.T) {
	parser := csdl.NewParser()
	searcher := csdl.NewSearcher(parser)
	result := searcher.GetParser()
	if result != parser {
		t.Error("GetParser did not return the same parser")
	}
}

func TestSearcherGetFiles(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	files := searcher.GetFiles()
	if len(files) == 0 {
		t.Error("GetFiles returned empty map")
	}
}

func TestSearcherGetSchemas(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test2.v1_0_0">
			<EntityType Name="TestEntity2">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	schemas := searcher.GetSchemas("")
	if len(schemas) != 2 {
		t.Errorf("GetSchemas(\"\") returned %d schemas, want 2", len(schemas))
	}

	schemas = searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) != 1 {
		t.Errorf("GetSchemas(\"Test.v1_0_0\") returned %d schemas, want 1", len(schemas))
	}
}

func TestSearcherGetEntityTypes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="Entity1">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
			<EntityType Name="Entity2">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	entities := searcher.GetEntityTypes("")
	if len(entities) != 2 {
		t.Errorf("GetEntityTypes(\"\") returned %d entities, want 2", len(entities))
	}

	entities = searcher.GetEntityTypes("Entity1")
	if len(entities) != 1 {
		t.Errorf("GetEntityTypes(\"Entity1\") returned %d entities, want 1", len(entities))
	}
	if len(entities) > 0 && entities[0].Name != "Entity1" {
		t.Errorf("GetEntityTypes(\"Entity1\") returned wrong entity: %s", entities[0].Name)
	}

	entities = searcher.GetEntityTypes("*")
	if len(entities) != 2 {
		t.Errorf("GetEntityTypes(\"*\") returned %d entities, want 2", len(entities))
	}
}

func TestSearcherGetComplexTypes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="Complex1">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
			<ComplexType Name="Complex2">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	complexTypes := searcher.GetComplexTypes("")
	if len(complexTypes) != 2 {
		t.Errorf("GetComplexTypes(\"\") returned %d types, want 2", len(complexTypes))
	}

	complexTypes = searcher.GetComplexTypes("Complex1")
	if len(complexTypes) != 1 {
		t.Errorf("GetComplexTypes(\"Complex1\") returned %d types, want 1", len(complexTypes))
	}
	if len(complexTypes) > 0 && complexTypes[0].Name != "Complex1" {
		t.Errorf("GetComplexTypes(\"Complex1\") returned wrong type: %s", complexTypes[0].Name)
	}
}

func TestSearcherGetEnumTypes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EnumType Name="Enum1">
				<Member Name="Value1"/>
			</EnumType>
			<EnumType Name="Enum2">
				<Member Name="Value2"/>
			</EnumType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	enumTypes := searcher.GetEnumTypes("")
	if len(enumTypes) != 2 {
		t.Errorf("GetEnumTypes(\"\") returned %d types, want 2", len(enumTypes))
	}

	enumTypes = searcher.GetEnumTypes("Enum1")
	if len(enumTypes) != 1 {
		t.Errorf("GetEnumTypes(\"Enum1\") returned %d types, want 1", len(enumTypes))
	}
	if len(enumTypes) > 0 && enumTypes[0].Name != "Enum1" {
		t.Errorf("GetEnumTypes(\"Enum1\") returned wrong type: %s", enumTypes[0].Name)
	}
}

func TestSearcherGetProperties(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	properties := searcher.GetProperties("")
	if len(properties) < 2 {
		t.Errorf("GetProperties(\"\") returned %d properties, want at least 2", len(properties))
	}

	properties = searcher.GetProperties("Prop1")
	if len(properties) < 1 {
		t.Errorf("GetProperties(\"Prop1\") returned %d properties, want at least 1", len(properties))
	}
}

func TestSearcherGetNavigationProperties(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
				<NavigationProperty Name="Nav2" Type="Test.v1_0_0.Other"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	navProps := searcher.GetNavigationProperties("")
	if len(navProps) < 2 {
		t.Errorf("GetNavigationProperties(\"\") returned %d properties, want at least 2", len(navProps))
	}

	navProps = searcher.GetNavigationProperties("Nav1")
	if len(navProps) < 1 {
		t.Errorf("GetNavigationProperties(\"Nav1\") returned %d properties, want at least 1", len(navProps))
	}
}

func TestSearcherGetActions(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<Action Name="Action1">
				<Parameter Name="Param1" Type="Edm.String"/>
			</Action>
			<Action Name="Action2">
				<Parameter Name="Param2" Type="Edm.String"/>
			</Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	actions := searcher.GetActions("")
	if len(actions) != 2 {
		t.Errorf("GetActions(\"\") returned %d actions, want 2", len(actions))
	}

	actions = searcher.GetActions("Action1")
	if len(actions) != 1 {
		t.Errorf("GetActions(\"Action1\") returned %d actions, want 1", len(actions))
	}
}

func TestSearcherGetAnnotations(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
					<Annotation Term="Test.Annotation2"/>
				</Property>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	annotations := searcher.GetAnnotations("")
	if len(annotations) < 2 {
		t.Errorf("GetAnnotations(\"\") returned %d annotations, want at least 2", len(annotations))
	}

	annotations = searcher.GetAnnotations("Test.Annotation1")
	if len(annotations) < 1 {
		t.Errorf("GetAnnotations(\"Test.Annotation1\") returned %d annotations, want at least 1", len(annotations))
	}
}

func TestSearcherGetSchemaByNamespace(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	schema := searcher.GetSchemaByNamespace("Test.v1_0_0")
	if schema == nil {
		t.Fatal("GetSchemaByNamespace returned nil")
	}
	if schema.Namespace != "Test.v1_0_0" {
		t.Errorf("Schema namespace = %s, want Test.v1_0_0", schema.Namespace)
	}

	schema = searcher.GetSchemaByNamespace("Test.v1_0_0")
	if schema == nil {
		t.Error("Second call to GetSchemaByNamespace returned nil (cache issue)")
	}

	schema = searcher.GetSchemaByNamespace("NonExistent")
	if schema != nil {
		t.Error("GetSchemaByNamespace returned non-nil for non-existent namespace")
	}
}

func TestSearcherGetByTypeName(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
			<ComplexType Name="TestComplex">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	result := searcher.GetByTypeName("Test.v1_0_0.TestEntity")
	if result == nil {
		t.Error("GetByTypeName returned nil for existing entity type")
	}

	result = searcher.GetByTypeName("Test.v1_0_0.TestComplex")
	if result == nil {
		t.Error("GetByTypeName returned nil for existing complex type")
	}

	result = searcher.GetByTypeName("Collection(Test.v1_0_0.TestEntity)")
	if result == nil {
		t.Error("GetByTypeName returned nil for collection type")
	}

	result = searcher.GetByTypeName("Edm.String")
	if result != nil {
		t.Error("GetByTypeName returned non-nil for Edm type")
	}

	result = searcher.GetByTypeName("NonExistent.Type")
	if result != nil {
		t.Error("GetByTypeName returned non-nil for non-existent type")
	}
}

func TestSearcherGetPath(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Id" Type="Edm.String"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	schemas := searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) == 0 {
		t.Fatal("No schemas found")
	}
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}
	props := searcher.GetPropertiesUnder(entities[0])
	if len(props) == 0 {
		t.Fatal("No properties found")
	}

	fileName, xpath := searcher.GetPath(schemas[0])
	if fileName != "test.xml" {
		t.Errorf("GetPath returned wrong file name: got %s, want test.xml", fileName)
	}
	if xpath != "/edmx:Edmx/edmx:DataServices/Schema[@Namespace='Test.v1_0_0']" {
		t.Errorf("GetPath returned wrong xpath: got %s, want /edmx:Edmx/edmx:DataServices/Schema[@Namespace='Test.v1_0_0']", xpath)
	}

	fileName, xpath = searcher.GetPath(entities[0])
	if fileName != "test.xml" {
		t.Errorf("GetPath returned wrong file name: got %s, want test.xml", fileName)
	}
	if xpath != "/edmx:Edmx/edmx:DataServices/Schema[@Namespace='Test.v1_0_0']/EntityType[@Name='TestEntity']" {
		t.Errorf("GetPath returned wrong xpath: got %s, want /edmx:Edmx/edmx:DataServices/Schema[@Namespace='Test.v1_0_0']/EntityType[@Name='TestEntity']", xpath)
	}

	fileName, xpath = searcher.GetPath(props[0])
	if fileName != "test.xml" {
		t.Errorf("GetPath returned wrong file name: got %s, want test.xml", fileName)
	}
	if xpath != "/edmx:Edmx/edmx:DataServices/Schema[@Namespace='Test.v1_0_0']/EntityType[@Name='TestEntity']/Property[@Name='Id']" {
		t.Errorf("GetPath returned wrong xpath: got %s, want /edmx:Edmx/edmx:DataServices/Schema[@Namespace='Test.v1_0_0']/EntityType[@Name='TestEntity']/Property[@Name='Id']", xpath)
	}
}

func TestSearcherGetPropertiesUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}

	props := searcher.GetPropertiesUnder(entities[0])
	if len(props) != 2 {
		t.Errorf("GetPropertiesUnder returned %d properties, want 2", len(props))
	}
}

func TestSearcherGetNavigationPropertiesUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
				<NavigationProperty Name="Nav2" Type="Test.v1_0_0.Other"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}

	navProps := searcher.GetNavigationPropertiesUnder(entities[0])
	if len(navProps) != 2 {
		t.Errorf("GetNavigationPropertiesUnder returned %d properties, want 2", len(navProps))
	}
}

func TestSearcherGetActionsUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<Action Name="Action1">
				<Parameter Name="Param1" Type="Edm.String"/>
			</Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	schemas := searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) == 0 {
		t.Fatal("No schemas found")
	}

	actions := searcher.GetActionsUnder(schemas[0])
	if len(actions) != 1 {
		t.Errorf("GetActionsUnder returned %d actions, want 1", len(actions))
	}
}

func TestSearcherGetAnnotaionsUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
				</Property>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}

	annotations := searcher.GetAnnotaionsUnder(entities[0])
	if len(annotations) < 1 {
		t.Errorf("GetAnnotaionsUnder returned %d annotations, want at least 1", len(annotations))
	}
}

func BenchmarkSearcherGetEntityTypes(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="Entity1"><Property Name="Id" Type="Edm.String"/></EntityType>
			<EntityType Name="Entity2"><Property Name="Id" Type="Edm.String"/></EntityType>
			<EntityType Name="Entity3"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetEntityTypes("")
	}
}

func BenchmarkSearcherGetSchemaByNamespace(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetSchemaByNamespace("Test.v1_0_0")
	}
}

func BenchmarkSearcherGetByTypeName(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetByTypeName("Test.v1_0_0.TestEntity")
	}
}

func BenchmarkSearcherGetFiles(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test2.v1_0_0">
			<EntityType Name="TestEntity2"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetFiles()
	}
}

func BenchmarkSearcherGetParser(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetParser()
	}
}

func BenchmarkSearcherGetAnnotations(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
					<Annotation Term="Test.Annotation2"/>
				</Property>
				<Property Name="Prop2" Type="Edm.String">
					<Annotation Term="Test.Annotation3"/>
				</Property>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetAnnotations("")
	}
}

func BenchmarkSearcherGetSchemas(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test2.v1_0_0">
			<EntityType Name="TestEntity2"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test3.v1_0_0">
			<EntityType Name="TestEntity3"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetSchemas("")
	}
}

func BenchmarkSearcherGetComplexTypes(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="Complex1"><Property Name="Value" Type="Edm.String"/></ComplexType>
			<ComplexType Name="Complex2"><Property Name="Value" Type="Edm.String"/></ComplexType>
			<ComplexType Name="Complex3"><Property Name="Value" Type="Edm.String"/></ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetComplexTypes("")
	}
}

func BenchmarkSearcherGetEnumTypes(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EnumType Name="Enum1"><Member Name="Value1"/></EnumType>
			<EnumType Name="Enum2"><Member Name="Value1"/></EnumType>
			<EnumType Name="Enum3"><Member Name="Value1"/></EnumType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetEnumTypes("")
	}
}

func BenchmarkSearcherGetProperties(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
				<Property Name="Prop3" Type="Edm.Boolean"/>
			</EntityType>
			<EntityType Name="TestEntity2">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetProperties("")
	}
}

func BenchmarkSearcherGetNavigationProperties(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
				<NavigationProperty Name="Nav2" Type="Test.v1_0_0.Other"/>
				<NavigationProperty Name="Nav3" Type="Test.v1_0_0.Another"/>
			</EntityType>
			<EntityType Name="TestEntity2">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetNavigationProperties("")
	}
}

func BenchmarkSearcherGetActions(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<Action Name="Action1"><Parameter Name="Param1" Type="Edm.String"/></Action>
			<Action Name="Action2"><Parameter Name="Param2" Type="Edm.String"/></Action>
			<Action Name="Action3"><Parameter Name="Param3" Type="Edm.String"/></Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetActions("")
	}
}

func BenchmarkSearcherGetPath(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test2.v1_0_0">
			<EntityType Name="TestEntity2"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	schemas := searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) == 0 {
		b.Fatal("No schemas found")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = searcher.GetPath(schemas[0])
	}
}

func BenchmarkSearcherGetPropertiesUnder(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
				<Property Name="Prop3" Type="Edm.Boolean"/>
				<Property Name="Prop4" Type="Edm.String"/>
				<Property Name="Prop5" Type="Edm.Int32"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		b.Fatal("No entities found")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetPropertiesUnder(entities[0])
	}
}

func BenchmarkSearcherGetNavigationPropertiesUnder(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
				<NavigationProperty Name="Nav2" Type="Test.v1_0_0.Other"/>
				<NavigationProperty Name="Nav3" Type="Test.v1_0_0.Another"/>
				<NavigationProperty Name="Nav4" Type="Test.v1_0_0.More"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		b.Fatal("No entities found")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetNavigationPropertiesUnder(entities[0])
	}
}

func BenchmarkSearcherGetActionsUnder(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<Action Name="Action1"><Parameter Name="Param1" Type="Edm.String"/></Action>
			<Action Name="Action2"><Parameter Name="Param2" Type="Edm.String"/></Action>
			<Action Name="Action3"><Parameter Name="Param3" Type="Edm.String"/></Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	schemas := searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) == 0 {
		b.Fatal("No schemas found")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetActionsUnder(schemas[0])
	}
}

func BenchmarkSearcherGetAnnotaionsUnder(b *testing.B) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
					<Annotation Term="Test.Annotation2"/>
				</Property>
				<Property Name="Prop2" Type="Edm.String">
					<Annotation Term="Test.Annotation3"/>
					<Annotation Term="Test.Annotation4"/>
				</Property>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		b.Fatal("No entities found")
	}

	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_ = searcher.GetAnnotaionsUnder(entities[0])
	}
}

// Concurrency tests for all public Get functions

func TestSearcherConcurrencyGetFiles(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test2.v1_0_0">
			<EntityType Name="TestEntity2"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				files := searcher.GetFiles()
				if len(files) == 0 {
					t.Error("GetFiles returned empty map")
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetAnnotations(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
					<Annotation Term="Test.Annotation2"/>
				</Property>
				<Property Name="Prop2" Type="Edm.String">
					<Annotation Term="Test.Annotation3"/>
				</Property>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				annotations := searcher.GetAnnotations("")
				if len(annotations) < 2 {
					t.Errorf("GetAnnotations returned %d annotations, want at least 2", len(annotations))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetSchemas(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test2.v1_0_0">
			<EntityType Name="TestEntity2"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				schemas := searcher.GetSchemas("")
				if len(schemas) != 2 {
					t.Errorf("GetSchemas returned %d schemas, want 2", len(schemas))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetEntityTypes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="Entity1"><Property Name="Id" Type="Edm.String"/></EntityType>
			<EntityType Name="Entity2"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				entities := searcher.GetEntityTypes("")
				if len(entities) != 2 {
					t.Errorf("GetEntityTypes returned %d entities, want 2", len(entities))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetComplexTypes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<ComplexType Name="Complex1"><Property Name="Value" Type="Edm.String"/></ComplexType>
			<ComplexType Name="Complex2"><Property Name="Value" Type="Edm.String"/></ComplexType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				types := searcher.GetComplexTypes("")
				if len(types) != 2 {
					t.Errorf("GetComplexTypes returned %d types, want 2", len(types))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetEnumTypes(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EnumType Name="Enum1"><Member Name="Value1"/></EnumType>
			<EnumType Name="Enum2"><Member Name="Value1"/></EnumType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				types := searcher.GetEnumTypes("")
				if len(types) != 2 {
					t.Errorf("GetEnumTypes returned %d types, want 2", len(types))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetProperties(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				props := searcher.GetProperties("")
				if len(props) < 2 {
					t.Errorf("GetProperties returned %d properties, want at least 2", len(props))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetNavigationProperties(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
				<NavigationProperty Name="Nav2" Type="Test.v1_0_0.Other"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				props := searcher.GetNavigationProperties("")
				if len(props) < 2 {
					t.Errorf("GetNavigationProperties returned %d properties, want at least 2", len(props))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetActions(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<Action Name="Action1"><Parameter Name="Param1" Type="Edm.String"/></Action>
			<Action Name="Action2"><Parameter Name="Param2" Type="Edm.String"/></Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				actions := searcher.GetActions("")
				if len(actions) != 2 {
					t.Errorf("GetActions returned %d actions, want 2", len(actions))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetSchemaByNamespace(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				schema := searcher.GetSchemaByNamespace("Test.v1_0_0")
				if schema == nil {
					t.Error("GetSchemaByNamespace returned nil")
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetByTypeName(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				result := searcher.GetByTypeName("Test.v1_0_0.TestEntity")
				if result == nil {
					t.Error("GetByTypeName returned nil for existing entity type")
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetPath(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity"><Property Name="Id" Type="Edm.String"/></EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	schemas := searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) == 0 {
		t.Fatal("No schemas found")
	}

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				fileName, xpath := searcher.GetPath(schemas[0])
				if fileName == "" {
					t.Error("GetPath returned empty file name")
				}
				if xpath == "" {
					t.Error("GetPath returned empty xpath")
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetPropertiesUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String"/>
				<Property Name="Prop2" Type="Edm.Int32"/>
				<Property Name="Prop3" Type="Edm.Boolean"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				props := searcher.GetPropertiesUnder(entities[0])
				if len(props) != 3 {
					t.Errorf("GetPropertiesUnder returned %d properties, want 3", len(props))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetNavigationPropertiesUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
				<NavigationProperty Name="Nav2" Type="Test.v1_0_0.Other"/>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				props := searcher.GetNavigationPropertiesUnder(entities[0])
				if len(props) != 2 {
					t.Errorf("GetNavigationPropertiesUnder returned %d properties, want 2", len(props))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetActionsUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<Action Name="Action1"><Parameter Name="Param1" Type="Edm.String"/></Action>
			<Action Name="Action2"><Parameter Name="Param2" Type="Edm.String"/></Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	schemas := searcher.GetSchemas("Test.v1_0_0")
	if len(schemas) == 0 {
		t.Fatal("No schemas found")
	}

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				actions := searcher.GetActionsUnder(schemas[0])
				if len(actions) != 2 {
					t.Errorf("GetActionsUnder returned %d actions, want 2", len(actions))
				}
			}
		}()
	}

	wg.Wait()
}

func TestSearcherConcurrencyGetAnnotaionsUnder(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
					<Annotation Term="Test.Annotation2"/>
				</Property>
				<Property Name="Prop2" Type="Edm.String">
					<Annotation Term="Test.Annotation3"/>
				</Property>
			</EntityType>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)
	entities := searcher.GetEntityTypes("TestEntity")
	if len(entities) == 0 {
		t.Fatal("No entities found")
	}

	const numGoroutines = 10
	const numIterations = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				annotations := searcher.GetAnnotaionsUnder(entities[0])
				if len(annotations) < 3 {
					t.Errorf("GetAnnotaionsUnder returned %d annotations, want at least 3", len(annotations))
				}
			}
		}()
	}

	wg.Wait()
}

// Mixed concurrency test - multiple goroutines calling different functions
func TestSearcherConcurrencyMixed(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" Version="4.0">
	<edmx:DataServices>
		<Schema xmlns="http://docs.oasis-open.org/odata/ns/edm" Namespace="Test.v1_0_0">
			<EntityType Name="TestEntity">
				<Property Name="Prop1" Type="Edm.String">
					<Annotation Term="Test.Annotation1"/>
				</Property>
				<NavigationProperty Name="Nav1" Type="Test.v1_0_0.Related"/>
			</EntityType>
			<ComplexType Name="TestComplex">
				<Property Name="Value" Type="Edm.String"/>
			</ComplexType>
			<EnumType Name="TestEnum"><Member Name="Value1"/></EnumType>
			<Action Name="TestAction"><Parameter Name="Param1" Type="Edm.String"/></Action>
		</Schema>
	</edmx:DataServices>
</edmx:Edmx>`

	parser := createTestParser(xmlData)
	searcher := csdl.NewSearcher(parser)

	const numGoroutines = 20
	const numIterations = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Different goroutines call different functions
	functions := []func(){
		func() {
			files := searcher.GetFiles()
			if len(files) == 0 {
				t.Error("GetFiles returned empty map")
			}
		},
		func() {
			p := searcher.GetParser()
			if p == nil {
				t.Error("GetParser returned nil")
			}
		},
		func() {
			schemas := searcher.GetSchemas("")
			if len(schemas) != 1 {
				t.Errorf("GetSchemas returned %d schemas, want 1", len(schemas))
			}
		},
		func() {
			entities := searcher.GetEntityTypes("")
			if len(entities) != 1 {
				t.Errorf("GetEntityTypes returned %d entities, want 1", len(entities))
			}
		},
		func() {
			types := searcher.GetComplexTypes("")
			if len(types) != 1 {
				t.Errorf("GetComplexTypes returned %d types, want 1", len(types))
			}
		},
		func() {
			types := searcher.GetEnumTypes("")
			if len(types) != 1 {
				t.Errorf("GetEnumTypes returned %d types, want 1", len(types))
			}
		},
		func() {
			props := searcher.GetProperties("")
			if len(props) < 1 {
				t.Errorf("GetProperties returned %d properties, want at least 1", len(props))
			}
		},
		func() {
			props := searcher.GetNavigationProperties("")
			if len(props) < 1 {
				t.Errorf("GetNavigationProperties returned %d properties, want at least 1", len(props))
			}
		},
		func() {
			actions := searcher.GetActions("")
			if len(actions) != 1 {
				t.Errorf("GetActions returned %d actions, want 1", len(actions))
			}
		},
		func() {
			annotations := searcher.GetAnnotations("")
			if len(annotations) < 1 {
				t.Errorf("GetAnnotations returned %d annotations, want at least 1", len(annotations))
			}
		},
		func() {
			schema := searcher.GetSchemaByNamespace("Test.v1_0_0")
			if schema == nil {
				t.Error("GetSchemaByNamespace returned nil")
			}
		},
		func() {
			result := searcher.GetByTypeName("Test.v1_0_0.TestEntity")
			if result == nil {
				t.Error("GetByTypeName returned nil for existing entity type")
			}
		},
	}

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				functions[index%len(functions)]()
			}
		}(i)
	}

	wg.Wait()
}
