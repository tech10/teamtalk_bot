package main

// Compares two maps of type map[string]map[string]string
// and returns added, changed, removed

import "fmt"

func mapCompare(first, second map[string]map[string]string) (map[string]map[string]string, map[string]map[string]string, map[string]map[string]string) {
	defer func() {
		pd := recover()
		if pd != nil {
			fmt.Println("PANIC CAUGHT:", pd)
		}
	}()
	added := make(map[string]map[string]string)
	changed := make(map[string]map[string]string)
	removed := make(map[string]map[string]string)

outerloop:
	for k, v := range first {
		if _, exists := second[k]; !exists {
			added[k] = v
			continue outerloop
		}
	innerloop:
		for sk, sv := range v {
			if sv != second[k][sk] {
				if _, exists := changed[k]; !exists {
					changed[k] = make(map[string]string)
				}
				changed[k][sk] = sv
				continue innerloop
			}
		}

	}
	for k, v := range second {
		if _, exists := first[k]; !exists {
			removed[k] = v
		}
	}
	return added, changed, removed
}
