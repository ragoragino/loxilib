// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

package loxilib

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"net"

	"golang.org/x/sys/unix"
)

// ip6ZoneToInt converts an IP6 Zone net string to a unix int
// returns 0 if zone is not a valid int
func ip6ZoneToInt(zone string) int {
	if zone == "" {
		return 0
	}

	if ifi, err := net.InterfaceByName(zone); err == nil {
		return ifi.Index
	}

	n, _ := strconv.Atoi(zone)

	return n
}

// TODO: Add support for IPv6 zones.
func parseAddr(addrPort string) (string, uint16, error) {
	colonIx := strings.LastIndex(addrPort, ":")
	if colonIx == -1 {
		return "", 0, errors.New("sctp-missing-port-err")
	}

	host := addrPort[:colonIx]
	if len(host) > 0 && host[0] == '[' {
		if len(host) < 2 || host[len(host)-1] != ']' {
			return "", 0, errors.New("sctp-addr-err")
		}

		host = host[1 : len(host)-1]
	}

	if len(host) == 0 {
		return "", 0, errors.New("sctp-addr-err")
	}

	port, err := strconv.ParseInt(addrPort[colonIx+1:], 10, 16)
	if err != nil {
		return "", 0, errors.New("sctp-port-err")
	}

	return host, uint16(port), nil
}

func parseAddrPort(addrPort string) (unix.Sockaddr, error) {
	host, port, err := parseAddr(addrPort)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("sctp-addr-err")
	}

	if ip.To4() == nil {
		sockAddr := &unix.SockaddrInet6{
			Port: int(port),
		}
		copy(sockAddr.Addr[:], ip.To16())

		return sockAddr, nil
	}

	sockAddr := &unix.SockaddrInet4{
		Port: int(port),
	}
	copy(sockAddr.Addr[:], ip.To4())

	return sockAddr, nil
}

type SCTPConn struct {
	fd int
}

func (c *SCTPConn) Close() error {
	if c.fd > 0 {
		return unix.Close(c.fd)
	}

	return nil
}

func DialSCTP(addressPort string, timeout time.Duration) (*SCTPConn, error) {
	sockAddr, err := parseAddrPort(addressPort)
	if err != nil {
		return nil, err
	}

	fd, err := sctpConnect(sockAddr, timeout)
	if err != nil {
		// Try to close the file descriptor if an error occurred
		if fd > 0 {
			unix.Close(fd)
		}

		return nil, err
	}

	return &SCTPConn{
		fd: fd,
	}, nil
}

// TODO: Some more error information / or logging?
// TODO: Shouldn't we handle EINTR?
func sctpConnect(sa unix.Sockaddr, timeout time.Duration) (int, error) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_SCTP)
	if err != nil {
		return fd, err
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return fd, err
	}

	err = unix.Connect(fd, sa)
	if err != nil {
		// These errors signal that connection might still finish async
		if err != unix.EINPROGRESS && err != unix.EINTR {
			return fd, err
		}
	}

	// Wait for the fd to become ready until the timeout expires
	timeval := unix.NsecToTimeval(int64(timeout))

	writeSet := new(unix.FdSet)
	writeSet.Zero()
	writeSet.Set(fd)

	ready, err := unix.Select(fd+1, nil, writeSet, nil, &timeval)
	if err != nil {
		return fd, err
	}

	if ready == 0 {
		return fd, errors.New("sctp-connect-err")
	}

	return fd, nil
}
