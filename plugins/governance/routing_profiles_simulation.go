package governance

import (
	"fmt"
	"sort"
	"strings"

	"github.com/maximhq/bifrost/core/schemas"
)

type SimulatedRoutingCandidate struct {
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
	Priority int     `json:"priority"`
	Weight   float64 `json:"weight"`
}

type SimulatedRoutingDecision struct {
	Profile    *RoutingProfile             `json:"profile"`
	Primary    string                      `json:"primary"`
	Fallbacks  []string                    `json:"fallbacks"`
	Candidates []SimulatedRoutingCandidate `json:"candidates"`
}

func SimulateRoutingProfileDecision(profiles []RoutingProfile, model string, requestType string, capabilities []string) (*SimulatedRoutingDecision, error) {
	providerAlias, virtualModel := schemas.ParseModelString(model, "")
	if providerAlias == "" || strings.TrimSpace(virtualModel) == "" {
		return nil, fmt.Errorf("model must be in virtual_provider/virtual_model format")
	}

	var profile *RoutingProfile
	for i := range profiles {
		if !profiles[i].Enabled {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(profiles[i].VirtualProvider), strings.TrimSpace(string(providerAlias))) {
			profile = &profiles[i]
			break
		}
	}
	if profile == nil {
		return nil, fmt.Errorf("no routing profile found for virtual provider %s", providerAlias)
	}

	candidates := make([]SimulatedRoutingCandidate, 0, len(profile.Targets))
	for _, target := range profile.Targets {
		if !target.Enabled || strings.TrimSpace(target.Provider) == "" {
			continue
		}
		if !matchesVirtualModel(profile.VirtualModel, virtualModel) {
			continue
		}
		if len(target.RequestTypes) > 0 && strings.TrimSpace(requestType) != "" && !containsFold(target.RequestTypes, requestType) {
			continue
		}
		if len(target.Capabilities) > 0 && !matchesCapabilities(target.Capabilities, capabilities) {
			continue
		}

		modelName := strings.TrimSpace(target.Model)
		if modelName == "" {
			modelName = strings.TrimSpace(virtualModel)
		}
		weight := 1.0
		if target.Weight != nil {
			weight = *target.Weight
		}

		candidates = append(candidates, SimulatedRoutingCandidate{
			Provider: strings.TrimSpace(target.Provider),
			Model:    modelName,
			Priority: target.Priority,
			Weight:   weight,
		})
	}

	strategy := profile.Strategy
	if strategy == "" {
		strategy = RoutingProfileStrategyOrdered
	}

	switch strategy {
	case RoutingProfileStrategyWeighted:
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Weight == candidates[j].Weight {
				return candidates[i].Priority < candidates[j].Priority
			}
			return candidates[i].Weight > candidates[j].Weight
		})
	default:
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].Priority == candidates[j].Priority {
				return candidates[i].Weight > candidates[j].Weight
			}
			return candidates[i].Priority < candidates[j].Priority
		})
	}

	primary := ""
	fallbacks := make([]string, 0, len(candidates))
	for i, candidate := range candidates {
		modelRef := candidate.Provider + "/" + candidate.Model
		if i == 0 {
			primary = modelRef
			continue
		}
		fallbacks = append(fallbacks, modelRef)
	}

	return &SimulatedRoutingDecision{
		Profile:    profile,
		Primary:    primary,
		Fallbacks:  fallbacks,
		Candidates: candidates,
	}, nil
}
