/*
Package genjs provides a goa generator for a javascript client module.
The module exposes functions for calling the API actions. It relies on the
axios (https://github.com/mzabriskie/axios) javascript library to perform the actual HTTP requests.

The generator also produces an example controller and index HTML that shows how to use the module.
The controller simply serves all the files under the "js" directory so that loading "/js" in a
browser triggers the example code.
*/
package genjs
