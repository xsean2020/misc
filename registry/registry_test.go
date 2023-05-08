package registry

import (
	"testing"
)

type User struct {
	ID int
}

func (u *User) Primary() int {
	return u.ID
}

func TestRegister(t *testing.T) {
	r := &Registry[int, *User]{}
	r.Init(10)

	// add a user
	user := &User{ID: 1}
	r.Register(user)

	// check if the user is added
	if r.Query(user.ID) != user {
		t.Error("User not registered")
	}
}

func TestUnregister(t *testing.T) {
	r := &Registry[int, *User]{}

	r.Init(10)

	// add a user
	user := &User{ID: 1}
	r.Register(user)

	// remove the user
	r.Unregister(user)

	// check if the user is removed
	if r.Query(user.ID) != nil {
		t.Error("User not unregistered")
	}

	if r.Count() != 0 {

		t.Error("User not unregistered")

	}
}

func TestQuery(t *testing.T) {
	r := &Registry[int, *User]{}

	r.Init(10)

	// add a user
	user := &User{ID: 1}
	r.Register(user)

	// query the user
	user2 := r.Query(user.ID)

	// check if the user is found
	if user2 != user {
		t.Error("User not found")
	}
}

func TestAll(t *testing.T) {
	r := &Registry[int, *User]{}

	r.Init(10)

	// add some users
	user1 := &User{ID: 1}
	user2 := &User{ID: 2}
	user3 := &User{ID: 3}
	r.Register(user1)
	r.Register(user2)
	r.Register(user3)

	// query all users
	all := r.All()

	// check if all users are found
	if len(all) != 3 || !contains(all, user1.ID) || !contains(all, user2.ID) || !contains(all, user3.ID) {
		t.Error("All users not found")
	}
}

func TestCount(t *testing.T) {
	r := &Registry[int, *User]{}

	r.Init(10)

	// add some users
	user1 := &User{ID: 1}
	user2 := &User{ID: 2}
	user3 := &User{ID: 3}
	r.Register(user1)
	r.Register(user2)
	r.Register(user3)

	// query the count
	count := r.Count()

	// check if the count is correct
	if count != 3 {
		t.Error("Incorrect count")
	}
}

func contains(ids []int, id int) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}
