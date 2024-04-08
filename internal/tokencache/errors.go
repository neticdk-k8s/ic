package tokencache

// CacheMissError is an error used on cache misses
type CacheMissError struct{}

func (e *CacheMissError) Error() string {
	return "no cache entry found"
}
