package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"goa.design/goa.v2/codegen"
	"goa.design/goa.v2/codegen/generators/client"
	"goa.design/goa.v2/codegen/generators/openapi"
	"goa.design/goa.v2/codegen/generators/server"
	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/rest/design"
)

func main() {
	var writers []codegen.FileWriter
	{
		writers = append(writers, server.Writers(design.Root, rest.Root)...)
		writers = append(writers, client.Writers(design.Root, rest.Root)...)
		writers = append(writers, openapi.Writers(design.Root, rest.Root)...)
	}
	outputs := make([]string, len(writers))
	for i, w := range writers {
		if err := codegen.Render(w); err != nil {
			fail(err)
		}
		outputs[i] = w.OutputPath()
	}
	sort.Strings(outputs)
	fmt.Println(strings.Join(outputs, "\n"))
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, err.Error())
	os.Exit(1)
}
