import React from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { FileText, Database, Sparkles } from 'lucide-react';
import type { RetrievalChunk } from '@/types/retrieval';
import { cn } from '@/lib/utils';

interface RetrievalResultItemProps {
    chunk: RetrievalChunk;
    onClick: (chunk: RetrievalChunk) => void;
}

export const RetrievalResultItem: React.FC<RetrievalResultItemProps> = ({ chunk, onClick }) => {

    const getSourceIcon = (source: string) => {
        switch (source) {
            case 'vector': return <Database className="w-3 h-3 mr-1" />;
            case 'keyword': return <FileText className="w-3 h-3 mr-1" />;
            case 'rerank': return <Sparkles className="w-3 h-3 mr-1" />;
            default: return <Database className="w-3 h-3 mr-1" />;
        }
    };

    const getScoreColor = (score: number) => {
        if (score >= 0.8) return 'text-green-600';
        if (score >= 0.5) return 'text-yellow-600';
        return 'text-red-600';
    };

    const getSourceLabel = (source: string) => {
        switch (source) {
            case 'vector': return '向量';
            case 'keyword': return '关键词';
            case 'rerank': return '重排序';
            default: return source;
        }
    };

    return (
        <Card
            className="hover:shadow-md transition-all cursor-pointer hover:border-blue-300 active:scale-[0.99]"
            onClick={() => onClick(chunk)}
        >
            <CardContent className="p-4">
                <div className="flex justify-between items-start mb-2">
                    <div className="flex items-center space-x-2">
                        <Badge variant="outline" className="flex items-center text-xs capitalize bg-background">
                            {getSourceIcon(chunk.source)}
                            {getSourceLabel(chunk.source)}
                        </Badge>
                        <span className="text-xs font-mono text-muted-foreground truncate max-w-[120px]" title={chunk.chunk_id}>
                            {chunk.chunk_id}
                        </span>
                    </div>
                    <div className="flex items-center space-x-2">
                        <span className={cn("text-xs font-bold font-mono", getScoreColor(chunk.score))}>
                            {chunk.score.toFixed(4)}
                        </span>
                    </div>
                </div>

                <div className="text-sm bg-muted/30 p-3 rounded-md mb-2 line-clamp-3 max-h-[100px] overflow-hidden text-muted-foreground/90">
                    {chunk.content}
                </div>

                <div className="flex justify-between items-center mt-2">
                    <div className="text-xs text-muted-foreground flex items-center bg-muted/50 px-2 py-0.5 rounded-full max-w-full" title={chunk.doc_name}>
                        <FileText className="w-3 h-3 mr-1 flex-shrink-0" />
                        <span className="truncate">{chunk.doc_name || "未知文档"}</span>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
};
