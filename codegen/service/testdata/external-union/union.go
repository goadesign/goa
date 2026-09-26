// Package nestedalpha provides public union methods for external conversion
// tests. Its name deliberately matches the separate package supplying Child.
package nestedalpha

import leaf "goa.design/goa/v3/codegen/service/testdata/nested-alpha"

type (
	// Text_Value deliberately differs from a Goified name and cannot be
	// derived from Choice.
	Text_Value string

	// Text_List keeps a named element inside a named collection.
	Text_List []Text_Value

	// Text_Table keeps named keys and named collection values.
	Text_Table map[Text_Value]Text_List

	// Byte_Data keeps the name of a byte collection.
	Byte_Data []byte

	// Detail includes named scalar fields below an object branch and an
	// ordinary recursive object pointer. Leaf comes from another package.
	Detail struct {
		Value    Text_Value
		Optional *Text_Value
		Next     *Detail
		Leaf     *leaf.Child
	}

	// ChoiceKind identifies the branch stored by Choice.
	ChoiceKind string

	// Choice stores one selected value behind public methods.
	Choice struct {
		kind  ChoiceKind
		text  Text_Value
		child *Detail
		words Text_List
		table Text_Table
	}

	// ChoiceBox has one pointer union branch, as generated branch methods do.
	ChoiceBox struct {
		selected bool
		choice   *Choice
	}

	// Entry includes both value and pointer union fields.
	Entry struct {
		Choice   Choice
		Optional *Choice
		Choices  []Choice
		ByName   map[Text_Value]Choice
		Wrapped  ChoiceBox
		Blob     Byte_Data
	}

	// Envelope contains nested union values converted by a shared receiver.
	Envelope struct {
		Name    string
		Label   *Text_Value
		Words   Text_List
		Table   Text_Table
		Entries []*Entry
	}

	// DoublePointerEnvelope cannot be represented by one optional union field.
	DoublePointerEnvelope struct{ Choice **Choice }
	// TriplePointerEnvelope cannot be represented by one optional union field.
	TriplePointerEnvelope struct{ Choice ***Choice }
	// SlicePointerEnvelope stores union pointers where conversion stores values.
	SlicePointerEnvelope struct{ Choices []*Choice }
	// MapPointerEnvelope stores union pointers where conversion stores values.
	MapPointerEnvelope struct{ Choices map[string]*Choice }
	// MapKeyPointerEnvelope would lose the pointer identity of its keys.
	MapKeyPointerEnvelope struct{ Choices map[*Choice]string }
	// NestedPointerEnvelope exposes a union pointer below nested collections.
	NestedPointerEnvelope struct{ Choices []map[string]*Choice }

	// BadKind returns a discriminator with an incompatible Go kind.
	BadKind struct{ Choice }
	// BadGetter returns a nonboolean branch-presence result.
	BadGetter struct{ Choice }
	// BadSetter accepts a different type from its getter.
	BadSetter struct{ Choice }
	// ValueSetter cannot update the stored union value.
	ValueSetter struct{ Choice }
	// ExtraBranch adds a branch absent from the authored schema.
	ExtraBranch struct{ Choice }
	// PointerText uses a pointer for a scalar branch.
	PointerText struct{ Choice }
	// ValueChild uses a value for an object branch.
	ValueChild struct{ Choice }

	// KindEnvelope exposes BadKind as a converted field.
	KindEnvelope struct{ Choice BadKind }
	// GetterEnvelope exposes BadGetter as a converted field.
	GetterEnvelope struct{ Choice BadGetter }
	// SetterEnvelope exposes BadSetter as a converted field.
	SetterEnvelope struct{ Choice BadSetter }
	// ValueSetterEnvelope exposes ValueSetter as a converted field.
	ValueSetterEnvelope struct{ Choice ValueSetter }
	// ExtraEnvelope exposes ExtraBranch as a converted field.
	ExtraEnvelope struct{ Choice ExtraBranch }
	// PointerTextEnvelope exposes PointerText as a converted field.
	PointerTextEnvelope struct{ Choice PointerText }
	// ValueChildEnvelope exposes ValueChild as a converted field.
	ValueChildEnvelope struct{ Choice ValueChild }
	// ChoiceEnvelope exposes a valid Choice to negative schema tests.
	ChoiceEnvelope struct{ Choice Choice }
)

// Kind reports the selected branch.
func (u Choice) Kind() ChoiceKind {
	return u.kind
}

// AsText returns the exact named value when text is selected.
func (u Choice) AsText() (Text_Value, bool) {
	return u.text, u.kind == "text"
}

// SetText replaces the selected branch with text.
func (u *Choice) SetText(value Text_Value) {
	*u = Choice{kind: "text", text: value}
}

// AsChild returns the external object when child is selected.
func (u Choice) AsChild() (*Detail, bool) {
	return u.child, u.kind == "child"
}

// SetChild replaces the selected branch with child.
func (u *Choice) SetChild(value *Detail) {
	*u = Choice{kind: "child", child: value}
}

// AsWords returns the named collection when words is selected.
func (u Choice) AsWords() (Text_List, bool) {
	return u.words, u.kind == "words"
}

// SetWords replaces the selected branch with words.
func (u *Choice) SetWords(value Text_List) {
	*u = Choice{kind: "words", words: value}
}

// AsTable returns the named map when table is selected.
func (u Choice) AsTable() (Text_Table, bool) {
	return u.table, u.kind == "table"
}

// SetTable replaces the selected branch with table.
func (u *Choice) SetTable(value Text_Table) {
	*u = Choice{kind: "table", table: value}
}

// Kind reports whether the choice branch is selected.
func (u ChoiceBox) Kind() string {
	if u.selected {
		return "choice"
	}
	return ""
}

// AsChoice returns the selected union pointer.
func (u ChoiceBox) AsChoice() (*Choice, bool) {
	return u.choice, u.selected
}

// SetChoice replaces the selected branch with a union pointer.
func (u *ChoiceBox) SetChoice(value *Choice) {
	*u = ChoiceBox{selected: true, choice: value}
}

// Kind deliberately violates the discriminator result contract.
func (BadKind) Kind() int {
	return 0
}

// AsText deliberately violates the presence-result contract.
func (BadGetter) AsText() (Text_Value, int) {
	return "", 0
}

// SetText deliberately disagrees with the getter's named parameter type.
func (*BadSetter) SetText(string) {}

// SetText deliberately uses a value receiver.
func (ValueSetter) SetText(Text_Value) {}

// AsOther exposes an extra branch.
func (ExtraBranch) AsOther() (int, bool) {
	return 0, false
}

// SetOther exposes the extra branch's setter.
func (*ExtraBranch) SetOther(int) {}

// AsText deliberately returns a scalar pointer.
func (PointerText) AsText() (*Text_Value, bool) {
	return nil, false
}

// SetText agrees with the getter but not with the schema's scalar form.
func (*PointerText) SetText(*Text_Value) {}

// AsChild deliberately returns an object value.
func (ValueChild) AsChild() (Detail, bool) {
	return Detail{}, false
}

// SetChild agrees with the getter but not with the schema's object form.
func (*ValueChild) SetChild(Detail) {}
