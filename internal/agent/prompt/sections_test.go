package prompt

import "testing"

func TestSectionCache(t *testing.T) {
	c := newSectionCache()
	if _, ok := c.get("x"); ok {
		t.Fatalf("expected miss")
	}
	c.set("x", "y")
	if v, ok := c.get("x"); !ok || v != "y" {
		t.Fatalf("expected cached value")
	}
	c.clear()
	if _, ok := c.get("x"); ok {
		t.Fatalf("expected cache to be cleared")
	}
}
