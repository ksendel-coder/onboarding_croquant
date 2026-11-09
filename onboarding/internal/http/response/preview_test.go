package response

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/stretchr/testify/require"
)

func TestNewPreviewIncludesDraftHintsAndSettings(t *testing.T) {
	delay := 1500
	id := uuid.New()
	versionId := uuid.New()
	result := NewPreview(&domains.Tour{
		Id:            id,
		VersionId:     versionId,
		Version:       2,
		Status:        domains.TourDraft,
		TriggerType:   domains.TriggerDelay,
		TriggerConfig: &domains.TriggerConfig{DelayMs: &delay},
		Audience:      &domains.Audience{MaxShows: 3},
		Hints:         []domains.Hint{{Id: uuid.New(), Step: 1, Title: "Шаг"}},
		StepsCount:    1,
	})
	require.True(t, result.Preview)
	require.Equal(t, "draft", result.Tour.Status)
	require.Equal(t, versionId, result.Tour.VersionId)
	require.Len(t, result.Tour.Hints, 1)
	require.Equal(t, 1500, *result.TriggerConfig.DelayMs)
	require.Equal(t, 3, result.Audience.MaxShows)
}
