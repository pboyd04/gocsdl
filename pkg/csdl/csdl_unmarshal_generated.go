package csdl

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/orisano/gosax"
)

type generatedCSDLDecoder struct {
	reader *gosax.Reader
}

func newGeneratedCSDLDecoder() *generatedCSDLDecoder {
	return &generatedCSDLDecoder{reader: gosax.NewReaderSize(nil, gosaxBufferSize)}
}

var generatedCSDLDecoderPool = sync.Pool{
	New: func() any { return newGeneratedCSDLDecoder() },
}

func decodeGeneratedCSDL(reader io.Reader, value any) error {
	edmx, ok := value.(*Edmx)
	if !ok {
		return fmt.Errorf("generated CSDL decoder requires *csdl.Edmx, got %T", value)
	}
	decoder := generatedCSDLDecoderPool.Get().(*generatedCSDLDecoder)
	err := decoder.decode(reader, edmx)
	generatedCSDLDecoderPool.Put(decoder)
	return err
}

func (d *generatedCSDLDecoder) decode(reader io.Reader, edmx *Edmx) error {
	d.reader.Reset(reader)
	d.reader.EmitSelfClosingTag = true
	defer d.reader.Reset(nil)

	for {
		event, err := d.reader.Event()
		if err != nil {
			return err
		}
		switch event.Type() {
		case gosax.EventStart:
			if localName(event.Bytes) != "Edmx" {
				return fmt.Errorf("expected Edmx root element, got %s", localName(event.Bytes))
			}
			return d.parseEdmx(edmx)
		case gosax.EventEOF:
			return io.EOF
		}
	}
}

func localName(raw []byte) string {
	name, _ := gosax.Name(raw)
	if index := bytes.IndexByte(name, ':'); index >= 0 {
		name = name[index+1:]
	}
	return string(name)
}

func visitAttributes(raw []byte, visit func(string, []byte) error) error {
	_, remaining := gosax.Name(raw)
	for len(remaining) > 0 {
		attribute, rest, err := gosax.NextAttribute(remaining)
		if err != nil {
			return err
		}
		remaining = rest
		if len(attribute.Key) == 0 {
			return nil
		}
		if len(attribute.Value) < 2 {
			return fmt.Errorf("invalid value for attribute %s", attribute.Key)
		}
		name := attribute.Key
		if index := bytes.IndexByte(name, ':'); index >= 0 {
			name = name[index+1:]
		}
		value, err := gosax.Unescape(attribute.Value[1 : len(attribute.Value)-1])
		if err != nil {
			return err
		}
		if err := visit(string(name), value); err != nil {
			return err
		}
	}
	return nil
}

func parseBool(value []byte) (bool, error) {
	return strconv.ParseBool(string(value))
}

func (d *generatedCSDLDecoder) nextChild(parent string) (string, []byte, bool, error) {
	for {
		event, err := d.reader.Event()
		if err != nil {
			return "", nil, false, err
		}
		switch event.Type() {
		case gosax.EventStart:
			return localName(event.Bytes), event.Bytes, false, nil
		case gosax.EventEnd:
			name := localName(event.Bytes)
			if name != parent {
				return "", nil, false, fmt.Errorf("element <%s> closed by </%s>", parent, name)
			}
			return "", nil, true, nil
		case gosax.EventEOF:
			return "", nil, false, fmt.Errorf("unexpected EOF in <%s>", parent)
		}
	}
}

func (d *generatedCSDLDecoder) skipElement(parent string) error {
	stack := []string{parent}
	for len(stack) > 0 {
		event, err := d.reader.Event()
		if err != nil {
			return err
		}
		switch event.Type() {
		case gosax.EventStart:
			stack = append(stack, localName(event.Bytes))
		case gosax.EventEnd:
			name := localName(event.Bytes)
			last := len(stack) - 1
			if name != stack[last] {
				return fmt.Errorf("element <%s> closed by </%s>", stack[last], name)
			}
			stack = stack[:last]
		case gosax.EventEOF:
			return fmt.Errorf("unexpected EOF in <%s>", stack[len(stack)-1])
		}
	}
	return nil
}

