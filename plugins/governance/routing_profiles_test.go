package governance

import (
	"testing"

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
