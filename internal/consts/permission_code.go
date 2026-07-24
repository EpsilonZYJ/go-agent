package consts

type PermissionCode int

const (
	PermissionDeny PermissionCode = iota
	PermissionAllow
	PermissionAskUser
	PermissionInputInvalid
)
