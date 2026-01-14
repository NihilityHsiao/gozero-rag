// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gozero-rag/internal/jwt"
	"gozero-rag/internal/model/user"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/metric"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {

	start := time.Now()

	defer func() {
		status := "success"
		if err != nil {
			status = "fail"
		}
		metric.UserLoginCount.Inc(status)
		duration := time.Since(start).Milliseconds()

		metric.ApiLatencyHistogram.Observe(duration, "login_logic")

	}()

	// 1. 参数校验
	if len(req.Email) == 0 || len(req.Password) == 0 {
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "邮箱或密码不能为空")
	}

	// 2. 根据邮箱查询用户
	userInfo, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			metric.LoginFail()
			return nil, xerr.NewErrCodeMsg(xerr.UserNotFoundError, "邮箱或密码错误")
		}
		l.Errorf("查询用户失败: %v, email: %s", err, req.Email)
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "登录失败")
	}

	// 3. 比对密码

	err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(req.Password))
	if err != nil {
		metric.LoginFail()
		return nil, xerr.NewErrCode(xerr.UserPasswordError)
	}

	// 4. 检查用户是否激活
	if userInfo.IsActive != 1 {
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "账户已被禁用")
	}

	// 5. 查询用户关联的租户列表
	userTenants, err := l.svcCtx.UserTenantModel.FindByUserId(l.ctx, userInfo.Id)
	if err != nil {
		l.Errorf("查询用户租户关联失败: %v, userId: %s", err, userInfo.Id)
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "登录失败")
	}

	// 构建租户列表
	tenants := make([]types.TenantInfo, 0, len(userTenants))
	for _, ut := range userTenants {
		tenants = append(tenants, types.TenantInfo{
			TenantId: ut.TenantId,
			Name:     ut.TenantName,
			Role:     ut.Role,
		})
	}

	// 默认使用第一个租户作为当前租户
	var currentTenant types.TenantInfo
	if len(tenants) > 0 {
		currentTenant = tenants[0]
	}

	// 7. 生成JWT Token
	jwtCfg := jwt.JwtConfig{
		AccessSecret: l.svcCtx.Config.Auth.AccessSecret,
		AccessExpire: l.svcCtx.Config.Auth.AccessExpire,
	}

	accessToken, expireAt, err := jwt.GenerateToken(jwtCfg, userInfo.Id, currentTenant.TenantId, userInfo.Nickname)
	if err != nil {
		l.Errorf("生成JWT Token失败: %v, email: %s", err, req.Email)
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "登录失败")
	}
	refreshToken, err := jwt.GenerateRefreshToken(jwtCfg, userInfo.Id, currentTenant.TenantId, userInfo.Nickname)
	if err != nil {
		l.Errorf("生成RefreshToken失败: %v, email: %s", err, req.Email)
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "登录失败")
	}

	// 6. 更新用户状态 (LastLoginTime 和 AccessToken)
	userInfo.LastLoginTime = sql.NullTime{Time: time.Now(), Valid: true}
	err = l.svcCtx.UserModel.Update(l.ctx, userInfo)
	if err != nil {
		l.Errorf("更新用户登录信息失败: %v, userId: %s", err, userInfo.Id)
		// 不影响登录流程，只记录日志
	}

	// 8. 返回响应
	return &types.LoginResponse{
		Token: types.JwtToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpireAt:     expireAt,
		},
		User: types.UserInfo{
			UserId:   userInfo.Id,
			Nickname: userInfo.Nickname,
			Email:    userInfo.Email,
			Avatar:   userInfo.Avatar.String,
			Language: userInfo.Language,
		},
		CurrentTenant: currentTenant,
		Tenants:       tenants,
	}, nil
}
