package skills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type persistedState struct {
	Disabled []string `json:"disabled"`
}

func (r *SkillRegistry) SetStateFile(path string) error {
	r.mu.Lock()
	r.stateFile = path
	r.mu.Unlock()
	return r.loadState()
}

func (r *SkillRegistry) loadState() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stateFile == "" {
		return nil
	}

	data, err := os.ReadFile(r.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var st persistedState
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}

	disabled := map[string]bool{}
	for _, name := range st.Disabled {
		disabled[name] = true
	}
	r.disabledSet = disabled

	for name, s := range r.skills {
		if disabled[name] {
			s.Enabled = false
		}
	}
	return nil
}

func (r *SkillRegistry) saveStateLocked() error {
	if r.stateFile == "" {
		return nil
	}

	disabled := []string{}
	for name := range r.disabledSet {
		disabled = append(disabled, name)
	}

	data, err := json.MarshalIndent(persistedState{Disabled: disabled}, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(r.stateFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(r.stateFile, data, 0644)
}

func (r *SkillRegistry) Enable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.skills[name]
	if !ok {
		return errors.New("skill not found")
	}
	s.Enabled = true
	delete(r.disabledSet, name)
	return r.saveStateLocked()
}

func (r *SkillRegistry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, ok := r.skills[name]
	if !ok {
		return errors.New("skill not found")
	}
	s.Enabled = false
	r.disabledSet[name] = true
	return r.saveStateLocked()
}

func (r *SkillRegistry) Delete(name string) error {
	r.mu.Lock()

	s, ok := r.skills[name]
	if !ok {
		r.mu.Unlock()
		return errors.New("skill not found")
	}

	if s.Source == SourceInternal {
		r.mu.Unlock()
		return errors.New("cannot delete internal skill")
	}

	path := s.Path
	delete(r.skills, name)
	delete(r.disabledSet, name)
	saveErr := r.saveStateLocked()
	r.mu.Unlock()

	if path != "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return saveErr
}
