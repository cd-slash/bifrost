package governance

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// UpdateProviderGovernanceRequest represents a request to update provider governance
type UpdateProviderGovernanceRequest struct {
	Budget    *BudgetRequest          `json:"budget,omitempty"`
	RateLimit *CreateRateLimitRequest `json:"rate_limit,omitempty"`
}

// GetErrorMessage extracts error message from API response
func GetErrorMessage(resp *APIResponse) string {
	if msg, ok := resp.Body["message"].(string); ok {
		return msg
	}
	if err, ok := resp.Body["error"].(string); ok {
		return err
	}
	return string(resp.RawBody)
}

// ExtractCostFromResponse extracts the cost from a successful response
func ExtractCostFromResponse(resp *APIResponse) float64 {
	// Look for usage in response
	if usage, ok := resp.Body["usage"].(map[string]interface{}); ok {
		if cost, ok := usage["cost"].(float64); ok {
			return cost
		}
	}
	// Default to small cost if not found
	return 0.001
}

// ExtractUsageFromResponse extracts token usage from response
func ExtractUsageFromResponse(resp *APIResponse) struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
} {
	usage := struct {
		PromptTokens     int
		CompletionTokens int
		TotalTokens      int
	}{}

	if usageMap, ok := resp.Body["usage"].(map[string]interface{}); ok {
		if pt, ok := usageMap["prompt_tokens"].(float64); ok {
			usage.PromptTokens = int(pt)
		}
		if ct, ok := usageMap["completion_tokens"].(float64); ok {
			usage.CompletionTokens = int(ct)
		}
		if tt, ok := usageMap["total_tokens"].(float64); ok {
			usage.TotalTokens = int(tt)
		}
	}

	return usage
}

