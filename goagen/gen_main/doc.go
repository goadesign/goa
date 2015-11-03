/*
Package genmain provides a generator for a skeleton goa application.
This generator generates the code for a basic "main" package and is mainly intended as a way to
bootstrap new applications.
The generator creates a main.go file and one file per resource listed in the API metadata.
If a file already exists it skips its creation unless the flag --force is provided on the command
line in which case it overrides the content of existing files.
*/
package genmain
