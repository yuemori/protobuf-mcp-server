package templates

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed *.yml
var promptFiles embed.FS

// PromptData represents the data for template rendering
type PromptData struct {
	ProjectRoot string
}

// GetOnboardingPrompt returns the onboarding prompt with template data
func GetOnboardingPrompt(projectRoot string) (string, error) {
	// Read the template file
	templateContent, err := promptFiles.ReadFile("onboarding_prompt.yml")
	if err != nil {
		return "", fmt.Errorf("failed to read onboarding prompt template: %w", err)
	}

	// Parse the template
	tmpl, err := template.New("onboarding_prompt").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse onboarding prompt template: %w", err)
	}

	// Prepare data for template rendering
	data := PromptData{
		ProjectRoot: projectRoot,
	}

	// Render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute onboarding prompt template: %w", err)
	}

	return buf.String(), nil
}
