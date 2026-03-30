package csdl_test

import (
	"testing"

	"github.com/pboyd04/gocsdl/pkg/csdl"
)

func TestSchemaGetChildren(t *testing.T) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
		Action: []*csdl.Action{
			{Name: "Action1"},
		},
		ComplexType: []*csdl.ComplexType{
			{Name: "Complex1"},
		},
		EntityType: []*csdl.EntityType{
			{Name: "Entity1"},
		},
		EnumType: []*csdl.EnumType{
			{Name: "Enum1"},
		},
		Term: []*csdl.Term{
			{Name: "Term1"},
		},
		TypeDefinition: []*csdl.TypeDefinition{
			{Name: "TypeDef1"},
		},
	}

	children := schema.GetChildren()
	if len(children) == 0 {
		t.Error("Expected children, got none")
	}
}

func TestSchemaGetName(t *testing.T) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
	}
	if schema.GetName() != "Test.v1_0_0" {
		t.Errorf("GetName() = %v, want %v", schema.GetName(), "Test.v1_0_0")
	}
}

func TestSchemaGetXPath(t *testing.T) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
	}
	if schema.GetXPath() != "Test.v1_0_0" {
		t.Errorf("GetXPath() = %v, want %v", schema.GetXPath(), "Test.v1_0_0")
	}
}

func TestActionGetChildren(t *testing.T) {
	action := csdl.Action{
		Name: "TestAction",
		Annotation: []*csdl.Annotation{
			{Term: "Test.Term"},
		},
		Parameter: []*csdl.Parameter{
			{Name: "Param1"},
		},
	}

	children := action.GetChildren()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestActionGetName(t *testing.T) {
	action := csdl.Action{Name: "TestAction"}
	if action.GetName() != "TestAction" {
		t.Errorf("GetName() = %v, want %v", action.GetName(), "TestAction")
	}
}

func TestActionGetXPath(t *testing.T) {
	action := csdl.Action{Name: "TestAction"}
	if action.GetXPath() != "TestAction" {
		t.Errorf("GetXPath() = %v, want %v", action.GetXPath(), "TestAction")
	}
}

func TestAnnotationGetName(t *testing.T) {
	annotation := csdl.Annotation{Term: "Test.Term"}
	if annotation.GetName() != "Test.Term" {
		t.Errorf("GetName() = %v, want %v", annotation.GetName(), "Test.Term")
	}
}

func TestAnnotationGetXPath(t *testing.T) {
	annotation := csdl.Annotation{Term: "Test.Term"}
	expected := "Annotation[@Term='Test.Term']"
	if annotation.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", annotation.GetXPath(), expected)
	}
}

func TestAnnotationGetChildren(t *testing.T) {
	annotation := csdl.Annotation{Term: "Test.Term"}
	children := annotation.GetChildren()
	if children != nil {
		t.Error("Expected nil children for Annotation")
	}
}

func TestAnnotationsGetName(t *testing.T) {
	annotations := csdl.Annotations{Target: "Test.Target"}
	if annotations.GetName() != "Test.Target" {
		t.Errorf("GetName() = %v, want %v", annotations.GetName(), "Test.Target")
	}
}

func TestAnnotationsGetXPath(t *testing.T) {
	annotations := csdl.Annotations{Target: "Test.Target"}
	expected := "Annotations[@Target='Test.Target']"
	if annotations.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", annotations.GetXPath(), expected)
	}
}

func TestAnnotationsGetChildren(t *testing.T) {
	annotations := csdl.Annotations{
		Target: "Test.Target",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
			{Term: "Term2"},
		},
	}
	children := annotations.GetChildren()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestComplexTypeGetName(t *testing.T) {
	complexType := csdl.ComplexType{Name: "TestComplex"}
	if complexType.GetName() != "TestComplex" {
		t.Errorf("GetName() = %v, want %v", complexType.GetName(), "TestComplex")
	}
}

func TestComplexTypeGetXPath(t *testing.T) {
	complexType := csdl.ComplexType{Name: "TestComplex"}
	expected := "ComplexType[@Name='TestComplex']"
	if complexType.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", complexType.GetXPath(), expected)
	}
}

