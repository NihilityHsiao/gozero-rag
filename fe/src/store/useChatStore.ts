import { create } from 'zustand';
import type { ChatConversation, ChatMessage, ChatModelConfig } from '@/types/chat';

interface ChatState {
    conversations: ChatConversation[];
    currentConversationId: string | null;
    messages: ChatMessage[];
    isStreaming: boolean;
    streamingContent: string; // Temporary content buffer for streaming

    // Config
    config: ChatModelConfig;

    // Actions
    setConfig: (config: Partial<ChatModelConfig>) => void;
    setConversations: (conversations: ChatConversation[]) => void;
    setCurrentConversationId: (id: string | null) => void;
    setMessages: (messages: ChatMessage[]) => void;
    addMessage: (message: ChatMessage) => void;
    updateLastMessage: (content: string, extra?: any) => void;
    setIsStreaming: (isStreaming: boolean) => void;
    setStreamingContent: (content: string) => void;
    resetStreaming: () => void;
}

export const useChatStore = create<ChatState>((set) => ({
    conversations: [],
    currentConversationId: null,
    messages: [],
    isStreaming: false,
    streamingContent: '',

    config: {
        temperature: 0.7,
        retrieval_mode: 'hybrid',
        top_k: 10,
        score_threshold: 0.5,
        hybrid_strategy_type: 'weighted',
        weight_vector: 0.7,
        weight_keyword: 0.3
    } as ChatModelConfig,
    setConfig: (config) => set((state) => ({ config: { ...state.config, ...config } })),

    setConversations: (conversations) => set({ conversations }),
    setCurrentConversationId: (id) => set({ currentConversationId: id }),
    setMessages: (messages) => set({ messages }),
    addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),

    updateLastMessage: (content, extra) => set((state) => {
        const newMessages = [...state.messages];
        if (newMessages.length > 0) {
            const lastMsg = newMessages[newMessages.length - 1];
            lastMsg.content = content;
            if (extra) {
                lastMsg.extra = { ...lastMsg.extra, ...extra };
            }
        }
        return { messages: newMessages };
    }),

    setIsStreaming: (isStreaming) => set({ isStreaming }),
    setStreamingContent: (content) => set({ streamingContent: content }),
    resetStreaming: () => set({ isStreaming: false, streamingContent: '' }),
}));
