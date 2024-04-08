package tokencache

type CacheMissError struct{}

func (e *CacheMissError) Error() string {
	return "no cache entry found"
}
