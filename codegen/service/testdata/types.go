package testdata

import "time"

type StringT struct {
	String string
}

type StringPointerT struct {
	String *string
}

type ExternalNameT struct {
	String string
}

type ExternalNamePointerT struct {
	String *string
}

type ArrayStringT struct {
	ArrayString []string
}

type ObjectT struct {
	Object *ObjectFieldT
}

type ObjectExtraT struct {
	Object *ObjectFieldT
	t      *time.Time
}

type ObjectFieldT struct {
	Bool    bool
	Int     int
	Int32   int32
	Int64   int64
	UInt    uint
	UInt32  uint32
	UInt64  uint64
	Float32 float32
	Float64 float64
	Bytes   []byte
	String  string
	Array   []bool
	Map     map[string]bool
}