func (d *generatedCSDLDecoder) parseEdmx(edmx *Edmx) error {
	for {
		name, raw, done, err := d.nextChild("Edmx")
		if err != nil || done {
			return err
		}
		switch name {
		case "Reference":
			value := &Reference{}
			if err := d.parseReference(raw, value); err != nil {
				return err
			}
			edmx.Reference = append(edmx.Reference, value)
		case "DataServices":
			if err := d.parseDataServices(&edmx.DataServices); err != nil {
				return err
			}
		default:
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseReference(raw []byte, value *Reference) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		if name == "Uri" {
			value.Uri = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("Reference")
		if err != nil || done {
			return err
		}
		if name == "Include" {
			var include Include
			if err := d.parseInclude(childRaw, &include); err != nil {
				return err
			}
			value.Include = append(value.Include, include)
		} else if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseInclude(raw []byte, value *Include) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Namespace":
			value.Namespace = string(data)
		case "Alias":
			value.Alias = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, _, done, err := d.nextChild("Include")
		if err != nil || done {
			return err
		}
		value.Other = append(value.Other, nil)
		if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseDataServices(value *DataServices) error {
	for {
		name, raw, done, err := d.nextChild("DataServices")
		if err != nil || done {
			return err
		}
		if name == "Schema" {
			schema := &Schema{}
			if err := d.parseSchema(raw, schema); err != nil {
				return err
			}
			value.Schema = append(value.Schema, schema)
		} else if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseSchema(raw []byte, value *Schema) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Namespace":
			value.Namespace = string(data)
		case "Alias":
			value.Alias = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("Schema")
		if err != nil || done {
			return err
		}
		switch name {
		case "Action":
			child := &Action{}
			if err := d.parseAction(childRaw, child); err != nil {
				return err
			}
			value.Action = append(value.Action, child)
		case "Annotations":
			child := &Annotations{}
			if err := d.parseAnnotations(childRaw, child); err != nil {
				return err
			}
			value.Annotations = append(value.Annotations, child)
		case "Annotation":
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		case "ComplexType":
			child := &ComplexType{}
			if err := d.parseComplexType(childRaw, child); err != nil {
				return err
			}
			value.ComplexType = append(value.ComplexType, child)
		case "EntityContainer":
			child := &EntityContainer{}
			if err := d.parseEntityContainer(childRaw, child); err != nil {
				return err
			}
			value.EntityContainer = append(value.EntityContainer, child)
		case "EntityType":
			child := &EntityType{}
			if err := d.parseEntityType(childRaw, child); err != nil {
				return err
			}
			value.EntityType = append(value.EntityType, child)
		case "EnumType":
			child := &EnumType{}
			if err := d.parseEnumType(childRaw, child); err != nil {
				return err
			}
			value.EnumType = append(value.EnumType, child)
		case "Function":
			child := &Function{}
			if err := d.parseFunction(child); err != nil {
				return err
			}
			value.Function = append(value.Function, child)
		case "Term":
			child := &Term{}
			if err := d.parseTerm(childRaw, child); err != nil {
				return err
			}
			value.Term = append(value.Term, child)
		case "TypeDefinition":
			child := &TypeDefinition{}
			if err := d.parseTypeDefinition(childRaw, child); err != nil {
				return err
			}
			value.TypeDefinition = append(value.TypeDefinition, child)
		default:
			value.Other = append(value.Other, nil)
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseAction(raw []byte, value *Action) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Name":
			value.Name = string(data)
		case "IsBound":
			value.IsBound, err = parseBool(data)
		}
		return err
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("Action")
		if err != nil || done {
			return err
		}
		switch name {
		case "Annotation":
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		case "Parameter":
			child := &Parameter{}
			if err := d.parseParameter(childRaw, child); err != nil {
				return err
			}
			value.Parameter = append(value.Parameter, child)
		case "ReturnType":
			child := &ReturnType{}
			if err := d.parseReturnType(childRaw, child); err != nil {
				return err
			}
			value.ReturnType = child
		default:
			value.Other = append(value.Other, nil)
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseAnnotation(raw []byte, value *Annotation) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Term":
			value.Term = string(data)
		case "Qualifier":
			value.Qualifier = string(data)
		case "String":
			value.String = string(data)
		case "EnumMember":
			value.EnumMember = string(data)
		case "Bool":
			value.Bool, err = parseBool(data)
		case "Int":
			value.Int, err = strconv.ParseInt(string(data), 10, 64)
		case "Decimal":
			value.Decimal, err = strconv.ParseFloat(string(data), 64)
		}
		return err
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("Annotation")
		if err != nil || done {
			return err
		}
		if name == "Record" {
			if err := d.parseRecord(childRaw, &value.Record); err != nil {
				return err
			}
		} else if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseAnnotations(raw []byte, value *Annotations) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Target":
			value.Target = string(data)
		case "Qualifier":
			value.Qualifier = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("Annotations")
		if err != nil || done {
			return err
		}
		if name == "Annotation" {
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		} else if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseComplexType(raw []byte, value *ComplexType) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Name":
			value.Name = string(data)
		case "BaseType":
			value.BaseType = string(data)
		case "Abstract":
			value.Abstract, err = parseBool(data)
		case "OpenType":
			value.OpenType, err = parseBool(data)
		}
		return err
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("ComplexType")
		if err != nil || done {
			return err
		}
		switch name {
		case "Property":
			child := &Property{}
			if err := d.parseProperty(childRaw, child); err != nil {
				return err
			}
			value.Property = append(value.Property, child)
		case "NavigationProperty":
			child := &NavigationProperty{}
			if err := d.parseNavigationProperty(childRaw, child); err != nil {
				return err
			}
			value.NavigationProperty = append(value.NavigationProperty, child)
		case "Annotation":
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		default:
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseEntityContainer(raw []byte, value *EntityContainer) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Name":
			value.Name = string(data)
		case "Extends":
			value.Extends = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, _, done, err := d.nextChild("EntityContainer")
		if err != nil || done {
			return err
		}
		value.Other = append(value.Other, nil)
		if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseEntityType(raw []byte, value *EntityType) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Name":
			value.Name = string(data)
		case "BaseType":
			value.BaseType = string(data)
		case "Abstract":
			value.Abstract, err = parseBool(data)
		case "OpenType":
			value.OpenType, err = parseBool(data)
		case "HasStream":
			value.HasStream, err = parseBool(data)
		}
		return err
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("EntityType")
		if err != nil || done {
			return err
		}
		switch name {
		case "Key":
			if err := d.skipElement(name); err != nil {
				return err
			}
		case "Property":
			child := &Property{}
			if err := d.parseProperty(childRaw, child); err != nil {
				return err
			}
			value.Property = append(value.Property, child)
		case "NavigationProperty":
			child := &NavigationProperty{}
			if err := d.parseNavigationProperty(childRaw, child); err != nil {
				return err
			}
			value.NavigationProperty = append(value.NavigationProperty, child)
		case "Annotation":
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		default:
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseEnumType(raw []byte, value *EnumType) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Name":
			value.Name = string(data)
		case "UnderlyingType":
			value.UnderlyingType = string(data)
		case "IsFlags":
			value.IsFlags, err = parseBool(data)
		}
		return err
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("EnumType")
		if err != nil || done {
			return err
		}
		switch name {
		case "Annotation":
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		case "Member":
			child := &Member{}
			if err := d.parseMember(childRaw, child); err != nil {
				return err
			}
			value.Member = append(value.Member, child)
		default:
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseFunction(value *Function) error {
	for {
		name, _, done, err := d.nextChild("Function")
		if err != nil || done {
			return err
		}
		value.Other = append(value.Other, nil)
		if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseMember(raw []byte, value *Member) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Name":
			value.Name = string(data)
		case "Value":
			parsed := string(data)
			value.Value = &parsed
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, childRaw, done, err := d.nextChild("Member")
		if err != nil || done {
			return err
		}
		if name == "Annotation" {
			child := &Annotation{}
			if err := d.parseAnnotation(childRaw, child); err != nil {
				return err
			}
			value.Annotation = append(value.Annotation, child)
		} else if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseNavigationProperty(raw []byte, value *NavigationProperty) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Name":
			value.Name = string(data)
		case "Type":
			value.Type = string(data)
		case "Nullable":
			parsed, parseErr := parseBool(data)
			value.Nullable, err = &parsed, parseErr
		case "Partner":
			value.Partner = string(data)
		case "ContainsTarget":
			value.ContainsTarget, err = parseBool(data)
		}
		return err
	}); err != nil {
		return err
	}
	return d.parseAnnotationChildren("NavigationProperty", &value.Annotation, nil)
}

func (d *generatedCSDLDecoder) parseParameter(raw []byte, value *Parameter) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Name":
			value.Name = string(data)
		case "Type":
			value.Type = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	return d.parseAnnotationChildren("Parameter", &value.Annotation, &value.Other)
}

func (d *generatedCSDLDecoder) parseProperty(raw []byte, value *Property) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Name":
			value.Name = string(data)
		case "Type":
			value.Type = string(data)
		case "Nullable":
			parsed, parseErr := parseBool(data)
			value.Nullable, err = &parsed, parseErr
		case "MaxLength":
			value.MaxLength, err = strconv.Atoi(string(data))
		case "Unicode":
			value.Unicode, err = parseBool(data)
		case "Precision":
			value.Precision, err = strconv.Atoi(string(data))
		case "Scale":
			value.Scale, err = strconv.Atoi(string(data))
		case "SRID":
			value.SRID = string(data)
		case "DefaultValue":
			value.DefaultValue = string(data)
		}
		return err
	}); err != nil {
		return err
	}
	return d.parseAnnotationChildren("Property", &value.Annotation, nil)
}

