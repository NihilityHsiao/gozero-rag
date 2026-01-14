package user_tenant

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var ErrNotFound = sqlx.ErrNotFound

type Role = string

const (
	RoleOwner  = "owner"  // 绑定租户计费、注销租户、邀请/踢出 Admin 和 Normal。
	RoleNormal = "normal" // 仅允许 Chat 回话、查看授权范围内的知识库、新建自己的标注。
	RoleAdmin  = "admin"  // 管理知识库、上传/切片文档、配置对话参数。
)
