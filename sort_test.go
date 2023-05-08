package misc

import (
	"reflect"
	"sort"
	"testing"
)

func TestSortEmptySlice(t *testing.T) {
	arr := []int{}
	Sort(arr, func(a, b int) bool {
		return a < b
	})

	expected := []int{}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func TestSortSingleElementSlice(t *testing.T) {
	arr := []int{42}
	Sort(arr, func(a, b int) bool {
		return a < b
	})

	expected := []int{42}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func TestSortEqualElementsSlice(t *testing.T) {
	arr := []int{5, 5, 5, 5, 5}
	Sort(arr, func(a, b int) bool {
		return a < b
	})

	expected := []int{5, 5, 5, 5, 5}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func TestSortSortedSlice(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8}
	Sort(arr, func(a, b int) bool {
		return a < b
	})

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func TestSortReverseSlice(t *testing.T) {
	arr := []int{8, 7, 6, 5, 4, 3, 2, 1}
	Sort(arr, func(a, b int) bool {
		return a < b
	})

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func TestSortStrings(t *testing.T) {
	arr := []string{"apple", "banana", "orange", "pear", "grape"}
	Sort(arr, func(a, b string) bool {
		return a < b
	})

	expected := []string{"apple", "banana", "grape", "orange", "pear"}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func TestSortInts(t *testing.T) {
	arr := []int{5, 3, 2, 7, 1, 8, 4, 6}
	Sort(arr, func(a, b int) bool {
		return a < b
	})

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8}
	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

type Person struct {
	Name string
	Age  int
}

func TestSortStructs(t *testing.T) {
	arr := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 20},
		{"David", 25},
	}

	Sort(arr, func(a, b Person) bool {
		if a.Age != b.Age {
			return a.Age < b.Age
		}
		return a.Name < b.Name
	})

	expected := []Person{
		{"Charlie", 20},
		{"Alice", 25},
		{"David", 25},
		{"Bob", 30},
	}

	if !reflect.DeepEqual(arr, expected) {
		t.Errorf("Expected %v but got %v", expected, arr)
	}
}

func BenchmarkSortInts(b *testing.B) {
	arr := []int{5, 3, 2, 7, 1, 8, 4, 6}
	for n := 0; n < b.N; n++ {
		Sort(arr, func(a, b int) bool {
			return a < b
		})
	}
}

func BenchmarkSortStrings(b *testing.B) {
	arr := []string{"apple", "banana", "orange", "pear", "grape"}
	for n := 0; n < b.N; n++ {
		Sort(arr, func(a, b string) bool {
			return a < b
		})
	}
}

func BenchmarkSortStructs(b *testing.B) {
	arr := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Charlie", 20},
		{"David", 25},
	}

	for n := 0; n < b.N; n++ {
		sort.Slice(arr, func(i, j int) bool {
			if arr[i].Age != arr[j].Age {
				return arr[i].Age < arr[j].Age
			}
			return arr[i].Name < arr[j].Name
		})
	}
}
