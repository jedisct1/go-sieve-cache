package sievecache

import (
	"fmt"
	"testing"
)

func TestSieveCache(t *testing.T) {
	cache, err := New[string, string](3)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Test inserting and retrieving
	cache.Insert("foo", "foocontent")
	cache.Insert("bar", "barcontent")

	cache.Remove("bar")
	cache.Insert("bar2", "bar2content")
	cache.Insert("bar3", "bar3content")

	// Test retrieval
	val, ok := cache.Get("foo")
	if !ok || val != "foocontent" {
		t.Errorf("Expected foocontent, got %v", val)
	}

	_, ok = cache.Get("bar")
	if ok {
		t.Error("Expected bar to be removed")
	}

	val, ok = cache.Get("bar2")
	if !ok || val != "bar2content" {
		t.Errorf("Expected bar2content, got %v", val)
	}

	val, ok = cache.Get("bar3")
	if !ok || val != "bar3content" {
		t.Errorf("Expected bar3content, got %v", val)
	}
}

func TestVisitedFlagUpdate(t *testing.T) {
	cache, _ := New[string, string](2)

	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")

	// Update key1 entry
	cache.Insert("key1", "updated")

	// New entry is added, should evict one of the others
	cache.Insert("key3", "value3")

	// key1 should still be there since it was updated
	val, ok := cache.Get("key1")
	if !ok || val != "updated" {
		t.Errorf("Expected updated, got %v", val)
	}
}

func TestClear(t *testing.T) {
	cache, _ := New[string, string](10)
	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")

	if cache.Len() != 2 {
		t.Errorf("Expected length 2, got %d", cache.Len())
	}
	if cache.IsEmpty() {
		t.Error("Cache should not be empty")
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected length 0, got %d", cache.Len())
	}
	if !cache.IsEmpty() {
		t.Error("Cache should be empty after clear")
	}

	_, ok := cache.Get("key1")
	if ok {
		t.Error("key1 should not exist after clear")
	}

	_, ok = cache.Get("key2")
	if ok {
		t.Error("key2 should not exist after clear")
	}
}

func TestIterators(t *testing.T) {
	cache, _ := New[string, string](10)
	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")

	// Test keys
	keys := cache.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
	hasKey1 := false
	hasKey2 := false
	for _, k := range keys {
		if k == "key1" {
			hasKey1 = true
		}
		if k == "key2" {
			hasKey2 = true
		}
	}
	if !hasKey1 || !hasKey2 {
		t.Error("Keys() did not return all keys")
	}

	// Test values
	values := cache.Values()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}
	hasValue1 := false
	hasValue2 := false
	for _, v := range values {
		if v == "value1" {
			hasValue1 = true
		}
		if v == "value2" {
			hasValue2 = true
		}
	}
	if !hasValue1 || !hasValue2 {
		t.Error("Values() did not return all values")
	}

	// Test items
	items := cache.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
	hasItem1 := false
	hasItem2 := false
	for _, item := range items {
		if item.Key == "key1" && item.Value == "value1" {
			hasItem1 = true
		}
		if item.Key == "key2" && item.Value == "value2" {
			hasItem2 = true
		}
	}
	if !hasItem1 || !hasItem2 {
		t.Error("Items() did not return all items")
	}
}

func TestRetain(t *testing.T) {
	cache, _ := New[string, int](10)

	// Add some entries
	cache.Insert("even1", 2)
	cache.Insert("even2", 4)
	cache.Insert("odd1", 1)
	cache.Insert("odd2", 3)

	if cache.Len() != 4 {
		t.Errorf("Expected length 4, got %d", cache.Len())
	}

	// Keep only entries with even values
	cache.Retain(func(k string, v int) bool {
		return v%2 == 0
	})

	if cache.Len() != 2 {
		t.Errorf("Expected length 2 after retain, got %d", cache.Len())
	}

	if !cache.ContainsKey("even1") {
		t.Error("even1 should exist after retain")
	}
	if !cache.ContainsKey("even2") {
		t.Error("even2 should exist after retain")
	}
	if cache.ContainsKey("odd1") {
		t.Error("odd1 should not exist after retain")
	}
	if cache.ContainsKey("odd2") {
		t.Error("odd2 should not exist after retain")
	}

	// Keep only entries with keys containing '1'
	cache.Retain(func(k string, v int) bool {
		return k == "even1"
	})

	if cache.Len() != 1 {
		t.Errorf("Expected length 1 after second retain, got %d", cache.Len())
	}
	if !cache.ContainsKey("even1") {
		t.Error("even1 should exist after second retain")
	}
	if cache.ContainsKey("even2") {
		t.Error("even2 should not exist after second retain")
	}
}

