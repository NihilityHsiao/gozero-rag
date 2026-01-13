// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserApiListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户模型列表
func NewGetUserApiListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserApiListLogic {
	return &GetUserApiListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserApiListLogic) GetUserApiList(req *types.GetUserApiListReq) (resp *types.GetUserApiListResp, err error) {
	// 调用 Model 层查询
	list, total, err := l.svcCtx.UserApiModel.FindListByUserId(
		l.ctx,
		req.UserId,
		req.ModelType,
		req.Status,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		l.Errorf("GetUserApiList error: %v", err)
		return nil, err
	}

	// 转换为响应结构
	var apiList []types.UserApiInfo
	for _, item := range list {
		apiList = append(apiList, types.UserApiInfo{
			Id:          int64(item.Id),
			UserId:      item.UserId,
			ConfigName:  item.ConfigName,
			ApiKey:      item.ApiKey,
			BaseUrl:     item.BaseUrl,
			ModelName:   item.ModelName,
			ModelType:   item.ModelType,
			ModelDim:    int(item.ModelDim),
			MaxTokens:   int(item.MaxTokens),
			Temperature: item.Temperature,
			TopP:        item.TopP,
			Timeout:     int(item.Timeout),
			Status:      int(item.Status),
			IsDefault:   int(item.IsDefault),
			CreatedAt:   item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.GetUserApiListResp{
		Total: total,
		List:  apiList,
	}, nil
}
