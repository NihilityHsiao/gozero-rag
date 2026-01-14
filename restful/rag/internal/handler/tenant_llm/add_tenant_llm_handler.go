// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant_llm

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gozero-rag/restful/rag/internal/logic/tenant_llm"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"
)

// 批量添加租户LLM配置
func AddTenantLlmHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddTenantLlmReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tenant_llm.NewAddTenantLlmLogic(r.Context(), svcCtx)
		resp, err := l.AddTenantLlm(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
