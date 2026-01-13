// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserApiByModelTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定类型的模型
func NewGetUserApiByModelTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserApiByModelTypeLogic {
	return &GetUserApiByModelTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserApiByModelTypeLogic) GetUserApiByModelType(req *types.GetUserModelReq) (resp *types.GetUserModelResp, err error) {
	// 调用 Model 层查询指定类型的模型
	list, err := l.svcCtx.UserApiModel.FindByUserIdAndModelType(l.ctx, req.UserId, req.ModelType)
	if err != nil {
		l.Errorf("GetUserApiByModelType error: %v", err)
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

	return &types.GetUserModelResp{
		Total: int64(len(apiList)),
		List:  apiList,
	}, nil
}
