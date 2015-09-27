## Generic generator
Add `--genPkg [PACKAGE_PATH]` and `--genFac [FACTORY]` command line options to
goagen.
Upon execution goagen calls the FACTORY global function implemented in
PACKAGE_PATH. It calls `Generate` on the result passing in the metadata.
