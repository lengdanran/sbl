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
    @date: 2023/5/7 18:34
    @note: --
**/

package balancer

import (
	"hash/crc32"
	"log"
)

// init will add random balancer into balancer factories
func init() {
	factories[IPHashBalancer] = NewIPHash
}

type IPHash struct {
	BaseBalancer
}

func NewIPHash(hosts []string) Balancer {
	return &IPHash{
		BaseBalancer{
			hosts: hosts,
		},
	}
}

// getCrc returns unit32 crc hashcode of key
func getCrc(key string) uint32 {
	if len(key) < 64 {
		var scratch [64]byte
		copy(scratch[:], key)
		return crc32.ChecksumIEEE(scratch[:len(key)])
	}
	return crc32.ChecksumIEEE([]byte(key))
}

func (ipHash *IPHash) Balance(cliIp string) (string, error) {
	log.Printf("Client Ip = %s\n", cliIp)
	ipHash.RLock()
	defer ipHash.RUnlock()
	if len(ipHash.hosts) == 0 {
		return "", NoHostError
	}
	hashCode := getCrc(cliIp)
	log.Printf("HashCode of ClientIp[%s] is %d\n", cliIp, hashCode)
	// Take the remainder as the host index subscript
	targetHostIndex := hashCode % uint32(len(ipHash.hosts))
	return ipHash.hosts[targetHostIndex], nil
}
