/*
 * Copyright (c) 2020-2021 The SeerLink developers
 */

package store

import (
	"context"

	"github.com/SeerLink/seerlink/core/store/models"
)

// HeadTrackable represents any object that wishes to respond to ethereum events,
// after being attached to HeadTracker.
//go:generate mockery --name HeadTrackable --output ../internal/mocks/ --case=underscore
type HeadTrackable interface {
	Connect(head *models.Head) error
	Disconnect()
	OnNewLongestChain(ctx context.Context, head models.Head)
}
