package ring

import (
	"fmt"
	"testing"
)

func TestRingBoundary(t *testing.T) {
	// 测试创建新的Ring
	rb := NewRing[int](5)
	if len(rb.data) != 5 {
		t.Errorf("Expected length of data to be 5, got %d", len(rb.data))
	}
	if rb.tail != 0 {
		t.Errorf("Expected tail to be -1 initially, got %d", rb.tail)
	}
	if rb.head != 0 {
		t.Errorf("Expected head to be 0 initially, got %d", rb.head)
	}

	// 测试Push方法
	for i := 1; i <= 5; i++ {
		rb.Push(i)
	}
	if rb.tail != 0 {
		t.Errorf("Expected tail to be 4, got %d", rb.tail)
	}
	if rb.head != 0 {
		t.Errorf("Expected head to be 0, got %d", rb.head)
	}

	// 测试Range方法
	var result []int
	rb.Range(func(val int) bool {
		result = append(result, val)
		return false
	})
	expected := []int{1, 2, 3, 4, 5}
	for i, val := range result {
		if val != expected[i] {
			t.Errorf("Expected value %d at index %d, got %d", expected[i], i, val)
		}
	}

	// 测试RevRange方法
	result = nil
	rb.RevRange(func(val int) bool {
		result = append(result, val)
		return false
	})
	expected = []int{5, 4, 3, 2, 1}
	for i, val := range result {
		if val != expected[i] {
			t.Errorf("Expected value %d at index %d, got %d", expected[i], i, val)
		}
	}

	// 测试环形缓冲区满时的覆盖行为
	rb.Push(6)
	rb.Push(7)
	rb.Push(8)

	if *rb.Tail() != 8 {

		t.Fatalf("tail error")
	}

	result = nil
	rb.Range(func(val int) bool {
		// fmt.Printf("%d ", val)
		result = append(result, val)
		return false
	})

	expected = []int{4, 5, 6, 7, 8}
	for i, val := range result {
		if val != expected[i] {
			t.Errorf("Expected value %d at index %d, got %d", expected[i], i, val)
		}
	}

	result = nil
	rb.RevRange(func(val int) bool {
		fmt.Printf("%d ", val)
		result = append(result, val)
		return false
	})

	expected = []int{8, 7, 6, 5, 4}
	for i, val := range result {
		if val != expected[i] {
			t.Errorf("Expected value %d at index %d, got %d", expected[i], i, val)
		}
	}
	fmt.Println()
	fmt.Println("-----", rb.Len())

	// 期望输出: 4 5 6 7 8
}
