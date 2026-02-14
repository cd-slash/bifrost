package governance

import (
	"testing"
)

// TestAllLimitsEndpoint tests the consolidated /api/governance/all-limits endpoint
func TestAllLimitsEndpoint(t *testing.T) {
	t.Parallel()
	testData := NewGlobalTestData()
	defer testData.Cleanup(t)

	// Create test resources
	// 1. Create a provider with governance
	providerGovResp := MakeRequest(t, APIRequest{
		Method: "PUT",
		Path:   "/api/governance/providers/openai",
		Body: UpdateProviderGovernanceRequest{
			Budget: &BudgetRequest{
				MaxLimit:      100.0,
				ResetDuration: "1M",
			},
		},
	})
	if providerGovResp.StatusCode != 200 {
		t.Fatalf("Failed to create provider governance: status %d", providerGovResp.StatusCode)
	}

	// 2. Create a VK with limits
	createVKResp := MakeRequest(t, APIRequest{
		Method: "POST",
		Path:   "/api/governance/virtual-keys",
		Body: CreateVirtualKeyRequest{
			Name: "test-vk-limits-" + generateRandomID(),
			Budget: &BudgetRequest{
				MaxLimit:      50.0,
				ResetDuration: "1d",
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

	// Test the all-limits endpoint
	t.Run("AllLimitsEndpointReturnsData", func(t *testing.T) {
		resp := MakeRequest(t, APIRequest{
			Method: "GET",
			Path:   "/api/governance/all-limits?from_memory=true",
		})

		if resp.StatusCode != 200 {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify response structure
		if _, ok := resp.Body["providers"]; !ok {
			t.Error("Response missing 'providers' field")
		}
		if _, ok := resp.Body["virtual_keys"]; !ok {
			t.Error("Response missing 'virtual_keys' field")
		}
		if _, ok := resp.Body["model_configs"]; !ok {
			t.Error("Response missing 'model_configs' field")
		}
		if _, ok := resp.Body["teams"]; !ok {
			t.Error("Response missing 'teams' field")
		}
		if _, ok := resp.Body["customers"]; !ok {
			t.Error("Response missing 'customers' field")
		}
		if _, ok := resp.Body["budgets"]; !ok {
			t.Error("Response missing 'budgets' field")
		}
		if _, ok := resp.Body["rate_limits"]; !ok {
			t.Error("Response missing 'rate_limits' field")
		}

		t.Log("All-limits endpoint returned complete data structure")
	})

	// Clean up provider governance
	MakeRequest(t, APIRequest{
		Method: "DELETE",
		Path:   "/api/governance/providers/openai",
	})
}

// TestAllLimitsWithFlexibleDuration tests that flexible duration formats work via API
func TestAllLimitsWithFlexibleDuration(t *testing.T) {
	t.Parallel()
	testData := NewGlobalTestData()
	defer testData.Cleanup(t)

	// Test creating budget with flexible duration (90m)
	createVKResp := MakeRequest(t, APIRequest{
		Method: "POST",
		Path:   "/api/governance/virtual-keys",
		Body: CreateVirtualKeyRequest{
			Name: "test-vk-flexible-duration-" + generateRandomID(),
			Budget: &BudgetRequest{
				MaxLimit:      10.0,
				ResetDuration: "90m", // 90 minutes - flexible format
			},
			ProviderConfigs: []ProviderConfigRequest{
				{
					Provider: "openai",
					Weight:   1.0,
					RateLimit: &CreateRateLimitRequest{
						TokenMaxLimit:        int64Ptr(5000),
						TokenResetDuration:   strPtr("4h"), // 4 hours - flexible format
						RequestMaxLimit:      int64Ptr(100),
						RequestResetDuration: strPtr("45m"), // 45 minutes - flexible format
					},
				},
			},
		},
	})

	if createVKResp.StatusCode != 200 {
		t.Fatalf("Failed to create VK with flexible duration: status %d, body: %v",
			createVKResp.StatusCode, createVKResp.Body)
	}

	vkID := ExtractIDFromResponse(t, createVKResp)
	testData.AddVirtualKey(vkID)

	t.Log("Successfully created VK with flexible duration formats (90m, 4h, 45m)")

	// Verify the VK was created with correct duration values
	getVKResp := MakeRequest(t, APIRequest{
		Method: "GET",
		Path:   "/api/governance/virtual-keys/" + vkID,
	})

	if getVKResp.StatusCode != 200 {
		t.Fatalf("Failed to retrieve VK: status %d", getVKResp.StatusCode)
	}

	vk := getVKResp.Body["virtual_key"].(map[string]interface{})

	// Check budget reset duration
	if budget, ok := vk["budget"].(map[string]interface{}); ok {
		if resetDuration, ok := budget["reset_duration"].(string); ok {
			if resetDuration != "90m" {
				t.Errorf("Expected budget reset_duration to be '90m', got '%s'", resetDuration)
			}
			t.Logf("Budget reset duration correctly stored as: %s", resetDuration)
		}
	}

	// Check rate limit reset durations
	if providerConfigs, ok := vk["provider_configs"].([]interface{}); ok && len(providerConfigs) > 0 {
		if config, ok := providerConfigs[0].(map[string]interface{}); ok {
			if rateLimit, ok := config["rate_limit"].(map[string]interface{}); ok {
				if tokenResetDuration, ok := rateLimit["token_reset_duration"].(string); ok {
					if tokenResetDuration != "4h" {
						t.Errorf("Expected token_reset_duration to be '4h', got '%s'", tokenResetDuration)
					}
					t.Logf("Token reset duration correctly stored as: %s", tokenResetDuration)
				}
				if requestResetDuration, ok := rateLimit["request_reset_duration"].(string); ok {
					if requestResetDuration != "45m" {
						t.Errorf("Expected request_reset_duration to be '45m', got '%s'", requestResetDuration)
					}
					t.Logf("Request reset duration correctly stored as: %s", requestResetDuration)
				}
			}
		}
	}
}
