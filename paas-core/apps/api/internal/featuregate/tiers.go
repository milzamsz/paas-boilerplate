package featuregate

import "encoding/json"

// TierSlug identifies a built-in subscription tier.
const (
	TierFree       = "free"
	TierPro        = "pro"
	TierEnterprise = "enterprise"
)

// PlanDefaults holds the default limits + feature flags for each tier.
type PlanDefaults struct {
	Name           string
	Slug           string
	PriceMonthly   int64
	PriceYearly    int64
	Currency       string
	MaxProjects    int // -1 = unlimited
	MaxDeployments int
	MaxMembers     int
	Features       []string // feature flag names
}

// DefaultPlans returns the three built-in tier definitions.
func DefaultPlans() []PlanDefaults {
	return []PlanDefaults{
		{
			Name:           "Free",
			Slug:           TierFree,
			PriceMonthly:   0,
			PriceYearly:    0,
			Currency:       "IDR",
			MaxProjects:    1,
			MaxDeployments: 5,
			MaxMembers:     1,
			Features:       []string{},
		},
		{
			Name:           "Pro",
			Slug:           TierPro,
			PriceMonthly:   299000,  // IDR 299k
			PriceYearly:    2990000, // IDR 2.99M
			Currency:       "IDR",
			MaxProjects:    10,
			MaxDeployments: 50,
			MaxMembers:     10,
			Features:       []string{"custom_domain", "priority_support"},
		},
		{
			Name:           "Enterprise",
			Slug:           TierEnterprise,
			PriceMonthly:   999000,  // IDR 999k
			PriceYearly:    9990000, // IDR 9.99M
			Currency:       "IDR",
			MaxProjects:    -1, // unlimited
			MaxDeployments: -1,
			MaxMembers:     -1,
			Features:       []string{"custom_domain", "priority_support", "sso", "audit_logs", "sla"},
		},
	}
}

// FreeTierLimits returns the default limits when no subscription exists.
var FreeTierLimits = struct {
	MaxProjects    int
	MaxDeployments int
	MaxMembers     int
}{1, 5, 1}

// MarshalFeatures converts a feature list to JSONB-compatible string.
func MarshalFeatures(features []string) string {
	b, _ := json.Marshal(features)
	return string(b)
}

// UnmarshalFeatures parses a JSONB feature list.
func UnmarshalFeatures(raw string) []string {
	if raw == "" || raw == "null" {
		return nil
	}
	var features []string
	_ = json.Unmarshal([]byte(raw), &features)
	return features
}
