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

package golang

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/openshift-online/ocm-api-metamodel/pkg/reporter"
)

// BufferBuilder is used to create a new Go buffer. Don't create it directly, use the
// NewBufferBuilder function instead.
type BufferBuilder struct {
	reporter  *reporter.Reporter
	output    string
	base      string
	pkg       string
	file      string
	functions map[string]interface{}
}

// Buffer is a type that simplifies the generation of Go code.
type Buffer struct {
	reporter  *reporter.Reporter
	pkg       string
	file      string
	functions map[string]interface{}
	imports   map[string]string
	code      *bytes.Buffer
}

// NewBufferBuilder creates a builder for Golang buffers.
func NewBufferBuilder() *BufferBuilder {
	return new(BufferBuilder)
}

// Reporter sets the object that will be used to report errors and other relevant information.
func (b *BufferBuilder) Reporter(reporter *reporter.Reporter) *BufferBuilder {
	b.reporter = reporter
	return b
}

// Output sets the directory where the source code will be generated.
func (b *BufferBuilder) Output(value string) *BufferBuilder {
	b.output = value
	return b
}

// Base sets the import path of the base package where the code will be generated.
func (b *BufferBuilder) Base(value string) *BufferBuilder {
	b.base = value
	return b
}

// Package sets the package name that will be added to the base import path to calculate the
// complete import path where the source code will be generated. The last segment of this path will
// also be used as the 'package' statement of the generated code.
func (b *BufferBuilder) Package(value string) *BufferBuilder {
	b.pkg = value
	return b
}

// File sets the name of the file, without the .go extension.
func (b *BufferBuilder) File(value string) *BufferBuilder {
	b.file = value
	return b
}

// Function adds a function that can then be used in the templates.
func (b *BufferBuilder) Function(name string, function interface{}) *BufferBuilder {
	if b.functions == nil {
		b.functions = make(map[string]interface{})
	}
	b.functions[name] = function
	return b
}

// Build creates a new buffer using the configuration stored in the builder.
func (b *BufferBuilder) Build() (buffer *Buffer, err error) {
	// Check that the mandatory parameters have been provided:
	if b.reporter == nil {
		err = fmt.Errorf("reporter is mandatory")
		return
	}
	if b.output == "" {
		err = fmt.Errorf("output is mandatory")
		return
	}
	if b.base == "" {
		err = fmt.Errorf("base is mandatory")
		return
	}
	if b.file == "" {
		err = fmt.Errorf("file is mandatory")
		return
	}

	// Allocate and populate the buffer:
	buffer = new(Buffer)
	buffer.reporter = b.reporter
	buffer.pkg = path.Join(b.base, b.pkg)
	buffer.file = filepath.Join(b.output, b.pkg, b.file+".go")
	buffer.functions = make(map[string]interface{})
	buffer.functions["lineComment"] = buffer.lineComment
	for name, function := range b.functions {
		buffer.functions[name] = function
	}
	buffer.imports = make(map[string]string)
	buffer.code = new(bytes.Buffer)

	return
}

// Import adds an import statement for the given package. The value of imprt should be the complete
// import path of the package, and the value of selector should be the selector used to import it.
func (b *Buffer) Import(imprt, selector string) {
	if imprt != "" && imprt != b.pkg {
		b.imports[imprt] = selector
	}
}

// Emit writes to the code buffer, using the given template and arguments. The syntax of the
// template is the one used by the text/template package. The arguments should be a set nave/value
// pairs. Names should be strings, and values can be anything. These names and values will be put in
// a map that will then be the data object used to execute the template. For example, a template
// that generates the code of a Go method could be used like this:
//
//	// Calculate the name of the method and of the type:
//	typeName := ...
//	methodName := ...
//
//	// Generate the code of the method:
//	buffer.Emit(`
//		func (c *{{ .type }}) {{ .method }} {
//			...
//		}
//		`,
//		"type", typeName,
//		"method", methodName,
//	)
//
// All the facilities defined in the text/template package will also be available in these
// templates.
func (b *Buffer) Emit(tmpl string, args ...interface{}) {
	// Check that there is an even number of args, and that the first of each pair is
	// an string:
	count := len(args)
	if count%2 != 0 {
		b.reporter.Errorf(
			"Template '%s' should have an even number of arguments, but it has %d",
			tmpl, count,
		)
		return
	}
	for i := 0; i < count; i = i + 2 {
		name := args[i]
		_, ok := name.(string)
		if !ok {
			b.reporter.Errorf(
				"Argument %d of template '%s' is a key, so it should be a string, "+
					"but its type is %T",
				i, tmpl, name,
			)
		}
	}
	if b.reporter.Errors() > 0 {
		return
	}

	// Put the variables in the map that will be passed as the data object for the execution of
	// the template:
	data := make(map[string]interface{})
	for i := 0; i < count; i = i + 2 {
		name := args[i].(string)
		value := args[i+1]
		data[name] = value
	}

	// Process qualified names and type references in the data:
	for name, value := range data {
		data[name] = b.replaceValue(value)
	}

	// Wrap each function with another function that replaces qualified names and type
	// references in the returned values:
	functions := make(map[string]interface{})
	for name, function := range b.functions {
		value := reflect.ValueOf(function)
		wrapper := func(data interface{}) (result interface{}, err error) {
			results := value.Call([]reflect.Value{reflect.ValueOf(data)})
			if len(results) > 2 {
				err = fmt.Errorf(
					"expected at most 2 return values from '%s' but got %d",
					name, len(results),
				)
				return
			}
			if len(results) > 1 {
				switch typed := results[1].Interface().(type) {
				case error:
					err = typed
					return
				}
			}
			result = b.replaceValue(results[0].Interface())
			return
		}
		functions[name] = wrapper
	}

	// Parse the template:
	obj, err := template.New("").
		Funcs(functions).
		Parse(tmpl)
	if err != nil {
		b.reporter.Errorf("Can't parse template '%s': %v", tmpl, err)
		return
	}

	// Execute the template:
	buffer := new(bytes.Buffer)
	err = obj.Execute(buffer, data)
	if err != nil {
		b.reporter.Errorf("Can't execute template '%s': %v", tmpl, err)
		return
	}

	// Add the generated text to the code:
	text := buffer.String()
	_, err = b.code.WriteString(text)
	if err != nil {
		b.reporter.Errorf("Can't add generated text '%s' to the buffer: %v", text, err)
		return
	}

	// Add a blank line:
	b.code.WriteString("\n")
}