// ContainsSubstring checks if a string contains a substring (case-insensitive)
func ContainsSubstring(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// TestProviderLimitPrecedence tests that provider limits take precedence over API key limits
// When a provider limit is exceeded, requests should be blocked even if the API key limit hasn't been hit
func TestProviderLimitPrecedence(t *testing.T) {
	t.Parallel()
	testData := NewGlobalTestData()
	defer testData.Cleanup(t)

	// Create a provider with a very low budget limit
	// This will be set at the provider level (global provider governance)
	t.Log("Setting up provider-level budget limit")

	// First, create a provider governance configuration with a strict budget
	providerGovResp := MakeRequest(t, APIRequest{
		Method: "PUT",
		Path:   "/api/governance/providers/openai",
		Body: UpdateProviderGovernanceRequest{
			Budget: &BudgetRequest{
				MaxLimit:      0.05, // Very low provider budget - $0.05
				ResetDuration: "1h",
			},
		},
	})

	if providerGovResp.StatusCode != 200 {
		t.Fatalf("Failed to set provider governance: status %d, body: %v", providerGovResp.StatusCode, providerGovResp.Body)
	}

	t.Log("Created provider governance with $0.05 budget")

	// Create a VK with a high budget limit
	// This tests that provider limits override VK limits
	createVKResp := MakeRequest(t, APIRequest{
		Method: "POST",
		Path:   "/api/governance/virtual-keys",
		Body: CreateVirtualKeyRequest{
			Name: "test-vk-high-budget-" + generateRandomID(),
			Budget: &BudgetRequest{
				MaxLimit:      10.0, // High VK budget - $10.00
				ResetDuration: "1h",
			},
			ProviderConfigs: []ProviderConfigRequest{
				{
					Provider: "openai",
					Weight:   1.0,
					// No provider-specific budget here - relying on global provider governance
				},
			},
		},
	})

	if createVKResp.StatusCode != 200 {
		t.Fatalf("Failed to create VK: status %d, body: %v", createVKResp.StatusCode, createVKResp.Body)
	}

	vkID := ExtractIDFromResponse(t, createVKResp)
	testData.AddVirtualKey(vkID)

	vk := createVKResp.Body["virtual_key"].(map[string]interface{})
	vkValue := vk["value"].(string)

	t.Logf("Created VK with $10.00 budget, but provider limit is $0.05")

	// Test that provider limit is enforced before VK limit
	t.Run("ProviderBudgetBlocksRequestsBeforeVKLimit", func(t *testing.T) {
		providerBudget := 0.05
		consumedBudget := 0.0
		requestNum := 1
		budgetExceeded := false

		for requestNum <= 100 && !budgetExceeded {
			longPrompt := "Please provide a comprehensive and detailed response. " +
				"I need extensive information covering all aspects of the topic. " +
				"Request number " + strconv.Itoa(requestNum) + ". " +
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris. " +
				"Nisi ut aliquip ex ea commodo consequat. " +
				"Duis aute irure dolor in reprehenderit. " +
				"In voluptate velit esse cillum dolore eu fugiat nulla pariatur."

			resp := MakeRequest(t, APIRequest{
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: ChatCompletionRequest{
					Model: "openai/gpt-4o",
					Messages: []ChatMessage{
						{
							Role:    "user",
							Content: longPrompt,
						},
					},
				},
				VKHeader: &vkValue,
			})

			if resp.StatusCode >= 400 {
				// Request failed - check if it's due to budget exceeded
				errorMsg := GetErrorMessage(resp)
				t.Logf("Request %d failed with status %d: %s", requestNum, resp.StatusCode, errorMsg)

				if resp.StatusCode == 429 || resp.StatusCode == 403 {
					// Check that we hit the provider limit, not the VK limit
					if consumedBudget >= providerBudget*0.8 { // Should have consumed most of provider budget
						budgetExceeded = true
						t.Logf("SUCCESS: Provider budget ($%.2f) was enforced before VK budget ($%.2f)",
							providerBudget, 10.0)
						t.Logf("Requests stopped at $%.2f consumed (provider limit: $%.2f, VK limit: $%.2f)",
							consumedBudget, providerBudget, 10.0)
						break
					} else {
						t.Fatalf("Budget exceeded too early: consumed $%.2f, provider limit $%.2f",
							consumedBudget, providerBudget)
					}
				} else {
					t.Fatalf("Request failed with unexpected error: %s", errorMsg)
				}
			} else {
				// Request succeeded
				cost := ExtractCostFromResponse(resp)
				consumedBudget += cost
				t.Logf("Request %d succeeded, cost: $%.4f, total consumed: $%.4f",
					requestNum, cost, consumedBudget)
			}

			requestNum++
			time.Sleep(100 * time.Millisecond) // Small delay between requests
		}

		if !budgetExceeded {
			t.Fatalf("Provider budget was not enforced. Consumed $%.2f, provider limit $%.2f, VK limit $%.2f",
				consumedBudget, providerBudget, 10.0)
		}

		// Verify that we stopped well before the VK budget was hit
		if consumedBudget >= 1.0 {
			t.Fatalf("VK budget limit ($10.00) was not respected - consumed $%.2f", consumedBudget)
		}
	})

	// Clean up provider governance
	MakeRequest(t, APIRequest{
		Method: "DELETE",
		Path:   "/api/governance/providers/openai",
	})
}

// TestProviderRateLimitPrecedence tests that provider rate limits take precedence over API key rate limits
func TestProviderRateLimitPrecedence(t *testing.T) {
	t.Parallel()
	testData := NewGlobalTestData()
	defer testData.Cleanup(t)

	// Set up provider-level rate limit
	providerGovResp := MakeRequest(t, APIRequest{
		Method: "PUT",
		Path:   "/api/governance/providers/openai",
		Body: UpdateProviderGovernanceRequest{
			RateLimit: &CreateRateLimitRequest{
				TokenMaxLimit:      int64Ptr(1000), // Very low token limit
				TokenResetDuration: strPtr("1h"),
			},
		},
	})

	if providerGovResp.StatusCode != 200 {
		t.Fatalf("Failed to set provider governance: status %d", providerGovResp.StatusCode)
	}

	// Create VK with high rate limit
	createVKResp := MakeRequest(t, APIRequest{
		Method: "POST",
		Path:   "/api/governance/virtual-keys",
		Body: CreateVirtualKeyRequest{
			Name: "test-vk-high-ratelimit-" + generateRandomID(),
			RateLimit: &CreateRateLimitRequest{
				TokenMaxLimit:      int64Ptr(100000), // High VK token limit
				TokenResetDuration: strPtr("1h"),
			},
			ProviderConfigs: []ProviderConfigRequest{
				{
					Provider: "openai",
					Weight:   1.0,
				},
			},
		},
	})

	if createVKResp.StatusCode != 200 {
		t.Fatalf("Failed to create VK: status %d", createVKResp.StatusCode)
	}

	vkID := ExtractIDFromResponse(t, createVKResp)
	testData.AddVirtualKey(vkID)

	vk := createVKResp.Body["virtual_key"].(map[string]interface{})
	vkValue := vk["value"].(string)

	t.Logf("Created VK with 100K token limit, but provider limit is 1K tokens")

	// Test that provider rate limit is enforced
	t.Run("ProviderRateLimitBlocksRequestsBeforeVKLimit", func(t *testing.T) {
		requestNum := 1
		rateLimitExceeded := false

		for requestNum <= 50 && !rateLimitExceeded {
			resp := MakeRequest(t, APIRequest{
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: ChatCompletionRequest{
					Model: "openai/gpt-4o",
					Messages: []ChatMessage{
						{
							Role: "user",
							Content: "Write a long essay about artificial intelligence and its impact on society. " +
								"Discuss the ethical implications, economic effects, and future prospects. " +
								"Make it comprehensive and detailed.",
						},
					},
				},
				VKHeader: &vkValue,
			})

			if resp.StatusCode >= 400 {
				errorMsg := GetErrorMessage(resp)
				t.Logf("Request %d failed with status %d: %s", requestNum, resp.StatusCode, errorMsg)

				if resp.StatusCode == 429 {
					// Check for rate limit error
					if ContainsSubstring(errorMsg, "rate limit") ||
						ContainsSubstring(errorMsg, "token limit") {
						rateLimitExceeded = true
						t.Logf("SUCCESS: Provider rate limit (1000 tokens) was enforced before VK rate limit (100000 tokens)")
						break
					}
				}
			} else {
				usage := ExtractUsageFromResponse(resp)
				t.Logf("Request %d succeeded, tokens used: %d", requestNum, usage.TotalTokens)
			}

			requestNum++
			time.Sleep(100 * time.Millisecond)
		}

		if !rateLimitExceeded {
			t.Fatalf("Provider rate limit was not enforced within 50 requests")
		}
	})

	// Clean up
	MakeRequest(t, APIRequest{
		Method: "DELETE",
		Path:   "/api/governance/providers/openai",
	})
}

// TestProviderConfigPrecedenceOverVK tests that provider configs within a VK take precedence
// This tests the VK hierarchy: Provider Config -> VK -> Team -> Customer
func TestProviderConfigPrecedenceOverVK(t *testing.T) {
	t.Parallel()
	testData := NewGlobalTestData()
	defer testData.Cleanup(t)

	// Create a VK with both provider config budget AND VK budget
	// The provider config budget is lower, so it should be enforced first
	createVKResp := MakeRequest(t, APIRequest{
		Method: "POST",
		Path:   "/api/governance/virtual-keys",
		Body: CreateVirtualKeyRequest{
			Name: "test-vk-provider-precedence-" + generateRandomID(),
			Budget: &BudgetRequest{
				MaxLimit:      5.0, // VK budget: $5.00
				ResetDuration: "1h",
			},
			ProviderConfigs: []ProviderConfigRequest{
				{
					Provider: "openai",
					Weight:   1.0,
					Budget: &BudgetRequest{
						MaxLimit:      0.05, // Provider config budget: $0.05 (much lower)
						ResetDuration: "1h",
					},
				},
			},
		},
	})

	if createVKResp.StatusCode != 200 {
		t.Fatalf("Failed to create VK: status %d", createVKResp.StatusCode)
	}

	vkID := ExtractIDFromResponse(t, createVKResp)
	testData.AddVirtualKey(vkID)

	vk := createVKResp.Body["virtual_key"].(map[string]interface{})
	vkValue := vk["value"].(string)

	t.Logf("Created VK with VK budget $5.00 and provider config budget $0.05")

	// Verify provider config budget is enforced before VK budget
	t.Run("ProviderConfigBudgetEnforcedBeforeVKBudget", func(t *testing.T) {
		providerBudget := 0.05
		consumedBudget := 0.0
		requestNum := 1
		budgetExceeded := false

		for requestNum <= 100 && !budgetExceeded {
			longPrompt := "Please write a detailed analysis. " +
				"Request number " + strconv.Itoa(requestNum) + ". " +
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."

			resp := MakeRequest(t, APIRequest{
				Method: "POST",
				Path:   "/v1/chat/completions",
				Body: ChatCompletionRequest{
					Model: "openai/gpt-4o",
					Messages: []ChatMessage{
						{
							Role:    "user",
							Content: longPrompt,
						},
					},
				},
				VKHeader: &vkValue,
			})

			if resp.StatusCode >= 400 {
				errorMsg := GetErrorMessage(resp)
				t.Logf("Request %d failed with status %d: %s", requestNum, resp.StatusCode, errorMsg)

				if resp.StatusCode == 429 || resp.StatusCode == 403 {
					if consumedBudget >= providerBudget*0.7 {
						budgetExceeded = true
						t.Logf("SUCCESS: Provider config budget ($%.2f) enforced before VK budget ($%.2f)",
							providerBudget, 5.0)
						break
					}
				}
			} else {
				cost := ExtractCostFromResponse(resp)
				consumedBudget += cost
				t.Logf("Request %d succeeded, cost: $%.4f, total: $%.4f",
					requestNum, cost, consumedBudget)
			}

			requestNum++
			time.Sleep(100 * time.Millisecond)
		}

		if !budgetExceeded {
			t.Fatalf("Provider config budget was not enforced. Consumed $%.2f, provider config limit $%.2f, VK limit $%.2f",
				consumedBudget, providerBudget, 5.0)
		}
	})
}

// Helper functions
func int64Ptr(i int64) *int64 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
