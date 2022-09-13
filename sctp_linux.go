// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"errors"
	"strconv"
	"strings"
	"sys/unix"
	"time"

	"net"

	"golang.org/x/sys/unix"
)

// Implements net.Conn on SCTP
type SCTPConn struct {
	fd int
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

}

func (c *SCTPConn) RemoteAddr() net.Addr {

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

// TODO: Some more error information / or logging?
// TODO: Shouldn't we handle EINTR?
func DialSCTP(address string, timeout time.Duration) (net.Conn, error) {
	addressComponents := strings.Split(address, ":")
	if len(addressComponents) != 2 {
		return nil, errors.New("sctp-address-format-err")
	}

	port, err := strconv.ParseInt(addressComponents[1], 10, 32)
	if err != nil {
		return nil, err
	}

	host := net.ParseIP(addressComponents[0])
	if host == nil {
		return nil, errors.New("ip-address-err")
	}

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, unix.IPPROTO_SCTP)
	if err != nil {
		return nil, err
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return nil, err
	}

	sockAddr := &unix.SockaddrInet4{
		Port: port,
		Addr: host.To4(),
	}
	copy(sockAddr.Addr[:], host.To4())

	err = unix.Connect(fd, sockAddr)
	if err != nil {
		if err != unix.EINPROGRESS && err != unix.EINTR {
			return nil, err
		}
	}

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
		fd: fd,
	}, nil
}
