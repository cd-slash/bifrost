package governance

import (
	"reflect"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore"
)

type testInMemoryStore struct {
	providers map[schemas.ModelProvider]configstore.ProviderConfig
}

func (t *testInMemoryStore) GetConfiguredProviders() map[schemas.ModelProvider]configstore.ProviderConfig {
	return t.providers
}

func TestMatchesVirtualModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		targetAlias   string
		requestAlias  string
		expectMatched bool
	}{
		{name: "empty target alias matches all", targetAlias: "", requestAlias: "light", expectMatched: true},
		{name: "wildcard target alias matches all", targetAlias: "*", requestAlias: "light", expectMatched: true},
		{name: "exact alias matches", targetAlias: "light", requestAlias: "light", expectMatched: true},
		{name: "case-insensitive alias matches", targetAlias: "LiGhT", requestAlias: "light", expectMatched: true},
		{name: "different alias does not match", targetAlias: "fast", requestAlias: "light", expectMatched: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := matchesVirtualModel(tt.targetAlias, tt.requestAlias)
			if got != tt.expectMatched {
				t.Fatalf("expected match=%v, got %v", tt.expectMatched, got)
			}
		})
	}
}

func TestValidateRoutingProfilesRejectsConflicts(t *testing.T) {
	t.Parallel()

	plugin := &GovernancePlugin{
		inMemoryStore: &testInMemoryStore{providers: map[schemas.ModelProvider]configstore.ProviderConfig{
			"openai": {},
		}},
		routingProfiles: []RoutingProfile{
			{
				Name:            "Fast",
				VirtualProvider: "fast",
				Enabled:         true,
				Targets: []RoutingProfileTarget{{
					Provider: "cerebras",
					Enabled:  true,
				}},
			},
			{
				Name:            "Fast Duplicate",
				VirtualProvider: "FAST",
				Enabled:         true,
				Targets: []RoutingProfileTarget{{
					Provider: "openai",
					Enabled:  true,
				}},
			},
		},
	}

	if err := plugin.validateRoutingProfiles(); err == nil {
		t.Fatalf("expected duplicate virtual_provider validation error")
	}
}

func TestValidateRoutingProfilesRejectsWildcardMix(t *testing.T) {
	t.Parallel()

	plugin := &GovernancePlugin{
		routingProfiles: []RoutingProfile{{
			Name:            "Light",
			VirtualProvider: "light",
			Enabled:         true,
			Targets: []RoutingProfileTarget{
				{Provider: "cerebras", VirtualModel: "*", Model: "glm-4.7-flash", Enabled: true},
				{Provider: "anthropic", VirtualModel: "light", Model: "claude-3-5-haiku-latest", Enabled: true},
			},
		}},
	}

	if err := plugin.validateRoutingProfiles(); err == nil {
		t.Fatalf("expected wildcard and named virtual_model mix validation error")
	}
}

func TestValidateRoutingProfilesRejectsDuplicateVirtualModelAliases(t *testing.T) {
	t.Parallel()

	plugin := &GovernancePlugin{
		routingProfiles: []RoutingProfile{{
			Name:            "Light",
			VirtualProvider: "light",
			Enabled:         true,
			Targets: []RoutingProfileTarget{
				{Provider: "cerebras", VirtualModel: "light", Model: "glm-4.7-flash", Enabled: true},
				{Provider: "anthropic", VirtualModel: "LIGHT", Model: "claude-3-5-haiku-latest", Enabled: true},
			},
		}},
	}

	if err := plugin.validateRoutingProfiles(); err == nil {
		t.Fatalf("expected duplicate virtual_model alias validation error")
	}
}

func TestValidateRoutingProfilesRejectsInvalidStrategy(t *testing.T) {
	t.Parallel()

	plugin := &GovernancePlugin{
		routingProfiles: []RoutingProfile{{
			Name:            "Light",
			VirtualProvider: "light",
			Strategy:        "round_robin",
			Enabled:         true,
			Targets: []RoutingProfileTarget{{
				Provider: "cerebras",
				Enabled:  true,
			}},
		}},
	}

	if err := plugin.validateRoutingProfiles(); err == nil {
		t.Fatalf("expected invalid strategy validation error")
	}
}

func TestValidateRoutingProfilesRejectsDuplicateNames(t *testing.T) {
	t.Parallel()

	plugin := &GovernancePlugin{
		routingProfiles: []RoutingProfile{
			{
				Name:            "Light",
				VirtualProvider: "light",
				Enabled:         true,
				Targets: []RoutingProfileTarget{{
					Provider: "cerebras",
					Enabled:  true,
				}},
			},
			{
				Name:            "LIGHT",
				VirtualProvider: "fast",
				Enabled:         true,
				Targets: []RoutingProfileTarget{{
					Provider: "openai",
					Enabled:  true,
				}},
			},
		},
	}

	if err := plugin.validateRoutingProfiles(); err == nil {
		t.Fatalf("expected duplicate profile name validation error")
	}
}

