// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package team

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"gozero-rag/internal/model/user"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
)

// InviteData Redis 中存储的邀请信息
type InviteData struct {
	TenantId     string `json:"tenant_id"`
	InviterId    string `json:"inviter_id"`
	InviteeEmail string `json:"invitee_email"`
	Role         string `json:"role"`
}

type CreateInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发起邀请
func NewCreateInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateInviteLogic {
	return &CreateInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateInviteLogic) CreateInvite(req *types.CreateInviteReq) (resp *types.CreateInviteResp, err error) {
	// 1. 从 context 获取当前用户ID和租户ID
	userId, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := common.GetTenantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 权限校验: 只有 owner 或 admin 可以发起邀请
	userTenant, err := l.svcCtx.UserTenantModel.FindByUserIdAndTenantId(l.ctx, userId, tenantId)
	if err != nil {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "无权限操作")
	}
	if userTenant.Role != "owner" && userTenant.Role != "admin" {
		return nil, xerr.NewErrCodeMsg(xerr.ForbiddenError, "只有管理员可以邀请成员")
	}

	// 3. 检查被邀请用户是否存在
	invitee, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		if err == user.ErrNotFound || err == sqlc.ErrNotFound {
			return nil, xerr.NewErrCodeMsg(xerr.UserNotFoundError, "该邮箱用户未注册")
		}
		return nil, err
	}

	// 4. 检查用户是否已加入该租户
	_, err = l.svcCtx.UserTenantModel.FindByUserIdAndTenantId(l.ctx, invitee.Id, tenantId)
	if err == nil {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "该用户已是团队成员")
	}

	// 5. 生成邀请码
	inviteCode := generateInviteCode(8)

	// 6. 存储到 Redis
	inviteData := InviteData{
		TenantId:     tenantId,
		InviterId:    userId,
		InviteeEmail: req.Email,
		Role:         "member",
	}
	data, _ := json.Marshal(inviteData)
	redisKey := fmt.Sprintf("rag:invite:code:%s", inviteCode)
	err = l.svcCtx.RedisClient.SetexCtx(l.ctx, redisKey, string(data), 86400) // 24小时过期
	if err != nil {
		l.Errorf("存储邀请码到Redis失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "生成邀请码失败")
	}

	// 7. 生成邀请链接
	inviteLink := fmt.Sprintf("http://localhost:5173/join/%s", inviteCode)

	return &types.CreateInviteResp{
		InviteCode: inviteCode,
		InviteLink: inviteLink,
	}, nil
}

// generateInviteCode 生成指定长度的随机邀请码
func generateInviteCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rng.Intn(len(charset))]
	}
	return string(code)
}
