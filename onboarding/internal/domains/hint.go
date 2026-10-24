package domains

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/ksendel-coder/onboarding_croquant/platform/errs"
)

type (
	HintDomain struct {
		tours TourRepositoryInterface
		hints HintRepositoryInterface
	}

	HintDomainInterface interface {
		Create(
			ctx context.Context,
			tourId uuid.UUID,
			h *Hint,
		) (*Hint, error)

		Update(
			ctx context.Context,
			tourId uuid.UUID,
			hintId uuid.UUID,
			patch HintPatch,
		) (*Hint, error)

		Delete(
			ctx context.Context,
			tourId uuid.UUID,
			hintId uuid.UUID,
		) error

		Reorder(
			ctx context.Context,
			tourId uuid.UUID,
			ids []uuid.UUID,
		) ([]Hint, error)

		ListByTour(
			ctx context.Context,
			tourId uuid.UUID,
		) ([]Hint, error)
	}

	HintRepositoryInterface interface {
		Create(
			ctx context.Context,
			h *Hint,
		) error

		GetById(
			ctx context.Context,
			id uuid.UUID,
		) (*Hint, error)

		ListByVersion(
			ctx context.Context,
			versionId uuid.UUID,
		) ([]Hint, error)

		Update(
			ctx context.Context,
			h *Hint,
		) error

		Delete(
			ctx context.Context,
			versionId uuid.UUID,
			id uuid.UUID,
		) error

		Reorder(
			ctx context.Context,
			versionId uuid.UUID,
			ids []uuid.UUID,
		) error
	}
)

func NewHintDomain(
	tours TourRepositoryInterface,
	hints HintRepositoryInterface,
) *HintDomain {
	return &HintDomain{
		tours: tours,
		hints: hints,
	}
}

func (hd *HintDomain) Create(
	ctx context.Context,
	tourId uuid.UUID,
	h *Hint,
) (*Hint, error) {
	if err := validateHint(h); err != nil {
		return nil, err
	}

	draft, err := hd.draftOf(
		ctx,
		tourId,
	)
	if err != nil {
		return nil, err
	}

	existing, err := hd.hints.ListByVersion(
		ctx,
		draft.Id,
	)
	if err != nil {
		return nil, err
	}

	maxStep := 0

	for _, e := range existing {
		if e.Step > maxStep {
			maxStep = e.Step
		}
	}

	h.TourVersionId = draft.Id
	h.Step = maxStep + 1

	if err := hd.hints.Create(
		ctx,
		h,
	); err != nil {
		return nil, err
	}

	return h, nil
}

func (hd *HintDomain) Update(
	ctx context.Context,
	tourId uuid.UUID,
	hintId uuid.UUID,
	patch HintPatch,
) (*Hint, error) {
	draft, err := hd.draftOf(
		ctx,
		tourId,
	)
	if err != nil {
		return nil, err
	}

	current, err := hd.hints.GetById(
		ctx,
		hintId,
	)
	if err != nil {
		return nil, err
	}

	if current.TourVersionId != draft.Id {
		return nil, logicErr(
			ErrVersionImmutable,
			"hint update",
			http.StatusConflict,
		)
	}

	if patch.Selector != nil {
		current.Selector = *patch.Selector
	}

	if patch.Title != nil {
		current.Title = *patch.Title
	}

	if patch.Content != nil {
		current.Content = *patch.Content
	}

	if patch.Placement != nil {
		current.Placement = *patch.Placement
	}

	if err := validateHint(current); err != nil {
		return nil, err
	}

	if err := hd.hints.Update(
		ctx,
		current,
	); err != nil {
		return nil, err
	}

	return current, nil
}

func (hd *HintDomain) Delete(
	ctx context.Context,
	tourId uuid.UUID,
	hintId uuid.UUID,
) error {
	draft, err := hd.draftOf(
		ctx,
		tourId,
	)
	if err != nil {
		return err
	}

	return hd.hints.Delete(
		ctx,
		draft.Id,
		hintId,
	)
}

func (hd *HintDomain) Reorder(
	ctx context.Context,
	tourId uuid.UUID,
	ids []uuid.UUID,
) ([]Hint, error) {
	draft, err := hd.draftOf(
		ctx,
		tourId,
	)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, logicErr(
			ErrReorderMismatch,
			"hint reorder",
			http.StatusBadRequest,
		)
	}

	seen := make(
		map[uuid.UUID]struct{},
		len(ids),
	)

	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			return nil, logicErr(
				ErrReorderMismatch,
				"hint reorder",
				http.StatusBadRequest,
			)
		}

		seen[id] = struct{}{}
	}

	if err := hd.hints.Reorder(
		ctx,
		draft.Id,
		ids,
	); err != nil {
		if errors.Is(
			err,
			ErrReorderMismatch,
		) {
			return nil, logicErr(
				ErrReorderMismatch,
				"hint reorder",
				http.StatusBadRequest,
			)
		}

		return nil, err
	}

	return hd.hints.ListByVersion(
		ctx,
		draft.Id,
	)
}

func (hd *HintDomain) ListByTour(
	ctx context.Context,
	tourId uuid.UUID,
) ([]Hint, error) {
	draft, err := hd.draftOf(
		ctx,
		tourId,
	)
	if err != nil {
		return nil, err
	}

	return hd.hints.ListByVersion(
		ctx,
		draft.Id,
	)
}

func (hd *HintDomain) draftOf(
	ctx context.Context,
	tourId uuid.UUID,
) (*TourVersion, error) {
	draft, err := hd.tours.VersionByStatus(
		ctx,
		tourId,
		TourDraft,
	)
	if err != nil {
		if errs.IsNotFound(err) {
			return nil, logicErr(
				ErrNoDraft,
				"hint edit",
				http.StatusConflict,
			)
		}

		return nil, err
	}

	return draft, nil
}

func validateHint(
	h *Hint,
) error {
	switch {
	case h.Selector == "":
		return logicErr(
			ErrSelectorRequired,
			"hint validation",
			http.StatusBadRequest,
		)

	case !h.Placement.Valid():
		return logicErr(
			ErrBadPlacement,
			"hint validation",
			http.StatusBadRequest,
		)

	default:
		return nil
	}
}
