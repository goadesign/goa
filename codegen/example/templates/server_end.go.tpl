

    {{ comment "Wait for signal." }}
	logger.Printf("exiting (%v)", <-errc)

	{{ comment "Send cancellation signal to the goroutines." }}
	cancel()

	wg.Wait()
	logger.Println("exited")
}