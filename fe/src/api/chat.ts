import request from '@/utils/request';
import type { ChatReq, ChatResp, StartNewChatResp, GetConversationListResp, ChatMessage } from '@/types/chat';

// 开启新对话
// 开启新对话
export async function startNewChat(data: any): Promise<StartNewChatResp> {
    return request.post<any, StartNewChatResp>('/chat/new', data);
}

// 获取会话列表
export async function getConversationList(params: { page: number; page_size: number }): Promise<GetConversationListResp> {
    return request.get<any, GetConversationListResp>('/chat/conversations', { params });
}

// 获取会话历史
export async function getConversationHistory(conversationId: string): Promise<{ list: ChatMessage[] }> {
    return request.get<any, { list: ChatMessage[] }>(`/chat/conversations/${conversationId}/messages`);
}

// 删除会话
export async function deleteConversation(conversationId: string): Promise<void> {
    return request.delete<any, void>(`/chat/conversations/${conversationId}`);
}

// 更新会话
export async function updateConversation(conversationId: string, data: { title: string }): Promise<void> {
    return request.put<any, void>(`/chat/conversations/${conversationId}`, data);
}

// SSE 对话接口
export async function chatSse(
    data: ChatReq,
    onMessage: (resp: ChatResp) => void,
    onError: (err: any) => void
) {
    try {
        const response = await fetch('/api/chat/sse', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                // 如果有鉴权token，需在此处添加
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
            },
            body: JSON.stringify(data),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        if (!response.body) {
            throw new Error('Response body is null');
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (true) {
            const { value, done } = await reader.read();
            if (done) break;

            buffer += decoder.decode(value, { stream: true });

            // 处理 SSE 格式: data: {...}\n\n
            const lines = buffer.split('\n');
            buffer = lines.pop() || ''; // 保留未完整的流

            for (const line of lines) {
                if (line.trim() === '') continue;
                if (line.startsWith('data: ')) {
                    const jsonStr = line.slice(6);
                    if (jsonStr.trim() === '[DONE]') continue;

                    try {
                        const parsed = JSON.parse(jsonStr) as ChatResp;
                        onMessage(parsed);
                    } catch (e) {
                        console.error('Failed to parse SSE message:', e);
                    }
                }
            }
        }
    } catch (err) {
        console.error('SSE Error:', err);
        onError(err);
    }
}
