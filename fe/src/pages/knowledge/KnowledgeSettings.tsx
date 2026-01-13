import { useEffect, useState, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    FormDescription,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';

import { getKnowledgeDetail, updateKnowledgeBase } from '@/api/knowledge';
import { getUserApiList } from '@/api/user_model';
import { useKnowledgeStore } from '@/store/useKnowledgeStore';
import { useAuthStore } from '@/store/useAuthStore';
import type { KnowledgeBaseInfo, UserApiInfo } from '@/types';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';

// Schema Definition
const settingsSchema = z.object({
    name: z.string().min(1, '请输入名称').max(50, '名称过长'),
    description: z.string().max(500, '描述过长').optional(),
    status: z.boolean(),
    qa_model_id: z.number().optional(),
    chat_model_id: z.number().optional(),
    rerank_model_id: z.number().optional(),
    rewrite_model_id: z.number().optional(),
});

type SettingsFormValues = z.infer<typeof settingsSchema>;

export default function KnowledgeSettings() {
    const { id } = useParams<{ id: string }>();
    const knowledgeId = Number(id);
    const { fetchList } = useKnowledgeStore();

    // Additional state for read-only fields that are not in the form
    // or just use form.watch() if we put them in form (but model shouldn't be edited)
    // Better to store full info to display created_at etc.
    // We can use a separate state or just rely on form for editable ones.
    // Let's store full info in a state.
    const [info, setInfo] = useState<KnowledgeBaseInfo | null>(null);
    const [loading, setLoading] = useState(true);
    const [models, setModels] = useState<UserApiInfo[]>([]);
    const user = useAuthStore(state => state.userInfo);

    // 按类型分组模型
    const modelGroups = useMemo(() => {
        const groups: Record<string, UserApiInfo[]> = {
            qa: [],
            chat: [],
            embedding: [],
            rerank: [],
            rewrite: [],
        };
        models.forEach(m => {
            if (groups[m.model_type]) {
                groups[m.model_type].push(m);
            }
        });
        return groups;
    }, [models]);

    const form = useForm<SettingsFormValues>({
        resolver: zodResolver(settingsSchema),
        defaultValues: {
            name: '',
            description: '',
            status: true,
            qa_model_id: undefined,
            chat_model_id: undefined,
            rerank_model_id: undefined,
            rewrite_model_id: undefined,
        },
    });

    useEffect(() => {
        if (!knowledgeId) return;
        const load = async () => {
            try {
                setLoading(true);
                const res = await getKnowledgeDetail(knowledgeId);
                setInfo(res);

                // Parse model_ids to find specific model type IDs
                const modelMap = new Map<string, number>();
                if (res.model_ids) {
                    res.model_ids.forEach((m: any) => {
                        modelMap.set(m.model_type, m.model_id);
                    });
                }

                form.reset({
                    name: res.name,
                    description: res.description,
                    status: res.status === 1,
                    qa_model_id: modelMap.get('qa'),
                    chat_model_id: modelMap.get('chat'),
                    rerank_model_id: modelMap.get('rerank'),
                    rewrite_model_id: modelMap.get('rewrite'),
                });
            } catch (error) {
                console.error(error);
                toast.error('加载设置失败');
            } finally {
                setLoading(false);
            }
        };
        load();
    }, [knowledgeId, form]);

    // 加载用户可用模型列表
    useEffect(() => {
        if (!user?.user_id) return;
        getUserApiList(user.user_id, { status: 1 })
            .then(res => setModels(res.list || []))
            .catch(err => console.error('Failed to load models:', err));
    }, [user?.user_id]);

    const onSubmit = async (data: SettingsFormValues) => {
        if (!knowledgeId) return;
        try {
            await updateKnowledgeBase(knowledgeId, {
                name: data.name,
                description: data.description,
                status: data.status ? 1 : 0,
                qa_model_id: data.qa_model_id,
                chat_model_id: data.chat_model_id,
                rerank_model_id: data.rerank_model_id,
                rewrite_model_id: data.rewrite_model_id,
            });
            toast.success('设置更新成功');
            // Refresh detail
            const res = await getKnowledgeDetail(knowledgeId);
            setInfo(res);
            // Refresh global list to update sidebar name
            fetchList();
        } catch (error) {
            console.error(error);
            toast.error('更新设置失败');
        }
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div>
                <h3 className="text-lg font-medium">设置</h3>
                <p className="text-sm text-gray-500">
                    配置知识库设置。
                </p>
            </div>

            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>基本信息</CardTitle>
                            <CardDescription>
                                知识库的基本信息。
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6 max-w-2xl">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>知识库名称</FormLabel>
                                        <FormControl>
                                            <Input placeholder="例如：产品文档" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            <FormField
                                control={form.control}
                                name="description"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>描述</FormLabel>
                                        <FormControl>
                                            <Textarea
                                                placeholder="描述该知识库包含的内容..."
                                                className="min-h-[100px]"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {info && (
                                <div className="space-y-2">
                                    <FormLabel>Embedding 模型</FormLabel>
                                    <Input value={"类型: " + (models.find(m => m.id === info.embedding_model_id)?.config_name || '加载中...')} disabled className="bg-gray-50 text-gray-500" />
                                    <p className="text-[0.8rem] text-gray-500">
                                        Embedding 模型创建后无法更改。
                                    </p>
                                </div>
                            )}

                            <FormField
                                control={form.control}
                                name="status"
                                render={({ field }) => (
                                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm">
                                        <div className="space-y-0.5">
                                            <FormLabel className="text-base">
                                                启用状态
                                            </FormLabel>
                                            <FormDescription>
                                                禁用后该知识库将不会被检索。
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

                            {info && (
                                <div className="pt-4 flex items-center gap-6 text-xs text-gray-400">
                                    <span>创建时间: {new Date(info.created_at).toLocaleString()}</span>
                                    <span>更新时间: {new Date(info.updated_at).toLocaleString()}</span>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    {/* Model Configuration Card */}
                    <Card>
                        <CardHeader>
                            <CardTitle>模型配置</CardTitle>
                            <CardDescription>
                                配置知识库在 RAG 流程中使用的各类模型
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-6 max-w-2xl">
                            {/* Embedding Model - Read Only */}
                            <div className="space-y-2">
                                <div className="flex items-center gap-2">
                                    <FormLabel>Embedding 模型</FormLabel>
                                    <Badge variant="secondary" className="text-xs">不可变</Badge>
                                </div>
                                <div className="p-3 bg-gray-50 rounded-md border text-sm text-gray-600">
                                    {models.find(m => m.id === info?.embedding_model_id)?.config_name || '加载中...'}
                                </div>
                                <p className="text-xs text-gray-500">
                                    Embedding 模型在创建时确定，变更会导致索引失效。
                                </p>
                            </div>

                            {/* QA Model */}
                            <FormField
                                control={form.control}
                                name="qa_model_id"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>QA 生成模型</FormLabel>
                                        <Select
                                            onValueChange={(val) => field.onChange(val ? Number(val) : undefined)}
                                            value={field.value?.toString() ?? ''}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="选择 QA 模型（可选）" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {modelGroups.qa.map(m => (
                                                    <SelectItem key={m.id} value={m.id.toString()}>
                                                        {m.config_name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Chat Model */}
                            <FormField
                                control={form.control}
                                name="chat_model_id"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Chat 模型</FormLabel>
                                        <Select
                                            onValueChange={(val) => field.onChange(val ? Number(val) : undefined)}
                                            value={field.value?.toString() ?? ''}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="选择 Chat 模型（可选）" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {modelGroups.chat.map(m => (
                                                    <SelectItem key={m.id} value={m.id.toString()}>
                                                        {m.config_name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Rerank Model */}
                            <FormField
                                control={form.control}
                                name="rerank_model_id"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Rerank 模型</FormLabel>
                                        <Select
                                            onValueChange={(val) => field.onChange(val ? Number(val) : undefined)}
                                            value={field.value?.toString() ?? ''}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="选择 Rerank 模型（可选）" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {modelGroups.rerank.map(m => (
                                                    <SelectItem key={m.id} value={m.id.toString()}>
                                                        {m.config_name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {/* Rewrite Model */}
                            <FormField
                                control={form.control}
                                name="rewrite_model_id"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Rewrite 模型</FormLabel>
                                        <Select
                                            onValueChange={(val) => field.onChange(val ? Number(val) : undefined)}
                                            value={field.value?.toString() ?? ''}
                                        >
                                            <SelectTrigger>
                                                <SelectValue placeholder="选择 Rewrite 模型（可选）" />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {modelGroups.rewrite.map(m => (
                                                    <SelectItem key={m.id} value={m.id.toString()}>
                                                        {m.config_name}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                        </CardContent>
                    </Card>

                    <div className="flex justify-end">
                        <Button type="submit" disabled={form.formState.isSubmitting}>
                            {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            保存设置
                        </Button>
                    </div>
                </form>
            </Form>
        </div>
    );
}
