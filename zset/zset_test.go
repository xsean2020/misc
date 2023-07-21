package zset

import (
	"testing"
)

func TestSortedSet(t *testing.T) {
	// Initialize the SortedSet
	ss := New[int, int, float64]()

	// Test adding elements
	ss.Set(1, 10, 3.14)
	ss.Set(2, 15, 2.71)
	ss.Set(3, 5, 1.618)

	// Test Length()
	if length := ss.Length(); length != 3 {
		t.Errorf("Expected length: 3, got: %d", length)
	}

	// Test GetRank()
	rank, score, data := ss.GetRank(2, false)
	if rank != 2 || score != 15 || data != 2.71 {
		t.Errorf("GetRank failed. Expected: (1, 15, 2.71), Got: (%d, %d, %f)", rank, score, data)
	}

	// Test GetData()
	data, ok := ss.GetData(3)
	if !ok || data != 1.618 {
		t.Errorf("GetData failed. Expected: (true, 1.618), Got: (%t, %f)", ok, data)
	}

	// Test GetScore()
	score, ok = ss.GetScore(1)
	if !ok || score != 10 {
		t.Errorf("GetScore failed. Expected: (true, 10), Got: (%t, %d)", ok, score)
	}

	// Test Range()
	results := make(map[int]int)
	ss.Range(0, 1, false, func(key int, score int, data float64) {
		results[key] = score
	})

	expectedResults := map[int]int{3: 5, 1: 10}
	if !compareMaps(results, expectedResults) {
		t.Errorf("Range failed. Expected: %v, Got: %v", expectedResults, results)
	}

	// Test RangeByScore()
	results = make(map[int]int)
	ss.RangeByScore(6, 15, false, func(key int, score int, data float64) {
		results[key] = score
	})
	expectedResults = map[int]int{1: 10, 2: 15}
	if !compareMaps(results, expectedResults) {
		t.Errorf("RangeByScore failed. Expected: %v, Got: %v", expectedResults, results)
	}

	// Test Increment score
	newScore, newData := ss.Incr(1, 5)
	if newScore != 15 || newData != 3.14 {
		t.Errorf("Incr failed. Expected: (15, 3.14), Got: (%v, %f)", newScore, newData)
	}

	// Test deleting an element
	ok = ss.Delete(2)
	if !ok {
		t.Errorf("Delete failed. Expected: true, Got: false")
	}

	// Test Length() after deletion
	if length := ss.Length(); length != 2 {
		t.Errorf("Expected length: 2, got: %d", length)
	}
}

func compareMaps(map1, map2 map[int]int) bool {
	if len(map1) != len(map2) {
		return false
	}
	for k, v1 := range map1 {
		v2, ok := map2[k]
		if !ok || v1 != v2 {
			return false
		}
	}
	return true
}
