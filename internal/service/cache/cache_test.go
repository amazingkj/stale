package cache

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New[string](time.Minute)

	if c == nil {
		t.Fatal("expected non-nil cache")
	}
	if c.items == nil {
		t.Error("items map should be initialized")
	}
	if c.ttl != time.Minute {
		t.Errorf("ttl = %v, want %v", c.ttl, time.Minute)
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := New[string](time.Minute)

	// Test set and get
	c.Set("key1", "value1")

	val, found := c.Get("key1")
	if !found {
		t.Error("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Get(key1) = %q, want %q", val, "value1")
	}
}

func TestCache_GetNotFound(t *testing.T) {
	c := New[string](time.Minute)

	val, found := c.Get("nonexistent")
	if found {
		t.Error("expected not to find nonexistent key")
	}
	if val != "" {
		t.Errorf("Get(nonexistent) = %q, want empty string", val)
	}
}

func TestCache_GetExpired(t *testing.T) {
	c := New[string](10 * time.Millisecond)

	c.Set("key1", "value1")

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	val, found := c.Get("key1")
	if found {
		t.Error("expected expired item to not be found")
	}
	if val != "" {
		t.Errorf("Get(expired) = %q, want empty string", val)
	}
}

func TestCache_Delete(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("key1", "value1")
	c.Delete("key1")

	_, found := c.Get("key1")
	if found {
		t.Error("expected deleted key to not be found")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	c.Clear()

	_, found1 := c.Get("key1")
	_, found2 := c.Get("key2")
	_, found3 := c.Get("key3")

	if found1 || found2 || found3 {
		t.Error("expected all keys to be cleared")
	}
}

func TestCache_OverwriteValue(t *testing.T) {
	c := New[string](time.Minute)

	c.Set("key1", "value1")
	c.Set("key1", "value2")

	val, found := c.Get("key1")
	if !found {
		t.Error("expected to find key1")
	}
	if val != "value2" {
		t.Errorf("Get(key1) = %q, want %q", val, "value2")
	}
}

func TestCache_IntType(t *testing.T) {
	c := New[int](time.Minute)

	c.Set("count", 42)

	val, found := c.Get("count")
	if !found {
		t.Error("expected to find count")
	}
	if val != 42 {
		t.Errorf("Get(count) = %d, want %d", val, 42)
	}
}

func TestCache_StructType(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	c := New[Person](time.Minute)

	person := Person{Name: "John", Age: 30}
	c.Set("person1", person)

	val, found := c.Get("person1")
	if !found {
		t.Error("expected to find person1")
	}
	if val.Name != "John" || val.Age != 30 {
		t.Errorf("Get(person1) = %+v, want %+v", val, person)
	}
}

func TestCache_PointerType(t *testing.T) {
	type Stats struct {
		Count int
	}

	c := New[*Stats](time.Minute)

	stats := &Stats{Count: 100}
	c.Set("stats", stats)

	val, found := c.Get("stats")
	if !found {
		t.Error("expected to find stats")
	}
	if val == nil || val.Count != 100 {
		t.Errorf("Get(stats) = %+v, want %+v", val, stats)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New[int](time.Minute)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			c.Set("key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			c.Get("key")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

func TestCache_SliceType(t *testing.T) {
	c := New[[]string](time.Minute)

	names := []string{"Alice", "Bob", "Charlie"}
	c.Set("names", names)

	val, found := c.Get("names")
	if !found {
		t.Error("expected to find names")
	}
	if len(val) != 3 {
		t.Errorf("Get(names) has %d items, want 3", len(val))
	}
}
