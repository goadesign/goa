

{{ comment "Setup logger. Replace logger with your own log package of choice." }}
	var (
		logger *log.Logger
	)
	{
		logger = log.New(os.Stderr, "[{{ .APIPkg }}] ", log.Ltime)
	}