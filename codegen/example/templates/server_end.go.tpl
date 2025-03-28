

    {{ comment "Wait for signal." }}
	log.Printf(ctx, "exiting (%v)", <-errc)

	{{ comment "Send cancellation signal to the goroutines." }}
	cancel()

	wg.Wait()
	log.Printf(ctx, "exited")
}
