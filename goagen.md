# goagen

`goagen` is the tool that generates the various artifacts from the user provided metadata.

Each type of artifact is associated with a command that exposes it own set of flags. Internally
these commands map to "generators" that contain the logic for generating the artifacts.

`goagen` works something like this:
* It first builds a temporary executable that contains the DSL. This temporary executable also
  contains code that runs the DSL and invokes the target generator with the resulting data
  structures.
* `goagen` then runs that temporary executable and reports any error or the list of generated files
  in case of success.
* Finally, `goagen` cleans up the temporary files.

Each generator registers a sub-command with the `goagen` tool, `goagen --help` lists all the
available commands. At the time of writing these are:
* `app`: generates the application code including controllers, contexts, media types and user types.
* `main`: generates a skeleton file for each resource controller as well as a default `main`.
* `schema`: generates the API Hyper-schema JSON.
The command `goa --help-long` lists all the supported commands and their flags.

The default command - also run when no command is provided - invokes each generator one by one.

## Common flags

The following flags apply to all the `goagen` sub-commands:

* `--design|-d=DESIGN` sets the Go package path to the application design package (the package containing
  the application DSL).
* `--out|-o=OUT` specifies where to generate the files, defaults to the current directory.
* `--debug` enables `goagen` debug. This causes `goagen` to print the content of the temporary
  files and to leave them around.
* `--help|--help-long|--help-man` prints contextual help.

## `goagen app`

The `app` subcommand is probably the most critical. It generates all the supporting code for the
goa application. This command suports an additional flag:
* `--pkg=app` specifies the name of the generated Go package, defaults to `app`. That's also the
  name of the sub-directory that gets created to store the generated Go files.
This command always deletes and re-creates any pre-existing directory with the same name. The idea
being that these files should never be edited.

## `goagen main`

The `main` subcommand helps bootstrap a new `goa` application by generating a default `main.go` as
well as a default (empty) implementation for each resource controller defined in the DSL. By default
this command only generates the files if they don't exist yet in the output directory. This
subcommand accepts one additional flag:
* `--force` causes the files to be generated even if files with the same name already exist (in
  which case they get overwritten).

## `goagen schema`

The `schema` subcommand generates a JSON hyper-schema representation of the API a la Heroku.
* `--url|-u=URL` specifies the base URL used to build the JSON schema ID.
