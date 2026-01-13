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

type SetDefaultModelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 将模型设为默认模型
func NewSetDefaultModelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultModelLogic {
	return &SetDefaultModelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultModelLogic) SetDefaultModel(req *types.SetDefaultModelReq) (resp *types.SetDefaultModelResp, err error) {
	// 1. 获取当前用户ID
	uid, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 检查模型是否存在
	apiInfo, err := l.svcCtx.UserApiModel.FindOne(l.ctx, uint64(req.ModelId))
	if err != nil {
		if errors.Is(err, user_api.ErrNotFound) {
			return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "模型不存在")
		}
		l.Errorf("查询用户API配置失败: %v, id: %d", err, req.ModelId)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	// 3. 校验权限
	if apiInfo.UserId != uid {
		return nil, xerr.NewErrCodeMsg(xerr.Unauthorized, "无权操作此模型")
	}

	// 4. 校验模型类型是否匹配 (防止误传)
	if apiInfo.ModelType != req.ModelType {
		return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "模型类型不匹配")
	}

	// 5. 更新默认状态 (事务)
	err = l.svcCtx.UserApiModel.UpdateDefaultStatus(l.ctx, uid, req.ModelType, req.ModelId)
	if err != nil {
		l.Errorf("设置默认模型失败: %v, uid: %d, modelId: %d", err, uid, req.ModelId)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	return &types.SetDefaultModelResp{}, nil
}
