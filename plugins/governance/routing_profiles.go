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

	genaiRequestSuffix := ""
	if strings.Contains(req.Path, "/genai") {
		for _, sfx := range gemini.GeminiRequestSuffixPaths {
			if before, ok := strings.CutSuffix(modelStr, sfx); ok {
				modelStr = before
				genaiRequestSuffix = sfx
				break
			}
		}
	}

	providerAlias, baseModel := schemas.ParseModelString(modelStr, "")
	if providerAlias == "" {
		return body, false, nil
	}

	profile := p.findRoutingProfile(providerAlias)
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

	candidates := p.profileCandidates(ctx, profile, baseModel, requestType, virtualKey)
	if len(candidates) == 0 {
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

func (p *GovernancePlugin) profileCandidates(ctx *schemas.BifrostContext, profile *RoutingProfile, baseModel, requestType string, virtualKey *configstoreTables.TableVirtualKey) []profileCandidate {
	if profile == nil {
		return nil
	}

	out := make([]profileCandidate, 0, len(profile.Targets))
	for _, target := range profile.Targets {
		if !target.Enabled || target.Provider == "" {
			continue
		}
		if !matchesVirtualModel(target.VirtualModel, baseModel) {
			continue
		}
		if len(target.RequestTypes) > 0 && requestType != "" && !containsFold(target.RequestTypes, requestType) {
			continue
		}

		provider := schemas.ModelProvider(target.Provider)
		if p.inMemoryStore != nil {
			if _, ok := p.inMemoryStore.GetConfiguredProviders()[provider]; !ok {
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
				p.logger.Debug("[RoutingProfile] skip target %s/%s refine failed: %v", target.Provider, candidateModel, err)
				continue
			}
			candidateModel = refined
		}

		status := p.store.GetBudgetAndRateLimitStatus(ctx, candidateModel, provider, virtualKey, nil, nil, nil)
		if status != nil {
			if !withinThreshold(status.RateLimitRequestPercentUsed, target.RateLimit, "request") {
				continue
			}
			if !withinThreshold(status.RateLimitTokenPercentUsed, target.RateLimit, "token") {
				continue
			}
			if !withinThreshold(status.BudgetPercentUsed, target.RateLimit, "budget") {
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
		return out
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

	return out
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

		hasWildcardVirtualModel := false
		hasNamedVirtualModel := false
		seenVirtualModels := map[string]struct{}{}
		for _, target := range profile.Targets {
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
