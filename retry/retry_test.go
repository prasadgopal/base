// Copyright 2018 GRAIL, Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package retry

import (
	"context"
	"testing"
	"time"

	"github.com/grailbio/base/errors"
)

func TestBackoff(t *testing.T) {
	policy := Backoff(time.Second, 10*time.Second, 2)
	expect := []time.Duration{
		time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		10 * time.Second,
		10 * time.Second,
	}
	for retries, wait := range expect {
		keepgoing, dur := policy.Retry(retries)
		if !keepgoing {
			t.Fatal("!keepgoing")
		}
		if got, want := dur, wait; got != want {
			t.Errorf("retry %d: got %v, want %v", retries, got, want)
		}
	}
}

func TestWaitCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	policy := Backoff(time.Hour, time.Hour, 1)
	cancel()
	if got, want := Wait(ctx, policy, 0), context.Canceled; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWaitDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	policy := Backoff(time.Hour, time.Hour, 1)
	if got, want := Wait(ctx, policy, 0), errors.E(errors.Timeout); !errors.Match(want, got) {
		t.Errorf("got %v, want %v", got, want)
	}
}
