/*
 * Copyright (c) 2020-2021 The SeerLink developers
 */

package utils_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/SeerLink/seerlink/core/utils"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

type countingWorker struct {
	numJobsPerformed int32
	delay            time.Duration
}

func (t *countingWorker) Work() {
	if t.delay != 0 {
		time.Sleep(t.delay)
	}
	// Without an atomic, the race detector fails
	atomic.AddInt32(&t.numJobsPerformed, 1)
}

func (t *countingWorker) getNumJobsPerformed() int {
	return int(atomic.LoadInt32(&t.numJobsPerformed))
}

func TestSleeperTask_WakeupAfterStopPanics(t *testing.T) {
	t.Parallel()

	worker := &countingWorker{}
	sleeper := utils.NewSleeperTask(worker)

	require.NoError(t, sleeper.Stop())

	require.Panics(t, func() {
		sleeper.WakeUp()
	})
	gomega.NewGomegaWithT(t).Eventually(worker.getNumJobsPerformed).Should(gomega.Equal(0))
}

func TestSleeperTask_CallingStopTwicePanics(t *testing.T) {
	t.Parallel()

	worker := &countingWorker{}
	sleeper := utils.NewSleeperTask(worker)
	require.NoError(t, sleeper.Stop())
	require.Panics(t, func() {
		require.NoError(t, sleeper.Stop())
	})
}

func TestSleeperTask_WakeupPerformsWork(t *testing.T) {
	t.Parallel()

	worker := &countingWorker{}
	sleeper := utils.NewSleeperTask(worker)

	sleeper.WakeUp()
	gomega.NewGomegaWithT(t).Eventually(worker.getNumJobsPerformed).Should(gomega.Equal(1))
	require.NoError(t, sleeper.Stop())
}

type controllableWorker struct {
	countingWorker
	awaitWorkStarted chan struct{}
	allowResumeWork  chan struct{}
	ignoreSignals    bool
}

func (w *controllableWorker) Work() {
	if !w.ignoreSignals {
		w.awaitWorkStarted <- struct{}{}
		<-w.allowResumeWork
	}
	w.countingWorker.Work()
}

func TestSleeperTask_WakeupEnqueuesMaxTwice(t *testing.T) {
	t.Parallel()

	worker := &controllableWorker{awaitWorkStarted: make(chan struct{}), allowResumeWork: make(chan struct{})}
	sleeper := utils.NewSleeperTask(worker)

	sleeper.WakeUp()
	<-worker.awaitWorkStarted
	sleeper.WakeUp()
	sleeper.WakeUp()
	sleeper.WakeUp()
	sleeper.WakeUp()
	sleeper.WakeUp()
	worker.ignoreSignals = true
	worker.allowResumeWork <- struct{}{}

	gomega.NewGomegaWithT(t).Eventually(worker.getNumJobsPerformed).Should(gomega.Equal(2))
	gomega.NewGomegaWithT(t).Consistently(worker.getNumJobsPerformed).Should(gomega.BeNumerically("<", 3))
	require.NoError(t, sleeper.Stop())
}

func TestSleeperTask_StopWaitsUntilWorkFinishes(t *testing.T) {
	t.Parallel()

	worker := &controllableWorker{awaitWorkStarted: make(chan struct{}), allowResumeWork: make(chan struct{})}
	sleeper := utils.NewSleeperTask(worker)

	sleeper.WakeUp()
	<-worker.awaitWorkStarted
	require.Equal(t, 0, worker.getNumJobsPerformed())
	worker.allowResumeWork <- struct{}{}

	require.NoError(t, sleeper.Stop())
	require.Equal(t, worker.getNumJobsPerformed(), 1)
}
