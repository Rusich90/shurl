package testdata

import (
	"testing"
)

func TestResetableStruct_Reset(t *testing.T) {
	// Создаем строку, на которую будем ссылаться
	strVal := "hello world"

	// Создаем вложенный объект
	child := &ResetableStruct{
		i:     42,
		str:   "child string",
		strP:  &strVal,
		s:     []int{1, 2, 3, 4, 5},
		m:     map[string]string{"key": "value"},
		child: nil,
	}

	// Создаем основной объект со значениями
	rs := &ResetableStruct{
		i:     100,
		str:   "test string",
		strP:  &strVal,
		s:     []int{1, 2, 3, 4, 5},
		m:     map[string]string{"key1": "value1", "key2": "value2"},
		child: child,
	}

	// Проверяем начальные значения
	if rs.i != 100 {
		t.Errorf("expected i=100, got %d", rs.i)
	}
	if rs.str != "test string" {
		t.Errorf("expected str='test string', got '%s'", rs.str)
	}
	if rs.strP == nil || *rs.strP != "hello world" {
		t.Errorf("expected strP='hello world', got nil or wrong value")
	}
	if len(rs.s) != 5 {
		t.Errorf("expected len(s)=5, got %d", len(rs.s))
	}
	if len(rs.m) != 2 {
		t.Errorf("expected len(m)=2, got %d", len(rs.m))
	}
	if rs.child == nil || rs.child.i != 42 {
		t.Errorf("expected child.i=42, got nil or wrong value")
	}

	// Вызываем Reset()
	rs.Reset()

	// Проверяем сброс значений
	if rs.i != 0 {
		t.Errorf("expected i=0 after reset, got %d", rs.i)
	}
	if rs.str != "" {
		t.Errorf("expected str='', got '%s'", rs.str)
	}
	if rs.strP == nil {
		t.Errorf("expected strP != nil after reset")
	} else if *rs.strP != "" {
		t.Errorf("expected *strP='', got '%s'", *rs.strP)
	}
	if len(rs.s) != 0 {
		t.Errorf("expected len(s)=0 after reset, got %d", len(rs.s))
	}
	if len(rs.m) != 0 {
		t.Errorf("expected len(m)=0 after reset, got %d", len(rs.m))
	}
	if rs.child == nil {
		t.Errorf("expected child != nil after reset")
	} else if rs.child.i != 0 {
		t.Errorf("expected child.i=0 after reset, got %d", rs.child.i)
	}
}

func TestResetableStruct2_Reset(t *testing.T) {
	// Создаем вложенный объект
	child := ResetableStruct{
		i:     42,
		str:   "child string",
		strP:  nil,
		s:     []int{1, 2, 3, 4, 5},
		m:     map[string]string{"key": "value"},
		child: nil,
	}

	// Создаем основной объект со значениями
	rs := &ResetableStruct2{
		i:     100,
		str:   "test string",
		child: child,
	}

	// Проверяем начальные значения
	if rs.i != 100 {
		t.Errorf("expected i=100, got %d", rs.i)
	}
	if rs.str != "test string" {
		t.Errorf("expected str='test string', got '%s'", rs.str)
	}
	if rs.child.i != 42 {
		t.Errorf("expected child.i=42, got %d", rs.child.i)
	}

	// Вызываем Reset()
	rs.Reset()

	// Проверяем сброс значений
	if rs.i != 0 {
		t.Errorf("expected i=0 after reset, got %d", rs.i)
	}
	if rs.str != "" {
		t.Errorf("expected str='', got '%s'", rs.str)
	}
	if rs.child.i != 0 {
		t.Errorf("expected child.i=0 after reset, got %d", rs.child.i)
	}
}

func TestNilReset(t *testing.T) {
	// Проверяем, что вызов Reset() на nil не паникует
	var rs *ResetableStruct
	rs.Reset() // не должно паниковать

	var rs2 *ResetableStruct2
	rs2.Reset() // не должно паниковать
}
