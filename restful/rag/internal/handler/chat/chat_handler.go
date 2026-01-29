// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gozero-rag/restful/rag/internal/logic/chat"
	"gozero-rag/restful/rag/internal/svc"
	"gozero-rag/restful/rag/internal/types"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/threading"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// sse对话
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Buffer size of 16 is chosen as a reasonable default to balance throughput and memory usage.
		// You can change this based on your application's needs.
		// if your go-zero version less than 1.8.1, you need to add 3 lines below.
		// w.Header().Set("Content-Type", "text/event-stream")
		// w.Header().Set("Cache-Control", "no-cache")
		// w.Header().Set("Connection", "keep-alive")
		client := make(chan *types.ChatResp, 16)

		l := chat.NewChatLogic(r.Context(), svcCtx)
		threading.GoSafeCtx(r.Context(), func() {
			defer close(client)
			err := l.Chat(&req, client)
			if err != nil {
				logc.Errorw(r.Context(), "ChatHandler", logc.Field("error", err))
				return
			}
		})

		for {
			select {
			case data, ok := <-client:
				if !ok {
					return
				}
				if data.Type != "" && data.Type != "text" {
					if _, err := fmt.Fprintf(w, "event: %s\n", data.Type); err != nil {
						logc.Errorw(r.Context(), "ChatHandler: failed into write event type", logc.Field("error", err))
						return
					}
				} else if data.Type == "text" {
					// 显式把 text 类型映射为 standard message event (optional, but good for clarity)
					// Or just omit event: for text to use default 'message' event.
					// The plan says: event: message for text.
					if _, err := fmt.Fprintf(w, "event: message\n"); err != nil {
						logc.Errorw(r.Context(), "ChatHandler: failed into write event type", logc.Field("error", err))
						return
					}
				}

				output, err := json.Marshal(data)
				if err != nil {
					logc.Errorw(r.Context(), "ChatHandler", logc.Field("error", err))
					continue
				}

				if _, err := fmt.Fprintf(w, "data: %s\n\n", string(output)); err != nil {
					logc.Errorw(r.Context(), "ChatHandler", logc.Field("error", err))
					return
				}
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
			case <-r.Context().Done():
				return
			}
		}
	}
}
