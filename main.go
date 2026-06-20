package main

import (
	"fmt"
	"github.com/rickyschuster2zz/Cilium/pkg/identity/cache"
)

func main() {
	fmt.Println("Hello, Bounty Hunter!")
	c := cache.NewIdentityCache()
	pod := cache.Pod{
		UID:             "uid-a",
		IP:              "10.0.0.1",
		Labels:          []string{"app=a"},
		ResourceVersion: 100,
	}
	c.Upsert(pod)
	labels, _ := c.Lookup("10.0.0.1")
	fmt.Printf("Cached labels: %v\n", labels)
}
