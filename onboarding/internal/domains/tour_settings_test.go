package domains

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type settingsTourRepositoryStub struct {
	TourRepositoryInterface
	called bool
	patch  TourSettingsPatch
}

func (s *settingsTourRepositoryStub) UpdateSettings(_ context.Context, _ uuid.UUID, patch TourSettingsPatch) error {
	s.called = true
	s.patch = patch
	return nil
}

func TestTourUpdateSettingsValidation(t *testing.T) {
	blank := "  "
	negative := -1
	tooLargeDelay := 600001
	tooLargeScroll := 101
	tooLargeInactivity := 86401
	invalidTrigger := TriggerType("unknown")

	tests := []struct {
		name  string
		patch TourSettingsPatch
	}{
		{name: "blank target path", patch: TourSettingsPatch{TargetPath: &blank}},
		{name: "negative max shows", patch: TourSettingsPatch{AudienceMaxShows: &negative}},
		{name: "invalid trigger type", patch: TourSettingsPatch{TriggerType: &invalidTrigger}},
		{name: "delay above limit", patch: TourSettingsPatch{TriggerConfig: &TriggerConfig{DelayMs: &tooLargeDelay}}},
		{name: "scroll above limit", patch: TourSettingsPatch{TriggerConfig: &TriggerConfig{ScrollDepth: &tooLargeScroll}}},
		{name: "inactivity above limit", patch: TourSettingsPatch{TriggerConfig: &TriggerConfig{InactivitySecs: &tooLargeInactivity}}},
		{name: "blank selector", patch: TourSettingsPatch{TriggerConfig: &TriggerConfig{ElementSelector: &blank}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repository := &settingsTourRepositoryStub{}
			domain := NewTourDomain(repository)
			err := domain.UpdateSettings(context.Background(), uuid.New(), tc.patch)

			var logicError *LogicError
			require.ErrorAs(t, err, &logicError)
			assert.Equal(t, http.StatusBadRequest, logicError.Code)
			assert.False(t, repository.called)
		})
	}
}

func TestTourUpdateSettingsDelegatesValidPatch(t *testing.T) {
	targetPath := "/dashboard"
	triggerType := TriggerElementVisible
	delayMs := 1500
	scrollDepth := 60
	inactivity := 30
	selector := "#welcome"
	maxShows := 2
	patch := TourSettingsPatch{
		TargetPath:       &targetPath,
		TriggerType:      &triggerType,
		TriggerConfig:    &TriggerConfig{DelayMs: &delayMs, ScrollDepth: &scrollDepth, InactivitySecs: &inactivity, ElementSelector: &selector},
		AudienceMaxShows: &maxShows,
	}
	repository := &settingsTourRepositoryStub{}
	domain := NewTourDomain(repository)

	require.NoError(t, domain.UpdateSettings(context.Background(), uuid.New(), patch))
	require.True(t, repository.called)
	assert.Equal(t, patch, repository.patch)
}