func TestRoutingProfilesFromConfigPrefersPluginConfig(t *testing.T) {
	t.Parallel()

	profiles := routingProfilesFromConfig(&Config{
		RoutingProfiles: []RoutingProfile{{
			Name:            "Light",
			VirtualProvider: "light",
			Enabled:         true,
			Targets: []RoutingProfileTarget{{
				Provider: "cerebras",
				Enabled:  true,
			}},
		}},
	}, &configstore.GovernanceConfig{})

	if len(profiles) != 1 {
		t.Fatalf("expected 1 routing profile from plugin config, got %d", len(profiles))
	}
	if profiles[0].VirtualProvider != "light" {
		t.Fatalf("expected virtual provider light, got %s", profiles[0].VirtualProvider)
	}
}

func TestRoutingProfilesFromConfigFallsBackToGovernanceConfig(t *testing.T) {
	t.Parallel()

	governanceConfig := &configstore.GovernanceConfig{}
	field := reflect.ValueOf(governanceConfig).Elem().FieldByName("RoutingProfiles")
	if !field.IsValid() {
		t.Skip("governance config in this module does not expose RoutingProfiles")
	}

	encoded := []map[string]any{{
		"id":               "p1",
		"name":             "Light",
		"virtual_provider": "light",
		"enabled":          true,
		"strategy":         "ordered_failover",
		"targets": []map[string]any{{
			"provider": "cerebras",
			"enabled":  true,
		}},
	}}

	payload, err := sonic.Marshal(encoded)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}
	ptr := reflect.New(field.Type())
	if err := sonic.Unmarshal(payload, ptr.Interface()); err != nil {
		t.Fatalf("failed to unmarshal test data into routing profiles field: %v", err)
	}
	field.Set(ptr.Elem())

	profiles := routingProfilesFromConfig(nil, governanceConfig)
	if len(profiles) != 1 {
		t.Fatalf("expected 1 routing profile from governance config, got %d", len(profiles))
	}
	if profiles[0].VirtualProvider != "light" {
		t.Fatalf("expected virtual provider light, got %s", profiles[0].VirtualProvider)
	}
}

func TestMatchesCapabilities(t *testing.T) {
	t.Parallel()

	if !matchesCapabilities([]string{"vision"}, []string{"text", "vision"}) {
		t.Fatalf("expected capability match for vision")
	}
	if matchesCapabilities([]string{"audio"}, []string{"text", "vision"}) {
		t.Fatalf("did not expect capability match for audio")
	}
}

func TestExtractRequestCapabilitiesDetectsVision(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"messages": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "text", "text": "describe"},
					map[string]any{"type": "image_url", "image_url": map[string]any{"url": "https://example.com/img.png"}},
				},
			},
		},
	}

	capabilities := extractRequestCapabilities(body)
	if !containsFold(capabilities, "text") {
		t.Fatalf("expected text capability")
	}
	if !containsFold(capabilities, "vision") {
		t.Fatalf("expected vision capability")
	}
}

func TestSimulateRoutingProfileDecision(t *testing.T) {
	t.Parallel()

	profiles := []RoutingProfile{{
		Name:            "Light",
		VirtualProvider: "light",
		Enabled:         true,
		Strategy:        RoutingProfileStrategyOrdered,
		Targets: []RoutingProfileTarget{
			{Provider: "cerebras", VirtualModel: "light", Model: "glm-4.7-flash", Priority: 2, Enabled: true},
			{Provider: "anthropic", VirtualModel: "light", Model: "claude-3-5-haiku-latest", Priority: 1, Enabled: true},
		},
	}}

	decision, err := SimulateRoutingProfileDecision(profiles, "light/light", "chat", []string{"text"})
	if err != nil {
		t.Fatalf("unexpected simulation error: %v", err)
	}
	if decision.Primary != "anthropic/claude-3-5-haiku-latest" {
		t.Fatalf("unexpected primary: %s", decision.Primary)
	}
	if len(decision.Fallbacks) != 1 || decision.Fallbacks[0] != "cerebras/glm-4.7-flash" {
		t.Fatalf("unexpected fallbacks: %+v", decision.Fallbacks)
	}
}

func TestSimulateRoutingProfileDecisionErrorsOnUnknownAlias(t *testing.T) {
	t.Parallel()

	profiles := []RoutingProfile{{
		Name:            "Light",
		VirtualProvider: "light",
		Enabled:         true,
		Targets: []RoutingProfileTarget{{
			Provider: "cerebras",
			Enabled:  true,
		}},
	}}

	if _, err := SimulateRoutingProfileDecision(profiles, "fast/light", "chat", []string{"text"}); err == nil {
		t.Fatalf("expected error for unknown virtual provider")
	}
}

func TestSimulateRoutingProfileDecisionErrorsOnInvalidModel(t *testing.T) {
	t.Parallel()

	if _, err := SimulateRoutingProfileDecision(nil, "invalid", "chat", nil); err == nil {
		t.Fatalf("expected error for invalid model format")
	}
}

func TestSetRoutingProfilesValidatesInput(t *testing.T) {
	t.Parallel()

	plugin := &GovernancePlugin{}
	err := plugin.SetRoutingProfiles([]RoutingProfile{{
		Name:            "Invalid",
		VirtualProvider: "",
		Enabled:         true,
		Targets: []RoutingProfileTarget{{
			Provider: "openai",
			Enabled:  true,
		}},
	}})
	if err == nil {
		t.Fatalf("expected validation error when setting invalid routing profiles")
	}
}
