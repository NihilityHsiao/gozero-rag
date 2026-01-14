// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"
	"unicode"

	"gozero-rag/internal/jwt"
	"gozero-rag/internal/model/tenant"
	"gozero-rag/internal/model/user"
	"gozero-rag/internal/model/user_tenant"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.LoginResponse, err error) {
	// 1. 参数校验
	if err := l.validateParams(req); err != nil {
		return nil, err
	}

	// 2. 检查邮箱是否已存在
	existUser, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil && !errors.Is(user.ErrNotFound, err) {
		l.Errorf("查询用户失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}
	if existUser != nil {
		return nil, xerr.NewErrCode(xerr.UserAlreadyExistError)
	}

	// 3. 生成 UUID v7 作为用户ID
	userId, err := uuid.NewV7()
	if err != nil {
		l.Errorf("生成UUID失败: %v", err)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}
	userIdStr := userId.String()

	// 4. 密码加密
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("密码加密失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 5. 创建用户
	newUser := &user.User{
		Id:            userIdStr,
		Nickname:      req.Nickname,
		Password:      string(passwordHash),
		Email:         req.Email,
		Language:      "Chinese",
		ColorSchema:   "Bright",
		Timezone:      "UTC+8\tAsia/Shanghai",
		LastLoginTime: sql.NullTime{Time: time.Now(), Valid: true},
		IsActive:      1,
		Status:        1,
	}
	_, err = l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Errorf("插入用户失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.UserRegisterError, "注册失败")
	}

	// 6. 创建租户 (租户ID = 用户ID)
	tenantId := userIdStr
	tenantName := req.Nickname + "的工作空间"
	newTenant := &tenant.Tenant{
		Id:     tenantId,
		Name:   sql.NullString{String: tenantName, Valid: true},
		Status: 1,
	}
	_, err = l.svcCtx.TenantModel.Insert(l.ctx, newTenant)
	if err != nil {
		l.Errorf("创建租户失败: %v, userId: %s", err, userIdStr)
		// 注意: 这里可能需要回滚用户创建，但简单起见先不处理
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 7. 创建用户-租户关联
	userTenantId, _ := uuid.NewV7()
	newUserTenant := &user_tenant.UserTenant{
		Id:        userTenantId.String(),
		UserId:    userIdStr,
		TenantId:  tenantId,
		Role:      user_tenant.RoleOwner,
		InvitedBy: userIdStr, // 自己创建
		Status:    1,
	}
	_, err = l.svcCtx.UserTenantModel.Insert(l.ctx, newUserTenant)
	if err != nil {
		l.Errorf("创建用户租户关联失败: %v, userId: %s", err, userIdStr)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 8. 生成JWT Token
	jwtCfg := jwt.JwtConfig{
		AccessSecret: l.svcCtx.Config.Auth.AccessSecret,
		AccessExpire: l.svcCtx.Config.Auth.AccessExpire,
	}

	accessToken, expireAt, err := jwt.GenerateToken(jwtCfg, userIdStr, tenantId, req.Nickname)
	if err != nil {
		l.Errorf("生成JWT Token失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}
	refreshToken, err := jwt.GenerateRefreshToken(jwtCfg, userIdStr, tenantId, req.Nickname)
	if err != nil {
		l.Errorf("生成RefreshToken失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 在 Update 前确保所有字段都正确，这里 newUser 已经是完整的
	err = l.svcCtx.UserModel.Update(l.ctx, newUser)
	if err != nil {
		l.Errorf("更新用户AccessToken失败: %v, userId: %s", err, userIdStr)
		// 不阻断流程
	}

	// 9. 返回响应
	return &types.LoginResponse{
		Token: types.JwtToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpireAt:     expireAt,
		},
		User: types.UserInfo{
			UserId:   userIdStr,
			Nickname: req.Nickname,
			Email:    req.Email,
		},
		CurrentTenant: types.TenantInfo{
			TenantId: tenantId,
			Name:     tenantName,
			Role:     "owner",
		},
		Tenants: []types.TenantInfo{
			{
				TenantId: tenantId,
				Name:     tenantName,
				Role:     "owner",
			},
		},
	}, nil
}

// validateParams 参数校验
func (l *RegisterLogic) validateParams(req *types.RegisterRequest) error {
	// 昵称校验: 1-50位
	if len(req.Nickname) < 1 || len(req.Nickname) > 50 {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "昵称长度必须在1-50位之间")
	}

	// 邮箱校验
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "邮箱格式不正确")
	}

	// 密码校验: 至少6位
	if len(req.Password) < 6 {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "密码长度至少6位")
	}

	// 确认密码校验
	if req.Password != req.ConfirmPassword {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "两次输入的密码不一致")
	}

	return nil
}

// isValidPassword 检查密码是否包含大小写字母和数字
func (l *RegisterLogic) isValidPassword(password string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}
