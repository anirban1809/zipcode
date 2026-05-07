package skills

import (
	"sort"
	"strings"
	"sync"
)

type SkillSource string

const (
	SourceInternal SkillSource = "internal"
	SourceGlobal   SkillSource = "global"
	SourceProject  SkillSource = "project"
)

type Skill struct {
	Name           string
	Description    string
	PromptTemplate string
	Source         SkillSource
	Path           string
	Enabled        bool
}

type SkillRegistry struct {
	mu          sync.RWMutex
	skills      map[string]*Skill
	stateFile   string
	disabledSet map[string]bool
}

func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills:      map[string]*Skill{},
		disabledSet: map[string]bool{},
	}
}

func (r *SkillRegistry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	name = strings.TrimPrefix(name, "/")
	s, ok := r.skills[name]
	return s, ok
}

func (r *SkillRegistry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *SkillRegistry) ListEnabled() []*Skill {
	all := r.List()
	out := make([]*Skill, 0, len(all))
	for _, s := range all {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

func (r *SkillRegistry) replace(skills map[string]*Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range skills {
		if r.disabledSet[s.Name] {
			s.Enabled = false
		}
	}
	r.skills = skills
}
