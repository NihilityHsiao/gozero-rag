// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"errors"

	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除用户API配置
func NewDeleteUserApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserApiLogic {
	return &DeleteUserApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserApiLogic) DeleteUserApi(req *types.DeleteUserApiReq) (resp *types.DeleteUserApiResp, err error) {
	// 1. 获取当前用户ID
	uid, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 检查配置是否存在且属于该用户
	apiInfo, err := l.svcCtx.UserApiModel.FindOne(l.ctx, uint64(req.Id))
	if err != nil {
		if errors.Is(err, user_api.ErrNotFound) {
			return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "配置不存在")
		}
		l.Errorf("查询用户API配置失败: %v, id: %d", err, req.Id)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	if apiInfo.UserId != uid {
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "无权删除此配置")
	}

	// 3. 删除配置
	err = l.svcCtx.UserApiModel.Delete(l.ctx, uint64(req.Id))
	if err != nil {
		l.Errorf("删除用户API配置失败: %v, id: %d", err, req.Id)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	return &types.DeleteUserApiResp{}, nil
}
