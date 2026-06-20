package cache

import (
	"sync"
)

// Pod represents a Kubernetes Pod with its IP, UID, Labels, and ResourceVersion.
type Pod struct {
	UID             string
	IP              string
	Labels          []string
	ResourceVersion uint64
}

// cacheEntry represents a cached identity entry.
type cacheEntry struct {
	podUID          string
	labels          []string
	resourceVersion uint64
}

// IdentityCache manages the local cache of security identities.
type IdentityCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry // key: IP
}

// NewIdentityCache creates a new IdentityCache.
func NewIdentityCache() *IdentityCache {
	return &IdentityCache{
		entries: make(map[string]*cacheEntry),
	}
}

// Upsert updates or inserts a pod's identity/labels in the cache.
// It returns true if the cache was updated, false if the update was ignored (stale).
func (c *IdentityCache) Upsert(pod Pod) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.entries[pod.IP]
	if exists {
		// If the existing entry has a higher resource version, ignore the update.
		if existing.resourceVersion > pod.ResourceVersion {
			return false
		}
	}

	c.entries[pod.IP] = &cacheEntry{
		podUID:          pod.UID,
		labels:          pod.Labels,
		resourceVersion: pod.ResourceVersion,
	}
	return true
}

// Delete removes a pod's identity/labels from the cache.
// It returns true if the entry was deleted, false if the delete was ignored (stale/newer pod exists).
func (c *IdentityCache) Delete(pod Pod) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.entries[pod.IP]
	if !exists {
		return false
	}

	// Only delete if the cached entry matches the pod UID and the resource version is not newer.
	if existing.podUID != pod.UID {
		return false
	}
	if existing.resourceVersion > pod.ResourceVersion {
		return false
	}

	delete(c.entries, pod.IP)
	return true
}

// Lookup retrieves the labels for a given IP.
func (c *IdentityCache) Lookup(ip string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[ip]
	if !exists {
		return nil, false
	}
	// Return a copy of labels to avoid data race if caller modifies it
	labelsCopy := make([]string, len(entry.labels))
	copy(labelsCopy, entry.labels)
	return labelsCopy, true
}
