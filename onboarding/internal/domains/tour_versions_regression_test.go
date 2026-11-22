package domains

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type versionUnitRepo struct {
	TourRepositoryInterface
	publishCalls  int
	historyCalls  int
	restoreCalls  int
	previewCalls  int
	lastTourId    uuid.UUID
	lastSourceId  uuid.UUID
	err           error
	publishResult *PublishResult
	history       []TourVersionHistoryItem
	restoreResult *RestoreVersionResult
	preview       *Tour
}

func (s *versionUnitRepo) Publish(_ context.Context, id uuid.UUID) (*PublishResult, error) {
	s.publishCalls++
	s.lastTourId = id
	return s.publishResult, s.err
}
func (s *versionUnitRepo) ListVersions(_ context.Context, id uuid.UUID) ([]TourVersionHistoryItem, error) {
	s.historyCalls++
	s.lastTourId = id
	return s.history, s.err
}
func (s *versionUnitRepo) RestoreVersion(_ context.Context, tourId, versionId uuid.UUID) (*RestoreVersionResult, error) {
	s.restoreCalls++
	s.lastTourId = tourId
	s.lastSourceId = versionId
	return s.restoreResult, s.err
}
func (s *versionUnitRepo) Preview(_ context.Context, id uuid.UUID) (*Tour, error) {
	s.previewCalls++
	s.lastTourId = id
	return s.preview, s.err
}

func TestTourVersionUnitRejectsNilIdsBeforeRepository(t *testing.T) {
	repo := &versionUnitRepo{}
	domain := NewTourDomain(repo)
	_, err := domain.Publish(context.Background(), uuid.Nil)
	requireAuthHTTPCode(t, err, http.StatusBadRequest)
	_, err = domain.ListVersions(context.Background(), uuid.Nil)
	requireAuthHTTPCode(t, err, http.StatusBadRequest)
	_, err = domain.RestoreVersion(context.Background(), uuid.Nil, uuid.New())
	requireAuthHTTPCode(t, err, http.StatusBadRequest)
	_, err = domain.RestoreVersion(context.Background(), uuid.New(), uuid.Nil)
	requireAuthHTTPCode(t, err, http.StatusBadRequest)
	_, err = domain.Preview(context.Background(), uuid.Nil)
	requireAuthHTTPCode(t, err, http.StatusBadRequest)
	require.Zero(t, repo.publishCalls)
	require.Zero(t, repo.historyCalls)
	require.Zero(t, repo.restoreCalls)
	require.Zero(t, repo.previewCalls)
}

func TestTourVersionUnitDelegatesAndPreservesResults(t *testing.T) {
	tourId, versionId := uuid.New(), uuid.New()
	published := &PublishResult{TourId: tourId, PublishedVersionId: versionId, Version: 2}
	history := []TourVersionHistoryItem{{Id: versionId, Version: 2, Status: TourPublished}}
	restored := &RestoreVersionResult{TourId: tourId, RestoredFromVersionId: versionId, DraftVersionId: uuid.New(), DraftVersion: 3}
	preview := &Tour{Id: tourId, Title: "Preview"}
	repo := &versionUnitRepo{publishResult: published, history: history, restoreResult: restored, preview: preview}
	domain := NewTourDomain(repo)
	result, err := domain.Publish(context.Background(), tourId)
	require.NoError(t, err)
	require.Same(t, published, result)
	list, err := domain.ListVersions(context.Background(), tourId)
	require.NoError(t, err)
	require.Equal(t, history, list)
	restoredResult, err := domain.RestoreVersion(context.Background(), tourId, versionId)
	require.NoError(t, err)
	require.Same(t, restored, restoredResult)
	require.Equal(t, versionId, repo.lastSourceId)
	tour, err := domain.Preview(context.Background(), tourId)
	require.NoError(t, err)
	require.Same(t, preview, tour)
	require.Equal(t, tourId, repo.lastTourId)
	require.Equal(t, 1, repo.publishCalls)
	require.Equal(t, 1, repo.historyCalls)
	require.Equal(t, 1, repo.restoreCalls)
	require.Equal(t, 1, repo.previewCalls)
}

func TestTourVersionUnitPropagatesRepositoryErrors(t *testing.T) {
	sentinel := errors.New("database unavailable")
	repo := &versionUnitRepo{err: sentinel}
	domain := NewTourDomain(repo)
	id := uuid.New()
	_, err := domain.Publish(context.Background(), id)
	require.ErrorIs(t, err, sentinel)
	_, err = domain.ListVersions(context.Background(), id)
	require.ErrorIs(t, err, sentinel)
	_, err = domain.RestoreVersion(context.Background(), id, uuid.New())
	require.ErrorIs(t, err, sentinel)
	_, err = domain.Preview(context.Background(), id)
	require.ErrorIs(t, err, sentinel)
}
