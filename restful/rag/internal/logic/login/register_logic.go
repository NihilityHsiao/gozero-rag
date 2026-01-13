// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package login

import (
	"context"
	"errors"
	"gozero-rag/internal/jwt"
	"regexp"
	"unicode"

	"gozero-rag/internal/model/user"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

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

	// 2. 检查用户是否已存在
	existUser, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil && !errors.Is(user.ErrNotFound, err) {
		l.Errorf("查询用户失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}
	if existUser != nil {
		return nil, xerr.NewErrCode(xerr.UserAlreadyExistError)
	}

	// 3. 密码加密
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		l.Errorf("密码加密失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 4. 存储到数据库
	newUser := &user.User{
		Username:     req.Username,
		PasswordHash: string(passwordHash),
		Email:        req.Email,
	}
	result, err := l.svcCtx.UserModel.Insert(l.ctx, newUser)
	if err != nil {
		l.Errorf("插入用户失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.UserRegisterError, "注册失败")
	}

	userId, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取用户ID失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 5. 生成JWT Token
	// 生成 JWT token
	jwtCfg := jwt.JwtConfig{
		AccessSecret: l.svcCtx.Config.Auth.AccessSecret,
		AccessExpire: l.svcCtx.Config.Auth.AccessExpire,
	}

	accessToken, expireAt, err := jwt.GenerateToken(jwtCfg, userId, req.Username)
	if err != nil {
		l.Errorf("生成JWT Token失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}
	refreshToken, err := jwt.GenerateRefreshToken(jwtCfg, userId, req.Username)
	if err != nil {
		l.Errorf("生成RefreshToken失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "注册失败")
	}

	// 6. 返回响应
	return &types.LoginResponse{
		Token: types.JwtToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpireAt:     expireAt,
			Uid:          userId,
		},
		User: types.UserInfo{
			UserId:   userId,
			Username: req.Username,
		},
	}, nil
}

// validateParams 参数校验
func (l *RegisterLogic) validateParams(req *types.RegisterRequest) error {
	// 用户名校验: 3-20位，只能包含字母、数字、下划线
	if len(req.Username) < 3 || len(req.Username) > 20 {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "用户名长度必须在3-20位之间")
	}
	//usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	//if !usernameRegex.MatchString(req.Username) {
	//	return xerr.NewErrCodeMsg(xerr.BadRequest, "用户名只能包含字母、数字和下划线")
	//}

	// 邮箱校验
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "邮箱格式不正确")
	}

	// 密码校验: 至少8位，包含大小写字母和数字
	//if len(req.Password) < 8 {
	//	return xerr.NewErrCodeMsg(xerr.BadRequest, "密码长度至少8位")
	//}
	//if !l.isValidPassword(req.Password) {
	//	return xerr.NewErrCodeMsg(xerr.BadRequest, "密码必须包含大小写字母和数字")
	//}

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
