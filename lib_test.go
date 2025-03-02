// SPDX-License-Identifier: Apache 2.0
// Copyright Copyright (c) 2022 NetLOX Inc

package loxilib

import (
	"fmt"
	"testing"
)

type Tk struct {
}

func (tk *Tk) TrieNodeWalker(b string) {
	fmt.Printf("%s\n", b)
}

func (tk *Tk) TrieData2String(d TrieData) string {

	if data, ok := d.(int); ok {
		return fmt.Sprintf("%d", data)
	}

	return ""
}

var tk Tk

func BenchmarkTrie(b *testing.B) {
	var tk Tk
	trieR := TrieInit(false)

	i := 0
	j := 0
	k := 0
	pLen := 32

	for n := 0; n < b.N; n++ {
		i = n & 0xff
		j = n >> 8 & 0xff
		k = n >> 16 & 0xff

		/*if j > 0 {
		      pLen = 24
		  } else {
		      pLen = 32
		  }*/
		route := fmt.Sprintf("192.%d.%d.%d/%d", k, j, i, pLen)
		res := trieR.AddTrie(route, n)
		if res != 0 {
			b.Errorf("failed to add %s:%d - (%d)", route, n, res)
			trieR.Trie2String(&tk)
		}
	}
}

func TestTrie(t *testing.T) {
	trieR := TrieInit(false)
	route := "192.168.1.1/32"
	data := 1100
	res := trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/15"
	data = 100
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/16"
	data = 99
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/8"
	data = 1
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "192.168.1.0/16"
	data = 1
	res = trieR.AddTrie(route, data)
	if res == 0 {
		t.Errorf("re-added %s:%d", route, data)
	}

	route = "0.0.0.0/0"
	data = 222
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "8.8.8.8/32"
	data = 1200
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "10.10.10.10/32"
	data = 12
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "1.1.1.1/32"
	data = 1212
	res = trieR.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	// If we need to dump trie elements
	// Run # go test -v .
	trieR.Trie2String(&tk)

	ret, ipn, rdata := trieR.FindTrie("192.41.3.1")
	if ret != 0 || (*ipn).String() != "192.0.0.0/8" || rdata != 1 {
		t.Errorf("failed to find %s", "192.41.3.1")
	}

	ret1, ipn, rdata1 := trieR.FindTrie("195.41.3.1")
	if ret1 != 0 || (*ipn).String() != "0.0.0.0/0" || rdata1 != 222 {
		t.Errorf("failed to find %s", "195.41.3.1")
	}

	ret2, ipn, rdata2 := trieR.FindTrie("8.8.8.8")
	if ret2 != 0 || (*ipn).String() != "8.8.8.8/32" || rdata2 != 1200 {
		t.Errorf("failed to find %d %s %d", ret, "8.8.8.8", rdata)
	}

	route = "0.0.0.0/0"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	ret1, _, rdata1 = trieR.FindTrie("195.41.3.1")
	if ret1 == 0 {
		t.Errorf("failed to find %s", "195.41.3.1")
	}

	route = "192.168.1.1/32"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "192.168.1.0/15"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "192.168.1.0/16"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "192.168.1.0/8"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "0.0.0.0/0"
	res = trieR.DelTrie(route)
	if res == 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "8.8.8.8/32"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "10.10.10.10/32"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	route = "1.1.1.1/24"
	res = trieR.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}

	trieR6 := TrieInit(true)
	route = "2001:db8::/32"
	data = 5100
	res = trieR6.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	route = "2001:db8::1/128"
	data = 5200
	res = trieR6.AddTrie(route, data)
	if res != 0 {
		t.Errorf("failed to add %s:%d", route, data)
	}

	ret, ipn, rdata = trieR6.FindTrie("2001:db8::1")
	if ret != 0 || (*ipn).String() != "2001:db8::1/128" || rdata != 5200 {
		t.Errorf("failed to find %s", "2001:db8::1")
	}

	route = "2001:db8::1/128"
	res = trieR6.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to del %s", route)
	}

	route = "2001:db8::/32"
	res = trieR6.DelTrie(route)
	if res != 0 {
		t.Errorf("failed to delete %s", route)
	}
}

func TestCounter(t *testing.T) {

	cR := NewCounter(0, 10)

	for i := 0; i < 12; i++ {
		idx, err := cR.GetCounter()

		if err == nil {
			if i > 9 {
				t.Errorf("get Counter unexpected %d:%d", i, idx)
			}
		} else {
			if i <= 9 {
				t.Errorf("failed to get Counter %d:%d", i, idx)
			}
		}
	}

	err := cR.PutCounter(5)
	if err != nil {
		t.Errorf("failed to put valid Counter %d", 5)
	}

	err = cR.PutCounter(2)
	if err != nil {
		t.Errorf("failed to put valid Counter %d", 2)
	}

	err = cR.PutCounter(15)
	if err == nil {
		t.Errorf("Able to put invalid Counter %d", 15)
	}

	var idx int
	idx, err = cR.GetCounter()
	if idx != 5 || err != nil {
		t.Errorf("Counter get got %d of expected %d", idx, 5)
	}

	idx, err = cR.GetCounter()
	if idx != 2 || err != nil {
		t.Errorf("Counter get got %d of expected %d", idx, 2)
	}
}
