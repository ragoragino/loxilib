// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || zos

package loxilib

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"net"

	"golang.org/x/sys/unix"
)

// TODO: Add support for IPv6 zones.
func parseAddr(addrPort string) (string, uint16, error) {
	colonIx := strings.LastIndex(addrPort, ":")
	if colonIx == -1 {
		return "", 0, errors.New("sctp-missing-port-err")
	}

	host := addrPort[:colonIx]
	if len(host) > 0 && host[0] == '[' {
		if len(host) < 2 || host[len(host)-1] != ']' {
			return "", 0, errors.New("sctp-ipv6-addr-err")
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

func addressToSockAddr(addrPort string) (unix.Sockaddr, error) {
	host, port, err := parseAddr(addrPort)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("sctp-addr-err")
	}

	// IPv6 case
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

// DialSCTP creates a connection to an SCTP server. It errors if the connection
// couldn't be established within the timeout.
// Address can be an IPv4 or IPv6 address, and it must contain port.
func DialSCTP(address string, timeout time.Duration) (*SCTPConn, error) {
	sockAddr, err := addressToSockAddr(address)
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

func sctpConnect(sa unix.Sockaddr, timeout time.Duration) (int, error) {
	domain := unix.AF_INET
	if _, ok := sa.(*unix.SockaddrInet6); ok {
		domain = unix.AF_INET6
	}

	fd, err := unix.Socket(domain, unix.SOCK_STREAM, unix.IPPROTO_SCTP)
	if err != nil {
		return fd, err
	}

	err = unix.SetNonblock(fd, true)
	if err != nil {
		return fd, err
	}

	err = unix.Connect(fd, sa)
	if err != nil {
		// EINPROGRESS signals that connection might still finish async
		if err != unix.EINPROGRESS {
			return fd, err
		}
	}

	deadline := time.Now().Add(timeout)

	for {
		timeout = deadline.Sub(time.Now())
		if timeout <= 0 {
			return fd, errors.New("sctp-connect-err")
		}

		// Wait for the fd to become ready until the timeout expires
		pfds := []unix.PollFd{
			{
				Fd:     int32(fd),
				Events: unix.POLLOUT,
			},
		}
		ready, err := unix.Poll(pfds, int(timeout.Milliseconds()))
		if err != nil {
			// Syscall was interrupted, retry
			if err == unix.EINTR {
				continue
			}

			return fd, err
		}

		// Check whether the timeout expired or any of the errors occurred
		if ready == 0 {
			return fd, errors.New("sctp-poll-timeout-err")
		} else if (pfds[0].Revents & unix.POLLOUT) == 0 {
			errCode, _ := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)

			return fd, fmt.Errorf("sctp-poll-err-%d-%d", pfds[0].Revents, errCode)
		}

		return fd, nil
	}
}
