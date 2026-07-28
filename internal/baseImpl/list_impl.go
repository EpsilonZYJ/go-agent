// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package baseImpl

func ListContains[T comparable](list []T, item T) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
