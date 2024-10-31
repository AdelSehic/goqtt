package sliceiterator

type sliceIter[T any] struct {
	iterable []T
	max      int
	index    int
}

func NewIterator[T any](iterable []T) *sliceIter[T] {
	return &sliceIter[T]{
		iterable: iterable,
		max:      len(iterable),
		index:    0,
	}
}

func (it *sliceIter[T]) Next() *sliceIter[T] {
	it.index++
	if it.index == it.max {
		return nil
	}
	return it
}

func (it *sliceIter[T]) Prev() *sliceIter[T] {
	if it.index == 0 {
		return nil
	}
	it.index--
	return it
}

func (it *sliceIter[T]) Valid() bool {
	return it.index < it.max
}

func (it *sliceIter[T]) Value() T {
	return it.iterable[it.index]
}
