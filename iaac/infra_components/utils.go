package infracomponents

// merge copies key-value pairs from src to dest.
func merge(dest, src map[string]string) {
	for k, v := range src {
		dest[k] = v
	}
}
