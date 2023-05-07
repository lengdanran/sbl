/**
Copyright [2023] [name of copyright owner]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

    @author: lengdanran
    @date: 2023/5/7 16:41
    @note: --
**/

package balancer

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRandom_Add .
func TestRandom_Add(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	cases := []struct {
		name   string
		lb     Balancer
		args   string
		expect Balancer
	}{
		{
			"test-1",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012", "http://127.0.0.1:1013"},
				},
				rnd: rnd,
			},
			"http://127.0.0.1:1013",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012", "http://127.0.0.1:1013"},
				},
				rnd: rnd,
			},
		},
		{
			"test-2",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012"},
				},
				rnd: rnd,
			},
			"http://127.0.0.1:1012",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012"},
				},
				rnd: rnd,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.lb.Add(c.args)
			assert.Equal(t, c.expect, c.lb)
		})
	}
}

// TestRandom_Remove .
func TestRandom_Remove(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	cases := []struct {
		name   string
		lb     Balancer
		args   string
		expect Balancer
	}{
		{
			"test-1",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012", "http://127.0.0.1:1013"},
				},
				rnd: rnd,
			},
			"http://127.0.0.1:1013",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012"},
				},
				rnd: rnd,
			},
		},
		{
			"test-2",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012"},
				},
				rnd: rnd,
			},
			"http://127.0.0.1:1013",
			&Random{
				BaseBalancer: BaseBalancer{
					hosts: []string{"http://127.0.0.1:1011",
						"http://127.0.0.1:1012"},
				},
				rnd: rnd,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.lb.Remove(c.args)
			assert.Equal(t, c.expect, c.lb)
		})
	}
}

// TestRandom_Balance .
func TestRandom_Balance(t *testing.T) {
	type expect struct {
		reply string
		err   error
	}
	cases := []struct {
		name   string
		lb     Balancer
		args   string
		expect expect
	}{
		{
			"test-1",
			NewRandom([]string{"http://127.0.0.1:1011"}),
			"",
			expect{
				"http://127.0.0.1:1011",
				nil,
			},
		},
		{
			"test-2",
			NewRandom([]string{}),
			"",
			expect{
				"",
				NoHostError,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reply, err := c.lb.Balance(c.args)
			assert.Equal(t, c.expect.reply, reply)
			assert.Equal(t, c.expect.err, err)
		})
	}
}
