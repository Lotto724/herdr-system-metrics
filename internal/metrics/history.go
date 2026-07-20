package metrics

type history[T any] struct {
	values []T
	start  int
	count  int
}

func newHistory[T any](capacity int) history[T] {
	if capacity < 0 {
		capacity = 0
	}
	return history[T]{values: make([]T, capacity)}
}

func (h *history[T]) add(value T) {
	if len(h.values) == 0 {
		return
	}
	if h.count == len(h.values) {
		h.values[h.start] = value
		h.start = (h.start + 1) % len(h.values)
		return
	}
	h.values[(h.start+h.count)%len(h.values)] = value
	h.count++
}

func (h history[T]) all() []T {
	values := make([]T, h.count)
	for i := range values {
		values[i] = h.values[(h.start+i)%len(h.values)]
	}
	return values
}
