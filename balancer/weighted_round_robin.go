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
    @date: 2023/5/7 20:15
    @note: --
**/

package balancer

import "log"

// init will add WeightedRoundRobin balancer into balancer factories
func init() {
	factories[WeightedR2Balancer] = NewWeightedR2Balancer
}

type WeightedR2 struct {
	BaseBalancer
	currentWeights    []int
	hostsWeights      []int
	hostEffectWeights []int
}

func NewWeightedR2Balancer(hosts []string, hostsWeights []int) Balancer {
	var sum int
	sum = 0
	for _, weight := range hostsWeights {
		sum += weight
	}
	if sum == 0 {
		return nil
	}
	currentWeights := make([]int, len(hostsWeights))
	hostEffectWeights := make([]int, len(hostsWeights))
	copy(currentWeights, hostsWeights)
	copy(hostEffectWeights, hostsWeights)
	return &WeightedR2{
		BaseBalancer: BaseBalancer{
			hosts: hosts,
		},
		currentWeights:    currentWeights,
		hostsWeights:      hostsWeights,
		hostEffectWeights: hostEffectWeights,
	}
}

func (b *WeightedR2) Add(host string, weight int) {
	b.Lock()
	defer b.Unlock()
	for _, h := range b.hosts {
		if h == host {
			return
		}
	}
	b.hosts = append(b.hosts, host)
	b.hostsWeights = append(b.hostsWeights, weight)
	b.currentWeights = append(b.currentWeights, weight)
	b.hostEffectWeights = append(b.hostEffectWeights, weight)
}

// Remove new host from the balancer
func (b *WeightedR2) Remove(host string) {
	b.Lock()
	defer b.Unlock()
	for i, h := range b.hosts {
		if h == host {
			b.hosts = append(b.hosts[:i], b.hosts[i+1:]...)
			b.hostsWeights = append(b.hostsWeights[:i], b.hostsWeights[i+1:]...)
			b.currentWeights = append(b.currentWeights[:i], b.currentWeights[i+1:]...)
			b.hostEffectWeights = append(b.hostEffectWeights[:i], b.hostEffectWeights[i+1:]...)
			return
		}
	}
}

func (b *WeightedR2) Balance(cliIp string) (string, error) {
	log.Printf("Client Ip = %s\n", cliIp)
	b.RLock()
	defer b.RUnlock()
	// 1. 轮询所有节点，计算当前状态下所有的节点的 weight 之和 作为 totalWeight
	totalWeight := 0
	for _, weight := range b.hostEffectWeights {
		totalWeight += weight
	}
	log.Printf("Toatal = %d\n", totalWeight)
	// 2. 更新每个节点的 currentWeight ， currentWeight = currentWeight + effectiveWeight;
	// 选出所有节点 currentWeight 中最大的一个节点作为选中节点；
	// log.Printf("Before\neff=%v\ncur=%v\n", b.hostEffectWeights, b.currentWeights)
	selectedIndex := -1
	for i := range b.currentWeights {
		b.currentWeights[i] += b.hostEffectWeights[i]
		if selectedIndex == -1 || b.currentWeights[selectedIndex] < b.currentWeights[i] {
			selectedIndex = i
		}
	}
	// 3. 选择中的节点再次更新 currentWeight, currentWeight = currentWeight - totalWeight
	b.currentWeights[selectedIndex] -= totalWeight
	return b.hosts[selectedIndex], nil
}
