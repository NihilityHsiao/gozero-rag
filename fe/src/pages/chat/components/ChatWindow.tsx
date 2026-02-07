import { useEffect, useRef } from 'react';
import { useChatStore } from '@/store/useChatStore';
import MessageItem from './MessageItem';
import ChatInput from './ChatInput';
import { Bot } from 'lucide-react';
import { startNewChat, chatSse, getConversationHistory } from '@/api/chat';
import type { ChatReq } from '@/types/chat';
import { toast } from 'sonner';

export default function ChatWindow() {
    const {
        messages,
        isStreaming,
        addMessage,
        setIsStreaming,
        updateLastMessage,
        currentConversationId,
        setCurrentConversationId,
        setMessages,
        config
    } = useChatStore();
    const scrollRef = useRef<HTMLDivElement>(null);

    // Auto scroll to bottom
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [messages, isStreaming]);

    // Fetch History
    useEffect(() => {
        const fetchHistory = async () => {
            if (isStreaming) return; // Don't fetch if streaming (optimistic updates)

            if (!currentConversationId) {
                setMessages([]);
                return;
            }

            try {
                const res = await getConversationHistory(currentConversationId);
                setMessages(res.list || []);
            } catch (e) {
                console.error('Failed to fetch history', e);
                // toast.error('获取历史消息失败');
            }
        };
        fetchHistory();
    }, [currentConversationId, isStreaming, setMessages]);

    const handleSend = async (text: string) => {
        let conversationId = currentConversationId;

        // 1. Create Conversation if needed
        if (!conversationId) {
            try {
                // Validate Config
                const chatModelId = Number(config.model_id);
                if (!chatModelId) {
                    toast.error('请先在右侧配置面板选择模型');
                    return;
                }

                const startReq: any = { // Use 'any' or import StartNewChatReq type
                    llm_id: config.model_name && config.model_factory ? `${config.model_name}@${config.model_factory}` : undefined,
                    enable_quote_doc: true, // Default
                    enable_llm_keyword_extract: true, // Default
                    enable_tts: false,
                    system_prompt: config.system_prompt,
                    kb_ids: config.knowledge_base_ids,
                    temperature: config.temperature,
                    retrieval_config: {
                        mode: config.retrieval_mode,
                        rerank_mode: config.hybrid_strategy_type,
                        rerank_vector_weight: config.weight_vector,
                        top_n: 10, // Default or add to store
                        rerank_id: config.rerank_model_name && config.rerank_model_factory ? `${config.rerank_model_name}@${config.rerank_model_factory}` : undefined,
                        top_k: config.top_k,
                        score: config.score_threshold
                    }
                };

                // Fallback for ID if names missing (should be set by ConfigPanel)
                if (!startReq.llm_id) startReq.llm_id = chatModelId.toString();

                const resp = await startNewChat(startReq);
                conversationId = resp.conversation_id;
                setCurrentConversationId(conversationId);
            } catch (e) {
                console.error(e);
                toast.error('创建对话失败');
                return;
            }
        }

        // 2. Add User Message
        const userMsgId = Date.now().toString();
        addMessage({
            id: userMsgId,
            conversation_id: conversationId,
            seq_id: messages.length + 1,
            role: 'user',
            content: text,
            type: 'text',
            token_count: 0,
            created_at: new Date().toISOString()
        });

        // 3. Prepare AI Message Placeholder
        setIsStreaming(true);
        const aiMsgId = (Date.now() + 1).toString();
        addMessage({
            id: aiMsgId,
            conversation_id: conversationId,
            seq_id: messages.length + 2,
            role: 'assistant',
            content: '', // Start empty
            type: 'text',
            token_count: 0,
            created_at: new Date().toISOString()
        });

        // 4. Send SSE Request
        const chatModelId = Number(config.model_id);
        if (!chatModelId) {
            toast.error('请先在右侧配置面板选择模型');
            setIsStreaming(false);
            return;
        }

        const req: ChatReq = {
            conversation_id: conversationId,
            message: text,
            chat_model_id: chatModelId,
            prompt: config.system_prompt || '',
            temperature: config.temperature || 0.7,
            knowledge_base_ids: config.knowledge_base_ids || [],
            chat_retrieve_config: {
                mode: config.retrieval_mode || 'hybrid',
                top_k: config.top_k || 10,
                score: config.score_threshold || 0.5,
                rerank_mode: config.hybrid_strategy_type,
                rerank_model_id: config.rerank_model_id ? Number(config.rerank_model_id) : undefined,
                rerank_vector_weight: config.weight_vector,
                rerank_keyword_weight: config.weight_keyword
            }
        };

        console.log('Sending ChatReq:', req);

        let accumulatedContent = '';
        let accumulatedReasoning = '';

        await chatSse(req, (resp) => {
            if (resp.type === 'text' && resp.content) {
                accumulatedContent += resp.content;
                updateLastMessage(accumulatedContent);
            } else if (resp.type === 'reasoning' && resp.reasoning_content) {
                accumulatedReasoning += resp.reasoning_content;
                updateLastMessage(accumulatedContent, { reasoning: accumulatedReasoning });
            } else if (resp.type === 'citation' && resp.retrieval_docs) {
                updateLastMessage(accumulatedContent, { citations: resp.retrieval_docs });
            } else if (resp.type === 'error') {
                toast.error(resp.error_msg);
                updateLastMessage(accumulatedContent + `\n\n[Error: ${resp.error_msg}]`, { isError: true, errorMsg: resp.error_msg });
            } else if (resp.type === 'finish') {
                setIsStreaming(false);
            }
        }, (err) => {
            console.error(err);
            toast.error('请求失败');
            setIsStreaming(false);
        });
    };

    const handleStop = () => {
        setIsStreaming(false);
        // TODO: abort controller logic if needed later
    };

    return (
        <div className="flex-1 flex flex-col h-full bg-white relative">
            <div className="flex-1 overflow-y-auto scroll-smooth" ref={scrollRef}>
                {messages.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-gray-400 space-y-4">
                        <div className="w-16 h-16 bg-gray-100 rounded-2xl flex items-center justify-center">
                            <Bot className="w-8 h-8 text-blue-600" />
                        </div>
                        <p className="text-lg font-medium text-gray-600">今天有什么我可以帮你的吗？</p>
                    </div>
                ) : (
                    <div className="pb-4">
                        {messages.map((msg) => (
                            <MessageItem key={msg.id} message={msg} />
                        ))}
                    </div>
                )}
            </div>

            <ChatInput onSend={handleSend} onStop={handleStop} />
        </div>
    );
}
