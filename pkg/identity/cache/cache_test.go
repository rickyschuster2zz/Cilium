package cache

import (
	"reflect"
	"testing"
)

func TestIdentityCache_Basic(t *testing.T) {
	c := NewIdentityCache()

	podA := Pod{
		UID:             "uid-a",
		IP:              "10.0.0.1",
		Labels:          []string{"app=a"},
		ResourceVersion: 100,
	}

	// Add PodA
	if !c.Upsert(podA) {
		t.Error("Expected Upsert to succeed")
	}

	labels, found := c.Lookup("10.0.0.1")
	if !found || !reflect.DeepEqual(labels, []string{"app=a"}) {
		t.Errorf("Expected labels [app=a], got %v", labels)
	}

	// Delete PodA
	if !c.Delete(podA) {
		t.Error("Expected Delete to succeed")
	}

	_, found = c.Lookup("10.0.0.1")
	if found {
		t.Error("Expected IP to be evicted")
	}
}

func TestIdentityCache_OutOrOrderDelete(t *testing.T) {
	c := NewIdentityCache()

	podA := Pod{
		UID:             "uid-a",
		IP:              "10.0.0.1",
		Labels:          []string{"app=a"},
		ResourceVersion: 100,
	}

	podB := Pod{
		UID:             "uid-b",
		IP:              "10.0.0.1",
		Labels:          []string{"app=b"},
		ResourceVersion: 200,
	}

	// Simulate sequence: Add(PodA) -> Add(PodB) -> Delete(PodA)
	if !c.Upsert(podA) {
		t.Error("Expected Upsert PodA to succeed")
	}

	if !c.Upsert(podB) {
		t.Error("Expected Upsert PodB to succeed")
	}

	// Delete(PodA) arrives late
	if c.Delete(podA) {
		t.Error("Expected Delete PodA to be ignored because PodB is active")
	}

	// Verify PodB's labels are still present
	labels, found := c.Lookup("10.0.0.1")
	if !found || !reflect.DeepEqual(labels, []string{"app=b"}) {
		t.Errorf("Expected labels [app=b], got %v", labels)
	}
}

func TestIdentityCache_StaleUpsert(t *testing.T) {
	c := NewIdentityCache()

	podA1 := Pod{
		UID:             "uid-a",
		IP:              "10.0.0.1",
		Labels:          []string{"app=a", "version=1"},
		ResourceVersion: 100,
	}

	podA2 := Pod{
		UID:             "uid-a",
		IP:              "10.0.0.1",
		Labels:          []string{"app=a", "version=2"},
		ResourceVersion: 90, // older resource version
	}

	if !c.Upsert(podA1) {
		t.Error("Expected Upsert PodA1 to succeed")
	}

	if c.Upsert(podA2) {
		t.Error("Expected Upsert PodA2 to be ignored due to older resource version")
	}

	labels, found := c.Lookup("10.0.0.1")
	if !found || !reflect.DeepEqual(labels, []string{"app=a", "version=1"}) {
		t.Errorf("Expected labels [app=a, version=1], got %v", labels)
	}
}
