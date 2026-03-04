package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"babyagent/shared"
)

// SkillMetadata contains just the metadata for system prompt injection
type SkillMetadata struct {
	ID          string
	Name        string
	Description string
}

// Manager handles skill discovery and metadata loading
type Manager struct {
	skillsDir string
	metadata  []SkillMetadata
}

// NewManager creates a new skill manager
func NewManager() *Manager {
	skillsDir := filepath.Join(shared.GetWorkspaceDir(), ".babyagent", "skills")
	return &Manager{
		skillsDir: skillsDir,
		metadata:  make([]SkillMetadata, 0),
	}
}

// LoadAll discovers and loads all skill metadata from .babyagent/skills/
func (m *Manager) LoadAll() error {
	if _, err := os.Stat(m.skillsDir); os.IsNotExist(err) {
		// Skills directory doesn't exist, that's ok
		return nil
	}

	entries, err := os.ReadDir(m.skillsDir)
	if err != nil {
		return fmt.Errorf("failed to read skills directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillID := entry.Name()

		meta, _, err := LoadSkill(skillID)
		if err != nil {
			// Log but continue loading other skills
			fmt.Printf("warning: failed to load skill %s: %v\n", skillID, err)
			continue
		}

		m.metadata = append(m.metadata, meta)
	}

	return nil
}

// FormatForPrompt formats skill metadata for system prompt injection
func (m *Manager) FormatForPrompt() string {
	if len(m.metadata) == 0 {
		return "No skills available."
	}

	var sb strings.Builder
	sb.WriteString("You have access to the following skills. ")
	sb.WriteString("When a user request matches a skill's purpose, use the `load_skill` tool to load the full skill instructions.\n\n")

	for _, meta := range m.metadata {
		sb.WriteString(fmt.Sprintf("- **%s*:%s\n", meta.Name, meta.Description))
	}

	return sb.String()
}
