package tables

import (
	"strings"
	"time"

	"github.com/bytedance/sonic"
	bifrost "github.com/maximhq/bifrost/core"
	"gorm.io/gorm"
)

type TableRoutingProfileTarget struct {
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

// TableRoutingProfile persists virtual provider/model routing profile configuration.
type TableRoutingProfile struct {
	ID              string `gorm:"primaryKey;type:varchar(255)" json:"id"`
	ConfigHash      string `gorm:"type:varchar(255)" json:"config_hash"`
	Name            string `gorm:"type:varchar(255);not null;uniqueIndex:idx_routing_profile_name" json:"name"`
	Description     string `gorm:"type:text" json:"description"`
	VirtualProvider string `gorm:"type:varchar(255);not null;uniqueIndex:idx_routing_profile_virtual_provider" json:"virtual_provider"`
	Enabled         bool   `gorm:"not null;default:true" json:"enabled"`
	Strategy        string `gorm:"type:varchar(64);not null;default:'ordered_failover'" json:"strategy"`

	Targets       *string                     `gorm:"type:text" json:"-"`
	ParsedTargets []TableRoutingProfileTarget `gorm:"-" json:"targets"`

	CreatedAt time.Time `gorm:"index;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"index;not null" json:"updated_at"`
}

func (TableRoutingProfile) TableName() string { return "routing_profiles" }

func (r *TableRoutingProfile) BeforeSave(tx *gorm.DB) error {
	if len(r.ParsedTargets) > 0 {
		data, err := sonic.Marshal(r.ParsedTargets)
		if err != nil {
			return err
		}
		r.Targets = bifrost.Ptr(string(data))
	} else {
		r.Targets = nil
	}
	return nil
}

func (r *TableRoutingProfile) AfterFind(tx *gorm.DB) error {
	if r.Targets == nil || strings.TrimSpace(*r.Targets) == "" {
		return nil
	}
	if err := sonic.Unmarshal([]byte(*r.Targets), &r.ParsedTargets); err != nil {
		return err
	}
	return nil
}
