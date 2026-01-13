// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login

import (
	"context"
	"gozero-rag/internal/jwt"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/metric"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"
	"time"

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
	if len(req.Username) == 0 || len(req.Password) == 0 {
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "用户名或密码不能为空")
	}

	// 2. 查询用户
	userInfo, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil {
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.UserNotFoundError, "用户名或密码错误")
	}

	// 3. 比对密码
	err = bcrypt.CompareHashAndPassword([]byte(userInfo.PasswordHash), []byte(req.Password))
	if err != nil {
		metric.LoginFail()
		return nil, xerr.NewErrCode(xerr.UserPasswordError)
	}

	// 4. 生成JWT Token
	jwtCfg := jwt.JwtConfig{
		AccessSecret: l.svcCtx.Config.Auth.AccessSecret,
		AccessExpire: l.svcCtx.Config.Auth.AccessExpire,
	}

	accessToken, expireAt, err := jwt.GenerateToken(jwtCfg, userInfo.Id, userInfo.Username)
	if err != nil {
		l.Errorf("生成JWT Token失败: %v, req: %v", err, req)
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "登录失败")
	}
	refreshToken, err := jwt.GenerateRefreshToken(jwtCfg, userInfo.Id, userInfo.Username)
	if err != nil {
		l.Errorf("生成RefreshToken失败: %v, req: %v", err, req)
		metric.LoginFail()
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "登录失败")
	}

	return &types.LoginResponse{
		Token: types.JwtToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpireAt:     expireAt,
			Uid:          userInfo.Id,
		},
		User: types.UserInfo{
			UserId:   userInfo.Id,
			Username: userInfo.Username,
		},
	}, nil
}
