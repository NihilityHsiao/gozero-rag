import { useState, useEffect } from 'react';
import { Settings2, Database, Box, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Slider } from '@/components/ui/slider';
import { Textarea } from '@/components/ui/textarea';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useAuthStore } from '@/store/useAuthStore';
import { useChatStore } from '@/store/useChatStore';
import { getUserApiList } from '@/api/user_model';
import type { UserApiInfo, KnowledgeBaseInfo } from '@/types';
import SelectKnowledgeDialog from './SelectKnowledgeDialog';

export default function ConfigPanel() {
    const { userInfo } = useAuthStore();
    const { config, setConfig } = useChatStore();

    // Local state for UI display (e.g., lists), but core config comes from store
    const [models, setModels] = useState<UserApiInfo[]>([]);
    const [rerankModels, setRerankModels] = useState<UserApiInfo[]>([]);

    // Knowledge Selection State
    const [dialogOpen, setDialogOpen] = useState(false);
    const [selectedKnowledgeList, setSelectedKnowledgeList] = useState<KnowledgeBaseInfo[]>([]);

    // Initialize defaults if store is empty
    useEffect(() => {
        if (!config.temperature) {
            setConfig({
                temperature: 0.7,
                retrieval_mode: 'hybrid',
                top_k: 10,
                score_threshold: 0.5,
                hybrid_strategy_type: 'weighted',
                weight_vector: 0.7,
                weight_keyword: 0.3
            });
        }
    }, []);

    // Fetch Models (Chat + Rerank)
    useEffect(() => {
        const fetchModels = async () => {
            if (!userInfo?.user_id) return;
            try {
                // Fetch Chat Models
                const chatRes = await getUserApiList(userInfo.user_id, { model_type: 'chat' });
                setModels(chatRes.list || []);
                if (chatRes.list?.length > 0 && !config.model_id) {
                    const defaultModel = chatRes.list.find(m => m.is_default === 1);
                    setConfig({ model_id: defaultModel ? defaultModel.id : chatRes.list[0].id });
                }

                // Fetch Rerank Models
                const rerankRes = await getUserApiList(userInfo.user_id, { model_type: 'rerank' });
                setRerankModels(rerankRes.list || []);
                if (rerankRes.list?.length > 0 && !config.rerank_model_id) {
                    setConfig({ rerank_model_id: rerankRes.list[0].id });
                }
            } catch (e) {
                console.error('Failed to fetch models', e);
            }
        };
        fetchModels();
    }, [userInfo?.user_id]);

    const handleConfirmKnowledge = (kbs: KnowledgeBaseInfo[]) => {
        setSelectedKnowledgeList(kbs);
        setConfig({ knowledge_base_ids: kbs.map(k => k.id) });
    };

    const handleRemoveKnowledge = (id: number) => {
        const newList = selectedKnowledgeList.filter(k => k.id !== id);
        setSelectedKnowledgeList(newList);
        setConfig({ knowledge_base_ids: newList.map(k => k.id) });
    };

    return (
        <div className="w-[300px] border-l bg-gray-50/50 flex flex-col h-full">
            <div className="p-4 border-b bg-white flex items-center gap-2">
                <Settings2 className="w-4 h-4 text-gray-500" />
                <h3 className="font-semibold text-sm">编排配置</h3>
            </div>

            <ScrollArea className="flex-1">
                <div className="p-4 space-y-6">
                    {/* 1. Model Selection */}
                    <div className="space-y-3">
                        <Label className="text-xs font-medium text-gray-500 uppercase tracking-wider">模型 (Model)</Label>
                        <Select
                            value={config.model_id?.toString() || ''}
                            onValueChange={(val) => setConfig({ model_id: parseInt(val) })}
                        >
                            <SelectTrigger className="bg-white">
                                <SelectValue placeholder="选择模型" />
                            </SelectTrigger>
                            <SelectContent>
                                {models.map(m => (
                                    <SelectItem key={m.id} value={m.id.toString()}>
                                        <div className="flex items-center gap-2">
                                            <Box className="w-4 h-4 text-blue-500" />
                                            <span>{m.config_name}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>

                    <div className="h-px bg-gray-200" />

                    {/* 2. System Prompt */}
                    <div className="space-y-3">
                        <Label className="text-xs font-medium text-gray-500 uppercase tracking-wider">系统提示词 (System Prompt)</Label>
                        <Textarea
                            className="bg-white min-h-[100px] text-sm resize-none"
                            placeholder="你是一个乐于助人的 AI 助手..."
                            value={config.system_prompt || ''}
                            onChange={(e) => setConfig({ system_prompt: e.target.value })}
                        />
                    </div>

                    <div className="h-px bg-gray-200" />

                    {/* 3. Parameters */}
                    <div className="space-y-4">
                        <Label className="text-xs font-medium text-gray-500 uppercase tracking-wider">参数设置</Label>

                        <div className="space-y-3">
                            <div className="flex items-center justify-between">
                                <Label className="text-sm">随机性 (Temperature)</Label>
                                <span className="text-xs font-mono text-gray-500">{config.temperature}</span>
                            </div>
                            <Slider
                                value={[config.temperature || 0.7]}
                                max={1}
                                step={0.1}
                                onValueChange={([val]) => setConfig({ temperature: val })}
                            />
                        </div>
                    </div>

                    <div className="h-px bg-gray-200" />

                    {/* 4. Context / Knowledge */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between">
                            <Label className="text-xs font-medium text-gray-500 uppercase tracking-wider">上下文 (Context)</Label>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 text-xs text-blue-600 hover:text-blue-700 px-0"
                                onClick={() => setDialogOpen(true)}
                            >
                                + 添加
                            </Button>
                        </div>

                        {/* Selected Knowledge List */}
                        {selectedKnowledgeList.length === 0 ? (
                            <div className="bg-white border border-dashed rounded-lg p-3 text-center">
                                <Database className="w-4 h-4 text-gray-400 mx-auto mb-1" />
                                <p className="text-xs text-gray-500">未关联知识库</p>
                            </div>
                        ) : (
                            <div className="space-y-2 max-h-[200px] overflow-y-auto pr-1">
                                {selectedKnowledgeList.map(kb => (
                                    <div key={kb.id} className="bg-white border rounded-md p-2 flex items-center gap-2 group">
                                        <Database className="w-3.5 h-3.5 text-blue-500 shrink-0" />
                                        <span className="text-sm truncate flex-1" title={kb.name}>{kb.name}</span>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-5 w-5 opacity-0 group-hover:opacity-100 transition-opacity"
                                            onClick={() => handleRemoveKnowledge(kb.id)}
                                        >
                                            <X className="w-3 h-3 text-gray-400 hover:text-red-500" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    {/* 5. Retrieval Settings (Only if knowledge selected) */}
                    {selectedKnowledgeList.length > 0 && (
                        <>
                            <div className="h-px bg-gray-200" />
                            <div className="space-y-4">
                                <Label className="text-xs font-medium text-gray-500 uppercase tracking-wider">检索设置 (Retrieval)</Label>

                                {/* Retrieval Mode */}
                                <div className="space-y-2">
                                    <Label className="text-xs text-gray-500">检索模式</Label>
                                    <div className="grid grid-cols-3 gap-1 bg-gray-100 p-1 rounded-md">
                                        {['vector', 'fulltext', 'hybrid'].map((mode) => (
                                            <button
                                                key={mode}
                                                onClick={() => setConfig({ retrieval_mode: mode as any })}
                                                className={`text-xs py-1 px-2 rounded-sm transition-all ${config.retrieval_mode === mode
                                                    ? 'bg-white shadow text-blue-600 font-medium'
                                                    : 'text-gray-500 hover:text-gray-700'
                                                    }`}
                                            >
                                                {{
                                                    'vector': '向量',
                                                    'fulltext': '全文',
                                                    'hybrid': '混合'
                                                }[mode as string]}
                                            </button>
                                        ))}
                                    </div>
                                </div>

                                {/* Top K */}
                                <div className="space-y-3">
                                    <div className="flex items-center justify-between">
                                        <Label className="text-sm">Top K</Label>
                                        <span className="text-xs font-mono text-gray-500">{config.top_k}</span>
                                    </div>
                                    <Slider
                                        value={[config.top_k || 10]}
                                        max={20} step={1} min={1}
                                        onValueChange={([val]) => setConfig({ top_k: val })}
                                    />
                                </div>

                                {/* Score Threshold */}
                                <div className="space-y-3">
                                    <div className="flex items-center justify-between">
                                        <Label className="text-sm">Score 阈值</Label>
                                        <span className="text-xs font-mono text-gray-500">{config.score_threshold}</span>
                                    </div>
                                    <Slider
                                        value={[config.score_threshold || 0.5]}
                                        max={1} step={0.05}
                                        onValueChange={([val]) => setConfig({ score_threshold: val })}
                                    />
                                </div>

                                {/* Hybrid Settings */}
                                {config.retrieval_mode === 'hybrid' && (
                                    <div className="space-y-3 pt-2 border-t border-dashed">
                                        <Label className="text-xs text-gray-500">重排序策略 (Rerank)</Label>
                                        <div className="flex gap-4">
                                            <label className="flex items-center gap-2 text-sm cursor-pointer">
                                                <input
                                                    type="radio"
                                                    name="hybrid_strategy"
                                                    checked={config.hybrid_strategy_type === 'weighted'}
                                                    onChange={() => setConfig({ hybrid_strategy_type: 'weighted' })}
                                                    className="accent-blue-600"
                                                />
                                                加权
                                            </label>
                                            <label className="flex items-center gap-2 text-sm cursor-pointer">
                                                <input
                                                    type="radio"
                                                    name="hybrid_strategy"
                                                    checked={config.hybrid_strategy_type === 'rerank'}
                                                    onChange={() => setConfig({ hybrid_strategy_type: 'rerank' })}
                                                    className="accent-blue-600"
                                                />
                                                模型重排
                                            </label>
                                        </div>

                                        {config.hybrid_strategy_type === 'weighted' ? (
                                            <div className="space-y-3 bg-gray-50 p-2 rounded-md">
                                                <div className="flex items-center justify-between">
                                                    <Label className="text-xs text-gray-500">向量权重</Label>
                                                    <span className="text-xs font-mono text-gray-500">{config.weight_vector}</span>
                                                </div>
                                                <Slider
                                                    value={[config.weight_vector || 0.7]}
                                                    max={1}
                                                    step={0.1}
                                                    onValueChange={([val]) => setConfig({ weight_vector: val, weight_keyword: parseFloat((1 - val).toFixed(1)) })}
                                                />
                                                <div className="flex justify-between text-[10px] text-gray-400">
                                                    <span>Keyword: {config.weight_keyword}</span>
                                                </div>
                                            </div>
                                        ) : (
                                            <Select
                                                value={config.rerank_model_id?.toString() || ''}
                                                onValueChange={(val) => setConfig({ rerank_model_id: parseInt(val) })}
                                            >
                                                <SelectTrigger className="bg-white h-8 text-xs">
                                                    <SelectValue placeholder="选择重排序模型" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {rerankModels.map(m => (
                                                        <SelectItem key={m.id} value={m.id.toString()}>
                                                            {m.config_name}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        )}
                                    </div>
                                )}
                            </div>
                        </>
                    )}
                </div>
            </ScrollArea>

            <SelectKnowledgeDialog
                open={dialogOpen}
                onOpenChange={setDialogOpen}
                currentSelection={selectedKnowledgeList.map(k => k.id)}
                onConfirm={handleConfirmKnowledge}
            />
        </div>
    );
}
