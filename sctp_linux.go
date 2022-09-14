// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
	"net/netip"
	"strconv"
	"sys/unix"
	"time"

	"net"

	"golang.org/x/sys/unix"
)

// ip6ZoneToInt converts an IP6 Zone net string to a unix int
// returns 0 if zone is "" or not a valid int
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

func AddrPortToSockAddr(addrPort *netip.AddrPort) interface{} {
	if addrPort.Addr().Is6() {
		sockAddr := &unix.SockaddrInet6{
			Port:   addrPort.Port(),
			ZoneId: ip6ZoneToInt(addrPort.Addr().Zone()),
		}
		copy(sockAddr.Addr[:], addrPort.Addr.As16())

		return sockAddr
	}

	sockAddr := &unix.SockaddrInet4{
		Port: addrPort.Port(),
	}
	copy(sockAddr.Addr[:], addrPort.Addr.As4())

	return sockAddr
}

// Implements net.Conn on SCTP
type SCTPConn struct {
	fd         int
	remoteAddr *netip.AddrPort
}

func (c *SCTPConn) Read(b []byte) (n int, err error) {
	return 0, errors.New("Read not implemented")
}

func (c *SCTPConn) Write(b []byte) (n int, err error) {
	return 0, errors.New("Write not implemented")
}

func (c *SCTPConn) Close() error {
	if c.fd > 0 {
		return unix.Close(c.fd)
	}

	return nil
}

func (c *SCTPConn) LocalAddr() net.Addr {
	// TODO
	return nil
}

func (c *SCTPConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *SCTPConn) SetDeadline(t time.Time) error {
	return errors.New("SetDeadline not implemented")
}

func (c *SCTPConn) SetReadDeadline(t time.Time) error {
	return errors.New("SetReadDeadline not implemented")
}

func (c *SCTPConn) SetWriteDeadline(t time.Time) error {
	return errors.New("SetWriteDeadline not implemented")
}

func DialSCTP(address string, timeout time.Duration) (net.Conn, error) {
	// Try to close the file descriptor if an error occurs.
	c, err := dialSCTPInternal(address, timeout)
	if err != nil && c != nil {
		c.Close()
	}

	return c, err
}

// TODO: Some more error information / or logging?
// TODO: Shouldn't we handle EINTR?
func dialSCTPInternal(address string, timeout time.Duration) (net.Conn, error) {
	addrPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return nil, err
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_SCTP)
	if err != nil {
		return nil, err
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return nil, err
	}

	err = unix.Connect(fd, AddrPortToSockAddr(addrPort))
	if err != nil {
		// These errors signal that connection might still finish async
		if err != unix.EINPROGRESS && err != unix.EINTR {
			return nil, err
		}
	}

	// Wait for the fd to become ready until the timeout expires
	timespec := unix.NsecToTimespec(int64(timeout))

	writeSet := new(unix.FdSet)
	writeSet.Zero()
	writeSet.Set(fd)

	ready, err := unix.Select(fd+1, nil, writeSet, nil, &timespec)
	if err != nil {
		return nil, err
	}

	if ready == 0 {
		unix.Close(fd)

		return nil, errors.New("sctp-connect-err")
	}

	return &SCTPConn{
		fd:         fd,
		remoteAddr: addrPort,
	}, nil
}
