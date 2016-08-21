# grpc
This branch adds gRPC support to goa.

This is a work in progress...

## TODO

### Phase 1: move HTTP specific code to rest package

- [ ] Build new goa core design and dsl packages
- [ ] Build REST support leveraging new core packages
- [ ] Break out security into its own plugin
- [ ] Generalize middleware
- [ ] Generalize error handling
- [ ] Port generators

### Phase 2: implement gRPC support via grpc package

- [ ] Implement DSL
- [ ] Generate protobuf file
- [ ] Invoke protoc
- [ ] Generate code that integrates with protoc output
- [ ] Update goa libraries to add any necessary support

### Phase 3: go-kit plugin

- [ ] Generate go-kit gRPC transport
- [ ] Generate go-kit HTTP transport
- [ ] Generate go-kit endpoints
- [ ] Generate service interface
- [ ] Generate scaffolding main

