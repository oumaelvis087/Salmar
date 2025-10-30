package ethics

import (
	"strings"

	"github.com/elvisouma/salmar-ai/pkg/models"
)

// ContentFilter provides ethical filtering for AI responses
type ContentFilter struct {
	// Configuration options
	EnabledFilters map[string]bool
	SensitivityLevel string // "low", "medium", "high"
	
	// Banned terms and phrases
	bannedTerms []string
	
	// Categories of content to filter
	categories map[string][]string
}

// NewContentFilter creates a new content filter with default settings
func NewContentFilter() *ContentFilter {
	filter := &ContentFilter{
		EnabledFilters: map[string]bool{
			"harmful_content": true,
			"bias": true,
			"privacy": true,
			"security": true,
			"transparency": true,
		},
		SensitivityLevel: "medium",
		categories: make(map[string][]string),
	}
	
	// Initialize categories with example terms
	// In a production system, these would be more comprehensive
	filter.categories["harmful_content"] = []string{
		"harmful", "dangerous", "illegal", "exploit", "attack",
	}
	
	filter.categories["bias"] = []string{
		"biased", "discriminatory", "stereotype", "prejudice",
	}
	
	filter.categories["privacy"] = []string{
		"personal data", "private information", "confidential", "sensitive data",
	}
	
	filter.categories["security"] = []string{
		"vulnerability", "exploit", "hack", "breach", "backdoor",
	}
	
	// Flatten categories into banned terms
	filter.updateBannedTerms()
	
	return filter
}

// updateBannedTerms flattens enabled categories into a single list of banned terms
func (f *ContentFilter) updateBannedTerms() {
	f.bannedTerms = []string{}
	
	for category, terms := range f.categories {
		if f.EnabledFilters[category] {
			f.bannedTerms = append(f.bannedTerms, terms...)
		}
	}
}

// FilterResponse applies ethical filtering to a response
func (f *ContentFilter) FilterResponse(response *models.Response) *models.Response {
	// Add a reasoning step for ethical filtering
	ethicalStep := models.ReasoningStep{
		Description: "Applied ethical content filtering to ensure response adheres to ethical guidelines",
		Confidence:  0.95,
	}
	response.ReasoningSteps = append(response.ReasoningSteps, ethicalStep)
	
	// Filter text response
	if response.TextResponse != "" {
		response.TextResponse = f.filterText(response.TextResponse)
	}
	
	// Filter code explanation
	if response.CodeExplanation != "" {
		response.CodeExplanation = f.filterText(response.CodeExplanation)
	}
	
	// Add transparency information
	response = f.addTransparencyInfo(response)
	
	return response
}

// filterText checks and filters text for ethical concerns
func (f *ContentFilter) filterText(text string) string {
	// This is a simplified implementation
	// In a production system, this would use more sophisticated NLP techniques
	
	// Check for banned terms based on sensitivity level
	for _, term := range f.bannedTerms {
		if strings.Contains(strings.ToLower(text), strings.ToLower(term)) {
			// Replace or redact based on sensitivity level
			switch f.SensitivityLevel {
			case "high":
				// Completely remove the content
				return "Content filtered due to ethical concerns."
			case "medium":
				// Redact the specific terms
				text = strings.ReplaceAll(
					strings.ToLower(text),
					strings.ToLower(term),
					"[FILTERED]",
				)
			case "low":
				// Add a warning but keep the content
				text = "Note: This content may contain sensitive information. " + text
			}
		}
	}
	
	return text
}

// addTransparencyInfo adds transparency information to the response
func (f *ContentFilter) addTransparencyInfo(response *models.Response) *models.Response {
	// Add model information if not present
	if response.ModelInfo == nil {
		response.ModelInfo = make(map[string]interface{})
	}
	
	// Add ethical filtering information
	response.ModelInfo["ethical_filtering"] = map[string]interface{}{
		"applied": true,
		"sensitivity_level": f.SensitivityLevel,
		"enabled_filters": f.EnabledFilters,
	}
	
	// Ensure reasoning steps include transparency information
	transparencyStep := models.ReasoningStep{
		Description: "Added transparency information to response",
		Confidence:  1.0,
	}
	response.ReasoningSteps = append(response.ReasoningSteps, transparencyStep)
	
	return response
}