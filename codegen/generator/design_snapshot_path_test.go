// These tests keep recorded paths and errors independent of the reused walk buffer.
package generator

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreparedDesignPathBufferPreservesPaths(t *testing.T) {
	type nested struct {
		Value string
	}
	value := struct {
		First    string
		Nested   nested
		Array    [2]string
		Map      map[string]string
		Numbered map[int]string
		Last     string
	}{
		First:    "first",
		Nested:   nested{"nested"},
		Array:    [2]string{"one", "two"},
		Map:      map[string]string{"quoted\"key": "mapped"},
		Numbered: map[int]string{7: "numbered"},
		Last:     "last",
	}
	var paths []string
	walker := &designSnapshotter{
		visited: make(map[designVisit]struct{}),
		record: func(path []byte, _ reflect.Type, _ string) {
			paths = append(paths, string(path))
		},
	}
	require.NoError(t, walker.appendValue([]byte("root"), reflect.ValueOf(value)))
	require.Equal(t, []string{
		"root", "root.First", "root.Nested", "root.Nested.Value",
		"root.Array", "root.Array[0]", "root.Array[1]",
		"root.Map", "root.Map[\"quoted\\\"key\"].key", "root.Map[\"quoted\\\"key\"]",
		"root.Numbered", "root.Numbered{0}.key", "root.Numbered{0}", "root.Last",
	}, paths)
	require.Empty(t, walker.path)
	// A later walk overwrites the same buffer without changing earlier records.
	require.NoError(t, walker.appendValue(append(walker.path, "replacement"...), reflect.ValueOf("other")))
	require.Equal(t, "root.First", paths[1])
	require.Equal(t, "root.Last", paths[len(paths)-2])
}

func TestPreparedDesignPathBufferPreservesDeepError(t *testing.T) {
	type node struct {
		Next  *node
		Value any
	}
	root := &node{}
	leaf := root
	for range 32 {
		leaf.Next = &node{}
		leaf = leaf.Next
	}
	leaf.Value = make(chan int)
	walker := &designSnapshotter{
		visited: make(map[designVisit]struct{}),
		record: func([]byte, reflect.Type, string) {
		},
	}
	err := walker.appendValue([]byte("root"), reflect.ValueOf(root))
	require.EqualError(t, err, "snapshot prepared design at root"+strings.Repeat(".Next", 32)+".Value: unsupported non-nil channel chan int")
	require.Empty(t, walker.path)
	require.NoError(t, walker.appendValue(append(walker.path, "replacement"...), reflect.ValueOf("other")))
	require.EqualError(t, err, "snapshot prepared design at root"+strings.Repeat(".Next", 32)+".Value: unsupported non-nil channel chan int")
}
