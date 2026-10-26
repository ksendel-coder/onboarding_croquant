package domains

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type runtimeTourRepositoryStub struct {
	TourRepositoryInterface
	candidates []*Tour
}

func (s *runtimeTourRepositoryStub) ResolveCandidates(
	context.Context,
	uuid.UUID,
) ([]*Tour, error) {
	return s.candidates, nil
}

type runtimeHintRepositoryStub struct {
	HintRepositoryInterface
	hints map[uuid.UUID][]Hint
}

func (s *runtimeHintRepositoryStub) ListByVersion(
	_ context.Context,
	versionId uuid.UUID,
) ([]Hint, error) {
	return s.hints[versionId], nil
}

type runtimeRepositoryStub struct {
	RuntimeRepositoryInterface
	progress map[uuid.UUID]*Progress
}

func (s *runtimeRepositoryStub) SubjectProgress(
	context.Context,
	uuid.UUID,
	string,
) (map[uuid.UUID]*Progress, error) {
	return s.progress, nil
}

func (s *runtimeRepositoryStub) UpsertProgress(
	_ context.Context,
	p *Progress,
) error {
	copy := *p
	s.progress[p.TourId] = &copy
	return nil
}

func newRuntimeTestDomain(
	candidates []*Tour,
	hints map[uuid.UUID][]Hint,
) (*RuntimeDomain, *runtimeRepositoryStub) {
	runtime := &runtimeRepositoryStub{
		progress: make(map[uuid.UUID]*Progress),
	}

	return NewRuntimeDomain(
		&runtimeTourRepositoryStub{candidates: candidates},
		&runtimeHintRepositoryStub{hints: hints},
		runtime,
	), runtime
}

func publishedRuntimeTour() *Tour {
	return &Tour{
		Id:         uuid.New(),
		AppId:      uuid.New(),
		VersionId:  uuid.New(),
		Version:    2,
		Status:     TourPublished,
		Enabled:    true,
		TargetPath: "/dashboard",
	}
}

func TestRuntimeResolveReturnsPublishedEnabledTourWithHints(t *testing.T) {
	tour := publishedRuntimeTour()
	firstHintId := uuid.New()
	secondHintId := uuid.New()
	hints := []Hint{
		{Id: firstHintId, TourVersionId: tour.VersionId, Step: 1},
		{Id: secondHintId, TourVersionId: tour.VersionId, Step: 2},
	}
	domain, progress := newRuntimeTestDomain(
		[]*Tour{tour},
		map[uuid.UUID][]Hint{tour.VersionId: hints},
	)

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId:     tour.AppId,
		SubjectId: "user-1",
		Url:       "http://localhost:5173/dashboard",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Tour.Hints, 2)
	assert.Equal(t, firstHintId, result.Tour.Hints[0].Id)
	assert.Equal(t, secondHintId, result.Tour.Hints[1].Id)
	require.NotNil(t, result.CurrentHintId)
	assert.Equal(t, firstHintId, *result.CurrentHintId)

	stored := progress.progress[tour.Id]
	require.NotNil(t, stored)
	assert.Equal(t, ProgressInProgress, stored.Status)
	assert.Equal(t, 1, stored.ShowsCount)
	require.NotNil(t, stored.CurrentHintId)
	assert.Equal(t, firstHintId, *stored.CurrentHintId)
}

func TestRuntimeResolveSkipsUnavailableOrMismatchedTours(t *testing.T) {
	tests := []struct {
		name       string
		candidates func() []*Tour
		url        string
	}{
		{
			name: "draft tour",
			candidates: func() []*Tour {
				tour := publishedRuntimeTour()
				tour.Status = TourDraft
				return []*Tour{tour}
			},
			url: "/dashboard",
		},
		{
			name: "disabled tour",
			candidates: func() []*Tour {
				tour := publishedRuntimeTour()
				tour.Enabled = false
				return []*Tour{tour}
			},
			url: "/dashboard",
		},
		{
			name: "target path mismatch",
			candidates: func() []*Tour {
				return []*Tour{publishedRuntimeTour()}
			},
			url: "/settings",
		},
		{
			name:       "no matching tour",
			candidates: func() []*Tour { return nil },
			url:        "/dashboard",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidates := tc.candidates()
			domain, _ := newRuntimeTestDomain(candidates, nil)

			result, err := domain.Resolve(context.Background(), ResolveRequest{
				AppId:     uuid.New(),
				SubjectId: "user-1",
				Url:       tc.url,
			})

			require.NoError(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestRuntimeResolveIncrementsShowsCount(t *testing.T) {
	tour := publishedRuntimeTour()
	hint := Hint{Id: uuid.New(), TourVersionId: tour.VersionId, Step: 1}
	domain, progress := newRuntimeTestDomain(
		[]*Tour{tour},
		map[uuid.UUID][]Hint{tour.VersionId: {hint}},
	)
	req := ResolveRequest{
		AppId:     tour.AppId,
		SubjectId: "user-1",
		Url:       "/dashboard",
	}

	first, err := domain.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := domain.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, 2, progress.progress[tour.Id].ShowsCount)
	require.NotNil(t, second.CurrentHintId)
	assert.Equal(t, hint.Id, *second.CurrentHintId)
}
