package bootstrap

import "fmt"

// TBDCommand contains the implementation for commands that are TBD.
type TBDCommand struct {
	*BaseCommand

	// Desc describes the intent of the future command.
	Desc string

	// Example contains an output example.
	Example string
}

// NewTBDCommand returns a non implemented command using the given description and example.
func NewTBDCommand(desc, example string) *TBDCommand {
	b := BaseCommand{Factory: "unused"}
	return &TBDCommand{BaseCommand: &b, Desc: desc, Example: example}
}

// Description returns the command description.
func (t *TBDCommand) Description() string { return t.Desc }

// Run overrides the base command Run to simply print the description and example for a not
// implemented yet command.
func (t *TBDCommand) Run() ([]string, error) {
	fmt.Println("Work in progress: this command is not implemented yet.")
	fmt.Println("If if was it would:")
	fmt.Println(t.Desc)
	if len(t.Example) > 0 {
		fmt.Println("\nExample Output:")
		fmt.Println(t.Example)
	}
	return nil, nil
}
