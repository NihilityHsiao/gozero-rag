import { useState, useRef, type KeyboardEvent } from 'react';
import { SendHorizontal, Paperclip, StopCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { useChatStore } from '@/store/useChatStore';

interface ChatInputProps {
    onSend: (message: string) => void;
    onStop: () => void;
}

export default function ChatInput({ onSend, onStop }: ChatInputProps) {
    const [input, setInput] = useState('');
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const { isStreaming } = useChatStore();

    const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    const handleSend = () => {
        if (!input.trim() || isStreaming) return;
        onSend(input);
        setInput('');

        // Reset height
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
        }
    };

    const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setInput(e.target.value);
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
        }
    };

    return (
        <div className="p-4 border-t bg-white">
            <div className="max-w-4xl mx-auto relative flex items-end gap-2 bg-gray-50 border rounded-xl p-2 focus-within:ring-1 focus-within:ring-blue-600 focus-within:border-blue-600 ring-offset-2 transition-all">
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 text-gray-400 hover:text-gray-600 shrink-0 mb-0.5"
                    title="Upload file (coming soon)"
                >
                    <Paperclip className="w-5 h-5" />
                </Button>

                <Textarea
                    ref={textareaRef}
                    value={input}
                    onChange={handleInput}
                    onKeyDown={handleKeyDown}
                    placeholder="发送消息..."
                    className="min-h-[40px] max-h-[200px] border-0 bg-transparent focus-visible:ring-0 resize-none py-2.5 px-0"
                    rows={1}
                />

                {isStreaming ? (
                    <Button
                        size="icon"
                        onClick={onStop}
                        className="h-9 w-9 bg-red-100 text-red-600 hover:bg-red-200 shrink-0 mb-0.5 rounded-lg"
                    >
                        <StopCircle className="w-5 h-5" fill="currentColor" fillOpacity={0.2} />
                    </Button>
                ) : (
                    <Button
                        size="icon"
                        onClick={handleSend}
                        disabled={!input.trim()}
                        className="h-9 w-9 bg-blue-600 hover:bg-blue-700 shrink-0 mb-0.5 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        <SendHorizontal className="w-5 h-5" />
                    </Button>
                )}
            </div>
            <div className="text-center text-xs text-gray-400 mt-2">
                Antigravity AI 可能会犯错。请核实重要信息。
            </div>
        </div>
    );
}
