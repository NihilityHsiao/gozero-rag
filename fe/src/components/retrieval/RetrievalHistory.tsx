import React from 'react';
import type { RetrieveLog } from '@/types/retrieval';
import { Clock, List } from 'lucide-react';
import { Card } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { format } from 'date-fns';

interface RetrievalHistoryProps {
    logs: RetrieveLog[];
    onSelectLog: (log: RetrieveLog) => void;
    hasMore: boolean;
    onLoadMore: () => void;
    loading: boolean;
}

export const RetrievalHistory: React.FC<RetrievalHistoryProps> = ({ logs, onSelectLog, hasMore, onLoadMore, loading }) => {
    const getModeLabel = (mode: string) => {
        switch (mode) {
            case 'vector': return '向量检索';
            case 'fulltext': return '全文检索';
            case 'hybrid': return '混合检索';
            default: return mode;
        }
    };

    return (
        <ScrollArea className="h-full">
            <div className="space-y-2 p-1">
                {logs.length === 0 && !loading && (
                    <div className="text-center text-xs text-muted-foreground py-8">
                        暂无历史记录
                    </div>
                )}
                {logs.map((log) => (
                    <Card
                        key={log.id}
                        className="p-3 cursor-pointer hover:bg-muted/50 transition-colors border-l-2 border-l-transparent hover:border-l-primary"
                        onClick={() => onSelectLog(log)}
                    >
                        <div className="flex justify-between items-start mb-1">
                            <span className="text-xs font-bold uppercase text-primary bg-primary/10 px-1.5 py-0.5 rounded">
                                {getModeLabel(log.retrieval_mode)}
                            </span>
                            <div className="flex items-center text-[10px] text-muted-foreground">
                                <Clock className="w-3 h-3 mr-1" />
                                {format(new Date(log.created_at), 'MM-dd HH:mm')}
                            </div>
                        </div>
                        <div className="text-xs line-clamp-2 font-medium mb-2" title={log.query}>
                            {log.query}
                        </div>
                        <div className="flex justify-between items-center text-[10px] text-muted-foreground">
                            <div className="flex items-center">
                                <List className="w-3 h-3 mr-1" />
                                {log.chunk_count} 个切片
                            </div>
                            <div>
                                {log.time_cost_ms}ms
                            </div>
                        </div>
                    </Card>
                ))}

                {hasMore && (
                    <button
                        className="w-full text-xs text-muted-foreground py-2 hover:text-primary transition-colors disabled:opacity-50"
                        onClick={onLoadMore}
                        disabled={loading}
                    >
                        {loading ? '加载中...' : '加载更多'}
                    </button>
                )}
            </div>
        </ScrollArea>
    );
};
