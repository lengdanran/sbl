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
    @date: 2023/5/7 15:47
    @note: --
**/

package utils

import (
	"log"
	"reflect"
)

// Contains 使用了反射来判断切片的类型，并遍历切片中的每个元素，与给定的元素进行比较。
// 如果找到了相同的元素，则返回true，否则返回false。
func Contains(slice interface{}, item interface{}) bool {
	switch reflect.TypeOf(slice).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(slice)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(s.Index(i).Interface(), item) {
				return true
			}
		}
	}
	return false
}

// Remove uses reflection to determine the type of the slice,
// and iterates over each element in the slice to compare it with the given element.
// If the same element is found and removed successfully, the processed array is returned
func Remove(slice interface{}, item interface{}) interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		log.Panicln("remove: not a slice")
		return slice
	}
	for i := 0; i < s.Len(); i++ {
		if reflect.DeepEqual(s.Index(i).Interface(), item) {
			return reflect.AppendSlice(s.Slice(0, i), s.Slice(i+1, s.Len())).Interface()
		}
	}
	return slice
}
