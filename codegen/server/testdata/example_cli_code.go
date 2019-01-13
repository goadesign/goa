package testdata

const (
	NoServerCLIMainCode = `func main() {
	var (
		hostF = flag.String("host", "localhost", "Server host (valid values: localhost)")
		addrF = flag.String("url", "", "URL to service host")

		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF       = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
	var (
		addr    string
		timeout int
		debug   bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
			case "localhost":
				addr = "http://localhost:80"
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: localhost)", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host   string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
	var (
		endpoint goa.Endpoint
		payload  interface{}
		err      error
	)
	{
		switch scheme {
		case "http", "https":
			endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: grpc|http)", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the test api API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -host HOST:  server host (localhost). valid values: localhost
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent(httpUsageCommands()), os.Args[0], indent(httpUsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`

	SingleServerSingleHostCLIMainCode = `func main() {
	var (
		hostF = flag.String("host", "dev", "Server host (valid values: dev)")
		addrF = flag.String("url", "", "URL to service host")

		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF       = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
	var (
		addr    string
		timeout int
		debug   bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
			case "dev":
				addr = "http://example:8090"
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: dev)", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host   string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
	var (
		endpoint goa.Endpoint
		payload  interface{}
		err      error
	)
	{
		switch scheme {
		case "http", "https":
			endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: http|https)", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the SingleServerSingleHost API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -host HOST:  server host (dev). valid values: dev
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent(httpUsageCommands()), os.Args[0], indent(httpUsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`

	SingleServerSingleHostWithVariablesCLIMainCode = `func main() {
	var (
		hostF = flag.String("host", "dev", "Server host (valid values: dev)")
		addrF = flag.String("url", "", "URL to service host")

		int_F     = flag.String("int", "1", "")
		uintF     = flag.String("uint", "1", "")
		float32_F = flag.String("float32", "1.1", "")
		int32_F   = flag.String("int32", "1", "")
		int64_F   = flag.String("int64", "1", "")
		uint32_F  = flag.String("uint32", "1", "")
		uint64_F  = flag.String("uint64", "1", "")
		float64_F = flag.String("float64", "1", "")
		boolF     = flag.String("bool", "true", "")
		verboseF  = flag.Bool("verbose", false, "Print request and response details")
		vF        = flag.Bool("v", false, "Print request and response details")
		timeoutF  = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
	var (
		addr    string
		timeout int
		debug   bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
			case "dev":
				addr = "http://example-{int}-{uint}-{float32}:8090"
				addr = strings.Replace(addr, "{int}", *int_F, -1)
				addr = strings.Replace(addr, "{uint}", *uintF, -1)
				addr = strings.Replace(addr, "{float32}", *float32_F, -1)
				addr = strings.Replace(addr, "{int32}", *int32_F, -1)
				addr = strings.Replace(addr, "{int64}", *int64_F, -1)
				addr = strings.Replace(addr, "{uint32}", *uint32_F, -1)
				addr = strings.Replace(addr, "{uint64}", *uint64_F, -1)
				addr = strings.Replace(addr, "{float64}", *float64_F, -1)
				addr = strings.Replace(addr, "{bool}", *boolF, -1)
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: dev)", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host   string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
	var (
		endpoint goa.Endpoint
		payload  interface{}
		err      error
	)
	{
		switch scheme {
		case "http", "https":
			endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: http|https)", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the SingleServerSingleHostWithVariables API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v][-int INT][-uint UINT][-float32 FLOAT32][-int32 INT32][-int64 INT64][-uint32 UINT32][-uint64 UINT64][-float64 FLOAT64][-bool BOOL] SERVICE ENDPOINT [flags]

    -host HOST:  server host (dev). valid values: dev
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)
    -int:     (1)
    -uint:     (1)
    -float32:     (1.1)
    -int32:     (1)
    -int64:     (1)
    -uint32:     (1)
    -uint64:     (1)
    -float64:     (1)
    -bool:     (true)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent(httpUsageCommands()), os.Args[0], indent(httpUsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`

	SingleServerMultipleHostsCLIMainCode = `func main() {
	var (
		hostF = flag.String("host", "dev", "Server host (valid values: dev, stage)")
		addrF = flag.String("url", "", "URL to service host")

		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF       = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
	var (
		addr    string
		timeout int
		debug   bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
			case "dev":
				addr = "http://example:8090"
			case "stage":
				addr = "https://example"
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: dev|stage)", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host   string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
	var (
		endpoint goa.Endpoint
		payload  interface{}
		err      error
	)
	{
		switch scheme {
		case "http", "https":
			endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: http|https)", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the SingleServerMultipleHosts API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v] SERVICE ENDPOINT [flags]

    -host HOST:  server host (dev). valid values: dev, stage
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent(httpUsageCommands()), os.Args[0], indent(httpUsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`

	SingleServerMultipleHostsWithVariablesCLIMainCode = `func main() {
	var (
		hostF = flag.String("host", "dev", "Server host (valid values: dev, stage)")
		addrF = flag.String("url", "", "URL to service host")

		versionF = flag.String("version", "v1", "Version")
		domainF  = flag.String("domain", "test", "Domain")
		portF    = flag.String("port", "8080", "Port")
		verboseF = flag.Bool("verbose", false, "Print request and response details")
		vF       = flag.Bool("v", false, "Print request and response details")
		timeoutF = flag.Int("timeout", 30, "Maximum number of seconds to wait for response")
	)
	flag.Usage = usage
	flag.Parse()
	var (
		addr    string
		timeout int
		debug   bool
	)
	{
		addr = *addrF
		if addr == "" {
			switch *hostF {
			case "dev":
				addr = "http://example-{version}:8090"
				var versionSeen bool
				{
					for _, v := range []string{"v1", "v2"} {
						if v == *versionF {
							versionSeen = true
							break
						}
					}
				}
				if !versionSeen {
					fmt.Fprintf(os.Stderr, "invalid value for URL 'version' variable: %q (valid values: v1,v2)", *versionF)
					os.Exit(1)
				}
				addr = strings.Replace(addr, "{version}", *versionF, -1)
			case "stage":
				addr = "https://example-{domain}:{port}"
				addr = strings.Replace(addr, "{domain}", *domainF, -1)
				addr = strings.Replace(addr, "{port}", *portF, -1)
			default:
				fmt.Fprintf(os.Stderr, "invalid host argument: %q (valid hosts: dev|stage)", *hostF)
				os.Exit(1)
			}
		}
		timeout = *timeoutF
		debug = *verboseF || *vF
	}

	var (
		scheme string
		host   string
	)
	{
		u, err := url.Parse(addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid URL %#v: %s", addr, err)
			os.Exit(1)
		}
		scheme = u.Scheme
		host = u.Host
	}
	var (
		endpoint goa.Endpoint
		payload  interface{}
		err      error
	)
	{
		switch scheme {
		case "http", "https":
			endpoint, payload, err = doHTTP(scheme, host, timeout, debug)
		default:
			fmt.Fprintf(os.Stderr, "invalid scheme: %q (valid schemes: http|https)", scheme)
			os.Exit(1)
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err.Error())
		fmt.Fprintln(os.Stderr, "run '"+os.Args[0]+" --help' for detailed usage.")
		os.Exit(1)
	}

	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil && !debug {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, ` + "`" + `%s is a command line client for the SingleServerMultipleHostsWithVariables API.

Usage:
    %s [-host HOST][-url URL][-timeout SECONDS][-verbose|-v][-version VERSION][-domain DOMAIN][-port PORT] SERVICE ENDPOINT [flags]

    -host HOST:  server host (dev). valid values: dev, stage
    -url URL:    specify service URL overriding host URL (http://localhost:8080)
    -timeout:    maximum number of seconds to wait for response (30)
    -verbose|-v: print request and response details (false)
    -version:    Version (v1)
    -domain:    Domain (test)
    -port:    Port (8080)

Commands:
%s
Additional help:
    %s SERVICE [ENDPOINT] --help

Example:
%s
` + "`" + `, os.Args[0], os.Args[0], indent(httpUsageCommands()), os.Args[0], indent(httpUsageExamples()))
}

func indent(s string) string {
	if s == "" {
		return ""
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}
`
)
