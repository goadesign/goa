<p align="center">
  <p align="center">
    <a href="https://goa.design">
      <img alt="Goa" src="https://goa.design/img/social/goa-banner.png">
    </a>
  </p>
  <p align="center">
    <a href="https://github.com/goadesign/goa/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/goadesign/goa.svg?style=for-the-badge"></a>
    <a href="https://pkg.go.dev/goa.design/goa/v3@v3.26.0/dsl?tab=doc"><img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge"></a>
    <a href="https://github.com/goadesign/goa/actions/workflows/ci.yml"><img alt="GitHub Action: Test" src="https://img.shields.io/github/actions/workflow/status/goadesign/goa/test.yml?branch=v3&style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/goadesign/goa"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/goadesign/goa?style=for-the-badge"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge"></a>
    <a href="https://gurubase.io/g/goa"><img alt="Gurubase" src="https://img.shields.io/badge/Gurubase-Ask%20Goa%20Guru-006BFF?style=for-the-badge"></a>
    <a href="https://chat.openai.com/g/g-mLuQDGyro-goa-design-wizard"><img alt="Goa Design Wizard" src="https://img.shields.io/badge/Goa%20Design%20Wizard-ChatGPT-00A67D?logo=openai&logoColor=white&style=for-the-badge"></a>
    </br>
    <a href="https://goadesign.substack.com"><img alt="Substack: Design First" src="https://img.shields.io/badge/Design%20First-Substack-FF6719?logo=substack&logoColor=white&style=for-the-badge"></a>
    <a href="https://gophers.slack.com/messages/goa"><img alt="Slack: Goa" src="https://img.shields.io/badge/Goa-Slack-4A154B?logo=slack&logoColor=white&style=for-the-badge"></a>
    <a href="https://bsky.app/profile/goadesign.bsky.social"><img alt="Bluesky: Goa Design" src="https://img.shields.io/badge/Goa%20Design-Bluesky-0285FF?logo=bluesky&logoColor=white&style=for-the-badge"></a>
  </p>
</p>

<!-- Removed card; Wizard is now a badge button above -->

# Goa - Design First, Code With Confidence

## Overview

Goa transforms how you build APIs and microservices in Go with its powerful design-first approach. Instead of writing boilerplate code, you express your API's intent through a clear, expressive DSL. Goa then automatically generates production-ready code, comprehensive documentation, and client libraries—all perfectly aligned with your design.

The result? Dramatically reduced development time, consistent APIs, and the elimination of the documentation-code drift that plagues traditional development.

## Sponsors

<table width="100%">
    <tr>
        <td>
            <img width="1000" height="0" />
            <a href="https://www.incident.io">
                <img src="docs/incidentio.png" alt="incident.io" width="260" align="right" />
            </a>
            <h3>incident.io: Bounce back stronger after every incident</h3>
            <p>
                Use our platform to empower your team to run incidents end-to-end. Rapidly fix and
                learn from incidents, so you can build more resilient products.
            </p>
            <a href="https://incident.io">Learn more</a>
        </td>
    </tr>
    <tr>
        <td>
            <img width="1000" height="0" />
            <a href="https://www.speakeasy.com/editor?utm_source=goa+repo&utm_medium=github+sponsorship">
                <img src="docs/speakeasy.png" alt="Speakeasy" width="260" align="right" />
            </a>
            <h3>Speakeasy: Enterprise DevEx for your API</h3>
            <p>
                Our platform makes it easy to create feature-rich production ready SDKs.
                Speed up integrations and reduce errors by giving your API the DevEx it deserves.
            </p>
            <a href="https://www.speakeasy.com/docs/api-frameworks/goa?utm_source=goa+repo&utm_medium=github+sponsorship">Integrate with Goa</a>
        </td>
    </tr>
</table>

## Why Goa?

Traditional API development suffers from:
- **Inconsistency**: Manually maintained docs that quickly fall out of sync with code
- **Wasted effort**: Writing repetitive boilerplate and transport-layer code
- **Painful integrations**: Client packages that need constant updates
- **Design afterthoughts**: Documentation added after implementation, missing key details

Goa solves these problems by:
- Generating 30-50% of your codebase directly from your design
- Ensuring perfect alignment between design, code, and documentation
- Supporting multiple transports (HTTP, gRPC, and JSON-RPC) from a single design
- Maintaining a clean separation between business logic and transport details

## Key Features

- **Expressive Design Language**: Define your API with a clear, type-safe DSL that captures your intent
- **Comprehensive Code Generation**:
  - Type-safe server interfaces that enforce your design
  - Client packages with full error handling
  - Transport layer adapters (HTTP/gRPC/JSON-RPC) with routing and encoding
  - OpenAPI/Swagger documentation that's always in sync
  - CLI tools for testing your services
