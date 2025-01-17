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
	"fmt"
	"strings"

	"github.com/openshift-online/ocm-api-metamodel/pkg/names"
	"github.com/openshift-online/ocm-api-metamodel/pkg/reporter"
)

// NamesCalculatorBuilder is an object used to configure and build the Go names calculators. Don't
// create instances directly, use the NewNamesCalculator function instead.
type NamesCalculatorBuilder struct {
	reporter *reporter.Reporter
	base     string
}

// NamesCalculator is an object used to calculate Go names. Don't create instances directly, use the
// builder instead.
type NamesCalculator struct {
	reporter *reporter.Reporter
	base     string
}

// NewNamesCalculator creates a Go names calculator builder.
func NewNamesCalculator() *NamesCalculatorBuilder {
	builder := new(NamesCalculatorBuilder)
	return builder
}

// Reporter sets the object that will be used to report information about the calculation processes,
// including errors.
func (b *NamesCalculatorBuilder) Reporter(value *reporter.Reporter) *NamesCalculatorBuilder {
	b.reporter = value
	return b
}

// Base sets the import path of the base package were the code will be generated.
func (b *NamesCalculatorBuilder) Base(value string) *NamesCalculatorBuilder {
	b.base = value
	return b
}

// Build checks the configuration stored in the builder and, if it is correct, creates a new
// calculator using it.
func (b *NamesCalculatorBuilder) Build() (calculator *NamesCalculator, err error) {
	// Check that the mandatory parameters have been provided:
	if b.reporter == nil {
		err = fmt.Errorf("reporter is mandatory")
		return
	}
	if b.base == "" {
		err = fmt.Errorf("base package is mandatory")
		return
	}

	// Create the calculator:
	calculator = new(NamesCalculator)
	calculator.reporter = b.reporter
	calculator.base = b.base

	return
}

// Public converts the given name into an string, following the rules for Go public names.
func (c *NamesCalculator) Public(name *names.Name) string {
	words := name.Words()
	chunks := make([]string, len(words))
	for i, word := range words {
		chunks[i] = word.Capitalize()
	}
	public := strings.Join(chunks, "")
	public = AvoidReservedWord(public)
	return public
}

// Private converts the given name into an string, following the rules for Go private names.
func (c *NamesCalculator) Private(name *names.Name) string {
	words := name.Words()
	chunks := make([]string, len(words))
	for i, word := range words {
		if i == 0 {
			chunks[i] = strings.ToLower(word.String())
		} else {
			chunks[i] = word.Capitalize()
		}
	}
	private := strings.Join(chunks, "")
	private = AvoidReservedWord(private)
	return private
}

// File converts the given name into an string, following the rules for Go source files.
func (c *NamesCalculator) File(name *names.Name) string {
	words := name.Words()
	chunks := make([]string, len(words))
	for i, word := range words {
		chunks[i] = strings.ToLower(word.String())
	}
	file := strings.Join(chunks, "_")
	return file
}

// Package converts the given name into an string, following the rules for Go package names.
func (c *NamesCalculator) Package(name *names.Name) string {
	words := name.Words()
	chunks := make([]string, len(words))
	for i, word := range words {
		chunks[i] = strings.ToLower(word.String())
	}
	dir := strings.Join(chunks, "")
	dir = AvoidReservedWord(dir)
	return dir
}

// Tag converts the given name into an string, following the rules for JSON field names.
func (c *NamesCalculator) Tag(name *names.Name) string {
	words := name.Words()
	chunks := make([]string, len(words))
	for i, word := range words {
		chunks[i] = strings.ToLower(word.String())
	}
	file := strings.Join(chunks, "_")
	return file
}
