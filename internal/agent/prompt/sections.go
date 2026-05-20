package prompt

import "sync"

type sectionComputeFn func(PromptDat) string

type promptSection struct {
	name       string
	compute    sectionComputeFn
	cacheBreak bool
}

func systemPromptSection(name string, compute sectionComputeFn) promptSection {
	return promptSection{name: name, compute: compute}
}

func uncachedSystemPromptSection(name string, compute sectionComputeFn) promptSection {
	return promptSection{name: name, compute: compute, cacheBreak: true}
}

type sectionCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newSectionCache() *sectionCache {
	return &sectionCache{
		data: make(map[string]string),
	}
}

func (c *sectionCache) get(name string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[name]
	return v, ok
}

func (c *sectionCache) set(name, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[name] = value
}

