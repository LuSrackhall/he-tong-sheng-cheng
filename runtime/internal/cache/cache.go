package cache

import (
	"asset-leasing-system/runtime/internal/model"
	"sync"
)

// Cache 提供 Capability/Rule/Workflow/Plan 四种类型的内存缓存
type Cache struct {
	mu           sync.RWMutex
	capabilities map[string]*model.Capability
	rules        map[string]*model.Rule
	workflows    map[string]*model.Workflow
	plans        map[string]*model.ExecutionPlan
}

func New() *Cache {
	return &Cache{
		capabilities: make(map[string]*model.Capability),
		rules:        make(map[string]*model.Rule),
		workflows:    make(map[string]*model.Workflow),
		plans:        make(map[string]*model.ExecutionPlan),
	}
}

func (c *Cache) GetCapability(id string) *model.Capability {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.capabilities[id]
}

func (c *Cache) SetCapability(id string, cap *model.Capability) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.capabilities[id] = cap
}

func (c *Cache) GetRule(id string) *model.Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rules[id]
}

func (c *Cache) SetRule(id string, rule *model.Rule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules[id] = rule
}

func (c *Cache) GetWorkflow(id string) *model.Workflow {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.workflows[id]
}

func (c *Cache) SetWorkflow(id string, wf *model.Workflow) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workflows[id] = wf
}

func (c *Cache) GetPlan(hash string) *model.ExecutionPlan {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.plans[hash]
}

func (c *Cache) SetPlan(hash string, plan *model.ExecutionPlan) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.plans[hash] = plan
}

func (c *Cache) InvalidateCapability(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.capabilities, id)
}

func (c *Cache) InvalidateRule(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.rules, id)
}

func (c *Cache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.capabilities = make(map[string]*model.Capability)
	c.rules = make(map[string]*model.Rule)
	c.workflows = make(map[string]*model.Workflow)
	c.plans = make(map[string]*model.ExecutionPlan)
}

func (c *Cache) AllCapabilities() []*model.Capability {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*model.Capability, 0, len(c.capabilities))
	for _, v := range c.capabilities {
		result = append(result, v)
	}
	return result
}

func (c *Cache) AllRules() []*model.Rule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*model.Rule, 0, len(c.rules))
	for _, v := range c.rules {
		result = append(result, v)
	}
	return result
}

func (c *Cache) AllWorkflows() []*model.Workflow {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*model.Workflow, 0, len(c.workflows))
	for _, v := range c.workflows {
		result = append(result, v)
	}
	return result
}
