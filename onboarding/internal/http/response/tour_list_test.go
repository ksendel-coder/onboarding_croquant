package response

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/onboarding/internal/domains"
	"github.com/stretchr/testify/require"
)

func TestTourListFlagsJSON(t *testing.T) {
	for _, tc := range []struct {
		name               string
		published, changed bool
		status             domains.TourStatus
	}{
		{"draft only", false, true, domains.TourDraft},
		{"clean draft", true, false, domains.TourDraft},
		{"edited draft", true, true, domains.TourDraft},
		{"published only", true, false, domains.TourPublished},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.New()
			out := NewTour(&domains.Tour{VersionId: id, Status: tc.status, StepsCount: 3, HasPublished: tc.published, HasUnpublishedChanges: tc.changed})
			raw, err := json.Marshal(out)
			require.NoError(t, err)
			var payload map[string]any
			require.NoError(t, json.Unmarshal(raw, &payload))
			require.Equal(t, tc.published, payload["has_published"])
			require.Equal(t, tc.changed, payload["has_unpublished_changes"])
			require.Equal(t, string(tc.status), payload["status"])
			require.Equal(t, id.String(), payload["version_id"])
			require.Equal(t, float64(3), payload["steps_count"])
		})
	}
}
