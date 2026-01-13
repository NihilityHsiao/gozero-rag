// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"errors"
	"strings"

	"gozero-rag/internal/model/user_api"
	"gozero-rag/internal/xerr"
	"gozero-rag/restful/rag/internal/common"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

// 支持的模型类型
var validModelTypes = map[string]bool{
	"embedding": true,
	"chat":      true,
	"qa":        true,
	"rewrite":   true,
	"rerank":    true,
}

type AddUserApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 添加用户API配置
func NewAddUserApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserApiLogic {
	return &AddUserApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddUserApiLogic) AddUserApi(req *types.AddUserApiReq) (resp *types.AddUserApiResp, err error) {
	uid, err := common.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. 参数校验
	err = l.validateParams(req)
	if err != nil {
		return nil, err
	}

	// 2. config_name为空时，设置为model_name
	configName := req.ConfigName
	if configName == "" {
		configName = req.ModelName
	}

	// 3. 检查model_name是否已存在

	existApi, err := l.svcCtx.UserApiModel.FindOneByUserIdModelTypeModelName(l.ctx, uid, req.ModelType, req.ModelName)
	if err != nil && !errors.Is(err, user_api.ErrNotFound) {
		l.Errorf("查询模型名称失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}
	if existApi != nil {
		return nil, xerr.NewErrCode(xerr.UserApiModelNameExistError)
	}

	// 4. 插入数据库
	newUserApi := &user_api.UserApi{
		ConfigName:  configName,
		UserId:      uid,
		ApiKey:      req.ApiKey,
		BaseUrl:     req.BaseUrl,
		ModelName:   req.ModelName,
		ModelType:   req.ModelType,
		ModelDim:    int64(req.ModelDim),
		MaxTokens:   int64(req.MaxTokens),
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Timeout:     int64(req.Timeout),
		Status:      int64(req.Status),
		IsDefault:   int64(req.IsDefault),
	}

	result, err := l.svcCtx.UserApiModel.Insert(l.ctx, newUserApi)
	if err != nil {
		// 检查是否是唯一键冲突
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			// 判断是哪个唯一键冲突
			if strings.Contains(mysqlErr.Message, "config_name") {
				return nil, xerr.NewErrCodeMsg(xerr.BadRequest, "配置名称已存在")
			}
			if strings.Contains(mysqlErr.Message, "model_name") {
				return nil, xerr.NewErrCode(xerr.UserApiModelNameExistError)
			}
		}
		l.Errorf("插入用户API配置失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCode(xerr.UserApiAddError)
	}

	id, err := result.LastInsertId()
	if err != nil {
		l.Errorf("获取插入ID失败: %v, req: %v", err, req)
		return nil, xerr.NewErrCode(xerr.ServerCommonError)
	}

	// 5. 如果设为默认，则更新默认状态
	if req.IsDefault == 1 {
		err = l.svcCtx.UserApiModel.UpdateDefaultStatus(l.ctx, uid, req.ModelType, id)
		if err != nil {
			// 这里如果更新失败，不应该影响添加成功，只是默认设置失败
			// 但为了数据一致性，最好记录日志
			l.Errorf("设置默认模型失败: %v, uid: %d, modelId: %d", err, uid, id)
			// 可以选择返回错误，或者仅仅记录日志。这里选择记录日志，因为模型已经添加成功了。
			// 或者可以考虑在前端提示部分成功。
			// 为了严谨，这里不返回错误，因为主任务(添加模型)已完成。
		}
	}

	return &types.AddUserApiResp{
		Id: id,
	}, nil
}

// validateParams 参数校验
func (l *AddUserApiLogic) validateParams(req *types.AddUserApiReq) error {
	// api_key不能为空
	if req.ApiKey == "" {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "api_key不能为空")
	}

	// base_url不能为空
	if req.BaseUrl == "" {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "base_url不能为空")
	}

	// model_name不能为空
	if req.ModelName == "" {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "model_name不能为空")
	}

	// model_type不能为空
	if req.ModelType == "" {
		return xerr.NewErrCodeMsg(xerr.BadRequest, "model_type不能为空")
	}

	// model_type只能是embedding、chat、qa、rewrite、rerank
	if !validModelTypes[req.ModelType] {
		return xerr.NewErrCode(xerr.UserApiInvalidModelTypeError)
	}

	return nil
}