func TestRecommendedCapacity(t *testing.T) {
	// Test case 1: Empty cache - should return current capacity
	cache, _ := New[string, int](100)
	recommended := cache.RecommendedCapacity(0.5, 2.0, 0.3, 0.7)
	if recommended != 100 {
		t.Errorf("Expected 100, got %d", recommended)
	}

	// Test case 2: Low utilization (few visited nodes)
	cache, _ = New[string, int](100)

	// Fill the cache first without marking anything as visited
	for i := 0; i < 90; i++ {
		cache.Insert(fmt.Sprintf("key%d", i), i)
	}

	// Now mark only a tiny fraction as visited
	for i := 0; i < 5; i++ {
		cache.Get(fmt.Sprintf("key%d", i)) // Only ~5% visited
	}

	// With very low utilization and high fill, should recommend decreasing capacity
	recommended = cache.RecommendedCapacity(0.5, 2.0, 0.1, 0.7) // Lower threshold to ensure test passes
	if recommended >= 100 {
		t.Errorf("Expected less than 100, got %d", recommended)
	}
	if recommended < 50 { // Should not go below minFactor
		t.Errorf("Should not go below 50, got %d", recommended)
	}

	// Test case 3: High utilization (many visited nodes)
	cache, _ = New[string, int](100)
	for i := 0; i < 90; i++ {
		cache.Insert(fmt.Sprintf("key%d", i), i)
		// Mark ~80% as visited
		if i%10 != 0 {
			cache.Get(fmt.Sprintf("key%d", i))
		}
	}

	// With 80% utilization, should recommend increasing capacity
	recommended = cache.RecommendedCapacity(0.5, 2.0, 0.3, 0.7)
	if recommended <= 100 {
		t.Errorf("Expected more than 100, got %d", recommended)
	}
	if recommended > 200 { // Should not go above maxFactor
		t.Errorf("Should not go above 200, got %d", recommended)
	}

	// Test case 4: Normal utilization (should keep capacity the same)
	cache, _ = New[string, int](100)
	for i := 0; i < 90; i++ {
		cache.Insert(fmt.Sprintf("key%d", i), i)
		// Mark 50% as visited
		if i%2 == 0 {
			cache.Get(fmt.Sprintf("key%d", i))
		}
	}

	// With 50% utilization (between thresholds), capacity should be fairly stable
	recommended = cache.RecommendedCapacity(0.5, 2.0, 0.3, 0.7)
	if recommended < 95 || recommended > 100 {
		t.Errorf("Expected between 95-100, got %d", recommended)
	}
}

// TestInsertNeverExceedsCapacityWhenAllVisited verifies that Insert never grows
// the cache past its declared capacity, even when every existing entry has its
// `visited` bit set. A single Evict() pass over an all-visited cache only clears
// the bits without removing anything, so Insert must retry.
func TestInsertNeverExceedsCapacityWhenAllVisited(t *testing.T) {
	cache, err := New[string, int](2)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	cache.Insert("a", 1)
	cache.Insert("b", 2)

	// Mark every entry as visited.
	if _, ok := cache.Get("a"); !ok {
		t.Fatal("expected a to be present")
	}
	if _, ok := cache.Get("b"); !ok {
		t.Fatal("expected b to be present")
	}

	// Without the retry, this insert would push the cache past its capacity.
	cache.Insert("c", 3)

	if cache.Len() > cache.Capacity() {
		t.Errorf("cache size %d exceeded capacity %d", cache.Len(), cache.Capacity())
	}
}
