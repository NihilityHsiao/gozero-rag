// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoReq) (resp *types.UserInfo, err error) {
	// 从 JWT context 中获取用户ID
	uid, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("获取用户信息, uid: %d, 请求uid: %d", uid, req.Id)

	// 校验权限：只能查询自己的信息，或者请求的id为0表示查询当前用户
	queryUid := req.Id
	if queryUid == 0 {
		queryUid = uid
	} else if queryUid != uid {
		return nil, xerr.NewErrCode(xerr.InternalError)
	}

	findUser, err := l.svcCtx.UserModel.FindOne(l.ctx, queryUid)
	if err != nil {
		logx.Errorf("查询用户失败,sql error:%v, req:%v", err, req)
		return nil, xerr.NewErrCodeMsg(xerr.InternalError, "用户不存在")
	}

	return &types.UserInfo{
		UserId:   queryUid,
		Username: findUser.Username,
	}, nil
}
