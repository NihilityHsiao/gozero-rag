import React, { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
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
    FormMessage,
} from '@/components/ui/form';

import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';

import { Loader2 } from 'lucide-react';
import { getUserApiList } from '@/api/user_model';
import type { UserApiInfo } from '@/types';

// Schema Definition
const retrievalFormSchema = z.object({
    query: z.string().min(1, '请输入查询语句'),
    retrieval_mode: z.enum(['vector', 'fulltext', 'hybrid']),
    top_k: z.number().min(1).max(100),
    score_threshold: z.number().min(0).max(1.0),

    // Hybrid specific
    hybrid_strategy_type: z.enum(['weighted', 'rerank']).optional(),
    weight_vector: z.number().min(0).max(1.0).optional(),
    weight_keyword: z.number().min(0).max(1.0).optional(),
    rerank_model_id: z.string().optional(), // Select value is string
});

export type RetrievalFormValues = z.infer<typeof retrievalFormSchema>;

interface RetrievalConfigPanelProps {
    onSearch: (values: RetrievalFormValues) => void;
    isLoading: boolean;
    userId: string; // To fetch user models
    defaultValues?: Partial<RetrievalFormValues>;
}

export const RetrievalConfigPanel: React.FC<RetrievalConfigPanelProps> = ({
    onSearch,
    isLoading,
    userId,
    defaultValues
}) => {
    const [rerankModels, setRerankModels] = useState<UserApiInfo[]>([]);

    const form = useForm<RetrievalFormValues>({
        resolver: zodResolver(retrievalFormSchema),
        defaultValues: {
            query: '',
            retrieval_mode: 'hybrid',
            top_k: 10,
            score_threshold: 0.5,
            hybrid_strategy_type: 'weighted',
            weight_vector: 0.7,
            weight_keyword: 0.3,
            ...defaultValues
        },
    });

    // Watch values to conditionally render fields
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

    // Update form when defaultValues prop changes (e.g. from history)
    useEffect(() => {
        if (defaultValues) {
            form.reset({
                ...form.getValues(),
                ...defaultValues
            });
        }
    }, [defaultValues, form]);


    const onSubmit = (values: RetrievalFormValues) => {
        onSearch(values);
    };

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                <FormField
                    control={form.control}
                    name="query"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>查询语句 Query</FormLabel>
                            <FormControl>
                                <Textarea
                                    placeholder="请输入需要检索的问题或语句..."
                                    className="min-h-[100px] resize-none"
                                    {...field}
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <div className="grid grid-cols-2 gap-4">
                    <FormField
                        control={form.control}
                        name="retrieval_mode"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>检索模式 Mode</FormLabel>
                                <Select onValueChange={field.onChange} defaultValue={field.value} value={field.value}>
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select mode" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        <SelectItem value="vector">Vector Search</SelectItem>
                                        <SelectItem value="fulltext">Fulltext Search</SelectItem>
                                        <SelectItem value="hybrid">Hybrid Search</SelectItem>
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                    <FormField
                        control={form.control}
                        name="top_k"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Top K: {field.value}</FormLabel>
                                <FormControl>
                                    <Slider
                                        min={1}
                                        max={100}
                                        step={1}
                                        defaultValue={[field.value]}
                                        onValueChange={(vals: number[]) => field.onChange(vals[0])}
                                    />
                                </FormControl>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                </div>
                <FormField
                    control={form.control}
                    name="score_threshold"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Score Threshold: {field.value}</FormLabel>
                            <FormControl>
                                <Slider
                                    min={0}
                                    max={1.0}
                                    step={0.01}
                                    defaultValue={[field.value]}
                                    onValueChange={(vals: number[]) => field.onChange(vals[0])}
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {retrievalMode === 'hybrid' && (
                    <div className="p-4 border rounded-md bg-muted/20 space-y-4">
                        <FormField
                            control={form.control}
                            name="hybrid_strategy_type"
                            render={({ field }) => (
                                <FormItem className="space-y-3">
                                    <FormLabel>Hybrid Strategy</FormLabel>
                                    <FormControl>
                                        <RadioGroup
                                            onValueChange={field.onChange}
                                            defaultValue={field.value}
                                            className="flex flex-col space-y-1"
                                        >
                                            <FormItem className="flex items-center space-x-3 space-y-0">
                                                <FormControl>
                                                    <RadioGroupItem value="weighted" />
                                                </FormControl>
                                                <FormLabel className="font-normal">
                                                    Weighted Integration
                                                </FormLabel>
                                            </FormItem>
                                            <FormItem className="flex items-center space-x-3 space-y-0">
                                                <FormControl>
                                                    <RadioGroupItem value="rerank" />
                                                </FormControl>
                                                <FormLabel className="font-normal">
                                                    Rerank Model
                                                </FormLabel>
                                            </FormItem>
                                        </RadioGroup>
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        {hybridStrategy === 'weighted' && (
                            <div className="space-y-4 pl-6 border-l-2">
                                <FormField
                                    control={form.control}
                                    name="weight_vector"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className="text-xs">Vector Weight: {field.value}</FormLabel>
                                            <FormControl>
                                                <Slider
                                                    min={0}
                                                    max={1.0}
                                                    step={0.1}
                                                    defaultValue={[field.value || 0.7]}
                                                    onValueChange={(vals: number[]) => {
                                                        field.onChange(vals[0]);
                                                        // Auto adjust keyword weight? Optional but UX friendly
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
                                            <FormLabel className="text-xs">Keyword Weight: {field.value}</FormLabel>
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
                                    <FormItem className="pl-6 border-l-2">
                                        <FormLabel>Rerank Model</FormLabel>
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select a rerank model" />
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
                                                        No Rerank models found. Please configure in Settings.
                                                    </div>
                                                )}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        )}
                    </div>
                )}

                <Button type="submit" className="w-full" disabled={isLoading}>
                    {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                    开始召回测试 (Retrieve)
                </Button>
            </form>
        </Form>
    );
};
