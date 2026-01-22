import React, { useEffect, useState } from 'react';
import type { KnowledgeDocumentChunkInfo } from '@/types';
import { Button } from '@/components/ui/button';
import { X, Copy, Eye, FileText } from 'lucide-react';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { toast } from 'sonner';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';

class MarkdownErrorBoundary extends React.Component<
    { children: React.ReactNode; fallback: React.ReactNode },
    { hasError: boolean }
> {
    constructor(props: any) {
        super(props);
        this.state = { hasError: false };
    }

    static getDerivedStateFromError() {
        return { hasError: true };
    }

    componentDidCatch(error: any) {
        console.error('Markdown render error:', error);
    }

    render() {
        if (this.state.hasError) {
            return this.props.fallback;
        }
        return this.props.children;
    }
}

interface EditChunkDrawerProps {
    chunk: KnowledgeDocumentChunkInfo | null;
    isOpen: boolean;
    onClose: () => void;
    onSave?: (id: string, newText: string) => Promise<void>;
}

const EditChunkDrawer: React.FC<EditChunkDrawerProps> = ({ chunk, isOpen, onClose, onSave }) => {
    const [text, setText] = useState('');
    const [isSaving, setIsSaving] = useState(false);
    const [activeTab, setActiveTab] = useState<'content' | 'metadata'>('content');
    const [viewMode, setViewMode] = useState<'preview' | 'source'>('preview');

    useEffect(() => {
        if (chunk) {
            // 使用 content 字段，兼容旧的 chunk_text
            const fullText = chunk.content || chunk.chunk_text || '';
            setText(fullText);

            // 初始化视图模式：QA切片默认源码，文档切片默认预览
            const isQA = chunk.id.startsWith('qa-');
            setViewMode(isQA ? 'source' : 'preview');
        }
    }, [chunk]);

    if (!isOpen || !chunk) return null;

    // 计算字符数
    const charCount = text.length;
    const isQA = chunk.id.startsWith('qa-');
    const canPreview = !isQA && chunk.id.startsWith('chunk-');

    const handleSave = async () => {
        if (!onSave) {
            toast.info("保存切片功能即将上线");
            return;
        }
        setIsSaving(true);
        try {
            await onSave(chunk.id, text);
            toast.success("切片更新成功");
            onClose();
        } catch (e) {
            toast.error("更新切片失败");
        } finally {
            setIsSaving(false);
        }
    };

    const copyToClipboard = () => {
        navigator.clipboard.writeText(text);
        toast.success("已复制到剪贴板");
    };

    return (
        <>
            {/* Backdrop */}
            <div
                className="fixed inset-0 bg-black/20 z-40 transition-opacity"
                onClick={onClose}
            />

            {/* Drawer */}
            <div className="fixed inset-y-0 right-0 z-50 w-[600px] bg-white shadow-2xl transform transition-transform duration-300 ease-in-out flex flex-col border-l">
                {/* Header */}
                <div className="flex items-center justify-between px-6 py-4 border-b">
                    <div>
                        <div className="flex items-center gap-2">
                            <h2 className="text-lg font-semibold">切片详情</h2>
                            <Badge variant="outline" className="font-mono text-xs text-gray-500">
                                #{chunk.id.substring(0, 8)}
                            </Badge>
                        </div>
                        <p className="text-xs text-gray-500 mt-1">
                            状态: <span className={chunk.status === 1 ? "text-green-600 font-medium" : "text-gray-500"}>
                                {chunk.status === 1 ? '启用' : '禁用'}
                            </span>
                            {' · '}
                            {charCount} 字符
                        </p>
                    </div>
                    <div className="flex items-center gap-2">
                        <Button variant="ghost" size="icon" onClick={onClose}>
                            <X className="h-5 w-5" />
                        </Button>
                    </div>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto p-6 bg-gray-50/50">
                    <div className="w-full">
                        <div className="flex gap-1 bg-gray-100/50 p-1 rounded-lg mb-4 w-fit">
                            <button
                                onClick={() => setActiveTab('content')}
                                className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${activeTab === 'content'
                                    ? 'bg-white text-gray-900 shadow-sm'
                                    : 'text-gray-500 hover:text-gray-900'
                                    }`}
                            >
                                内容
                            </button>
                            <button
                                onClick={() => setActiveTab('metadata')}
                                className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors ${activeTab === 'metadata'
                                    ? 'bg-white text-gray-900 shadow-sm'
                                    : 'text-gray-500 hover:text-gray-900'
                                    }`}
                            >
                                元数据
                            </button>
                        </div>

                        {activeTab === 'content' && (
                            <div className="relative space-y-4 animate-in fade-in slide-in-from-left-2 duration-300">
                                <div className="flex items-center justify-between mb-2">
                                    <div className="flex items-center gap-2">
                                        <label className="text-sm font-medium text-gray-700">切片文本</label>
                                        {/* View Mode Toggle */}
                                        {canPreview && (
                                            <div className="flex items-center bg-gray-100 rounded-md p-0.5 ml-2">
                                                <button
                                                    onClick={() => setViewMode('preview')}
                                                    className={`p-1 rounded-sm transition-all ${viewMode === 'preview' ? 'bg-white shadow-sm text-primary' : 'text-gray-400 hover:text-gray-600'}`}
                                                    title="预览"
                                                >
                                                    <Eye className="h-3.5 w-3.5" />
                                                </button>
                                                <button
                                                    onClick={() => setViewMode('source')}
                                                    className={`p-1 rounded-sm transition-all ${viewMode === 'source' ? 'bg-white shadow-sm text-primary' : 'text-gray-400 hover:text-gray-600'}`}
                                                    title="源码"
                                                >
                                                    <FileText className="h-3.5 w-3.5" />
                                                </button>
                                            </div>
                                        )}
                                    </div>
                                    <Button variant="ghost" size="sm" className="h-6" onClick={copyToClipboard}>
                                        <Copy className="h-3 w-3 mr-1" /> 复制
                                    </Button>
                                </div>

                                {viewMode === 'preview' && canPreview ? (
                                    <div className="min-h-[400px] bg-white p-6 border rounded-md shadow-sm">
                                        <div className="prose prose-sm max-w-none text-gray-700">
                                            <MarkdownErrorBoundary
                                                fallback={<div className="whitespace-pre-wrap font-mono text-sm">{text}</div>}
                                            >
                                                <ReactMarkdown
                                                    remarkPlugins={[remarkGfm]}
                                                    components={{
                                                        img: ({ node, ...props }) => <span className="text-xs text-gray-400 block p-2 border border-dashed rounded bg-gray-50">[Image: {props.alt}]</span>,
                                                    }}
                                                >
                                                    {text}
                                                </ReactMarkdown>
                                            </MarkdownErrorBoundary>
                                        </div>
                                    </div>
                                ) : (
                                    <Textarea
                                        className="min-h-[400px] font-mono text-sm leading-relaxed resize-none bg-white p-4"
                                        value={text}
                                        onChange={(e) => setText(e.target.value)}
                                        placeholder="输入文本..."
                                    />
                                )}
                            </div>
                        )}

                        {activeTab === 'metadata' && (
                            <div className="bg-white rounded-lg border p-4 space-y-4 animate-in fade-in slide-in-from-right-2 duration-300">
                                <div>
                                    <h4 className="text-sm font-medium text-gray-500 mb-2">关键词</h4>
                                    <div className="flex flex-wrap gap-1.5">
                                        {chunk.important_keywords && chunk.important_keywords.length > 0 ? (
                                            chunk.important_keywords.map((kw, idx) => (
                                                <Badge key={idx} variant="outline" className="text-xs">
                                                    {kw}
                                                </Badge>
                                            ))
                                        ) : (
                                            <span className="text-sm text-gray-400">无关键词</span>
                                        )}
                                    </div>
                                </div>
                                {chunk.metadata && (
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-500 mb-2">JSON 元数据</h4>
                                        <pre className="bg-gray-50 p-3 rounded text-xs font-mono overflow-x-auto text-gray-700">
                                            {JSON.stringify(chunk.metadata, null, 2)}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </div>

                {/* Footer */}
                <div className="p-4 border-t bg-white flex justify-end gap-3">
                    <Button variant="outline" onClick={onClose}>取消</Button>
                    {/* 只有在源码模式下，或者处于预览模式但文本未变更（预览即最新）时才显示保存，
                        但因为 React state text 始终是最新的，所以保存始终保存 text。
                        如果用户在预览模式想保存，也是可以的。
                    */}
                    <Button onClick={handleSave} disabled={isSaving}>
                        {isSaving ? '保存中...' : '保存更改'}
                    </Button>
                </div>
            </div>
        </>
    );
};

export default EditChunkDrawer;

