package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"babyagent/shared"
)

// LoadSkill loads both metadata and content from a SKILLS.md file by id
// Returns (metadata, content, error)
func LoadSkill(id string) (SkillMetadata, string, error) {
	skillPath := filepath.Join(shared.GetWorkspaceDir(), ".babyagent", "skills", id, "SKILL.md")

	contentBytes, err := os.ReadFile(skillPath)
	if err != nil {
		return SkillMetadata{}, "", fmt.Errorf("failed to read skill file: %w", err)
	}

	text := string(contentBytes)
	text = strings.TrimLeft(text, "\n\r")

	// Check for front matter delimiter
	if !strings.HasPrefix(text, "---") {
		return SkillMetadata{}, "", errors.New("skill file must start with front matter (---)")
	}

	// Find the end of front matter
	endIdx := strings.Index(text[3:], "\n---")
	if endIdx == -1 {
		return SkillMetadata{}, "", errors.New("front matter not closed (---)")
	}
	endIdx += 3 // Account for the opening "---"

	frontMatter := text[3:endIdx]

	meta := SkillMetadata{ID: id}

	// Parse front matter
	for _, line := range strings.Split(frontMatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			meta.Name = value
		case "description":
			meta.Description = value
		}
	}

	if meta.Name == "" {
		return SkillMetadata{}, "", errors.New("skill must have a 'name' field in front matter")
	}

	if meta.Description == "" {
		return SkillMetadata{}, "", errors.New("skill must have a 'description' field in front matter")
	}

	// Extract body content (after front matter)
	endIdx += 4 // Skip closing "\n---"
	bodyContent := strings.TrimSpace(text[endIdx:])

	return meta, bodyContent, nil
}
