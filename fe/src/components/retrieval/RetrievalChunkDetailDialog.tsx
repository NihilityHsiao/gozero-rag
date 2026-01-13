import React from 'react';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { FileText, Database, Sparkles, Copy } from 'lucide-react';
import { Button } from '@/components/ui/button';
import type { RetrievalChunk } from '@/types/retrieval';
import { toast } from 'sonner';

interface RetrievalChunkDetailProps {
    chunk: RetrievalChunk | null;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export const RetrievalChunkDetailDialog: React.FC<RetrievalChunkDetailProps> = ({
    chunk,
    open,
    onOpenChange,
}) => {
    if (!chunk) return null;

    const getSourceIcon = (source: string) => {
        switch (source) {
            case 'vector': return <Database className="w-3 h-3 mr-1" />;
            case 'keyword': return <FileText className="w-3 h-3 mr-1" />;
            case 'rerank': return <Sparkles className="w-3 h-3 mr-1" />;
            default: return <Database className="w-3 h-3 mr-1" />;
        }
    };

    const getSourceLabel = (source: string) => {
        switch (source) {
            case 'vector': return '向量';
            case 'keyword': return '关键词';
            case 'rerank': return '重排序';
            default: return source;
        }
    };

    const handleCopy = () => {
        navigator.clipboard.writeText(chunk.content);
        toast.success("内容已复制");
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col">
                <DialogHeader>
                    <div className="flex items-center justify-between pr-8">
                        <DialogTitle className="flex items-center gap-2">
                            段落详情
                        </DialogTitle>
                        <Badge variant={chunk.score >= 0.8 ? "default" : "secondary"} className="text-sm font-mono">
                            分值: {chunk.score.toFixed(4)}
                        </Badge>
                    </div>
                </DialogHeader>

                <div className="flex-1 overflow-hidden flex flex-col gap-4">
                    {/* Metadata */}
                    <div className="flex flex-wrap gap-2 text-sm text-muted-foreground border-b pb-4">
                        <div className="flex items-center bg-muted px-2 py-1 rounded">
                            {getSourceIcon(chunk.source)}
                            <span className="capitalize">{getSourceLabel(chunk.source)}</span>
                        </div>
                        <div className="flex items-center bg-muted px-2 py-1 rounded font-mono">
                            ID: {chunk.chunk_id}
                        </div>
                        <div className="flex items-center bg-muted px-2 py-1 rounded">
                            <FileText className="w-3 h-3 mr-1" />
                            {chunk.doc_name || "未知文档"}
                        </div>
                    </div>

                    {/* Content */}
                    <div className="flex-1 min-h-0 relative bg-muted/30 rounded-md border">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="absolute top-2 right-2 h-8 w-8 z-10 bg-background/50 hover:bg-background"
                            onClick={handleCopy}
                        >
                            <Copy className="h-4 w-4" />
                        </Button>
                        <ScrollArea className="h-[400px] w-full p-4">
                            <div className="whitespace-pre-wrap text-sm leading-relaxed">
                                {chunk.content}
                            </div>
                        </ScrollArea>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
};