func TestComplexTypeGetChildren(t *testing.T) {
	complexType := csdl.ComplexType{
		Name: "TestComplex",
		Property: []*csdl.Property{
			{Name: "Prop1"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1"},
		},
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := complexType.GetChildren()
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	}
}

func TestComplexTypeSetGetSchema(t *testing.T) {
	schema := &csdl.Schema{Namespace: "Test.v1_0_0"}
	complexType := &csdl.ComplexType{Name: "TestComplex"}

	complexType.SetSchema(schema)
	result := complexType.GetSchema()

	if result != schema {
		t.Error("SetSchema/GetSchema failed")
	}
}

func TestEntityContainerGetName(t *testing.T) {
	entityContainer := &csdl.EntityContainer{Name: "ServiceRoot"}
	if entityContainer.GetName() != "ServiceRoot" {
		t.Errorf("GetName() = %v, want %v", entityContainer.GetName(), "ServiceRoot")
	}
}

func TestEntityContainerGetXPath(t *testing.T) {
	entityContainer := &csdl.EntityContainer{Name: "ServiceRoot"}
	expected := "EntityContainer[@Name='ServiceRoot']"
	if entityContainer.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", entityContainer.GetXPath(), expected)
	}
}

func TestEntityContainerGetChildren(t *testing.T) {
	entityContainer := &csdl.EntityContainer{Name: "ServiceRoot"}
	children := entityContainer.GetChildren()
	if children != nil {
		t.Error("Expected nil children for EntityContainer")
	}
}

func TestSchemaGetChildrenIncludesEntityContainer(t *testing.T) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
		EntityContainer: []*csdl.EntityContainer{
			{Name: "ServiceRoot"},
		},
	}

	children := schema.GetChildren()
	if len(children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(children))
	}

	container, ok := children[0].(*csdl.EntityContainer)
	if !ok {
		t.Fatalf("Expected child type *csdl.EntityContainer, got %T", children[0])
	}
	if container.Name != "ServiceRoot" {
		t.Errorf("EntityContainer.Name = %v, want %v", container.Name, "ServiceRoot")
	}
}

func TestEntityTypeGetName(t *testing.T) {
	entityType := csdl.EntityType{Name: "TestEntity"}
	if entityType.GetName() != "TestEntity" {
		t.Errorf("GetName() = %v, want %v", entityType.GetName(), "TestEntity")
	}
}

func TestEntityTypeGetXPath(t *testing.T) {
	entityType := csdl.EntityType{Name: "TestEntity"}
	expected := "EntityType[@Name='TestEntity']"
	if entityType.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", entityType.GetXPath(), expected)
	}
}

func TestEntityTypeGetChildren(t *testing.T) {
	entityType := csdl.EntityType{
		Name: "TestEntity",
		Property: []*csdl.Property{
			{Name: "Prop1"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1"},
		},
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := entityType.GetChildren()
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	}
}

func TestEntityTypeSetGetSchema(t *testing.T) {
	schema := &csdl.Schema{Namespace: "Test.v1_0_0"}
	entityType := &csdl.EntityType{Name: "TestEntity"}

	entityType.SetSchema(schema)
	result := entityType.GetSchema()

	if result != schema {
		t.Error("SetSchema/GetSchema failed")
	}
}

func TestEnumTypeGetName(t *testing.T) {
	enumType := csdl.EnumType{Name: "TestEnum"}
	if enumType.GetName() != "TestEnum" {
		t.Errorf("GetName() = %v, want %v", enumType.GetName(), "TestEnum")
	}
}

func TestEnumTypeGetXPath(t *testing.T) {
	enumType := csdl.EnumType{Name: "TestEnum"}
	expected := "EnumType[@Name='TestEnum']"
	if enumType.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", enumType.GetXPath(), expected)
	}
}

func TestEnumTypeGetChildren(t *testing.T) {
	enumType := csdl.EnumType{
		Name: "TestEnum",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
		Member: []*csdl.Member{
			{Name: "Member1"},
		},
	}
	children := enumType.GetChildren()
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestMemberGetName(t *testing.T) {
	member := csdl.Member{Name: "TestMember"}
	if member.GetName() != "TestMember" {
		t.Errorf("GetName() = %v, want %v", member.GetName(), "TestMember")
	}
}

func TestMemberGetXPath(t *testing.T) {
	member := csdl.Member{Name: "TestMember"}
	expected := "Member[@Name='TestMember']"
	if member.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", member.GetXPath(), expected)
	}
}

