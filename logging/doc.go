/*
Package logging contains logger adapters that make it possible for goa to log messages to various
logger backends. Each adapter exists in its own sub-package named after the corresponding logger
package.

Once instantiated adapters can be used by setting the goa service logger with WithLogger:

```go
  func main() {
    // ...

    // Setup logger adapter
    logger := log15.New()

    // Create service
    service := goa.New("my service")
    service.WithLogger(goalog15.New(logger))

    // ...
}
```

See http://goa.design/implement/logging/ for details.
*/
package logging