// Write creates the output file and writes the generated content.
func (b *Buffer) Write() error {
	var err error

	// Inform that we are writing the file:
	b.reporter.Infof("Writing file '%s'", b.file)

	// Check if there were errors:
	errors := b.reporter.Errors()
	if errors > 0 {
		if errors > 1 {
			err = fmt.Errorf("there were %d errors", errors)
		} else {
			err = fmt.Errorf("there was 1 error")
		}
		return err
	}

	// Make sure that the output directory exists:
	outputDir := filepath.Dir(b.file)
	err = os.MkdirAll(outputDir, 0777)
	if err != nil {
		return fmt.Errorf("can't create output directory '%s': %v", outputDir, err)
	}

	// Open the file:
	outputFd, err := os.OpenFile(b.file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0744)
	if err != nil {
		return fmt.Errorf("can't open output file '%s': %v", b.file, err)
	}

	// Write the header:
	fmt.Fprintf(outputFd, "%s\n", fileHeader)
	fmt.Fprintf(outputFd, "\n")

	// Write the package statement:
	fmt.Fprintf(outputFd, "package %s // %s\n", path.Base(b.pkg), b.pkg)
	fmt.Fprintf(outputFd, "\n")

	// Write the import statement:
	if len(b.imports) > 0 {
		fmt.Fprintf(outputFd, "import (\n")
		for imprt, selector := range b.imports {
			if selector != "" {
				fmt.Fprintf(outputFd, "%s \"%s\"\n", selector, imprt)
			} else {
				fmt.Fprintf(outputFd, "\"%s\"\n", imprt)
			}
		}
		fmt.Fprintf(outputFd, ")\n")
		fmt.Fprintf(outputFd, "\n")
	}

	// Remove extra blank lines from the code:
	err = b.removeExtraBlankLines()
	if err != nil {
		return fmt.Errorf("can't remove extra vlank lines: %v", err)
	}

	// Write the code:
	_, err = b.code.WriteTo(outputFd)
	if err != nil {
		return fmt.Errorf("can't write generated code to file '%s': %v", b.file, err)
	}

	// Close the file:
	err = outputFd.Close()
	if err != nil {
		return fmt.Errorf("can't close output file '%s': %v", b.file, err)
	}

	// Run goimport on the file:
	fmtCmd := exec.Command("goimports", "-w", b.file)
	err = fmtCmd.Run()
	if err != nil {
		return fmt.Errorf("can't format output file '%s': %v", b.file, err)
	}

	return nil
}

// removeExtraBlankLines removes the extra blank lines from the code, so that later `goimports` will
// generate nicer code.
func (b *Buffer) removeExtraBlankLines() error {
	clean := new(bytes.Buffer)
	var depth int
	var curr string
	var next string
	scanner := bufio.NewScanner(b.code)
	for scanner.Scan() {
		curr = next
		next = scanner.Text()
		currTrim := strings.TrimSpace(curr)
		nextTrim := strings.TrimSpace(next)
		if strings.HasSuffix(currTrim, " {") {
			depth++
		}
		if currTrim == "" {
			if depth > 0 && strings.HasPrefix(nextTrim, "//") {
				fmt.Fprintln(clean, curr)
			}
		} else {
			fmt.Fprintln(clean, curr)
		}
		if currTrim == "}" {
			depth--
		}
	}
	err := scanner.Err()
	if err != nil {
		return err
	}
	b.code = clean
	return nil
}

// replaceValue replaces type references with their corresponding text, and adds the required
// imports to the buffer.
func (b *Buffer) replaceValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case *TypeReference:
		imprt := typed.Import()
		selector := typed.Selector()
		text := typed.Text()
		b.Import(imprt, selector)
		if imprt == b.pkg {
			text = strings.Replace(text, selector+".", "", -1)
		}
		return text
	default:
		return value
	}
}

// lineComment converts the given text into a sequence of line comments.
func (b *Buffer) lineComment(value string) string {
	lines := strings.Split(value, "\n")
	for i, line := range lines {
		lines[i] = fmt.Sprintf("// %s", line)
	}
	return strings.Join(lines, "\n")
}

// Header that will be included in all generated files:
const fileHeader = `/*
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

// IMPORTANT: This file has been generated automatically, refrain from modifying it manually as all
// your changes will be lost when the file is generated again.
`
