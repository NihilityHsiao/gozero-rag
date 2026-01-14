import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormDescription,
} from '@/components/ui/form';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetFooter,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet"

import { getUserApiList } from '@/api/user_model';
import type { UserApiInfo } from '@/types';


// Schema Definition - Query removed
const retrievalSettingsSchema = z.object({
    retrieval_mode: z.enum(['vector', 'fulltext', 'hybrid']),
    top_k: z.number().min(1).max(100),
    score_threshold: z.number().min(0).max(1.0),

    // Hybrid specific
    hybrid_strategy_type: z.enum(['weighted', 'rerank']).optional(),
    weight_vector: z.number().min(0).max(1.0).optional(),
    weight_keyword: z.number().min(0).max(1.0).optional(),
    rerank_model_id: z.string().optional(),
});

export type RetrievalSettingsValues = z.infer<typeof retrievalSettingsSchema>;

interface RetrievalSettingsSheetProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    currentConfig: RetrievalSettingsValues;
    onSave: (values: RetrievalSettingsValues) => void;
    userId: string;
}

export const RetrievalSettingsSheet: React.FC<RetrievalSettingsSheetProps> = ({
    open,
    onOpenChange,
    currentConfig,
    onSave,
    userId
}) => {
    const [rerankModels, setRerankModels] = useState<UserApiInfo[]>([]);

    const form = useForm<RetrievalSettingsValues>({
        resolver: zodResolver(retrievalSettingsSchema),
        defaultValues: currentConfig,
    });

    // Reset form when currentConfig changes (e.g. from parent state)
    useEffect(() => {
        form.reset(currentConfig);
    }, [currentConfig, form]);

    const retrievalMode = form.watch('retrieval_mode');
    const hybridStrategy = form.watch('hybrid_strategy_type');

    useEffect(() => {
        const fetchModels = async () => {
            if (userId) {
                try {
                    const res = await getUserApiList(userId, { model_type: 'rerank' });
                    setRerankModels(res.list || []);
                } catch (error) {
                    console.error('Failed to fetch rerank models', error);
                }
            }
        };
        fetchModels();
    }, [userId]);

    const handleSave = (values: RetrievalSettingsValues) => {
        onSave(values);
        onOpenChange(false);
    };

    return (
        <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent className="overflow-y-auto sm:max-w-[500px]">
                <SheetHeader>
                    <SheetTitle>检索设置</SheetTitle>
                    <SheetDescription>
                        了解更多关于检索方法。
                    </SheetDescription>
                </SheetHeader>
                <div className="py-6">
                    <Form {...form}>
                        <form id="retrieval-settings-form" onSubmit={form.handleSubmit(handleSave)} className="space-y-6">

                            {/* Mode Selection */}
                            <FormField
                                control={form.control}
                                name="retrieval_mode"
                                render={({ field }) => (
                                    <FormItem className="border rounded-md p-4 space-y-4">
                                        <FormLabel className="font-semibold text-base block mb-2">检索方法</FormLabel>
                                        <FormControl>
                                            <RadioGroup
                                                onValueChange={field.onChange}
                                                defaultValue={field.value}
                                                className="flex flex-col space-y-3"
                                            >
                                                {/* Vector Search Item */}
                                                <FormItem className="flex items-start space-x-3 space-y-0">
                                                    <FormControl>
                                                        <RadioGroupItem value="vector" className="mt-1" />
                                                    </FormControl>
                                                    <div className="space-y-1">
                                                        <FormLabel className="font-medium">
                                                            向量检索
                                                        </FormLabel>
                                                        <FormDescription className="text-xs">
                                                            通过生成查询嵌入并查询与其向量表示最相似的文本片段
                                                        </FormDescription>
                                                    </div>
                                                </FormItem>

                                                {/* Fulltext Search Item */}
                                                <FormItem className="flex items-start space-x-3 space-y-0">
                                                    <FormControl>
                                                        <RadioGroupItem value="fulltext" className="mt-1" />
                                                    </FormControl>
                                                    <div className="space-y-1">
                                                        <FormLabel className="font-medium">
                                                            全文检索
                                                        </FormLabel>
                                                        <FormDescription className="text-xs">
                                                            索引文档中的所有词汇，从而允许用户查询任意词汇，并返回包含这些词汇的文本片段
                                                        </FormDescription>
                                                    </div>
                                                </FormItem>

                                                {/* Hybrid Search Item */}
                                                <FormItem className="flex items-start space-x-3 space-y-0">
                                                    <FormControl>
                                                        <RadioGroupItem value="hybrid" className="mt-1" />
                                                    </FormControl>
                                                    <div className="space-y-1">
                                                        <FormLabel className="font-medium">
                                                            混合检索 <span className="bg-blue-100 text-blue-600 text-[10px] px-1.5 py-0.5 rounded ml-1">推荐</span>
                                                        </FormLabel>
                                                        <FormDescription className="text-xs">
                                                            同时执行全文检索和向量检索，并应用重排序步骤，从两类查询结果中选择匹配用户问题的最佳结果。
                                                        </FormDescription>
                                                    </div>
                                                </FormItem>
                                            </RadioGroup>
                                        </FormControl>
                                    </FormItem>
                                )}
                            />

                            {/* Common Params (visible inside the selected mode block conceptually, but simplified here) */}
                            <div className="grid grid-cols-2 gap-6 p-4 border rounded-md bg-muted/20">
                                <FormField
                                    control={form.control}
                                    name="top_k"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Top K: {field.value}</FormLabel>
                                            <FormControl>
                                                <Slider
                                                    min={1}
                                                    max={20}
                                                    step={1}
                                                    defaultValue={[field.value]}
                                                    onValueChange={(vals: number[]) => field.onChange(vals[0])}
                                                />
                                            </FormControl>
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="score_threshold"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Score 阈值: {field.value}</FormLabel>
                                            <FormControl>
                                                <Slider
                                                    min={0}
                                                    max={1.0}
                                                    step={0.05}
                                                    defaultValue={[field.value]}
                                                    onValueChange={(vals: number[]) => field.onChange(vals[0])}
                                                />
                                            </FormControl>
                                        </FormItem>
                                    )}
                                />
                            </div>


                            {/* Hybrid Specifics */}
                            {retrievalMode === 'hybrid' && (
                                <div className="p-4 border rounded-md space-y-4">
                                    <FormField
                                        control={form.control}
                                        name="hybrid_strategy_type"
                                        render={({ field }) => (
                                            <FormItem className="space-y-3">
                                                <FormLabel>重排序模型 (Rerank)</FormLabel>
                                                <FormControl>
                                                    <RadioGroup
                                                        onValueChange={field.onChange}
                                                        defaultValue={field.value}
                                                        className="flex flex-col space-y-2"
                                                    >
                                                        <FormItem className="flex items-center space-x-2 space-y-0">
                                                            <FormControl>
                                                                <RadioGroupItem value="weighted" />
                                                            </FormControl>
                                                            <FormLabel className="font-normal">
                                                                加权融合 (Weighted)
                                                            </FormLabel>
                                                        </FormItem>
                                                        <FormItem className="flex items-center space-x-2 space-y-0">
                                                            <FormControl>
                                                                <RadioGroupItem value="rerank" />
                                                            </FormControl>
                                                            <FormLabel className="font-normal">
                                                                Rerank 模型
                                                            </FormLabel>
                                                        </FormItem>
                                                    </RadioGroup>
                                                </FormControl>
                                            </FormItem>
                                        )}
                                    />

                                    {hybridStrategy === 'weighted' && (
                                        <div className="space-y-4 pt-2">
                                            <FormField
                                                control={form.control}
                                                name="weight_vector"
                                                render={({ field }) => (
                                                    <FormItem>
                                                        <div className="flex justify-between">
                                                            <FormLabel className="text-xs">向量权重</FormLabel>
                                                            <span className="text-xs text-muted-foreground">{field.value}</span>
                                                        </div>
                                                        <FormControl>
                                                            <Slider
                                                                min={0}
                                                                max={1.0}
                                                                step={0.1}
                                                                defaultValue={[field.value || 0.7]}
                                                                onValueChange={(vals: number[]) => {
                                                                    field.onChange(vals[0]);
                                                                    form.setValue('weight_keyword', Number((1 - vals[0]).toFixed(1)));
                                                                }}
                                                            />
                                                        </FormControl>
                                                    </FormItem>
                                                )}
                                            />
                                            <FormField
                                                control={form.control}
                                                name="weight_keyword"
                                                render={({ field }) => (
                                                    <FormItem>
                                                        <div className="flex justify-between">
                                                            <FormLabel className="text-xs">关键词权重</FormLabel>
                                                            <span className="text-xs text-muted-foreground">{field.value}</span>
                                                        </div>
                                                        <FormControl>
                                                            <Slider
                                                                min={0}
                                                                max={1.0}
                                                                step={0.1}
                                                                defaultValue={[field.value || 0.3]}
                                                                onValueChange={(vals: number[]) => {
                                                                    field.onChange(vals[0]);
                                                                    form.setValue('weight_vector', Number((1 - vals[0]).toFixed(1)));
                                                                }}
                                                            />
                                                        </FormControl>
                                                    </FormItem>
                                                )}
                                            />
                                        </div>
                                    )}

                                    {hybridStrategy === 'rerank' && (
                                        <FormField
                                            control={form.control}
                                            name="rerank_model_id"
                                            render={({ field }) => (
                                                <FormItem className="pt-2">
                                                    <Select onValueChange={field.onChange} value={field.value}>
                                                        <FormControl>
                                                            <SelectTrigger>
                                                                <SelectValue placeholder="选择重排序模型" />
                                                            </SelectTrigger>
                                                        </FormControl>
                                                        <SelectContent>
                                                            {rerankModels.map(model => (
                                                                <SelectItem key={model.id} value={String(model.id)}>
                                                                    {model.model_name} ({model.config_name})
                                                                </SelectItem>
                                                            ))}
                                                            {rerankModels.length === 0 && (
                                                                <div className="p-2 text-xs text-muted-foreground text-center">
                                                                    未找到重排序模型
                                                                </div>
                                                            )}
                                                        </SelectContent>
                                                    </Select>
                                                </FormItem>
                                            )}
                                        />
                                    )}
                                </div>
                            )}

                        </form>
                    </Form>
                </div>
                <SheetFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
                    <Button type="submit" form="retrieval-settings-form">保存</Button>
                </SheetFooter>
            </SheetContent>
        </Sheet>
    );
};
