/*
 * Copyright (c) 2020-2021 The SeerLink developers
 */

package adapters_test

import (
	"encoding/json"
	"testing"

	"github.com/SeerLink/seerlink/core/adapters"
	"github.com/SeerLink/seerlink/core/internal/cltest"
	"github.com/SeerLink/seerlink/core/store/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSleep_Perform(t *testing.T) {
	store, cleanup := cltest.NewStore(t)
	defer cleanup()
	clock := cltest.NewTriggerClock(t)
	store.Clock = clock

	adapter := adapters.Sleep{}
	err := json.Unmarshal([]byte(`{"until": 2147483647}`), &adapter)
	assert.NoError(t, err)

	doneChan := make(chan struct{})
	go func() {
		result := adapter.Perform(models.RunInput{}, store)
		assert.Equal(t, string(models.RunStatusCompleted), string(result.Status()))
		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
		t.Error("Sleep adapter did not sleep")
	default:
	}

	clock.Trigger()

	_, ok := <-doneChan
	assert.True(t, ok)
}

func TestSleep_Perform_AlreadyElapsed(t *testing.T) {
	store, cleanup := cltest.NewStore(t)
	defer cleanup()

	adapter := adapters.Sleep{}
	err := json.Unmarshal([]byte(`{"until": 1332151919}`), &adapter)
	require.NoError(t, err)

	result := adapter.Perform(models.RunInput{}, store)
	require.NoError(t, result.Error())
	assert.Equal(t, string(models.RunStatusCompleted), string(result.Status()))
}
