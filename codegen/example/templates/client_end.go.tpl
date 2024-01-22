

	data, err := endpoint(context.Background(), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	if data != nil {
		m, _ := json.MarshalIndent(data, "", "    ")
		fmt.Println(string(m))
	}
}