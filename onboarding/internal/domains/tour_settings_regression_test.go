package domains

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type settingsBoundaryRepo struct {
	TourRepositoryInterface
	calls     int
	lastID    uuid.UUID
	lastPatch TourSettingsPatch
	err       error
}

func (s *settingsBoundaryRepo) UpdateSettings(_ context.Context, id uuid.UUID, patch TourSettingsPatch) error {
	s.calls++
	s.lastID, s.lastPatch = id, patch
	return s.err
}

func TestTourSettingsUnitAllowedTriggerTypes(t *testing.T) {
	for _, trigger := range []TriggerType{
		TriggerOnLoad, TriggerDelay, TriggerExitIntent, TriggerManual,
		TriggerScrollDepth, TriggerInactivity, TriggerElementVisible,
	} {
		t.Run(string(trigger), func(t *testing.T) {
			repo := &settingsBoundaryRepo{}
			id := uuid.New()
			patch := TourSettingsPatch{TriggerType: &trigger}
			require.NoError(t, NewTourDomain(repo).UpdateSettings(context.Background(), id, patch))
			require.Equal(t, 1, repo.calls)
			require.Equal(t, id, repo.lastID)
			require.Equal(t, trigger, *repo.lastPatch.TriggerType)
		})
	}
}

func TestTourSettingsUnitBoundaryValues(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		maxShows, delay, scroll, inactivity int
	}{
		{"minimum", 0, 0, 0, 0},
		{"maximum", 25, 600000, 100, 86400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &settingsBoundaryRepo{}
			path, selector := "/dashboard", "#start"
			once, onlyNew := true, true
			patch := TourSettingsPatch{
				TargetPath:       &path,
				TriggerConfig:    &TriggerConfig{DelayMs: &tc.delay, ScrollDepth: &tc.scroll, InactivitySecs: &tc.inactivity, ElementSelector: &selector},
				AudienceShowOnce: &once,
				AudienceOnlyNew:  &onlyNew,
				AudienceMaxShows: &tc.maxShows,
			}
			require.NoError(t, NewTourDomain(repo).UpdateSettings(context.Background(), uuid.New(), patch))
			require.Equal(t, 1, repo.calls)
			require.Equal(t, patch, repo.lastPatch)
		})
	}
}

func TestTourSettingsUnitRejectsNegativeBoundaries(t *testing.T) {
	negative := -1
	for _, tc := range []struct {
		name  string
		patch TourSettingsPatch
	}{
		{"negative delay", TourSettingsPatch{TriggerConfig: &TriggerConfig{DelayMs: &negative}}},
		{"negative scroll", TourSettingsPatch{TriggerConfig: &TriggerConfig{ScrollDepth: &negative}}},
		{"negative inactivity", TourSettingsPatch{TriggerConfig: &TriggerConfig{InactivitySecs: &negative}}},
		{"negative shows", TourSettingsPatch{AudienceMaxShows: &negative}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &settingsBoundaryRepo{}
			err := NewTourDomain(repo).UpdateSettings(context.Background(), uuid.New(), tc.patch)
			requireAuthHTTPCode(t, err, http.StatusBadRequest)
			require.Zero(t, repo.calls)
		})
	}
}

func TestTourSettingsUnitPropagatesRepositoryError(t *testing.T) {
	sentinel := errors.New("storage failed")
	repo := &settingsBoundaryRepo{err: sentinel}
	path := "/settings"
	err := NewTourDomain(repo).UpdateSettings(context.Background(), uuid.New(), TourSettingsPatch{TargetPath: &path})
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, 1, repo.calls)
}
