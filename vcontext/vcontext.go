// Copyright 2018 GRAIL, Inc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package vcontext creates a singleton vanadium Context object.
package vcontext

import (
	"sync"

	"github.com/grailbio/base/grail"
	"v.io/v23"
	"v.io/v23/context"
	_ "v.io/x/ref/runtime/factories/grail" // Needed to initialize v23
)

var (
	once = sync.Once{}
	ctx  *context.T
)

// Background returns the singleton Vanadium context for v23. It initializes v23
// on the first call.  GRAIL applications should always use this function to
// initialize and create a context instead of calling v23.Init() manually.
//
// Caution: this function is depended on by many services, specifically the
// production pipeline controller. Be extremely careful when changing it.
func Background() *context.T {
	once.Do(func() {
		var shutdown v23.Shutdown
		ctx, shutdown = v23.Init()
		grail.RegisterShutdownCallback(grail.Shutdown(shutdown))
	})
	return ctx
}
