import type { ChatMessage } from '@/types/chat';
import { User, Bot, Copy } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

interface MessageItemProps {
    message: ChatMessage;
}

export default function MessageItem({ message }: MessageItemProps) {
    const isUser = message.role === 'user';

    const copyToClipboard = () => {
        navigator.clipboard.writeText(message.content);
        toast.success('已复制');
    };

    return (
        <div className={cn("flex w-full gap-4 p-6", isUser ? "bg-white" : "bg-gray-50/50")}>
            <div className={cn(
                "w-8 h-8 rounded-full flex items-center justify-center shrink-0 border",
                isUser ? "bg-purple-100 border-purple-200" : "bg-blue-100 border-blue-200"
            )}>
                {isUser ? <User className="w-5 h-5 text-purple-600" /> : <Bot className="w-5 h-5 text-blue-600" />}
            </div>

            <div className="flex-1 space-y-2 overflow-hidden">
                <div className="flex items-center justify-between">
                    <span className="text-sm font-semibold text-gray-900">
                        {isUser ? 'You' : 'AI Assistant'}
                    </span>
                    {!isUser && (
                        <div className="flex gap-2">
                            <Button variant="ghost" size="icon" className="h-6 w-6 text-gray-400" onClick={copyToClipboard}>
                                <Copy className="w-3 h-3" />
                            </Button>
                        </div>
                    )}
                </div>

                {/* Reasoning Content (Thinking Process) */}
                {message.extra?.reasoning && (
                    <div className="mb-4">
                        <details className="group border rounded-lg bg-gray-50 open:bg-white transition-colors">
                            <summary className="flex items-center justify-between px-3 py-2 cursor-pointer text-xs font-medium text-gray-500 hover:text-gray-700 select-none">
                                <div className="flex items-center gap-2">
                                    <span className="w-1.5 h-1.5 rounded-full bg-orange-400 animate-pulse"></span>
                                    <span>思考过程 (Reasoning)</span>
                                </div>
                                <span className="opacity-0 group-hover:opacity-100 transition-opacity">
                                    {/* Chevron icon could go here */}
                                </span>
                            </summary>
                            <div className="px-3 pb-3 pt-1 text-xs text-gray-600 leading-relaxed border-t border-gray-100 whitespace-pre-wrap font-mono">
                                {message.extra.reasoning}
                            </div>
                        </details>
                    </div>
                )}

                <div className="prose prose-sm max-w-none prose-pre:bg-gray-100 prose-pre:p-0">
                    <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        components={{
                            code({ node, inline, className, children, ...props }: any) {
                                const match = /language-(\w+)/.exec(className || '');
                                return !inline && match ? (
                                    <div className="rounded-md border overflow-hidden my-2">
                                        <div className="bg-gray-100 px-3 py-1 text-xs text-gray-500 border-b flex justify-between items-center">
                                            <span>{match[1]}</span>
                                            <button
                                                className="hover:text-gray-900"
                                                onClick={() => navigator.clipboard.writeText(String(children))}
                                            >
                                                Copy
                                            </button>
                                        </div>
                                        <SyntaxHighlighter
                                            style={oneLight}
                                            language={match[1]}
                                            PreTag="div"
                                            customStyle={{ margin: 0, padding: '1rem', background: 'white' }}
                                            {...props}
                                        >
                                            {String(children).replace(/\n$/, '')}
                                        </SyntaxHighlighter>
                                    </div>
                                ) : (
                                    <code className={cn("bg-gray-100 px-1.5 py-0.5 rounded text-sm font-mono text-pink-600", className)} {...props}>
                                        {children}
                                    </code>
                                );
                            },
                            table({ children }) {
                                return (
                                    <div className="overflow-x-auto my-4 border rounded-lg">
                                        <table className="min-w-full divide-y divide-gray-200">
                                            {children}
                                        </table>
                                    </div>
                                );
                            },
                            thead({ children }) {
                                return <thead className="bg-gray-50">{children}</thead>;
                            },
                            th({ children }) {
                                return <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">{children}</th>;
                            },
                            td({ children }) {
                                return <td className="px-4 py-3 text-sm text-gray-500 border-t border-gray-100">{children}</td>;
                            },
                            a({ href, children }) {
                                const isCitation = href?.startsWith('#citation-');
                                if (isCitation) {
                                    return (
                                        <span className="inline-flex items-center justify-center w-5 h-5 ml-0.5 text-[10px] font-medium text-blue-600 bg-blue-50 rounded-full cursor-pointer hover:bg-blue-100 border border-blue-200 align-top transform -translate-y-0.5">
                                            {children}
                                        </span>
                                    );
                                }
                                return <a href={href} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">{children}</a>;
                            }
                        }}
                    >
                        {message.content}
                    </ReactMarkdown>
                </div>

                {/* Citations Preview Area */}
                {message.extra?.citations && message.extra.citations.length > 0 && (
                    <div className="mt-4 pt-4 border-t grid gap-2">
                        <div className="text-xs font-semibold text-gray-500 mb-1">参考来源:</div>
                        {message.extra.citations.map((cite: any, idx: number) => (
                            <div key={idx} className="bg-white border rounded p-3 text-xs text-gray-600 hover:bg-blue-50 transition-colors cursor-pointer group">
                                <div className="font-medium text-gray-900 mb-1 flex items-center gap-2">
                                    <span className="w-4 h-4 rounded-full bg-gray-100 flex items-center justify-center text-[10px]">{idx + 1}</span>
                                    {cite.doc_name}
                                </div>
                                <div className="line-clamp-2">{cite.content}</div>
                            </div>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
}
