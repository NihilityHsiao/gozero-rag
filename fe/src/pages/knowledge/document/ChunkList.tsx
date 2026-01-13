import React from 'react';
import type { KnowledgeDocumentChunkInfo } from '@/types';
import ChunkCard from './ChunkCard';
import { Button } from '@/components/ui/button';
import { ChevronLeft, ChevronRight, Loader2 } from 'lucide-react';

interface ChunkListProps {
    chunks: KnowledgeDocumentChunkInfo[];
    loading: boolean;
    selectedIds: string[];
    total: number;
    page: number;
    pageSize: number;
    onPageChange: (page: number) => void;
    onSelectChunk: (id: string, checked: boolean) => void;
    onEdit: (chunk: KnowledgeDocumentChunkInfo) => void;
}

const ChunkList: React.FC<ChunkListProps> = ({
    chunks,
    loading,
    selectedIds,
    total,
    page,
    pageSize,
    onPageChange,
    onSelectChunk,
    onEdit
}) => {
    if (loading && chunks.length === 0) {
        return (
            <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
            </div>
        );
    }

    if (chunks.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-64 text-gray-500">
                <p>暂无切片数据。</p>
            </div>
        );
    }

    const totalPages = Math.ceil(total / pageSize);

    return (
        <div className="space-y-4 pb-20">
            <div className="grid grid-cols-1 gap-4">
                {chunks.map((chunk) => (
                    <ChunkCard
                        key={chunk.id}
                        chunk={chunk}
                        selected={selectedIds.includes(chunk.id)}
                        onSelect={(checked) => onSelectChunk(chunk.id, checked)}
                        onEdit={onEdit}
                    />
                ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
                <div className="flex items-center justify-between pt-4 border-t">
                    <div className="text-sm text-gray-500">
                        显示第 {((page - 1) * pageSize) + 1} 到 {Math.min(page * pageSize, total)} 条，共 {total} 条
                    </div>
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onPageChange(page - 1)}
                            disabled={page <= 1}
                        >
                            <ChevronLeft className="h-4 w-4 mr-1" /> 上一页
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => onPageChange(page + 1)}
                            disabled={page >= totalPages}
                        >
                            下一页 <ChevronRight className="h-4 w-4 ml-1" />
                        </Button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ChunkList;
