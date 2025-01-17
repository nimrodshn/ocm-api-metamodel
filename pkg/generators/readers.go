/*
Copyright (c) 2019 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generators

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/openshift-online/ocm-api-metamodel/pkg/concepts"
	"github.com/openshift-online/ocm-api-metamodel/pkg/golang"
	"github.com/openshift-online/ocm-api-metamodel/pkg/names"
	"github.com/openshift-online/ocm-api-metamodel/pkg/nomenclator"
	"github.com/openshift-online/ocm-api-metamodel/pkg/reporter"
)

// ReadersGeneratorBuilder is an object used to configure and build the JSON readers generator.
// Don't create instances directly, use the NewReadersGenerator function instead.
type ReadersGeneratorBuilder struct {
	reporter *reporter.Reporter
	model    *concepts.Model
	output   string
	base     string
	names    *golang.NamesCalculator
	types    *golang.TypesCalculator
}

// ReadersGenerator generates code for the JSON readers. Don't create instances directly, use the
// builder instead.
type ReadersGenerator struct {
	reporter *reporter.Reporter
	errors   int
	model    *concepts.Model
	output   string
	base     string
	names    *golang.NamesCalculator
	types    *golang.TypesCalculator
	buffer   *golang.Buffer
}

// NewReadersGenerator creates a new builder JSON readers generators.
func NewReadersGenerator() *ReadersGeneratorBuilder {
	return new(ReadersGeneratorBuilder)
}

// Reporter sets the object that will be used to report information about the generation process,
// including errors.
func (b *ReadersGeneratorBuilder) Reporter(value *reporter.Reporter) *ReadersGeneratorBuilder {
	b.reporter = value
	return b
}

// Model sets the model that will be used by the types generator.
func (b *ReadersGeneratorBuilder) Model(value *concepts.Model) *ReadersGeneratorBuilder {
	b.model = value
	return b
}

// Output sets import path of the output package.
func (b *ReadersGeneratorBuilder) Output(value string) *ReadersGeneratorBuilder {
	b.output = value
	return b
}

// Base sets the import import path of the base output package.
func (b *ReadersGeneratorBuilder) Base(value string) *ReadersGeneratorBuilder {
	b.base = value
	return b
}

// Names sets the object that will be used to calculate names.
func (b *ReadersGeneratorBuilder) Names(value *golang.NamesCalculator) *ReadersGeneratorBuilder {
	b.names = value
	return b
}

// Types sets the object that will be used to calculate types.
func (b *ReadersGeneratorBuilder) Types(value *golang.TypesCalculator) *ReadersGeneratorBuilder {
	b.types = value
	return b
}

// Build checks the configuration stored in the builder and, if it is correct, creates a new types
// generator using it.
func (b *ReadersGeneratorBuilder) Build() (generator *ReadersGenerator, err error) {
	// Check that the mandatory parameters have been provided:
	if b.reporter == nil {
		err = fmt.Errorf("reporter is mandatory")
		return
	}
	if b.model == nil {
		err = fmt.Errorf("model is mandatory")
		return
	}
	if b.output == "" {
		err = fmt.Errorf("output is mandatory")
		return
	}
	if b.base == "" {
		err = fmt.Errorf("package is mandatory")
		return
	}
	if b.names == nil {
		err = fmt.Errorf("names is mandatory")
		return
	}
	if b.types == nil {
		err = fmt.Errorf("types is mandatory")
		return
	}

	// Create the generator:
	generator = new(ReadersGenerator)
	generator.reporter = b.reporter
	generator.model = b.model
	generator.output = b.output
	generator.base = b.base
	generator.names = b.names
	generator.types = b.types

	return
}

// Run executes the code generator.
func (g *ReadersGenerator) Run() error {
	var err error

	// Generate the helpers:
	err = g.generateHelpers()
	if err != nil {
		return err
	}

	// Generate the code for each type:
	for _, service := range g.model.Services() {
		for _, version := range service.Versions() {
			for _, typ := range version.Types() {
				switch {
				case typ.IsStruct():
					err = g.generateStructReader(typ)
				case typ.IsList() && typ.Element().IsStruct():
					err = g.generateListReader(typ)
				}
				if err != nil {
					return err
				}
			}
		}
	}

	// Check if there were errors:
	if g.errors > 0 {
		if g.errors > 1 {
			err = fmt.Errorf("there were %d errors", g.errors)
		} else {
			err = fmt.Errorf("there was 1 error")
		}
		return err
	}

	return nil
}

func (g *ReadersGenerator) generateHelpers() error {
	var err error

	// Calculate the package and file name:
	pkgName := g.helpersPkg()
	fileName := g.readersFile()

	// Create the buffer for the generated code:
	g.buffer, err = golang.NewBufferBuilder().
		Reporter(g.reporter).
		Output(g.output).
		Base(g.base).
		Package(pkgName).
		File(fileName).
		Build()
	if err != nil {
		return err
	}

	// Generate the code:
	g.buffer.Import("bytes", "")
	g.buffer.Import("encoding/json", "")
	g.buffer.Import("fmt", "")
	g.buffer.Import("io", "")
	g.buffer.Emit(`
		// NewEncoder creates a new JSON encoder from the given target. The target can be a
		// a writer or a JSON encoder.
		func NewEncoder(target interface{}) (encoder *json.Encoder, err error) {
			switch output := target.(type) {
			case io.Writer:
				encoder = json.NewEncoder(output)
			case *json.Encoder:
				encoder = output
			default:
				err = fmt.Errorf(
					"expected writer or JSON decoder, but got %T",
					output,
				)
			}
			return
		}

		// NewDecoder creates a new JSON decoder from the given source. The source can be a
		// slice of bytes, a string, a reader or a JSON decoder.
		func NewDecoder(source interface{}) (decoder *json.Decoder, err error) {
			switch input := source.(type) {
			case []byte:
				decoder = json.NewDecoder(bytes.NewBuffer(input))
			case string:
				decoder = json.NewDecoder(bytes.NewBufferString(input))
			case io.Reader:
				decoder = json.NewDecoder(input)
			case *json.Decoder:
				decoder = input
			default:
				err = fmt.Errorf(
					"expected bytes, string, reader or JSON decoder, but got %T",
					input,
				)
			}
			return
		}
	`)

	// Write the generated code:
	return g.buffer.Write()
}

func (g *ReadersGenerator) generateStructReader(typ *concepts.Type) error {
	var err error

	// Calculate the package and file name:
	pkgName := g.pkgName(typ.Owner())
	fileName := g.fileName(typ)

	// Create the buffer for the generated code:
	g.buffer, err = golang.NewBufferBuilder().
		Reporter(g.reporter).
		Output(g.output).
		Base(g.base).
		Package(pkgName).
		File(fileName).
		Function("dataFieldName", g.dataFieldName).
		Function("dataFieldType", g.dataFieldType).
		Function("dataStruct", g.dataStruct).
		Function("fieldTag", g.fieldTag).
		Function("marshalFunc", g.marshalFunc).
		Function("objectFieldName", g.objectFieldName).
		Function("objectName", g.objectName).
		Function("unmarshalFunc", g.unmarshalFunc).
		Function("valueType", g.valueType).
		Build()
	if err != nil {
		return err
	}

	// Generate the code:
	g.generateStructReaderSource(typ)

	// Write the generated code:
	return g.buffer.Write()
}

func (g *ReadersGenerator) generateStructReaderSource(typ *concepts.Type) {
	g.buffer.Import("fmt", "")
	g.buffer.Import(path.Join(g.base, g.helpersPkg()), "")
	g.buffer.Emit(`
		{{ $objectName := objectName .Type }}
		{{ $dataStruct := dataStruct .Type }}
		{{ $marshalFunc := marshalFunc .Type }}
		{{ $unmarshalFunc := unmarshalFunc .Type }}

		// {{ $dataStruct }} is the data structure used internally to marshal and unmarshal
		// objects of type '{{ .Type.Name }}'.
		type {{ $dataStruct }} struct {
			{{ if .Type.IsClass }}
				Kind *string "json:\"kind,omitempty\""
				ID   *string "json:\"id,omitempty\""
				HREF *string "json:\"href,omitempty\""
			{{ end }}
			{{ range .Type.Attributes }}
				{{ dataFieldName . }} {{ dataFieldType . }} "json:\"{{ fieldTag . }},omitempty\""
			{{ end }}
		}

		// {{ $marshalFunc }} writes a value of the '{{ .Type.Name }}' to the given target,
		// which can be a writer or a JSON encoder.
		func {{ $marshalFunc }}(object *{{ $objectName }}, target interface{}) error {
			encoder, err := helpers.NewEncoder(target)
			if err != nil {
				return err
			}
			data, err := object.wrap()
			if err != nil {
				return err
			}
			return encoder.Encode(data)
		}

		// wrap is the method used internally to convert a value of the '{{ .Type.Name }}'
		// value to a JSON document.
		func (o *{{ $objectName }}) wrap() (data *{{ $dataStruct }}, err error) {
			if o == nil {
				return
			}
			data = new({{ $dataStruct }})
			{{ if .Type.IsClass }}
				data.ID = o.id
				data.HREF = o.href
				data.Kind = new(string)
				if o.link {
					*data.Kind = {{ $objectName }}LinkKind
				} else {
					*data.Kind = {{ $objectName }}Kind
				}
			{{ end }}
			{{ range .Type.Attributes }}
				{{ $dataFieldName := dataFieldName . }}
				{{ $objectFieldName := objectFieldName . }}
				{{ if .Type.IsList }}
					{{ if .Type.Element.IsScalar }}
						data.{{ $dataFieldName }} = o.{{ $objectFieldName }}
					{{ else if .Type.Element.IsStruct }}
						{{ if and .Link .Type.Element.IsClass }}
							data.{{ $dataFieldName }}, err = o.{{ $objectFieldName }}.wrapLink()
						{{ else }}
							data.{{ $dataFieldName }}, err = o.{{ $objectFieldName }}.wrap()
						{{ end }}
						if err != nil {
							return
						}
					{{ end }}
				{{ else if .Type.IsStruct }}
					data.{{ $dataFieldName }}, err = o.{{ $objectFieldName }}.wrap()
					if err != nil {
						return
					}
				{{ else }}
					data.{{ $dataFieldName }} = o.{{ $objectFieldName }}
				{{ end }}
			{{ end }}
			return
		}

		// {{ $unmarshalFunc }} reads a value of the '{{ .Type.Name }}' type from the given
		// source, which can be an slice of bytes, a string, a reader or a JSON decoder.
		func {{ $unmarshalFunc }}(source interface{}) (object *{{ $objectName }}, err error) {
			decoder, err := helpers.NewDecoder(source)
			if err != nil {
				return
			}
			data := new({{ $dataStruct }})
			err = decoder.Decode(data)
			if err != nil {
				return
			}
			object, err = data.unwrap()
			return
		}

		// unwrap is the function used internally to convert the JSON unmarshalled data to a
		// value of the '{{ .Type.Name }}' type.
		func (d *{{ $dataStruct }}) unwrap() (object *{{ $objectName }}, err error) {
			if d == nil {
				return
			}
			object = new({{ $objectName }})
			{{ if .Type.IsClass }}
				object.id = d.ID
				object.href = d.HREF
				if d.Kind != nil {
					switch *d.Kind {
					case {{ $objectName }}Kind:
						object.link = false
					case {{ $objectName }}LinkKind:
						object.link = true
					default:
						err = fmt.Errorf(
							"expected kind '%s' or '%s' but got '%s'",
							{{ $objectName }}Kind,
							{{ $objectName }}LinkKind,
							*d.Kind,
						)
						return
					}
				}
			{{ end }}
			{{ range .Type.Attributes }}
				{{ $dataFieldName := dataFieldName . }}
				{{ $objectFieldName := objectFieldName . }}
				{{ if .Type.IsList }}
					{{ if .Type.Element.IsScalar }}
						object.{{ $objectFieldName }} = d.{{ $dataFieldName }}
					{{ else if .Type.Element.IsStruct }}
						{{ if and .Link .Type.Element.IsClass }}
							object.{{ $objectFieldName }}, err = d.{{ $dataFieldName }}.unwrapLink()
						{{ else }}
							object.{{ $objectFieldName }}, err = d.{{ $dataFieldName }}.unwrap()
						{{ end }}
						if err != nil {
							return
						}
					{{ end }}
				{{ else if .Type.IsStruct }}
					object.{{ $objectFieldName }}, err = d.{{ $dataFieldName }}.unwrap()
					if err != nil {
						return
					}
				{{ else }}
					object.{{ $objectFieldName }} = d.{{ $dataFieldName }}
				{{ end }}
			{{ end }}
			return
		}
		`,
		"Type", typ,
	)
}

func (g *ReadersGenerator) generateListReader(typ *concepts.Type) error {
	var err error

	// Calculate the package and file name:
	pkgName := g.pkgName(typ.Owner())
	fileName := g.fileName(typ)

	// Create the buffer for the generated code:
	g.buffer, err = golang.NewBufferBuilder().
		Reporter(g.reporter).
		Output(g.output).
		Base(g.base).
		Package(pkgName).
		File(fileName).
		Function("dataList", g.dataList).
		Function("dataStruct", g.dataStruct).
		Function("linkStruct", g.linkStruct).
		Function("marshalFunc", g.marshalFunc).
		Function("objectList", g.objectList).
		Function("objectName", g.objectName).
		Function("unmarshalFunc", g.unmarshalFunc).
		Build()
	if err != nil {
		return err
	}

	// Generate the code:
	g.generateListReaderSource(typ)

	// Write the generated code:
	return g.buffer.Write()
}

func (g *ReadersGenerator) generateListReaderSource(typ *concepts.Type) {
	g.buffer.Import("fmt", "")
	g.buffer.Import(path.Join(g.base, g.helpersPkg()), "")
	g.buffer.Emit(`
		{{ $objectName := objectName .Type.Element }}
		{{ $objectList := objectList .Type }}
		{{ $dataStruct := dataStruct .Type.Element }}
		{{ $dataList := dataList .Type }}
		{{ $marshalFunc := marshalFunc .Type }}
		{{ $unmarshalFunc := unmarshalFunc .Type }}

		// {{ $dataList }} is type used internally to marshal and unmarshal lists of objects
		// of type '{{ .Type.Element.Name }}'.
		type {{ $dataList }} []*{{ $dataStruct }}

		// {{ $unmarshalFunc }} reads a list of values of the '{{ .Type.Element.Name }}'
		// from the given source, which can be a slice of bytes, a string, an io.Reader or a
		// json.Decoder.
		func {{ $unmarshalFunc }}(source interface{}) (list *{{ $objectList }}, err error) {
			decoder, err := helpers.NewDecoder(source)
			if err != nil {
				return
			}
			var data {{ $dataList }}
			err = decoder.Decode(&data)
			if err != nil {
				return
			}
			list, err = data.unwrap()
			return
		}

		// wrap is the method used internally to convert a list of values of the
		// '{{ .Type.Element.Name }}' value to a JSON document.
		func (l *{{ $objectList }}) wrap() (data {{ $dataList }}, err error) {
			if l == nil {
				return
			}
			data = make({{ $dataList }}, len(l.items))
			for i, item := range l.items {
				data[i], err = item.wrap()
				if err != nil {
					return
				}
			}
			return
		}

		// unwrap is the function used internally to convert the JSON unmarshalled data to a
		// list of values of the '{{ .Type.Element.Name }}' type.
		func (d {{ $dataList }}) unwrap() (list *{{ $objectList }}, err error) {
			if d == nil {
				return
			}
			items := make([]*{{ $objectName }}, len(d))
			for i, item := range d {
				items[i], err = item.unwrap()
				if err != nil {
					return
				}
			}
			list = new({{ $objectList }})
			list.items = items
			return
		}

		{{ if .Type.Element.IsClass }}
			{{ $linkStruct := linkStruct .Type }}

			// {{ $linkStruct }} is type used internally to marshal and unmarshal links
			// to lists of objects of type '{{ .Type.Element.Name }}'.
			type {{ $linkStruct }} struct {
				Kind *string "json:\"kind,omitempty\""
				HREF *string "json:\"href,omitempty\""
				Items []*{{ $dataStruct }} "json:\"items,omitempty\""
			}

			// wrapLink is the method used internally to convert a list of values of the
			// '{{ .Type.Element.Name }}' value to a link.
			func (l *{{ $objectList }}) wrapLink() (data *{{ $linkStruct }}, err error) {
				if l == nil {
					return
				}
				items := make([]*{{ $dataStruct }}, len(l.items))
				for i, item := range l.items {
					items[i], err = item.wrap()
					if err != nil {
						return
					}
				}
				data = new({{ $linkStruct }})
				data.Items = items
				data.HREF = l.href
				data.Kind = new(string)
				if l.link {
					*data.Kind = {{ $objectName }}ListLinkKind
				} else {
					*data.Kind = {{ $objectName }}ListKind
				}
				return
			}

			// unwrapLink is the function used internally to convert a JSON link to a list
			// of values of the '{{ .Type.Element.Name }}' type to a list.
			func (d *{{ $linkStruct }}) unwrapLink() (list *{{ $objectList }}, err error) {
				if d == nil {
					return
				}
				items := make([]*{{ $objectName }}, len(d.Items))
				for i, item := range d.Items {
					items[i], err = item.unwrap()
					if err != nil {
						return
					}
				}
				list = new({{ $objectList }})
				list.items = items
				list.href = d.HREF
				if d.Kind != nil {
					switch *d.Kind {
					case {{ $objectName }}ListKind:
						list.link = false
					case {{ $objectName }}ListLinkKind:
						list.link = true
					default:
						err = fmt.Errorf(
							"expected kind '%s' or '%s' but got '%s'",
							{{ $objectName }}ListKind,
							{{ $objectName }}ListLinkKind,
							*d.Kind,
						)
						return
					}
				}
				return
			}
		{{ end }}
		`,
		"Type", typ,
	)
}

func (g *ReadersGenerator) readersFile() string {
	return g.names.File(nomenclator.Readers)
}

func (g *ReadersGenerator) helpersPkg() string {
	return g.names.Package(nomenclator.Helpers)
}

func (g *ReadersGenerator) pkgName(version *concepts.Version) string {
	servicePkg := g.names.Package(version.Owner().Name())
	versionPkg := g.names.Package(version.Name())
	return filepath.Join(servicePkg, versionPkg)
}

func (g *ReadersGenerator) fileName(typ *concepts.Type) string {
	return g.names.File(names.Cat(typ.Name(), nomenclator.Reader))
}

func (g *ReadersGenerator) marshalFunc(typ *concepts.Type) string {
	name := names.Cat(nomenclator.Marshal, typ.Name())
	return g.names.Public(name)
}

func (g *ReadersGenerator) unmarshalFunc(typ *concepts.Type) string {
	name := names.Cat(nomenclator.Unmarshal, typ.Name())
	return g.names.Public(name)
}

func (g *ReadersGenerator) objectName(typ *concepts.Type) string {
	return g.types.ValueReference(typ).Name()
}

func (g *ReadersGenerator) objectList(typ *concepts.Type) string {
	return g.types.ValueReference(typ).Name()
}

func (g *ReadersGenerator) dataStruct(typ *concepts.Type) string {
	return g.types.DataReference(typ).Name()
}

func (g *ReadersGenerator) linkStruct(typ *concepts.Type) string {
	return g.types.LinkDataReference(typ).Name()
}

func (g *ReadersGenerator) dataList(typ *concepts.Type) string {
	return g.types.DataReference(typ).Name()
}

func (g *ReadersGenerator) dataFieldName(attribute *concepts.Attribute) string {
	return g.names.Public(attribute.Name())
}

func (g *ReadersGenerator) dataFieldType(attribute *concepts.Attribute) *golang.TypeReference {
	if attribute.Link() {
		return g.types.LinkDataReference(attribute.Type())
	}
	return g.types.DataReference(attribute.Type())
}

func (g *ReadersGenerator) fieldTag(attribute *concepts.Attribute) string {
	return g.names.Tag(attribute.Name())
}

func (g *ReadersGenerator) objectFieldName(attribute *concepts.Attribute) string {
	return g.names.Private(attribute.Name())
}

func (g *ReadersGenerator) valueType(typ *concepts.Type) *golang.TypeReference {
	return g.types.ValueReference(typ)
}
