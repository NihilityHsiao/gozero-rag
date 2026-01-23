import { useEffect, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { Network, AlertTriangle } from 'lucide-react';
import {
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormDescription,
    FormMessage,
} from '@/components/ui/form';
import { Switch } from '@/components/ui/switch';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { X, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';

import { listTenantLlm } from '@/api/llm';
import type { TenantLlmInfo } from '@/types';

export const GraphSettings = () => {
    const form = useFormContext();
    const [graphLlms, setGraphLlms] = useState<TenantLlmInfo[]>([]);
    const [newType, setNewType] = useState('');

    useEffect(() => {
        const fetchModels = async () => {
            try {
                // Fetch models suitable for Graph extraction (usually LLM type)
                const res = await listTenantLlm({ model_type: 'LLM', status: 1 });
                setGraphLlms(res.list || []);
            } catch (error) {
                console.error('Failed to load graph llms:', error);
            }
        };
        fetchModels();
    }, []);

    const enableGraph = form.watch('graph_rag.enable_graph');

    const handleAddType = (field: any) => {
        if (newType && !field.value.includes(newType)) {
            field.onChange([...field.value, newType]);
            setNewType('');
        }
    };

    const handleRemoveType = (field: any, typeToRemove: string) => {
        field.onChange(field.value.filter((t: string) => t !== typeToRemove));
    };

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Network className="w-5 h-5" />
                        知识图谱配置
                    </CardTitle>
                    <CardDescription>
                        配置 GraphRAG 相关选项，提取文档中的实体与关系。
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-6">
                    {/* Enable Switch */}
                    <FormField
                        control={form.control}
                        name="graph_rag.enable_graph"
                        render={({ field }) => (
                            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm bg-card">
                                <div className="space-y-0.5">
                                    <FormLabel>启用知识图谱</FormLabel>
                                    <FormDescription>
                                        开启后将自动从文档中提取实体和关系构建图谱。
                                    </FormDescription>
                                </div>
                                <FormControl>
                                    <Switch
                                        checked={field.value}
                                        onCheckedChange={field.onChange}
                                    />
                                </FormControl>
                            </FormItem>
                        )}
                    />

                    {enableGraph && (
                        <div className="space-y-6 animate-in fade-in slide-in-from-top-4">
                            {/* Warning Alert */}
                            <Alert className="bg-yellow-50 border-yellow-200 text-yellow-800">
                                <AlertTriangle className="h-4 w-4 text-yellow-600" />
                                <AlertTitle>耗时与费用提示</AlertTitle>
                                <AlertDescription className="text-xs">
                                    知识图谱生成会显著增加文档解析耗时，并消耗大量 LLM Token。
                                    <br />
                                    建议仅在需要深度理解实体关系的场景下开启。
                                </AlertDescription>
                            </Alert>

                            {/* Model Selection */}
                            <FormField
                                control={form.control}
                                name="graph_rag.graph_llm_id"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>图谱抽取模型</FormLabel>
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="选择用于提取图谱的模型" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                {graphLlms.map((model) => (
                                                    <SelectItem
                                                        key={model.id}
                                                        value={`${model.llm_name}@${model.llm_factory}`}
                                                    >
                                                        {model.llm_name} ({model.llm_factory})
                                                    </SelectItem>
                                                ))}
                                                {graphLlms.length === 0 && (
                                                    <div className="p-2 text-xs text-muted-foreground text-center">
                                                        未找到可用模型
                                                    </div>
                                                )}
                                            </SelectContent>
                                        </Select>
                                        <FormDescription>建议选择上下文窗口较大且指令遵循能力强的模型 (如 GPT-4, DeepSeek V3)</FormDescription>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Entity Types */}
                            <FormField
                                control={form.control}
                                name="graph_rag.entity_types"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>实体类型</FormLabel>
                                        <FormControl>
                                            <div className="space-y-3">
                                                <div className="flex flex-wrap gap-2 p-3 border rounded-md bg-muted/20 min-h-[60px]">
                                                    {(field.value || []).map((type: string) => (
                                                        <Badge key={type} variant="secondary" className="px-2 py-1 flex items-center gap-1">
                                                            {type}
                                                            <X
                                                                className="w-3 h-3 cursor-pointer hover:text-destructive"
                                                                onClick={() => handleRemoveType(field, type)}
                                                            />
                                                        </Badge>
                                                    ))}
                                                    {(!field.value || field.value.length === 0) && (
                                                        <span className="text-xs text-muted-foreground self-center">暂无配置实体类型</span>
                                                    )}
                                                </div>
                                                <div className="flex gap-2">
                                                    <Input
                                                        value={newType}
                                                        onChange={(e) => setNewType(e.target.value)}
                                                        placeholder="输入新类型 (如 product)"
                                                        className="max-w-xs"
                                                        onKeyDown={(e) => {
                                                            if (e.key === 'Enter') {
                                                                e.preventDefault();
                                                                handleAddType(field);
                                                            }
                                                        }}
                                                    />
                                                    <Button
                                                        type="button"
                                                        variant="outline"
                                                        size="sm"
                                                        onClick={() => handleAddType(field)}
                                                        disabled={!newType}
                                                    >
                                                        <Plus className="w-4 h-4 mr-1" /> 添加
                                                    </Button>
                                                </div>
                                            </div>
                                        </FormControl>
                                        <FormDescription>定义需要提取的实体类别，有助于提高图谱质量。</FormDescription>
                                    </FormItem>
                                )}
                            />

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {/* Entity Resolution */}
                                <FormField
                                    control={form.control}
                                    name="graph_rag.enable_entity_resolution"
                                    render={({ field }) => (
                                        <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                                            <FormControl>
                                                <Switch
                                                    checked={field.value}
                                                    onCheckedChange={field.onChange}
                                                />
                                            </FormControl>
                                            <div className="space-y-1 leading-none">
                                                <FormLabel>实体归一化</FormLabel>
                                                <FormDescription>
                                                    自动合并具有相同含义的实体（如: 特朗普总统、唐纳德·特朗普 → 唐纳德·特朗普）
                                                </FormDescription>
                                            </div>
                                        </FormItem>
                                    )}
                                />

                                {/* Community Report */}
                                <FormField
                                    control={form.control}
                                    name="graph_rag.enable_community"
                                    render={({ field }) => (
                                        <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                                            <FormControl>
                                                <Switch
                                                    checked={field.value}
                                                    onCheckedChange={field.onChange}
                                                />
                                            </FormControl>
                                            <div className="space-y-1 leading-none">
                                                <FormLabel>社区报告生成</FormLabel>
                                                <FormDescription>
                                                    对图谱进行社区聚类，并生成每个社区的摘要报告。
                                                </FormDescription>
                                            </div>
                                        </FormItem>
                                    )}
                                />
                            </div>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
};
