package governance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/maximhq/bifrost/core/providers/gemini"
	"github.com/maximhq/bifrost/core/schemas"
	configstoreTables "github.com/maximhq/bifrost/framework/configstore/tables"
)

type RoutingProfileStrategy string

const (
	RoutingProfileStrategyOrdered  RoutingProfileStrategy = "ordered_failover"
	RoutingProfileStrategyWeighted RoutingProfileStrategy = "weighted"
)

type RoutingProfile struct {
	ID              string                 `json:"id,omitempty"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	VirtualProvider string                 `json:"virtual_provider"`
	VirtualKeyID    string                 `json:"virtual_key_id,omitempty"` // Route requests with this virtual key to this profile
	Enabled         bool                   `json:"enabled"`
	Strategy        RoutingProfileStrategy `json:"strategy,omitempty"`
	Targets         []RoutingProfileTarget `json:"targets"`
}

type RoutingProfileTarget struct {
	Provider     string    `json:"provider"`
	VirtualModel string    `json:"virtual_model,omitempty"`
	Model        string    `json:"model,omitempty"`
	Priority     int       `json:"priority,omitempty"`
	Weight       *float64  `json:"weight,omitempty"`
	RequestTypes []string  `json:"request_types,omitempty"`
	Capabilities []string  `json:"capabilities,omitempty"`
	Enabled      bool      `json:"enabled"`
	RateLimit    *RateHint `json:"rate_limit,omitempty"`
}

type RateHint struct {
	RequestPercentThreshold *float64 `json:"request_percent_threshold,omitempty"`
	TokenPercentThreshold   *float64 `json:"token_percent_threshold,omitempty"`
	BudgetPercentThreshold  *float64 `json:"budget_percent_threshold,omitempty"`
}

type profileCandidate struct {
	provider string
	model    string
	priority int
	weight   float64
}

func (p *GovernancePlugin) applyRoutingProfiles(ctx *schemas.BifrostContext, req *schemas.HTTPRequest, body map[string]any, virtualKey *configstoreTables.TableVirtualKey) (map[string]any, bool, error) {
	if len(p.getRoutingProfiles()) == 0 {
		return body, false, nil
	}

	var profile *RoutingProfile
	var genaiRequestSuffix string

	// First, check if there's a virtual key being used that has a routing profile
	var baseModel string
	if virtualKey != nil && virtualKey.ID != "" {
		profile = p.findRoutingProfileByVirtualKeyID(virtualKey.ID)
		if profile != nil {
			p.logger.Debug("[RoutingProfile] Found profile %s for virtual key ID %s", profile.Name, virtualKey.ID)
		}
	}

	// If no virtual key profile found, try to find by model alias
	if profile == nil {
		modelValue, hasModel := body["model"]
		if !hasModel {
			if strings.Contains(req.Path, "/genai") {
				modelValue = req.CaseInsensitivePathParamLookup("model")
			} else {
				return body, false, nil
			}
		}

		modelStr, ok := modelValue.(string)
		if !ok || modelStr == "" {
			return body, false, nil
		}

		if strings.Contains(req.Path, "/genai") {
			for _, sfx := range gemini.GeminiRequestSuffixPaths {
				if before, ok := strings.CutSuffix(modelStr, sfx); ok {
					modelStr = before
					genaiRequestSuffix = sfx
					break
				}
			}
		}

		providerAlias, bm := schemas.ParseModelString(modelStr, "")
		baseModel = bm
		if providerAlias == "" {
			return body, false, nil
		}

		profile = p.findRoutingProfile(providerAlias)
	}

	if profile == nil {
		return body, false, nil
	}

	requestType := ""
	if val := ctx.Value(schemas.BifrostContextKeyHTTPRequestType); val != nil {
		if requestTypeEnum, ok := val.(schemas.RequestType); ok {
			requestType = string(requestTypeEnum)
		} else if requestTypeStr, ok := val.(string); ok {
			requestType = requestTypeStr
		}
	}

	capabilities := extractRequestCapabilities(body)
	candidates, rejectionReasons := p.profileCandidates(ctx, profile, baseModel, requestType, capabilities, virtualKey)
	if len(candidates) == 0 {
		ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-rejections"), rejectionReasons)
		return body, false, nil
	}

	primary := candidates[0]
	if strings.Contains(req.Path, "/genai") {
		ctx.SetValue("model", primary.provider+"/"+primary.model+genaiRequestSuffix)
	} else {
		body["model"] = primary.provider + "/" + primary.model
	}

	fallbacks := make([]string, 0, len(candidates)-1)
	for i := 1; i < len(candidates); i++ {
		fallbacks = append(fallbacks, candidates[i].provider+"/"+candidates[i].model)
	}
	if len(fallbacks) > 0 {
		body["fallbacks"] = fallbacks
	}

	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-name"), profile.Name)
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-id"), profile.ID)
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-primary"), primary.provider+"/"+primary.model)
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-candidates"), len(candidates))
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-fallback-count"), len(fallbacks))
	ctx.SetValue(schemas.BifrostContextKey("bf-governance-routing-profile-rejections"), rejectionReasons)

	return body, true, nil
}

func (p *GovernancePlugin) getRoutingProfiles() []RoutingProfile {
	if p == nil {
		return nil
	}
	return p.routingProfiles
}

func (p *GovernancePlugin) findRoutingProfile(alias schemas.ModelProvider) *RoutingProfile {
	profiles := p.getRoutingProfiles()
	for i := range profiles {
		profile := &profiles[i]
		if !profile.Enabled {
			continue
		}
		if strings.EqualFold(profile.VirtualProvider, string(alias)) {
			return profile
		}
	}
	return nil
}

func (p *GovernancePlugin) findRoutingProfileByVirtualKeyID(virtualKeyID string) *RoutingProfile {
	if virtualKeyID == "" {
		return nil
	}
	profiles := p.getRoutingProfiles()
	for i := range profiles {
		profile := &profiles[i]
		if !profile.Enabled {
			continue
		}
		if profile.VirtualKeyID != "" && strings.EqualFold(profile.VirtualKeyID, virtualKeyID) {
			return profile
		}
	}
	return nil
}

func (p *GovernancePlugin) profileCandidates(ctx *schemas.BifrostContext, profile *RoutingProfile, baseModel, requestType string, capabilities []string, virtualKey *configstoreTables.TableVirtualKey) ([]profileCandidate, map[string]int) {
	if profile == nil {
		return nil, map[string]int{"profile_missing": 1}
	}

	rejectionReasons := map[string]int{}
	out := make([]profileCandidate, 0, len(profile.Targets))
	for _, target := range profile.Targets {
		if !target.Enabled || target.Provider == "" {
			rejectionReasons["target_disabled_or_provider_missing"]++
			continue
		}
		if !matchesVirtualModel(target.VirtualModel, baseModel) {
			rejectionReasons["virtual_model_mismatch"]++
			continue
		}
		if len(target.RequestTypes) > 0 && requestType != "" && !containsFold(target.RequestTypes, requestType) {
			rejectionReasons["request_type_mismatch"]++
			continue
		}
		if len(target.Capabilities) > 0 && !matchesCapabilities(target.Capabilities, capabilities) {
			rejectionReasons["capability_mismatch"]++
			continue
		}

		provider := schemas.ModelProvider(target.Provider)
		if p.inMemoryStore != nil {
			if _, ok := p.inMemoryStore.GetConfiguredProviders()[provider]; !ok {
				rejectionReasons["provider_not_configured"]++
				continue
			}
		}

		candidateModel := target.Model
		if candidateModel == "" {
			candidateModel = baseModel
		}
		if p.modelCatalog != nil {
			refined, err := p.modelCatalog.RefineModelForProvider(provider, candidateModel)
			if err != nil {
				rejectionReasons["model_refine_failed"]++
				p.logger.Debug("[RoutingProfile] skip target %s/%s refine failed: %v", target.Provider, candidateModel, err)
				continue
			}
			candidateModel = refined
		}

		status := p.store.GetBudgetAndRateLimitStatus(ctx, candidateModel, provider, virtualKey, nil, nil, nil)
		if status != nil {
			if !withinThreshold(status.RateLimitRequestPercentUsed, target.RateLimit, "request") {
				rejectionReasons["request_threshold_exceeded"]++
				continue
			}
			if !withinThreshold(status.RateLimitTokenPercentUsed, target.RateLimit, "token") {
				rejectionReasons["token_threshold_exceeded"]++
				continue
			}
			if !withinThreshold(status.BudgetPercentUsed, target.RateLimit, "budget") {
				rejectionReasons["budget_threshold_exceeded"]++
				continue
			}
		}

		weight := 1.0
		if target.Weight != nil {
			weight = *target.Weight
		}
		out = append(out, profileCandidate{
			provider: target.Provider,
			model:    candidateModel,
			priority: target.Priority,
			weight:   weight,
		})
	}

	if len(out) == 0 {
		return out, rejectionReasons
	}

	strategy := profile.Strategy
	if strategy == "" {
		strategy = RoutingProfileStrategyOrdered
	}

	switch strategy {
	case RoutingProfileStrategyWeighted:
		sort.SliceStable(out, func(i, j int) bool {
			if out[i].weight == out[j].weight {
				return out[i].priority < out[j].priority
			}
			return out[i].weight > out[j].weight
		})
	default:
		sort.SliceStable(out, func(i, j int) bool {
			if out[i].priority == out[j].priority {
				return out[i].weight > out[j].weight
			}
			return out[i].priority < out[j].priority
		})
	}

	return out, rejectionReasons
}

func containsFold(values []string, value string) bool {
	for _, item := range values {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func withinThreshold(percent float64, hint *RateHint, metric string) bool {
	if hint == nil {
		return percent < 100
	}
	var threshold *float64
	switch metric {
	case "request":
		threshold = hint.RequestPercentThreshold
	case "token":
		threshold = hint.TokenPercentThreshold
	case "budget":
		threshold = hint.BudgetPercentThreshold
	default:
		return percent < 100
	}
	if threshold == nil {
		return percent < 100
	}
	if *threshold < 0 || *threshold > 100 {
		return false
	}
	return percent < *threshold
}

func matchesCapabilities(targetCapabilities []string, requestCapabilities []string) bool {
	if len(targetCapabilities) == 0 {
		return true
	}
	if len(requestCapabilities) == 0 {
		return false
	}
	for _, capability := range targetCapabilities {
		if containsFold(requestCapabilities, capability) {
			return true
		}
	}
	return false
}

func extractRequestCapabilities(body map[string]any) []string {
	capabilities := []string{"text"}

	messages, ok := body["messages"].([]any)
	if !ok {
		return capabilities
	}

	for _, messageAny := range messages {
		message, ok := messageAny.(map[string]any)
		if !ok {
			continue
		}
		contentItems, ok := message["content"].([]any)
		if !ok {
			continue
		}
		for _, contentAny := range contentItems {
			content, ok := contentAny.(map[string]any)
			if !ok {
				continue
			}
			typeValue := strings.ToLower(strings.TrimSpace(fmt.Sprint(content["type"])))
			if typeValue == "image_url" || typeValue == "input_image" || typeValue == "image" {
				if !containsFold(capabilities, "vision") {
					capabilities = append(capabilities, "vision")
				}
			}
		}
	}

	return capabilities
}

func matchesVirtualModel(targetVirtualModel string, requestedVirtualModel string) bool {
	if strings.TrimSpace(targetVirtualModel) == "" {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(targetVirtualModel), "*") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(targetVirtualModel), strings.TrimSpace(requestedVirtualModel))
}

func (p *GovernancePlugin) validateRoutingProfiles() error {
	seenVirtualProviders := map[string]struct{}{}
	seenProfileNames := map[string]struct{}{}
	realProviders := map[string]struct{}{}
	if p.inMemoryStore != nil {
		for provider := range p.inMemoryStore.GetConfiguredProviders() {
			realProviders[strings.ToLower(strings.TrimSpace(string(provider)))] = struct{}{}
		}
	}

	for _, profile := range p.routingProfiles {
		if strings.TrimSpace(profile.Name) == "" {
			return fmt.Errorf("routing profile name is required")
		}
		nameKey := strings.ToLower(strings.TrimSpace(profile.Name))
		if _, exists := seenProfileNames[nameKey]; exists {
			return fmt.Errorf("routing profile name %s must be unique", profile.Name)
		}
		seenProfileNames[nameKey] = struct{}{}
		virtualProvider := strings.ToLower(strings.TrimSpace(profile.VirtualProvider))
		if virtualProvider == "" {
			return fmt.Errorf("routing profile %s virtual_provider is required", profile.Name)
		}
		if virtualProvider == "*" {
			return fmt.Errorf("routing profile %s virtual_provider '*' is reserved", profile.Name)
		}
		if _, exists := realProviders[virtualProvider]; exists {
			return fmt.Errorf("routing profile %s virtual_provider %s conflicts with a configured real provider", profile.Name, profile.VirtualProvider)
		}
		if _, exists := seenVirtualProviders[virtualProvider]; exists {
			return fmt.Errorf("routing profile virtual_provider %s must be unique", profile.VirtualProvider)
		}
		seenVirtualProviders[virtualProvider] = struct{}{}

		if len(profile.Targets) == 0 {
			return fmt.Errorf("routing profile %s must define at least one target", profile.Name)
		}
		if profile.Strategy != "" && profile.Strategy != RoutingProfileStrategyOrdered && profile.Strategy != RoutingProfileStrategyWeighted {
			return fmt.Errorf("routing profile %s has invalid strategy %s", profile.Name, profile.Strategy)
		}

		hasWildcardVirtualModel := false
		hasNamedVirtualModel := false
		seenVirtualModels := map[string]struct{}{}
		for _, target := range profile.Targets {
			if strings.TrimSpace(target.Provider) == "" {
				return fmt.Errorf("routing profile %s target provider is required", profile.Name)
			}
			if strings.TrimSpace(target.VirtualModel) != "" && strings.TrimSpace(target.Model) == "" {
				return fmt.Errorf("routing profile %s target for virtual_model %s must define model", profile.Name, target.VirtualModel)
			}
			if strings.EqualFold(strings.TrimSpace(target.VirtualModel), "*") {
				hasWildcardVirtualModel = true
			}
			if vm := strings.TrimSpace(target.VirtualModel); vm != "" && vm != "*" {
				hasNamedVirtualModel = true
				key := strings.ToLower(vm)
				if _, exists := seenVirtualModels[key]; exists {
					return fmt.Errorf("routing profile %s has duplicate virtual_model alias %s", profile.Name, vm)
				}
				seenVirtualModels[key] = struct{}{}
			}
		}
		if hasWildcardVirtualModel && hasNamedVirtualModel {
			return fmt.Errorf("routing profile %s mixes wildcard virtual_model '*' with named virtual models", profile.Name)
		}
	}
	return nil
}
