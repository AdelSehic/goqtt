package sliceiterator

import "fmt"

type SliceIter[T any] struct {
	iterable []T
	max      int
	index    int
}

func NewIterator[T any](iterable []T) *SliceIter[T] {
	return &SliceIter[T]{
		iterable: iterable,
		max:      len(iterable),
		index:    0,
	}
}

func (it *SliceIter[T]) Next() *SliceIter[T] {
	it.index++
	return it
}

func (it *SliceIter[T]) Prev() *SliceIter[T] {
	it.index--
	return it
}

func (it *SliceIter[T]) Valid() bool {
	fmt.Printf("IS IT VALID? %d < %d", it.index, it.max)
	return it.index < it.max
}

func (it *SliceIter[T]) Value() T {
	return it.iterable[it.index]
}

func (it *SliceIter[T]) IsLast() bool {
	return it.index == it.max
}
