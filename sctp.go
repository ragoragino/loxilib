// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos)

package loxilib

import (
	"errors"
	"sys/unix"
	"time"

	"net"

	"golang.org/x/sys/unix"
)

// Not implemented for non-Unix systems.
func DialSCTP(address string, timeout time.Duration) (net.Conn, error) {
	return nil, nil
}
