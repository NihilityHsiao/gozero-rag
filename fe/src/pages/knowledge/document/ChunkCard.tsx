import React from 'react';
import type { KnowledgeDocumentChunkInfo } from '@/types';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Badge } from '@/components/ui/badge';

interface ChunkCardProps {
    chunk: KnowledgeDocumentChunkInfo;
    selected: boolean;
    onSelect: (checked: boolean) => void;
    onEdit?: (chunk: KnowledgeDocumentChunkInfo) => void;
}

const ChunkCard: React.FC<ChunkCardProps> = ({ chunk, selected, onSelect, onEdit }) => {
    // 使用 content 字段，兼容旧的 chunk_text
    const displayContent = chunk.content || chunk.chunk_text || '';
    const charCount = displayContent.length;

    return (
        <Card className={`relative transition-all hover:shadow-md ${selected ? 'border-primary ring-1 ring-primary' : ''}`}>
            <div className="absolute top-4 left-4 z-10">
                <Checkbox
                    checked={selected}
                    onCheckedChange={(checked) => onSelect(checked === true)}
                />
            </div>

            <CardHeader className="pl-12 pb-2 pt-4">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <span className="text-xs font-mono text-gray-500">#{chunk.id.substring(0, 8)}...</span>
                        <Badge variant={chunk.status === 1 ? 'default' : 'secondary'} className="text-[10px] h-5">
                            {chunk.status === 1 ? '启用' : '启用'}
                        </Badge>
                    </div>
                    <span className="text-xs text-gray-400">{charCount} 字符</span>
                </div>
            </CardHeader>

            <CardContent className="pl-12 text-sm space-y-4 cursor-pointer" onClick={() => onEdit?.(chunk)}>
                {/* Chunk Content */}
                <div className="prose prose-sm max-w-none text-gray-700 whitespace-pre-wrap line-clamp-4 hover:text-gray-900 transition-colors">
                    {displayContent}
                </div>

                {/* Important Keywords */}
                {chunk.important_keywords && chunk.important_keywords.length > 0 && (
                    <div className="flex flex-wrap gap-1.5 pt-2 border-t">
                        <span className="text-xs text-gray-500 mr-1">关键词:</span>
                        {chunk.important_keywords.map((kw, idx) => (
                            <Badge key={idx} variant="outline" className="text-[10px] h-5 font-normal">
                                {kw}
                            </Badge>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

export default ChunkCard;