func (d *generatedCSDLDecoder) parseTerm(raw []byte, value *Term) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Name":
			value.Name = string(data)
		case "Type":
			value.Type = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	return d.parseAnnotationChildren("Term", &value.Annotation, &value.Other)
}

func (d *generatedCSDLDecoder) parseTypeDefinition(raw []byte, value *TypeDefinition) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		switch name {
		case "Name":
			value.Name = string(data)
		case "UnderlyingType":
			value.UnderlyingType = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	return d.parseAnnotationChildren("TypeDefinition", &value.Annotation, nil)
}

func (d *generatedCSDLDecoder) parseReturnType(raw []byte, value *ReturnType) error {
	if err := visitAttributes(raw, func(name string, data []byte) error {
		if name == "Type" {
			value.Type = string(data)
		}
		return nil
	}); err != nil {
		return err
	}
	for {
		name, _, done, err := d.nextChild("ReturnType")
		if err != nil || done {
			return err
		}
		if err := d.skipElement(name); err != nil {
			return err
		}
	}
}

func (d *generatedCSDLDecoder) parseAnnotationChildren(parent string, annotations *[]*Annotation, other *[]any) error {
	for {
		name, raw, done, err := d.nextChild(parent)
		if err != nil || done {
			return err
		}
		if name == "Annotation" {
			child := &Annotation{}
			if err := d.parseAnnotation(raw, child); err != nil {
				return err
			}
			*annotations = append(*annotations, child)
		} else {
			if other != nil {
				*other = append(*other, nil)
			}
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parseRecord(_ []byte, value *Record) error {
	for {
		name, raw, done, err := d.nextChild("Record")
		if err != nil || done {
			return err
		}
		if name == "PropertyValue" {
			child := &PropertyValue{}
			if err := d.parsePropertyValue(raw, child); err != nil {
				return err
			}
			value.PropertyValue = append(value.PropertyValue, child)
		} else {
			value.Other = append(value.Other, nil)
			if err := d.skipElement(name); err != nil {
				return err
			}
		}
	}
}

func (d *generatedCSDLDecoder) parsePropertyValue(raw []byte, value *PropertyValue) error {
	if err := visitAttributes(raw, func(name string, data []byte) (err error) {
		switch name {
		case "Property":
			value.Property = string(data)
		case "Bool":
			value.Bool, err = parseBool(data)
		}
		return err
	}); err != nil {
		return err
	}
	for {
		name, _, done, err := d.nextChild("PropertyValue")
		if err != nil || done {
			return err
		}
		value.Other = append(value.Other, nil)
		if err := d.skipElement(name); err != nil {
			return err
		}
	}
}
