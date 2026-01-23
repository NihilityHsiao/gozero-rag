import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { RetrievalSettingsSheet, type RetrievalSettingsValues } from '@/components/retrieval/RetrievalSettingsSheet';
import { RetrievalResultList } from '@/components/retrieval/RetrievalResultList';
import { RetrievalHistory } from '@/components/retrieval/RetrievalHistory';
import { retrievalApi } from '@/api/retrieval';
import { listTenantLlm } from '@/api/llm';
import type { RetrievalChunk, RetrieveLog, RetrieveReq } from '@/types/retrieval';
import type { TenantLlmInfo } from '@/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Settings2, Play, Loader2 } from 'lucide-react';
import { useAuthStore } from '@/store/useAuthStore';

// Use same type for simplicity where possible, config lacks query
type PageConfig = RetrievalSettingsValues;

const RetrievalTestPage: React.FC = () => {
    const { id } = useParams<{ id: string }>();
    const knowledgeBaseId = id || '';
    const { userInfo } = useAuthStore();

    const [loading, setLoading] = useState(false);
    const [query, setQuery] = useState('');
    const [results, setResults] = useState<RetrievalChunk[]>([]);
    const [logs, setLogs] = useState<RetrieveLog[]>([]);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [pageSize] = useState(20);
    const [logLoading, setLogLoading] = useState(false);

    const [isSettingsOpen, setIsSettingsOpen] = useState(false);
    const [config, setConfig] = useState<PageConfig>({
        retrieval_mode: 'hybrid',
        top_k: 10,
        score_threshold: 0.5,
        hybrid_strategy_type: 'weighted',
        weight_vector: 0.7,
        weight_keyword: 0.3,
    });

    const [rerankModels, setRerankModels] = useState<TenantLlmInfo[]>([]);

    // Fetch logs and models on mount
    useEffect(() => {
        if (knowledgeBaseId) {
            fetchLogs(1); // Load first page
            fetchRerankModels();
        }
    }, [knowledgeBaseId]);

    const fetchLogs = async (pageNum: number) => {
        setLogLoading(true);
        try {
            const res = await retrievalApi.getRetrievalLog(knowledgeBaseId, { page: pageNum, page_size: pageSize });
            if (res.logs) {
                if (pageNum === 1) {
                    setLogs(res.logs);
                } else {
                    setLogs(prev => [...prev, ...res.logs]);
                }
                setTotal(res.total || 0);
                setPage(pageNum);
            }
        } catch (error) {
            console.error('Failed to fetch logs', error);
        } finally {
            setLogLoading(false);
        }
    };

    const handleLoadMoreLogs = () => {
        fetchLogs(page + 1);
    };

    const fetchRerankModels = async () => {
        try {
            const res = await listTenantLlm({ model_type: 'rerank', page: 1, page_size: 100 });
            if (res.list && res.list.length > 0) {
                setRerankModels(res.list);
                // Set default rerank model if not set
                setConfig(prev => {
                    if (!prev.rerank_model_id) {
                        const first = res.list[0];
                        return { ...prev, rerank_model_id: `${first.llm_name}@${first.llm_factory}` };
                    }
                    return prev;
                });
            }
        } catch (error) {
            console.error('Failed to fetch rerank models', error);
        }
    };

    const handleSearch = async () => {
        if (!query.trim()) {
            toast.error('请输入查询语句');
            return;
        }

        // Ensure rerank model is selected (if available)
        if (!config.rerank_model_id && rerankModels.length > 0) {
            const first = rerankModels[0];
            config.rerank_model_id = `${first.llm_name}@${first.llm_factory}`;
        }

        setLoading(true);
        setResults([]); // Clear previous results

        try {
            const req: RetrieveReq = {
                knowledge_base_id: knowledgeBaseId,
                query: query,
                retrieval_mode: config.retrieval_mode as any,
                retrieval_config: {
                    top_k: config.top_k,
                    score_threshold: config.score_threshold,
                    hybrid_strategy: {
                        type: config.hybrid_strategy_type || 'weighted', // Ensure type is set
                        weights: config.hybrid_strategy_type === 'weighted' ? {
                            vector: config.weight_vector,
                            keyword: config.weight_keyword,
                        } : undefined,
                        // Always pass rerank_model_id if available
                        rerank_model_id: config.rerank_model_id || undefined
                    }
                }
            };

            const res = await retrievalApi.retrieve(req);
            setResults(res.chunks || []);
            toast.success(`召回成功，耗时 ${res.time_cost_ms}ms`);

            // Refresh logs (reload first page)
            fetchLogs(1);

        } catch (error) {
            console.error('Retrieval failed', error);
            toast.error('召回失败');
        } finally {
            setLoading(false);
        }
    };

    const handleSelectLog = (log: RetrieveLog) => {
        setQuery(log.query);

        // Restore config
        const newConfig: PageConfig = {
            ...config,
            retrieval_mode: log.retrieval_mode as any,
            top_k: log.retrieval_params.top_k || 10,
            score_threshold: log.retrieval_params.score_threshold || 0.5,
        };

        if (log.retrieval_params.hybrid_strategy) {
            const strategy = log.retrieval_params.hybrid_strategy;
            // Only restore if present, otherwise keep current defaults (like rerank_model_id)
            if (strategy.type) newConfig.hybrid_strategy_type = strategy.type as any;

            if (strategy.type === 'weighted' && strategy.weights) {
                newConfig.weight_vector = strategy.weights.vector;
                newConfig.weight_keyword = strategy.weights.keyword;
            }
            if (strategy.rerank_model_id) {
                newConfig.rerank_model_id = String(strategy.rerank_model_id);
            }
        }

        setConfig(newConfig);
        toast.info('配置已从历史记录恢复');
    };

    // Helper to get button label based on mode
    const getModeLabel = () => {
        if (config.retrieval_mode === 'vector') return '向量检索';
        if (config.retrieval_mode === 'fulltext') return '全文检索';
        return '混合检索';
    };

    return (
        <div className="h-[calc(100vh-8rem)] grid grid-cols-12 gap-6">
            <div className="col-span-7 flex flex-col space-y-6 h-full overflow-y-auto pr-1">
                {/* Configuration Area */}
                <Card className="flex-shrink-0 border-blue-500 border-2 shadow-sm">
                    <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                        <CardTitle className="text-base">源文本</CardTitle>
                        <Button
                            variant="secondary"
                            size="sm"
                            className="h-8 gap-2 bg-blue-50 text-blue-600 hover:bg-blue-100 border border-blue-100 shadow-sm"
                            onClick={() => setIsSettingsOpen(true)}
                        >
                            <Settings2 className="w-3.5 h-3.5" />
                            {getModeLabel()}
                        </Button>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="relative">
                            <Textarea
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                placeholder="请输入需要检索的问题或语句..."
                                className="min-h-[200px] resize-none pr-4 text-base"
                            />

                            <div className="absolute bottom-4 right-4">
                                <span className={query.length > 200 ? "text-red-500 text-xs" : "text-gray-400 text-xs"}>
                                    {query.length}/200
                                </span>
                            </div>
                        </div>

                        <div className="flex justify-end">
                            <Button onClick={handleSearch} disabled={loading} className="w-[120px]">
                                {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
                                测试
                            </Button>
                        </div>
                    </CardContent>
                </Card>

                {/* Settings Actions Sheet */}
                <RetrievalSettingsSheet
                    open={isSettingsOpen}
                    onOpenChange={setIsSettingsOpen}
                    currentConfig={config}
                    onSave={setConfig}
                    userId={userInfo?.user_id || ''} // TODO use real user id
                />

                {/* History Area */}
                <Card className="flex-1 flex-shrink-0 min-h-[300px] flex flex-col">
                    <CardHeader>
                        <CardTitle className="text-sm font-medium">历史记录</CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 overflow-hidden p-0">
                        <RetrievalHistory
                            logs={logs}
                            onSelectLog={handleSelectLog}
                            hasMore={logs.length < total}
                            onLoadMore={handleLoadMoreLogs}
                            loading={logLoading}
                        />
                    </CardContent>
                </Card>
            </div>

            <div className="col-span-5 h-full overflow-hidden">
                <Card className="h-full flex flex-col shadow-lg border-l-4 border-l-primary/20">
                    <CardHeader className="bg-muted/30 pb-4">
                        <CardTitle className="flex items-center justify-between">
                            <span>召回结果</span>
                            <span className="text-xs font-normal text-muted-foreground bg-background px-2 py-1 rounded-full border">
                                数量: {results.length}
                            </span>
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="flex-1 p-4 overflow-hidden bg-muted/10">
                        <RetrievalResultList chunks={results} loading={loading} />
                    </CardContent>
                </Card>
            </div>
        </div>
    );
};

export default RetrievalTestPage;
