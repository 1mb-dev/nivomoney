package service

import (
	"fmt"
	"regexp"
	"strings"
)

// TemplateEngine handles variable substitution in notification templates.
type TemplateEngine struct {
	// Regex to match {{variable_name}} patterns
	variablePattern *regexp.Regexp
}

// NewTemplateEngine creates a new template engine.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		variablePattern: regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`),
	}
}

// Render replaces all {{variable}} placeholders in the template with actual values.
// Returns the rendered text and a list of variables that were substituted.
func (e *TemplateEngine) Render(template string, variables map[string]interface{}) (string, []string) {
	usedVariables := make([]string, 0)
	rendered := template

	// Find all variable placeholders
	matches := e.variablePattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		placeholder := match[0]  // Full match: {{variable_name}}
		variableName := match[1] // Captured group: variable_name

		// Get variable value
		value, exists := variables[variableName]
		if !exists {
			// Keep placeholder if variable not provided
			continue
		}

		// Convert value to string
		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		case int:
			stringValue = fmt.Sprintf("%d", v)
		case int64:
			stringValue = fmt.Sprintf("%d", v)
		case float64:
			stringValue = fmt.Sprintf("%.2f", v)
		case bool:
			stringValue = fmt.Sprintf("%t", v)
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		// Replace placeholder with value
		rendered = strings.ReplaceAll(rendered, placeholder, stringValue)
		usedVariables = append(usedVariables, variableName)
	}

	return rendered, usedVariables
}

// ExtractVariables extracts all variable names from a template.
func (e *TemplateEngine) ExtractVariables(template string) []string {
	matches := e.variablePattern.FindAllStringSubmatch(template, -1)
	variables := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) >= 2 {
			variables = append(variables, match[1])
		}
	}

	return variables
}

// Validate checks if all required variables are present in the provided variables map.
func (e *TemplateEngine) Validate(template string, variables map[string]interface{}) []string {
	required := e.ExtractVariables(template)
	missing := make([]string, 0)

	for _, varName := range required {
		if _, exists := variables[varName]; !exists {
			missing = append(missing, varName)
		}
	}

	return missing
}
