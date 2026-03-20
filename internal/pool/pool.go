package internal

import "sync"

type Reseter[T any] interface {
	*T
	Reset()
}

type Pool[T any, R Reseter[T]] struct {
	pool *sync.Pool
}

func New[T any, R Reseter[T]]() *Pool[T, R] {
	return &Pool[T, R]{
		pool: &sync.Pool{
			New: func() any { return new(T) },
		},
	}
}

func (p *Pool[T, R]) Get() *T {
	return p.pool.Get().(*T)
}

func (p *Pool[T, R]) Put(st *T) {
	R(st).Reset()
	p.pool.Put(st)
}
