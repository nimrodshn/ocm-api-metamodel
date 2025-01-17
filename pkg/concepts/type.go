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

package concepts

import (
	"github.com/openshift-online/ocm-api-metamodel/pkg/names"
)

// TypeKind specifies the kind of a type. It can be scalar, enum, struct, list or class.
type TypeKind int

// Values of the TypeKind type:
const (
	UndefinedType TypeKind = iota
	ClassType
	EnumType
	ListType
	MapType
	ScalarType
	StructType
)

// String generates the string representation of a type kind.
func (k TypeKind) String() string {
	switch k {
	case ClassType:
		return "class"
	case EnumType:
		return "enum"
	case ListType:
		return "list"
	case MapType:
		return "map"
	case ScalarType:
		return "scalar"
	case StructType:
		return "struct"
	default:
		return "unknown"
	}
}

// NewType creates a new type.
func NewType() *Type {
	typ := new(Type)
	typ.kind = UndefinedType
	return typ
}

// Type specifies the data type of attributes of structs and method parameters.
type Type struct {
	owner      *Version
	doc        string
	kind       TypeKind
	name       *names.Name
	attributes []*Attribute
	values     []*EnumValue
	element    *Type
	index      *Type
}

// Owner returns the version that owns this type.
func (t *Type) Owner() *Version {
	return t.owner
}

// SetOwner sets the version that owns this type.
func (t *Type) SetOwner(value *Version) {
	t.owner = value
}

// Doc returns the documentation of this type.
func (t *Type) Doc() string {
	return t.doc
}

// SetDoc sets the documentation of this type.
func (t *Type) SetDoc(value string) {
	t.doc = value
}

// Kind returns the kind of this type.
func (t *Type) Kind() TypeKind {
	return t.kind
}

// SetKind sets the kind of this type.
func (t *Type) SetKind(value TypeKind) {
	t.kind = value
}

// IsClass returns true iff this type is a class type.
func (t *Type) IsClass() bool {
	return t.kind == ClassType
}

// IsEnum returns true iff this type is an enum type.
func (t *Type) IsEnum() bool {
	return t.kind == EnumType
}

// IsList returns true iff this type is a list type.
func (t *Type) IsList() bool {
	return t.kind == ListType
}

// IsMap returns true iff this type is a map type.
func (t *Type) IsMap() bool {
	return t.kind == MapType
}

// IsScalar returns true iff this type is an scalar type.
func (t *Type) IsScalar() bool {
	return t.kind == ScalarType || t.kind == EnumType
}

// IsStruct returns true iff this type is an struct type. Note that class types are also considered
// struct types.
func (t *Type) IsStruct() bool {
	return t.kind == ClassType || t.kind == StructType
}

// Name returns the name of this type.
func (t *Type) Name() *names.Name {
	return t.name
}

// SetName sets the name of this type.
func (t *Type) SetName(value *names.Name) {
	t.name = value
}

// Attributes returns the list of attributes of an struct type. If called for any other kind of type
// it will return nil.
func (t *Type) Attributes() []*Attribute {
	return t.attributes
}

// AddAttribute adds an attribute to the type, assuming hat it is an structured type.
func (t *Type) AddAttribute(attribute *Attribute) {
	if attribute != nil {
		t.attributes = append(t.attributes, attribute)
		attribute.SetOwner(t)
	}
}

// Values returns the list of values of an enumerated type. If called for any other kind of type it
// will return nil.
func (t *Type) Values() []*EnumValue {
	return t.values
}

// AddValue adds an enumerated value to the type, assuming that it is an enumerated type.
func (t *Type) AddValue(value *EnumValue) {
	if value != nil {
		t.values = append(t.values, value)
		value.SetType(t)
	}
}

// Element returns the element type for a list type. If called for any other kind of type it will
// return nil.
func (t *Type) Element() *Type {
	return t.element
}

// SetElement sets the element type for a list type.
func (t *Type) SetElement(value *Type) {
	t.element = value
}

// Index returns the index type for a list tpype. If called for any other kind of type it will
// return nil.
func (t *Type) Index() *Type {
	return t.index
}

// SetIndex sets the index type for a map type.
func (t *Type) SetIndex(value *Type) {
	t.index = value
}

// EnumValue represents each of the values of an enum type.
type EnumValue struct {
	typ  *Type
	doc  string
	name *names.Name
}

// NewEnumValue creates a new enumerated type value.
func NewEnumValue() *EnumValue {
	return new(EnumValue)
}

// Type returns the enum type that owns this value.
func (v *EnumValue) Type() *Type {
	return v.typ
}

// SetType sets the enum type that owns this value.
func (v *EnumValue) SetType(value *Type) {
	v.typ = value
}

// Doc returns the documentation of this value.
func (v *EnumValue) Doc() string {
	return v.doc
}

// SetDoc sets the documentation of this value.
func (v *EnumValue) SetDoc(value string) {
	v.doc = value
}

// Name return the name of this enum value.
func (v *EnumValue) Name() *names.Name {
	return v.name
}

// SetName sets the name of this enum value.
func (v *EnumValue) SetName(value *names.Name) {
	v.name = value
}
