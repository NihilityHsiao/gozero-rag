import React, { useState } from 'react';
import type { RetrievalChunk } from '@/types/retrieval';
import { RetrievalResultItem } from './RetrievalResultItem';
import { RetrievalChunkDetailDialog } from './RetrievalChunkDetailDialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Inbox } from 'lucide-react';

interface RetrievalResultListProps {
    chunks: RetrievalChunk[];
    loading?: boolean;
}

export const RetrievalResultList: React.FC<RetrievalResultListProps> = ({ chunks, loading }) => {
    const [selectedChunk, setSelectedChunk] = useState<RetrievalChunk | null>(null);
    const [dialogOpen, setDialogOpen] = useState(false);

    const handleChunkClick = (chunk: RetrievalChunk) => {
        setSelectedChunk(chunk);
        setDialogOpen(true);
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center h-full p-8 text-muted-foreground">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mb-4"></div>
                <p>正在召回...</p>
            </div>
        );
    }

    if (chunks.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full p-8 text-muted-foreground opacity-50">
                <Inbox className="w-12 h-12 mb-4" />
                <p>暂无召回结果</p>
                <p className="text-xs mt-1">请输入查询语句开始测试</p>
            </div>
        );
    }

    return (
        <>
            <ScrollArea className="h-full pr-4">
                <div className="space-y-4 pb-4">
                    {chunks.map((chunk, index) => (
                        <RetrievalResultItem
                            key={`${chunk.chunk_id}-${index}`}
                            chunk={chunk}
                            onClick={handleChunkClick}
                        />
                    ))}
                </div>
            </ScrollArea>

            <RetrievalChunkDetailDialog
                chunk={selectedChunk}
                open={dialogOpen}
                onOpenChange={setDialogOpen}
            />
        </>
    );
};
