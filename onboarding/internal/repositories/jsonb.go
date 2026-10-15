package repositories

import (
	"github.com/Vanady39/cluer/onboarding/internal/domains"
)

// The two shapes below are the on-disk contract of the jsonb columns
// trigger_config and audience_rules. They are deliberately separate from the
// domain structs they mirror: the domain is free to be renamed or restructured,
// and rows already written stay readable because this file — and only this
// file — decides what the database sees.
//
// Changing a tag here is a data migration. Changing a field name in
// domains.TriggerConfig is not.

type triggerConfigRow struct {
	ScrollDepth     *int    `json:"scroll_depth,omitempty"`
	InactivitySecs  *int    `json:"inactivity_secs,omitempty"`
	ElementSelector *string `json:"element_selector,omitempty"`
	DelayMs         *int    `json:"delay_ms,omitempty"`
}

type audienceRuleRow struct {
	Type      string `json:"type"`
	Key       string `json:"key"`
	Operator  string `json:"operator"`
	Value     any    `json:"value,omitempty"`
	Timeframe string `json:"timeframe,omitempty"`
}

func toTriggerConfigRow(cfg *domains.TriggerConfig) *triggerConfigRow {
	if cfg == nil {
		return nil
	}
	return &triggerConfigRow{
		ScrollDepth:     cfg.ScrollDepth,
		InactivitySecs:  cfg.InactivitySecs,
		ElementSelector: cfg.ElementSelector,
		DelayMs:         cfg.DelayMs,
	}
}

func (r *triggerConfigRow) toDomain() *domains.TriggerConfig {
	if r == nil {
		return nil
	}
	return &domains.TriggerConfig{
		ScrollDepth:     r.ScrollDepth,
		InactivitySecs:  r.InactivitySecs,
		ElementSelector: r.ElementSelector,
		DelayMs:         r.DelayMs,
	}
}

// toAudienceRuleRows keeps the nil/empty distinction intact: a nil slice must
// stay nil so marshalJSON writes SQL NULL, while an empty non-nil slice has to
// reach the column as [].
func toAudienceRuleRows(rules []domains.AudienceRule) []audienceRuleRow {
	if rules == nil {
		return nil
	}
	rows := make([]audienceRuleRow, 0, len(rules))
	for _, rule := range rules {
		rows = append(rows, audienceRuleRow{
			Type:      rule.Type,
			Key:       rule.Key,
			Operator:  rule.Operator,
			Value:     rule.Value,
			Timeframe: rule.Timeframe,
		})
	}
	return rows
}

func audienceRuleRowsToDomain(rows []audienceRuleRow) []domains.AudienceRule {
	if rows == nil {
		return nil
	}
	rules := make([]domains.AudienceRule, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, domains.AudienceRule{
			Type:      row.Type,
			Key:       row.Key,
			Operator:  row.Operator,
			Value:     row.Value,
			Timeframe: row.Timeframe,
		})
	}
	return rules
}

func marshalTriggerConfig(cfg *domains.TriggerConfig) ([]byte, error) {
	return marshalJSON(toTriggerConfigRow(cfg))
}

func marshalAudienceRules(rules []domains.AudienceRule) ([]byte, error) {
	return marshalJSON(toAudienceRuleRows(rules))
}

// unmarshalTriggerConfig leaves dst untouched when the column was NULL, which
// is what the caller relies on to tell "no config" from "empty config".
func unmarshalTriggerConfig(data []byte, dst **domains.TriggerConfig) error {
	if len(data) == 0 {
		return nil
	}
	var row *triggerConfigRow
	if err := unmarshalJSONB(data, &row); err != nil {
		return err
	}
	*dst = row.toDomain()
	return nil
}

func unmarshalAudienceRules(data []byte, dst *[]domains.AudienceRule) error {
	if len(data) == 0 {
		return nil
	}
	var rows []audienceRuleRow
	if err := unmarshalJSONB(data, &rows); err != nil {
		return err
	}
	*dst = audienceRuleRowsToDomain(rows)
	return nil
}
