package domains

import (
	"context"
	"testing"
	"time"

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
	progress    map[uuid.UUID]*Progress
	upsertCalls int
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
	s.upsertCalls++

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
	assert.Equal(t, 2, result.Tour.StepsCount)
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

func TestRuntimeResolveRepeatedRequestPreservesProgress(t *testing.T) {
	tour := publishedRuntimeTour()
	firstHint := Hint{Id: uuid.New(), TourVersionId: tour.VersionId, Step: 1}
	secondHint := Hint{Id: uuid.New(), TourVersionId: tour.VersionId, Step: 2}
	domain, progress := newRuntimeTestDomain(
		[]*Tour{tour},
		map[uuid.UUID][]Hint{tour.VersionId: {firstHint, secondHint}},
	)
	req := ResolveRequest{
		AppId:     tour.AppId,
		SubjectId: "user-1",
		Url:       "/dashboard",
	}

	first, err := domain.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, first)
	assert.Equal(t, 1, progress.upsertCalls)
	startedAt := progress.progress[tour.Id].StartedAt
	progress.progress[tour.Id].CurrentHintId = &secondHint.Id

	second, err := domain.Resolve(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, second)
	assert.Equal(t, 1, progress.upsertCalls,
		"повторный resolve не должен вызывать UpsertProgress")
	assert.Equal(t, 1, progress.progress[tour.Id].ShowsCount)
	assert.Equal(t, startedAt, progress.progress[tour.Id].StartedAt)
	require.NotNil(t, second.CurrentHintId)
	assert.Equal(t, secondHint.Id, *second.CurrentHintId)
}

func TestRuntimeResolveSkipsPathMismatchAndSelectsNextCandidate(t *testing.T) {
	mismatched := publishedRuntimeTour()
	mismatched.TargetPath = "/settings"
	matching := publishedRuntimeTour()
	matching.TargetPath = "/dashboard"
	hint := Hint{Id: uuid.New(), TourVersionId: matching.VersionId, Step: 1}
	domain, _ := newRuntimeTestDomain(
		[]*Tour{mismatched, matching},
		map[uuid.UUID][]Hint{matching.VersionId: {hint}},
	)

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: matching.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, matching.Id, result.Tour.Id)
}

func TestRuntimeResolveShowOnceSkipsCompletedTour(t *testing.T) {
	tour := publishedRuntimeTour()
	tour.Audience = &Audience{ShowOnce: true}
	domain, progress := newRuntimeTestDomain([]*Tour{tour}, nil)
	progress.progress[tour.Id] = &Progress{
		AppId: tour.AppId, SubjectId: "user-1", TourId: tour.Id,
		TourVersionId: tour.VersionId, Status: ProgressCompleted, ShowsCount: 1,
	}

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: tour.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRuntimeResolveOnlyNewSkipsExistingProgress(t *testing.T) {
	tour := publishedRuntimeTour()
	tour.Audience = &Audience{OnlyNew: true}
	domain, progress := newRuntimeTestDomain([]*Tour{tour}, nil)
	progress.progress[tour.Id] = &Progress{
		AppId: tour.AppId, SubjectId: "user-1", TourId: tour.Id,
		TourVersionId: uuid.New(), Status: ProgressCompleted, ShowsCount: 1,
	}

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: tour.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRuntimeResolveMaxShowsContinuesInProgressTour(t *testing.T) {
	tour := publishedRuntimeTour()
	tour.Audience = &Audience{MaxShows: 1}
	hint := Hint{Id: uuid.New(), TourVersionId: tour.VersionId, Step: 1}
	domain, progress := newRuntimeTestDomain(
		[]*Tour{tour},
		map[uuid.UUID][]Hint{tour.VersionId: {hint}},
	)
	startedAt := time.Now().Add(-time.Hour).UTC()
	progress.progress[tour.Id] = &Progress{
		AppId: tour.AppId, SubjectId: "user-1", TourId: tour.Id,
		TourVersionId: tour.VersionId, CurrentHintId: &hint.Id,
		Status: ProgressInProgress, ShowsCount: 1, StartedAt: startedAt,
	}

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: tour.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, progress.progress[tour.Id].ShowsCount)
	assert.Equal(t, startedAt, progress.progress[tour.Id].StartedAt)
	require.NotNil(t, result.CurrentHintId)
	assert.Equal(t, hint.Id, *result.CurrentHintId)
}

func TestRuntimeResolveDoesNotRestartCompletedVersion(t *testing.T) {
	tour := publishedRuntimeTour()
	domain, progress := newRuntimeTestDomain([]*Tour{tour}, nil)
	progress.progress[tour.Id] = &Progress{
		AppId: tour.AppId, SubjectId: "user-1", TourId: tour.Id,
		TourVersionId: tour.VersionId, Status: ProgressCompleted, ShowsCount: 1,
	}

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: tour.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, progress.progress[tour.Id].ShowsCount)
}

func TestRuntimeResolveDoesNotReplaceActiveOlderVersion(t *testing.T) {
	tour := publishedRuntimeTour()
	oldVersionId := uuid.New()
	currentHintId := uuid.New()
	domain, progress := newRuntimeTestDomain([]*Tour{tour}, nil)
	progress.progress[tour.Id] = &Progress{
		AppId: tour.AppId, SubjectId: "user-1", TourId: tour.Id,
		TourVersionId: oldVersionId, CurrentHintId: &currentHintId,
		Status: ProgressInProgress, ShowsCount: 1,
	}

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: tour.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	assert.Nil(t, result)
	stored := progress.progress[tour.Id]
	assert.Equal(t, oldVersionId, stored.TourVersionId)
	assert.Equal(t, 1, stored.ShowsCount)
	require.NotNil(t, stored.CurrentHintId)
	assert.Equal(t, currentHintId, *stored.CurrentHintId)
}

func TestRuntimeResolveStartsNewVersionAfterPreviousCompletion(t *testing.T) {
	tour := publishedRuntimeTour()
	oldVersionId := uuid.New()
	hint := Hint{Id: uuid.New(), TourVersionId: tour.VersionId, Step: 1}
	domain, progress := newRuntimeTestDomain(
		[]*Tour{tour},
		map[uuid.UUID][]Hint{tour.VersionId: {hint}},
	)
	progress.progress[tour.Id] = &Progress{
		AppId: tour.AppId, SubjectId: "user-1", TourId: tour.Id,
		TourVersionId: oldVersionId, Status: ProgressCompleted, ShowsCount: 1,
	}

	result, err := domain.Resolve(context.Background(), ResolveRequest{
		AppId: tour.AppId, SubjectId: "user-1", Url: "/dashboard",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	stored := progress.progress[tour.Id]
	assert.Equal(t, tour.VersionId, stored.TourVersionId)
	assert.Equal(t, 2, stored.ShowsCount)
	assert.Equal(t, ProgressInProgress, stored.Status)
	require.NotNil(t, stored.CurrentHintId)
	assert.Equal(t, hint.Id, *stored.CurrentHintId)
}
