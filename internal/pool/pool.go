package pool

import "sync"

// Resetter определяет интерфейс для объектов, которые могут быть сброшены в исходное состояние.
type Resetter interface {
	Reset()
}

// Pool — это контейнер для повторного использования объектов одного типа.
// T должен реализовывать интерфейс Resetter.
type Pool[T Resetter] struct {
	pool *sync.Pool
}

// New создаёт и возвращает новый экземпляр Pool.
// init — функция-фабрика, которая создаёт новый объект типа T при необходимости.
func New[T Resetter](init func() T) *Pool[T] {
	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return init()
			},
		},
	}
}

// Get извлекает объект из пула. Если пул пуст, создаётся новый объект через функцию инициализации.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put возвращает объект в пул для повторного использования. Перед возвратом объект должен быть сброшен через метод Reset().
func (p *Pool[T]) Put(obj T) {
	p.pool.Put(obj)
}
