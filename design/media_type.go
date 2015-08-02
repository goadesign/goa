package design

import (
	"fmt"
	"strings"
)

// NewMediaType creates new media type from its identifier, description and type.
// Initializes a default view that returns all the media type members.
func NewMediaType(id, desc string, o Object) *MediaTypeDefinition {
	mt := MediaTypeDefinition{Object: o, Identifier: id, Description: desc, Links: make(map[string]*LinkDefinition)}
	mt.Views = map[string]*ViewDefinition{"default": &ViewDefinition{Name: "default", Object: o}}
	return &mt
}

// View adds a new view to the media type.
// It returns the view so it can be modified further.
// This method ignore passed-in property names that do not exist in media type.
func (m *MediaTypeDefinition) View(name string, members ...string) *ViewDefinition {
	o := make(Object, len(members))
	i := 0
	for n, p := range m.Object {
		found := false
		for _, m := range members {
			if m == n {
				found = true
				break
			}
		}
		if found {
			o[n] = p
			i++
		}
	}
	view := ViewDefinition{Name: name, Object: o, MediaType: m}
	m.Views[name] = &view
	return &view
}

// With sets the list of member names rendered by view.
// If a member is a media type then the view used to render it defaults to the view with same name.
// The view used to renber media types members can be explicitely set using the syntax
// "<member name>:<view name>". For example:
//     m.View("expanded").As("id", "expensive_attribute:default")
func (v *ViewDefinition) With(members ...string) *ViewDefinition {
	o := Object{}
	for _, m := range members {
		elems := strings.SplitN(m, ":", 2)
		mm, ok := v.MediaType.Object[elems[0]]
		if !ok {
			panic(fmt.Sprintf("Invalid view member '%s', no such media type member.", m))
		}
		if len(elems) > 1 {
			if mm.Type.Kind() != ObjectType {
				panic(fmt.Sprintf("Cannot use view '%s' to render media type member '%s': not a media type", elems[1], elems[0]))
			}
		}
		o[m] = mm
	}
	v.Object = o
	return v
}

// Link specifies the list of links rendered with this media type.
func (v *ViewDefinition) Link(links ...string) *ViewDefinition {
	for _, l := range links {
		if _, ok := v.MediaType.Links[l]; !ok {
			panic(fmt.Sprintf("Invalid view link '%s', no such media type link.", l))
		}
	}
	v.Links = append(v.Links, links...)
	return v
}

// Link adds a new link to the given media type member.
// It returns the link so it can be modified further.
func (m *MediaTypeDefinition) Link(name string) *LinkDefinition {
	member, ok := m.Object[name]
	if !ok {
		panic(fmt.Sprintf("Invalid  link '%s', no such media type member.", name))
	}
	link := LinkDefinition{Name: name, Member: member, MediaType: m}
	m.Links[name] = &link
	return &link
}

// As overrides the link name.
// It returns the link so it can be modified further.
func (l *LinkDefinition) As(name string) *LinkDefinition {
	delete(l.MediaType.Links, l.Name)
	l.Name = name
	l.MediaType.Links[name] = l
	return l
}

// CollectionOf creates a collection media type from its element media type.
// A collection media type represents the content of responses that return a
// collection of resources such as "index" actions.
func CollectionOf(m *MediaTypeDefinition) *MediaTypeDefinition {
	col := *m
	col.isCollection = true
	return &col
}
