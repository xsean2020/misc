package misc

import (
	"testing"
)

type User struct {
	a int32
	b int32
}

func (u *User) A() int32 {
	return u.a
}

func (u *User) B() int32 {
	return u.b
}

func Benchmark_ternaray3(b *testing.B) {
	var u = &User{a: 10, b: 20}
	for i := 0; i < b.N; i++ {
		if true {
			u.A()
		} else {
			u.B()
		}
	}
}

func Benchmark_ternaray1(b *testing.B) {
	var u = &User{a: 10, b: 20}

	for i := 0; i < b.N; i++ {
		Ternary(true, u.A(), u.B())
	}
}

func Benchmark_ternaray4(b *testing.B) {
	var u = &User{a: 10, b: 20}

	for i := 0; i < b.N; i++ {
		Ternary(true, u, u).A()
	}
}

func Benchmark_ternaray2(b *testing.B) {
	var u = &User{a: 10, b: 20}

	for i := 0; i < b.N; i++ {
		Ternary(true, u.A, u.B)()
	}
}

func TestTernary(t *testing.T) {
	type args struct {
		expr bool
		a    interface{}
		b    interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "expr is true",
			args: args{expr: true, a: 1, b: 2},
			want: 1,
		},
		{
			name: "expr is false",
			args: args{expr: false, a: 1, b: 2},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ternary(tt.args.expr, tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Ternary() = %v, want %v", got, tt.want)
			}
		})
	}
}
