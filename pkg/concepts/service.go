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

// Service is the representation of service, containing potentiall mutiple versions, for example the
// clusters management service.
type Service struct {
	// Model that owns this service:
	owner *Model

	// Name of the service:
	name *names.Name

	// All the versions of the service, indexed by name:
	versions map[string]*Version
}

// NewVersion creates a new empty service.
func NewService() *Service {
	service := new(Service)
	service.versions = make(map[string]*Version)
	return service
}

// Onwer returns the model that owns this service.
func (s *Service) Owner() *Model {
	return s.owner
}

// SetOnwer sets the model that owns this service.
func (s *Service) SetOwner(value *Model) {
	s.owner = value
}

// Name returns the name of this service.
func (s *Service) Name() *names.Name {
	return s.name
}

// SetName sets the name of this service.
func (s *Service) SetName(value *names.Name) {
	s.name = value
}

// Versions returns the list of versions of the service.
func (s *Service) Versions() []*Version {
	count := len(s.versions)
	versions := make([]*Version, count)
	index := 0
	for _, version := range s.versions {
		versions[index] = version
		index++
	}
	return versions
}

// FindVersion returns the version with the given name, or nil of there is no such version.
func (s *Service) FindVersion(name *names.Name) *Version {
	if name == nil {
		return nil
	}
	return s.versions[name.String()]
}

// AddVersion adds the given version to the service.
func (s *Service) AddVersion(version *Version) {
	if version != nil {
		s.versions[version.Name().String()] = version
		version.SetOwner(s)
	}
}

// AddVersions adds the given versions to the service.
func (s *Service) AddVersions(versions []*Version) {
	if len(versions) > 0 {
		for _, version := range versions {
			s.AddVersion(version)
		}
	}
}
