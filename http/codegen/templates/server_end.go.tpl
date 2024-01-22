
	// Start HTTP server using default configuration, change the code to
	// configure the server as required by your service.
	srv := &http.Server{Addr: u.Host, Handler: handler, ReadHeaderTimeout: time.Second * 60}

	{{- range .Services }}
		for _, m := range {{ .Service.VarName }}Server.Mounts {
			logger.Printf("HTTP %q mounted on %s %s", m.Method, m.Verb, m.Pattern)
		}
	{{- end }}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		{{ comment "Start HTTP server in a separate goroutine." }}
		go func() {
			logger.Printf("HTTP server listening on %q", u.Host)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		logger.Printf("shutting down HTTP server at %q", u.Host)

		{{ comment "Shutdown gracefully with a 30s timeout." }}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			logger.Printf("failed to shutdown: %v", err)
		}
	}()
}
