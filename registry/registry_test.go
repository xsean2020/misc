package registry

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := New[int, *User](1)
	if r.Count() != 0 {
		t.Errorf("Expected 0 got %d", r.Count())
	}
}

type User struct {
	ID   int
	Name string
}

func (u *User) Primary() int {
	return u.ID
}

func TestRegisterAndQuery(t *testing.T) {
	r := New[int, *User](1)
	r.Register(&User{ID: 1, Name: "hello"})
	if r.Count() != 1 {
		t.Errorf("Expected 1, got %d", r.Count())
	}
	if v := r.Query(1); v == nil || v.Name != "hello" {
		t.Errorf("Expected 'hello', got '%v'", v)
	}
}

func TestUnregister(t *testing.T) {
	r := New[int, *User](1)
	u := &User{ID: 1, Name: "hello"}
	r.Register(u)
	if r.Count() != 1 {
		t.Errorf("Expected 1, got %d", r.Count())
	}
	r.Unregister(u)
	if r.Count() != 0 {
		t.Errorf("Expected 0, got %d", r.Count())
	}
}

func TestAll(t *testing.T) {
	r := New[int, *User](3)
	r.Register(&User{ID: 1, Name: "hello"})
	r.Register(&User{ID: 2, Name: "world"})
	r.Register(&User{ID: 3, Name: "!"})
	if r.Count() != 3 {
		t.Errorf("Expected 3, got %d", r.Count())
	}
	all := r.All()
	if len(all) != 3 {
		t.Errorf("Expected 3, got %d", len(all))
	}

	if all[0] != 1 || all[1] != 2 || all[2] != 3 {
		t.Errorf("Expected 1,2,3, got %d, %d, %d", all[0], all[1], all[2])
	}
}

func TestQueryNotExist(t *testing.T) {
	r := New[int, *User](1)
	v := r.Query(1)
	if v != nil {
		t.Errorf("Expected '', got '%v'", v)
	}
}
