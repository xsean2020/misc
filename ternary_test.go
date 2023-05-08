package misc

import (
	"testing"
)

func A() int32 {
	return 1
}

func B() int32 {
	return 2
}

func Benchmark_ternaray1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Ternary(true, A(), B())
	}
}

func Benchmark_ternaray2(b *testing.B) {

	for i := 0; i < b.N; i++ {
		Ternary(true, A, B)()
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
