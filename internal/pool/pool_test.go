package pool

import (
	"testing"
)

// testObject — тестовая реализация Resetter для проверки Pool.
type testObject struct {
	id    int
	used  bool
	calls int
}

func (o *testObject) Reset() {
	o.used = false
	o.calls++
}

func TestPool_GetAndPut(t *testing.T) {
	var createdCount int
	p := New(func() *testObject {
		createdCount++
		return &testObject{id: createdCount}
	})

	// Первый Get должен создать новый объект
	obj1 := p.Get()
	if obj1 == nil {
		t.Fatal("Get() вернул nil")
	}
	if obj1.id != 1 {
		t.Errorf("Ожидаем id=1, получили id=%d", obj1.id)
	}
	if obj1.used != false {
		t.Errorf("Ожидаем used=false, получили used=%v", obj1.used)
	}
	if obj1.calls != 0 {
		t.Errorf("Ожидаем calls=0 (Reset не должен вызываться при создании), получили calls=%d", obj1.calls)
	}
	if createdCount != 1 {
		t.Errorf("Ожидаем createdCount=1, получили createdCount=%d", createdCount)
	}

	// Перед Put пользователь должен вызвать Reset()
	obj1.Reset()

	// Возвращаем объект в пул
	p.Put(obj1)

	// Следующий Get может вернуть тот же объект (но это не гарантируется sync.Pool)
	obj2 := p.Get()
	if obj2.id != 1 {
		t.Errorf("Ожидаем id=1, получили id=%d", obj2.id)
	}
	if obj2.used != false {
		t.Errorf("Ожидаем used=false после Get, получили used=%v", obj2.used)
	}
	// Reset вызван пользователем перед Put
	if obj2.calls != 1 {
		t.Errorf("Ожидаем calls=1 (Reset вызван пользователем), получили calls=%d", obj2.calls)
	}
}

func TestPool_MultipleObjects(t *testing.T) {
	var createdCount int
	p := New(func() *testObject {
		createdCount++
		return &testObject{id: createdCount}
	})

	// Получаем и возвращаем несколько объектов
	obj1 := p.Get()
	obj2 := p.Get()
	obj3 := p.Get()

	if createdCount != 3 {
		t.Errorf("Ожидаем createdCount=3, получили createdCount=%d", createdCount)
	}

	// Перед возвратом вызываем Reset()
	obj1.Reset()
	obj2.Reset()
	obj3.Reset()

	// Возвращаем в обратном порядке (LIFO)
	p.Put(obj3)
	p.Put(obj2)
	p.Put(obj1)

	// Получаем снова — sync.Pool возвращает объекты в порядке LIFO
	obj4 := p.Get()
	obj5 := p.Get()
	obj6 := p.Get()

	// Проверяем, что все объекты были повторно использованы
	// (sync.Pool может вернуть их в порядке LIFO)
	ids := map[int]bool{obj4.id: true, obj5.id: true, obj6.id: true}
	if !ids[1] || !ids[2] || !ids[3] {
		t.Errorf("Ожидаем все три объекта (1, 2, 3), получили %d, %d, %d", obj4.id, obj5.id, obj6.id)
	}

	// Проверяем, что Reset был вызван для каждого Put
	if obj4.calls != 1 || obj5.calls != 1 || obj6.calls != 1 {
		t.Errorf("Ожидаем calls=1 для каждого объекта, получили calls=%d,%d,%d", obj4.calls, obj5.calls, obj6.calls)
	}
}

func TestPool_ResetCalledByUser(t *testing.T) {
	p := New(func() *testObject {
		return &testObject{}
	})

	obj := p.Get()
	if obj.calls != 0 {
		t.Errorf("Ожидаем calls=0 до Put, получили calls=%d", obj.calls)
	}

	// Пользователь должен вызвать Reset перед Put
	obj.Reset()

	p.Put(obj)

	if obj.calls != 1 {
		t.Errorf("Ожидаем calls=1 после Reset (вызван пользователем), получили calls=%d", obj.calls)
	}

	// Получаем снова
	obj2 := p.Get()
	if obj2.calls != 1 {
		t.Errorf("Ожидаем calls=1 (Reset не должен вызываться при Get), получили calls=%d", obj2.calls)
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	var createdCount int
	p := New(func() *testObject {
		createdCount++
		return &testObject{}
	})

	done := make(chan bool, 100)
	for i := 0; i < 100; i++ {
		go func() {
			obj := p.Get()
			obj.Reset() // Пользователь вызывает Reset
			p.Put(obj)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	// sync.Pool может создать больше одного объекта при конкурентном доступе
	if createdCount < 1 {
		t.Errorf("Ожидаем хотя бы 1 созданный объект, получили %d", createdCount)
	}
}

func TestPool_PoolEmptyCreatesNew(t *testing.T) {
	var createdCount int
	p := New(func() *testObject {
		createdCount++
		return &testObject{id: createdCount}
	})

	// Пул пуст, Get должен создать новый объект
	obj1 := p.Get()
	if createdCount != 1 {
		t.Errorf("Ожидаем createdCount=1, получили createdCount=%d", createdCount)
	}

	// Возвращаем объект
	obj1.Reset()
	p.Put(obj1)

	// Берем снова
	obj2 := p.Get()
	if obj2.id != 1 {
		t.Errorf("Ожидаем id=1 (повторное использование), получили id=%d", obj2.id)
	}
	if createdCount != 1 {
		t.Errorf("Ожидаем createdCount=1 (новый объект не должен создаваться), получили createdCount=%d", createdCount)
	}

	// Возвращаем и удаляем ссылку
	obj2.Reset()
	p.Put(obj2)
	obj2 = nil

	// Снова Get — sync.Pool может создать новый объект или вернуть старый
	obj3 := p.Get()
	if obj3.id != 1 {
		t.Errorf("Ожидаем id=1, получили id=%d", obj3.id)
	}
	// sync.Pool может уничтожить объект, если на него нет ссылок, и создать новый
	// или сохранить объект и вернуть его при следующем Get
	if createdCount != 1 {
		t.Errorf("Ожидаем createdCount=1, получили createdCount=%d", createdCount)
	}
}