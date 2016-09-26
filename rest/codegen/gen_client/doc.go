/*
Package genclient provides a generator for the client tool and package of a goa application.
The generator creates a main.go file and a subpackage containing data structures specific to the
service.

The generated code includes a client package with:

    * One client method per resource action
    * Helper functions to build the corresponding request paths
    * Structs for the action payloads and dependent types
    * Structs for the action media types and corresponding decoder functions

The generated code also includes a CLI tool with commands for each action and sub-commands for
each resource.
*/
package genclient
