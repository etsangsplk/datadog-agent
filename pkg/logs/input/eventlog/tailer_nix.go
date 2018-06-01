// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build !windows

package eventlog

import (
	log "github.com/cihub/seelog"
)

// Start does not do much
func (t *Tailer) Start(_ string) {
	log.Warn("Event Log not supported on this system")
	go t.tail()
}

// Stop stops the tailer
func (t *Tailer) Stop() {
	t.stop <- struct{}{}
	<-t.done
}

// tail does nothing
func (t *Tailer) tail() {
	<-t.stop
	t.done <- struct{}{}
}
