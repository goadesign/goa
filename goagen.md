# goagen

`goagen` is the tool that generates the various artifacts from the user provided design package.

Each type of artifact is associated with a tool command that exposes it own set of flags.
Internally these commands map to "generators" that contain the logic for generating the artifacts.

`goagen` works something like this:
* It first builds a temporary executable that contains the user design package. This temporary
  executable also contains code that runs the design package DSL and invokes the target generator
  with the resulting data structures.
* `goagen` then runs that temporary executable and reports any error. If the temporary executable
  succeeds `goagen` prints the list of generated files.
* Finally, `goagen` cleans up the temporary files.

Each generator registers a command with the `goagen` tool, `goagen --help` lists all the available
commands. At the time of writing these are:
* `app`: generates the application code including controllers, contexts, media types and user types.
* `main`: generates a skeleton file for each resource controller as well as a default `main`.
* `schema`: generates the API Hyper-schema JSON.
* `gen`: invokes a third party generator.

The command `goagen --help-long` lists all the supported commands and their flags.

The `bootstrap` command invokes the `app`, `main` and `schema` generators in this order.

## Common flags

The following flags apply to all the `goagen` commands:

* `--design|-d=DESIGN` defines the Go package path to the application design package (the package
  containing the application DSL).
* `--out|-o=OUT` specifies where to generate the files, defaults to the current directory.
* `--debug` enables `goagen` debug. This causes `goagen` to print the content of the temporary
  files and to leave them around.
* `--help|--help-long|--help-man` prints contextual help.

## `goagen app`

The `app` command is probably the most critical. It generates all the supporting code for the
goa application. This command supports an additional flag:
* `--pkg=app` specifies the name of the generated Go package, defaults to `app`. That's also the
  name of the subdirectory that gets created to store the generated Go files.
This command always deletes and re-creates any pre-existing directory with the same name. The idea
being that these files should never be edited.

## `goagen main`

The `main` command helps bootstrap a new `goa` application by generating a default `main.go` as
well as a default (empty) implementation for each resource controller defined in the DSL. By default
this command only generates the files if they don't exist yet in the output directory. This
command accepts two additional flags:
* `--force` causes the files to be generated even if files with the same name already exist (in
  which case they get overwritten).
* `--name=API` specifies the name of the application to be used in the generated call to `goa.New`.
  It defaults to "API".

## `goagen schema`

The `schema` command generates a [Heroku-like](https://blog.heroku.com/archives/2014/1/8/json_schema_for_heroku_platform_api)
JSON hyper-schema representation of the API. The command accepts an additional flag:
* `--url|-u=URL` specifies the base URL used to build the JSON schema ID.

## `goagen gen`

The `gen` command makes it possible to invoke third party generators. See the [gengen package
documentation](https://godoc.org/github.com/raphael/goa/goagen/gen_gen) for information on how to
write third party generators. This command accepts two flags:
* `--pkg-path=PKG-PATH` specifies the Go package path to the generator package.
* `--pkg-name=PKG-NAME` specifies the Go package name of the generator package. It defaults to the
   name of the inner most directory in the Go package path.
