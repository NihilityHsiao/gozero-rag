package response

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
	"gozero-rag/internal/xerr"
	"net/http"
)

type Response struct {
	Code    uint32 `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data"`
}

func Success(data any) *Response {
	return &Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	}
}
func Fail(code uint32, msg string) *Response {
	return &Response{
		Code:    code,
		Message: msg,
	}
}

func OkHandler(_ context.Context, v any) any {
	return Success(v)
}

func ErrHandler(name string) func(ctx context.Context, err error) (int, any) {

	return func(ctx context.Context, err error) (int, any) {

		//错误返回
		errcode := xerr.ServerCommonError
		errmsg := "服务器开小差啦，稍后再来试一试"

		causeErr := errors.Cause(err)                // err类型
		if e, ok := causeErr.(*xerr.CodeError); ok { //自定义错误类型
			//自定义CodeError
			errcode = e.GetErrCode()
			errmsg = e.GetErrMsg()
		} else {
			if gstatus, ok := status.FromError(causeErr); ok { // grpc err错误
				grpcCode := uint32(gstatus.Code())
				if xerr.IsCodeErr(grpcCode) { //区分自定义错误跟系统底层、db等错误，底层、db错误不能返回给前端
					errcode = grpcCode
					errmsg = gstatus.Message()
				}
			}
		}

		logx.WithContext(ctx).Errorf("【API-ERR】 : %+v ", err)

		return http.StatusOK, Fail(errcode, errmsg) // 还是返回200， 错误码和错误信息
	}
}
func NewResponse(r *http.Request, w http.ResponseWriter, resp any, err error) {
	if err == nil {
		// 成功返回
		r := &Response{
			Code:    0,
			Message: "",
			Data:    resp,
		}
		httpx.WriteJson(w, http.StatusOK, r)
		return
	}

	errCode := uint32(10086)
	errMsg := "服务器错误"

	httpx.WriteJson(w, http.StatusBadRequest, &Response{
		Code:    errCode,
		Message: errMsg,
		Data:    nil,
	})

}
