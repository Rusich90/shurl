package batch

func GenerateChunks(items []string, chunkSize int) <-chan []string {
	inputCh := make(chan []string)

	go func() {
		defer close(inputCh)

		for i := 0; i < len(items); i += chunkSize {
			end := i + chunkSize
			if end > len(items) {
				end = len(items)
			}
			chunk := items[i:end]
			inputCh <- chunk
		}
	}()

	return inputCh
}
