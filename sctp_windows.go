// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
	"sys/unix"
	"time"

	"net"

	"golang.org/x/sys/unix"
)

// Not implemented for Windows.
func DialSCTP(address string, timeout time.Duration) (net.Conn, error) {
	return nil, nil
}
