package domains

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type previewTourRepoStub struct {
	TourRepositoryInterface
	result *Tour
	err    error
	called bool
}

func (s *previewTourRepoStub) Preview(_ context.Context, _ uuid.UUID) (*Tour, error) {
	s.called = true
	return s.result, s.err
}

func TestPreviewRejectsEmptyID(t *testing.T) {
	stub := &previewTourRepoStub{}
	domain := NewTourDomain(stub)
	_, err := domain.Preview(context.Background(), uuid.Nil)
	require.Error(t, err)
	var logic *LogicError
	require.ErrorAs(t, err, &logic)
	require.Equal(t, http.StatusBadRequest, logic.Code)
	require.False(t, stub.called)
}

func TestPreviewReturnsRepositoryDraft(t *testing.T) {
	tour := &Tour{Id: uuid.New(), VersionId: uuid.New(), Status: TourDraft}
	stub := &previewTourRepoStub{result: tour}
	domain := NewTourDomain(stub)
	result, err := domain.Preview(context.Background(), tour.Id)
	require.NoError(t, err)
	require.Same(t, tour, result)
	require.True(t, stub.called)
}