func TestMemberGetChildren(t *testing.T) {
	member := csdl.Member{
		Name: "TestMember",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := member.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestNavigationPropertyGetName(t *testing.T) {
	navProp := csdl.NavigationProperty{Name: "TestNav"}
	if navProp.GetName() != "TestNav" {
		t.Errorf("GetName() = %v, want %v", navProp.GetName(), "TestNav")
	}
}

func TestNavigationPropertyGetXPath(t *testing.T) {
	navProp := csdl.NavigationProperty{Name: "TestNav"}
	expected := "NavigationProperty[@Name='TestNav']"
	if navProp.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", navProp.GetXPath(), expected)
	}
}

func TestNavigationPropertyGetChildren(t *testing.T) {
	navProp := csdl.NavigationProperty{
		Name: "TestNav",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := navProp.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestParameterGetName(t *testing.T) {
	param := csdl.Parameter{Name: "TestParam"}
	if param.GetName() != "TestParam" {
		t.Errorf("GetName() = %v, want %v", param.GetName(), "TestParam")
	}
}

func TestParameterGetXPath(t *testing.T) {
	param := csdl.Parameter{Name: "TestParam"}
	expected := "Parameter[@Name='TestParam']"
	if param.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", param.GetXPath(), expected)
	}
}

func TestParameterGetChildren(t *testing.T) {
	param := csdl.Parameter{
		Name: "TestParam",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := param.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestPropertyGetName(t *testing.T) {
	prop := csdl.Property{Name: "TestProp"}
	if prop.GetName() != "TestProp" {
		t.Errorf("GetName() = %v, want %v", prop.GetName(), "TestProp")
	}
}

func TestPropertyGetXPath(t *testing.T) {
	prop := csdl.Property{Name: "TestProp"}
	expected := "Property[@Name='TestProp']"
	if prop.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", prop.GetXPath(), expected)
	}
}

func TestPropertyGetChildren(t *testing.T) {
	prop := csdl.Property{
		Name: "TestProp",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := prop.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestTermGetName(t *testing.T) {
	term := csdl.Term{Name: "TestTerm"}
	if term.GetName() != "TestTerm" {
		t.Errorf("GetName() = %v, want %v", term.GetName(), "TestTerm")
	}
}

func TestTermGetXPath(t *testing.T) {
	term := csdl.Term{Name: "TestTerm"}
	expected := "Term[@Name='TestTerm']"
	if term.GetXPath() != expected {
		t.Errorf("GetXPath() = %v, want %v", term.GetXPath(), expected)
	}
}

func TestTermGetChildren(t *testing.T) {
	term := csdl.Term{
		Name: "TestTerm",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := term.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func TestTypeDefinitionGetName(t *testing.T) {
	typeDef := csdl.TypeDefinition{Name: "TestTypeDef"}
	if typeDef.GetName() != "TestTypeDef" {
		t.Errorf("GetName() = %v, want %v", typeDef.GetName(), "TestTypeDef")
	}
}

func TestTypeDefinitionGetXPath(t *testing.T) {
	typeDef := csdl.TypeDefinition{Name: "TestTypeDef"}
	if typeDef.GetXPath() != "TestTypeDef" {
		t.Errorf("GetXPath() = %v, want %v", typeDef.GetXPath(), "TestTypeDef")
	}
}

func TestTypeDefinitionGetChildren(t *testing.T) {
	typeDef := csdl.TypeDefinition{
		Name: "TestTypeDef",
		Annotation: []*csdl.Annotation{
			{Term: "Term1"},
		},
	}
	children := typeDef.GetChildren()
	if len(children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(children))
	}
}

func BenchmarkSchemaGetChildren(b *testing.B) {
	schema := &csdl.Schema{
		Namespace: "Test.v1_0_0",
		Action: []*csdl.Action{
			{Name: "Action1"},
			{Name: "Action2"},
		},
		ComplexType: []*csdl.ComplexType{
			{Name: "Complex1"},
			{Name: "Complex2"},
		},
		EntityType: []*csdl.EntityType{
			{Name: "Entity1"},
			{Name: "Entity2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.GetChildren()
	}
}

func BenchmarkComplexTypeGetChildren(b *testing.B) {
	complexType := csdl.ComplexType{
		Name: "TestComplex",
		Property: []*csdl.Property{
			{Name: "Prop1"},
			{Name: "Prop2"},
			{Name: "Prop3"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1"},
			{Name: "Nav2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = complexType.GetChildren()
	}
}

func BenchmarkEntityTypeGetChildren(b *testing.B) {
	entityType := csdl.EntityType{
		Name: "TestEntity",
		Property: []*csdl.Property{
			{Name: "Prop1"},
			{Name: "Prop2"},
			{Name: "Prop3"},
		},
		NavigationProperty: []*csdl.NavigationProperty{
			{Name: "Nav1"},
			{Name: "Nav2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entityType.GetChildren()
	}
}