- **Multi-Protocol Support**: Generate HTTP REST, gRPC, and JSON-RPC endpoints from a single design
- **Clean Architecture**: Business logic remains separate from transport concerns
- **Enterprise Ready**: Supports authentication, authorization, CORS, logging, and more
- **Comprehensive Testing**: Includes extensive unit and integration test suites ensuring quality and reliability

## How It Works

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────────┐
│ Design API  │────>│ Generate Code│────>│ Implement Business  │
│ using DSL   │     │ & Docs       │     │ Logic               │
└─────────────┘     └──────────────┘     └─────────────────────┘
```

1. **Design**: Express your API's intent in Goa's DSL
2. **Generate**: Run `goa gen` to create server interfaces, client code, and documentation
3. **Implement**: Focus solely on writing your business logic in the generated interfaces
4. **Evolve**: Update your design and regenerate code as your API evolves

## Quick Start

```bash
# Install Goa
go install goa.design/goa/v3/cmd/goa@latest

# Create a new module
mkdir hello && cd hello
go mod init hello

# Define a service in design/design.go
mkdir design
cat > design/design.go << EOF
package design

import . "goa.design/goa/v3/dsl"

var _ = Service("hello", func() {
    Method("say_hello", func() {
        Payload(func() {
            Field(1, "name", String)
            Required("name")
        })
        Result(String)

        HTTP(func() {
            GET("/hello/{name}")
        })
    })
})
EOF

# Generate the code
goa gen hello/design
goa example hello/design

# Build and run
go mod tidy
go run cmd/hello/*.go --http-port 8000

# In another terminal
curl http://localhost:8000/hello/world
```

The example above:
1. Defines a simple "hello" service with one method
2. Generates server and client code
3. Starts a server that logs requests server-side (without displaying any client output)

### JSON-RPC Alternative

For a JSON-RPC service, simply add a `JSONRPC` expression to the service and
method:

```go
var _ = Service("hello" , func() {
    JSONRPC(func() {
        Path("/jsonrpc")
    })
    Method("say_hello", func() {
        Payload(func() {
            Field(1, "name", String)
            Required("name")
        })
        Result(String)

        JSONRPC(func() {})
    })
}
```

Then test with:
```bash
curl -X POST http://localhost:8000/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"hello.say_hello","params":{"name":"world"},"id":"1"}'
```

## Documentation

Our documentation site at [goa.design](https://goa.design) provides comprehensive guides and references:

- **[Introduction](https://goa.design/docs/1-introduction/)**: Understand Goa's philosophy and benefits
- **[Getting Started](https://goa.design/docs/2-getting-started/)**: Build your first Goa service step-by-step
- **[Tutorials](https://goa.design/docs/3-tutorials/)**: Learn to create REST APIs, gRPC services, and more
- **[Core Concepts](https://goa.design/docs/4-concepts/)**: Master the design language and architecture
- **[Real-World Guide](https://goa.design/docs/5-real-world/)**: Follow best practices for production services
- **[Advanced Topics](https://goa.design/docs/6-advanced/)**: Explore advanced features and techniques

##  Real-World Examples

The [examples repository](https://github.com/goadesign/examples) contains complete, working examples demonstrating:

- **Basic**: Simple service showcasing core Goa concepts
- **Cellar**: A more complete REST API example
- **Cookies**: HTTP cookie management
- **Encodings**: Working with different content types
- **Error**: Comprehensive error handling strategies
- **Files & Upload/Download**: File handling capabilities
- **HTTP Status**: Custom status code handling
- **Interceptors**: Request/response processing middleware
- **Multipart**: Handling multipart form submissions
- **Security**: Authentication and authorization examples
- **Streaming**: Implementing streaming endpoints (HTTP, WebSocket, JSON-RPC SSE)
- **Tracing**: Integrating with observability tools
- **TUS**: Resumable file uploads implementation

## Community & Support

- Join the [#goa](https://gophers.slack.com/messages/goa/) channel on Gophers Slack
- Ask questions on [GitHub Discussions](https://github.com/goadesign/goa/discussions)
- Follow us on [Bluesky](https://goadesign.bsky.social)
- Report issues on [GitHub](https://github.com/goadesign/goa/issues)
- Find answers with the [Goa Guru](https://gurubase.io/g/goa) AI assistant
- Subscribe to our Substack, “Design First”: [Design First](https://goadesign.substack.com/subscribe?params=%5Bobject%20Object%5D)

## License

MIT License - see [LICENSE](LICENSE) for details.
