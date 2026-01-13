// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"errors"
	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/xerr"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserApiInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取模型信息
func NewGetUserApiInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserApiInfoLogic {
	return &GetUserApiInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserApiInfoLogic) GetUserApiInfo(req *types.GetUserApiInfoReq) (resp *types.UserApiInfo, err error) {
	if req.Id == 0 {
		return nil, xerr.NewBadRequestErrMsg("ID不能为空")
	}

	userApi, err := l.svcCtx.UserApiModel.FindOne(l.ctx, req.Id)
	if err != nil {
		if errors.Is(err, user_api.ErrNotFound) {
			return nil, xerr.NewErrCodeMsg(xerr.UserApiNotFoundError, "模型配置不存在")
		}

		return nil, xerr.NewErrCodeMsg(xerr.ServerCommonError, "查询模型配置失败")
	}

	resp = &types.UserApiInfo{
		Id:          int64(userApi.Id),
		UserId:      userApi.UserId,
		ConfigName:  userApi.ConfigName,
		ApiKey:      userApi.ApiKey,
		BaseUrl:     userApi.BaseUrl,
		ModelName:   userApi.ModelName,
		ModelType:   userApi.ModelType,
		ModelDim:    int(userApi.ModelDim),
		MaxTokens:   int(userApi.MaxTokens),
		Temperature: userApi.Temperature,
		TopP:        userApi.TopP,
		Timeout:     int(userApi.Timeout),
		Status:      int(userApi.Status),
		IsDefault:   int(userApi.IsDefault),
		CreatedAt:   userApi.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   userApi.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return
}
