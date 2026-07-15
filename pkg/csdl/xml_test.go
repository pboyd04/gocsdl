package csdl

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestXMLDecodersMatchStandardDecoder(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("ComputerSystem_v1.xml")
	if err != nil {
		t.Fatal(err)
	}

	var standard Edmx
	if err := decodeStandardXML(bytes.NewReader(data), &standard); err != nil {
		t.Fatalf("standard decoder: %v", err)
	}
	var sax Edmx
	if err := decodeGosax(bytes.NewReader(data), &sax); err != nil {
		t.Fatalf("gosax decoder: %v", err)
	}

	if !reflect.DeepEqual(standard, sax) {
		t.Fatal("gosax decoder produced a different CSDL object graph")
	}

	var generated Edmx
	if err := decodeGeneratedCSDL(bytes.NewReader(data), &generated); err != nil {
		t.Fatalf("generated decoder: %v", err)
	}
	if !reflect.DeepEqual(standard, generated) {
		t.Fatal("generated decoder produced a different CSDL object graph")
	}
}

func TestGeneratedCSDLDecoderRejectsMalformedXML(t *testing.T) {
	t.Parallel()

	var edmx Edmx
	err := decodeGeneratedCSDL(bytes.NewBufferString("<Edmx><DataServices></Edmx>"), &edmx)
	if err == nil {
		t.Fatal("malformed XML did not return an error")
	}
}

func TestGeneratedCSDLDecoderMatchesStandardDecoderForCSDLModel(t *testing.T) {
	t.Parallel()

	data := []byte(`<edmx:Edmx xmlns:edmx="urn:edmx">
  <edmx:Reference Uri="ref&amp;.xml"><edmx:Include Namespace="Ref" Alias="R"><Unknown /></edmx:Include></edmx:Reference>
  <edmx:DataServices><Schema Namespace="Example" Alias="E">
    <Action Name="Do" IsBound="true"><Parameter Name="target" Type="E.Entity"><Annotation Term="A.Parameter" /></Parameter><ReturnType Type="Edm.Boolean"/><Unknown /></Action>
    <Annotations Target="E.Entity" Qualifier="Q"><Annotation Term="A.Target" String="text" EnumMember="E.One" Bool="true" Int="42" Decimal="1.5"><Record><PropertyValue Property="Enabled" Bool="true"><Unknown /></PropertyValue><Unknown /></Record></Annotation></Annotations>
    <ComplexType Name="Complex" BaseType="E.Base" Abstract="true" OpenType="true"><Property Name="Name" Type="Edm.String" Nullable="false" MaxLength="10" Unicode="true" Precision="2" Scale="1" SRID="variable" DefaultValue="x&amp;y"><Annotation Term="A.Property" /></Property><NavigationProperty Name="Parent" Type="E.Entity" Nullable="false" Partner="Children" ContainsTarget="true"><Annotation Term="A.Navigation" /></NavigationProperty></ComplexType>
    <EntityContainer Name="Container" Extends="Base.Container"><EntitySet /></EntityContainer>
    <EntityType Name="Entity" BaseType="E.Base" Abstract="true" OpenType="true" HasStream="true"><Key><PropertyRef Name="ID" /></Key><Property Name="ID" Type="Edm.Int32"/><NavigationProperty Name="Children" Type="Collection(E.Entity)"/><Annotation Term="A.Entity" /></EntityType>
    <EnumType Name="Choice" UnderlyingType="Edm.Int64" IsFlags="true"><Member Name="One" Value="1"><Annotation Term="A.Member" /></Member><Annotation Term="A.Enum" /></EnumType>
    <Function Name="Ignored"><Parameter Name="value" Type="Edm.String" /></Function>
    <Term Name="Term" Type="Edm.String"><Annotation Term="A.Term" /><Unknown /></Term>
    <TypeDefinition Name="Code" UnderlyingType="Edm.String"><Annotation Term="A.Type" /></TypeDefinition>
    <Unknown />
  </Schema></edmx:DataServices>
</edmx:Edmx>`)

	var standard Edmx
	if err := decodeStandardXML(bytes.NewReader(data), &standard); err != nil {
		t.Fatalf("standard decoder: %v", err)
	}
	var generated Edmx
	if err := decodeGeneratedCSDL(bytes.NewReader(data), &generated); err != nil {
		t.Fatalf("generated decoder: %v", err)
	}
	if !reflect.DeepEqual(standard, generated) {
		t.Fatal("generated decoder produced a different CSDL object graph")
	}
}

func TestSaxXMLDecoderReuse(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("ComputerSystem_v1.xml")
	if err != nil {
		t.Fatal(err)
	}

	decoder := NewSaxXMLDecoder()
	var first Edmx
	if err := decoder.Decode(bytes.NewReader(data), &first); err != nil {
		t.Fatalf("first decode: %v", err)
	}
	var second Edmx
	if err := decoder.Decode(bytes.NewReader(data), &second); err != nil {
		t.Fatalf("reused decoder: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("reused decoder produced a different CSDL object graph")
	}

	if err := decoder.Decode(bytes.NewBufferString("<invalid>"), &Edmx{}); err == nil {
		t.Fatal("invalid XML did not return an error")
	}
	var afterError Edmx
	if err := decoder.Decode(bytes.NewReader(data), &afterError); err != nil {
		t.Fatalf("decode after error: %v", err)
	}
	if !reflect.DeepEqual(first, afterError) {
		t.Fatal("decoder produced a different graph after recovering from an error")
	}
}

func BenchmarkComputerSystemXMLDecoders(b *testing.B) {
	data, err := os.ReadFile("ComputerSystem_v1.xml")
	if err != nil {
		b.Fatal(err)
	}

	benchmarks := []struct {
		name   string
		decode xmlDecodeFunc
	}{
		{name: "standard", decode: decodeStandardXML},
		{name: "gosax", decode: decodeGosax},
		{name: "generated", decode: decodeGeneratedCSDL},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			reader := bytes.NewReader(data)
			for b.Loop() {
				var edmx Edmx
				if err := benchmark.decode(reader, &edmx); err != nil {
					b.Fatal(err)
				}
				reader.Reset(data)
			}
		})
	}
}

func BenchmarkComputerSystemParserDecoders(b *testing.B) {
	data, err := os.ReadFile("ComputerSystem_v1.xml")
	if err != nil {
		b.Fatal(err)
	}

	benchmarks := []struct {
		name   string
		decode xmlDecodeFunc
	}{
		{name: "standard", decode: decodeStandardXML},
		{name: "gosax", decode: decodeGosax},
		{name: "generated", decode: decodeGeneratedCSDL},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				parser := NewParser()
				parser.decodeXML = benchmark.decode
				parser.AddFile("ComputerSystem_v1.xml", io.NopCloser(bytes.NewReader(data)))
				types, err := parser.Parse()
				if err != nil {
					b.Fatal(err)
				}
				if len(types) != 434 {
					b.Fatalf("Parse() returned %d types, want 434", len(types))
				}
			}
		})
	}
}
