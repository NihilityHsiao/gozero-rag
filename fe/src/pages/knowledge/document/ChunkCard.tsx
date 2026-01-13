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
    const { metadata } = chunk;

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
                            {chunk.status === 1 ? '启用' : '禁用'}
                        </Badge>
                    </div>
                    <span className="text-xs text-gray-400">{chunk.chunk_size} 字符</span>
                </div>
            </CardHeader>

            <CardContent className="pl-12 text-sm space-y-4 cursor-pointer" onClick={() => onEdit?.(chunk)}>
                {/* Chunk Content */}
                <div className="prose prose-sm max-w-none text-gray-700 whitespace-pre-wrap line-clamp-4 hover:text-gray-900 transition-colors">
                    {chunk.chunk_text}
                </div>

                {/* QA Pairs */}
                {metadata?.qa_pairs && metadata.qa_pairs.length > 0 && (
                    <div className="bg-gray-50 rounded-md p-3 space-y-3 border">
                        <h5 className="text-xs font-semibold text-gray-500 uppercase tracking-wider">问答对</h5>
                        {metadata.qa_pairs.map((qa, idx) => (
                            <div key={idx} className="space-y-1">
                                <div className="font-medium text-gray-900">Q: {qa.question}</div>
                                <div className="text-gray-600 pl-4 border-l-2 border-gray-200">A: {qa.answer}</div>
                            </div>
                        ))}
                    </div>
                )}
            </CardContent>
        </Card>
    );
};

export default ChunkCard;
