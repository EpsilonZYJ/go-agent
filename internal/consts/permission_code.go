// Copyright (c) 2026 Yujie Zhou. Licensed under the MIT License.

package consts

type PermissionCode int

const (
	PermissionDeny PermissionCode = iota
	PermissionAllow
	PermissionAskUser
	PermissionInputInvalid
)
