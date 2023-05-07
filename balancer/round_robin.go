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
    @date: 2023/5/7 19:40
    @note: --
**/

package balancer

import (
	"log"
	"math"
)

// init will add RoundRobin balancer into balancer factories
func init() {
	factories[R2Balancer] = NewRoundRobin
}

type RoundRobin struct {
	BaseBalancer
	cnt int64
}

// NewRoundRobin will create RoundRobin balancer.
func NewRoundRobin(hosts []string) Balancer {
	return &RoundRobin{
		BaseBalancer: BaseBalancer{
			hosts: hosts,
		},
		cnt: 0,
	}
}

func (r *RoundRobin) Balance(_ string) (string, error) {
	log.Printf("RoundRobin Balancer, Current Cnt=%d\n", r.cnt)
	r.RLock()
	defer r.RUnlock()
	if len(r.hosts) == 0 {
		return "", NoHostError
	}
	if r.cnt == math.MaxInt64 {
		r.cnt = 0
	}
	selectedHost := r.hosts[r.cnt%int64(len(r.hosts))]
	r.cnt += 1
	return selectedHost, nil
}
