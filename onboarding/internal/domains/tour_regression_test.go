package domains

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type tourUnitRepo struct {
	TourRepositoryInterface
	createCalls   int
	updateCalls   int
	archiveCalls  int
	created       *Tour
	updated       TourMetaPatch
	errorToReturn error
}

func (s *tourUnitRepo) CreateWithDraft(_ context.Context, tour *Tour) (*Tour, error) {
	s.createCalls++
	s.created = tour
	if s.errorToReturn != nil {
		return nil, s.errorToReturn
	}
	return tour, nil
}
func (s *tourUnitRepo) UpdateMeta(_ context.Context, _ uuid.UUID, patch TourMetaPatch) (*Tour, error) {
	s.updateCalls++
	s.updated = patch
	if s.errorToReturn != nil {
		return nil, s.errorToReturn
	}
	return &Tour{Title: *patch.Title}, nil
}
func (s *tourUnitRepo) Archive(_ context.Context, _ uuid.UUID) error {
	s.archiveCalls++
	return s.errorToReturn
}

func TestTourUnitCreateDefaultsAndValidation(t *testing.T) {
	appId := uuid.New()
	for _, tc := range []struct {
		name string
		tour Tour
	}{
		{"empty title", Tour{TargetPath: "/welcome"}},
		{"empty path", Tour{Title: "Start"}},
		{"negative priority", Tour{Title: "Start", TargetPath: "/welcome", Priority: -1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &tourUnitRepo{}
			result, err := NewTourDomain(repo).Create(context.Background(), appId, &tc.tour)
			requireAuthHTTPCode(t, err, http.StatusBadRequest)
			require.Nil(t, result)
			require.Zero(t, repo.createCalls)
		})
	}
	repo := &tourUnitRepo{}
	input := &Tour{Title: "Welcome", TargetPath: "/welcome", Priority: 0}
	created, err := NewTourDomain(repo).Create(context.Background(), appId, input)
	require.NoError(t, err)
	require.Same(t, input, created)
	require.Equal(t, appId, created.AppId)
	require.Equal(t, TourDraft, created.Status)
	require.Equal(t, TriggerOnLoad, created.TriggerType)
	require.NotNil(t, created.Audience)
	require.Equal(t, 1, repo.createCalls)
}

func TestTourUnitUpdateMetaRejectsBadInputAndDelegates(t *testing.T) {
	id := uuid.New()
	blank := ""
	negative := -1
	for _, patch := range []TourMetaPatch{{Title: &blank}, {Priority: &negative}} {
		repo := &tourUnitRepo{}
		updated, err := NewTourDomain(repo).UpdateMeta(context.Background(), id, patch)
		requireAuthHTTPCode(t, err, http.StatusBadRequest)
		require.Nil(t, updated)
		require.Zero(t, repo.updateCalls)
	}
	title := "Updated"
	repo := &tourUnitRepo{}
	updated, err := NewTourDomain(repo).UpdateMeta(context.Background(), id, TourMetaPatch{Title: &title})
	require.NoError(t, err)
	require.Equal(t, "Updated", updated.Title)
	require.Equal(t, 1, repo.updateCalls)
	require.Equal(t, title, *repo.updated.Title)
}

func TestTourUnitRepositoryErrorsAreReturned(t *testing.T) {
	sentinel := errors.New("repository failed")
	repo := &tourUnitRepo{errorToReturn: sentinel}
	domain := NewTourDomain(repo)
	_, err := domain.Create(context.Background(), uuid.New(), &Tour{Title: "Welcome", TargetPath: "/"})
	require.ErrorIs(t, err, sentinel)
	title := "new title"
	_, err = domain.UpdateMeta(context.Background(), uuid.New(), TourMetaPatch{Title: &title})
	require.ErrorIs(t, err, sentinel)
	err = domain.Archive(context.Background(), uuid.New())
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, 1, repo.archiveCalls)
}
