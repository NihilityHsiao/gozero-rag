import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { Loader2, Settings, FileCog, X, Info, Network } from 'lucide-react';

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
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Slider } from '@/components/ui/slider';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';

import { getKnowledgeDetail, updateKnowledgeBase } from '@/api/knowledge';
import { getUserApiList } from '@/api/user_model';
import { listLlmFactories, listTenantLlm } from '@/api/llm';
import { useKnowledgeStore } from '@/store/useKnowledgeStore';
import { useAuthStore } from '@/store/useAuthStore';
import type { KnowledgeBaseInfo, UserApiInfo, LlmFactoryInfo, TenantLlmInfo } from '@/types';

// Import GraphSettings
import { GraphSettings } from './GraphSettings';

// Schema Definition
const settingsSchema = z.object({
    // General
    name: z.string().min(1, '请输入名称').max(50, '名称过长'),
    description: z.string().max(500, '描述过长').optional(),
    permission: z.enum(['me', 'team']),
    status: z.boolean(),

    // Models
    qa_model_id: z.number().optional(),
    chat_model_id: z.number().optional(),
    rerank_model_id: z.number().optional(),
    rewrite_model_id: z.number().optional(),

    // Parser Config
    parser_id: z.enum(['general', 'resume']),

    // Detailed Config Fields
    chunk_token_num: z.number().min(128).max(2048),
    chunk_overlap_token_num: z.number().min(0).max(200),
    separator: z.array(z.string()), // Changed to array
    layout_recognize: z.boolean(),
    qa_num: z.number().min(0).max(50).optional(),
    qa_llm_id: z.string().optional(),
    pdf_parser: z.enum(['pdfcpu', 'eino', 'deepdoc']),

    // Graph RAG Config
    graph_rag: z.object({
        enable_graph: z.boolean(),
        graph_llm_id: z.string().optional(),
        entity_types: z.array(z.string()).optional(),
        enable_entity_resolution: z.boolean().optional(),
        enable_community: z.boolean().optional(),
    }).optional(),

}).superRefine((val, ctx) => {
    if ((val.qa_num || 0) > 0 && !val.qa_llm_id) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "开启 QA 生成时必须选择模型",
            path: ["qa_llm_id"],
        });
    }

    if (val.graph_rag?.enable_graph && !val.graph_rag.graph_llm_id) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "启用知识图谱必须选择 LLM 模型",
            path: ["graph_rag", "graph_llm_id"],
        });
    }
});

type SettingsFormValues = z.infer<typeof settingsSchema>;

export default function KnowledgeSettings() {
    const { id } = useParams<{ id: string }>();
    const knowledgeId = id!;
    const { fetchList } = useKnowledgeStore();
    const [activeTab, setActiveTab] = useState<'general' | 'config' | 'graph'>('general');

    const [info, setInfo] = useState<KnowledgeBaseInfo | null>(null);
    const [loading, setLoading] = useState(true);
    const [models, setModels] = useState<UserApiInfo[]>([]);
    const [factories, setFactories] = useState<LlmFactoryInfo[]>([]);
    const [tenantLlms, setTenantLlms] = useState<TenantLlmInfo[]>([]);
    const user = useAuthStore(state => state.userInfo);

    // Tag Input State
    const [tagInput, setTagInput] = useState('');



    const form = useForm<SettingsFormValues>({
        resolver: zodResolver(settingsSchema),
        defaultValues: {
            name: '',
            description: '',
            permission: 'me',
            status: true,
            qa_model_id: undefined,
            chat_model_id: undefined,
            rerank_model_id: undefined,
            rewrite_model_id: undefined,

            parser_id: 'general',
            chunk_token_num: 512,
            chunk_overlap_token_num: 64,
            separator: ['\\n\\n', '\\n', ' ', ''], // Default separators as array
            layout_recognize: true,
            qa_num: 0,
            qa_llm_id: '',
            pdf_parser: 'eino',
            graph_rag: {
                enable_graph: false,
                entity_types: ['organization', 'person', 'geo', 'event', 'category'],
                enable_entity_resolution: false,
                enable_community: false,
                graph_llm_id: '',
            }
        },
    });

    useEffect(() => {
        if (!knowledgeId) return;
        const load = async () => {
            try {
                setLoading(true);
                const res = await getKnowledgeDetail(knowledgeId);
                setInfo(res);

                // Parse model_ids
                const modelMap = new Map<string, number>();
                if (res.model_ids) {
                    res.model_ids.forEach((m: any) => {
                        modelMap.set(m.model_type, m.model_id);
                    });
                }

                // Parse parser_config
                let parserConfig: any = {};
                try {
                    if (res.parser_config) {
                        parserConfig = JSON.parse(res.parser_config);
                    }
                } catch (e) {
                    console.error("Failed to parse parser_config", e);
                }

                // Handle separator (string[] or string or undefined)
                let separatorList = ['\\n\\n', '\\n', ' ', ''];
                if (parserConfig.separator) {
                    if (Array.isArray(parserConfig.separator)) {
                        separatorList = parserConfig.separator;
                    }
                }

                // Handle Graph Config
                const graphConfig = parserConfig.graph_rag || {
                    enable_graph: false,
                    entity_types: ['organization', 'person', 'geo', 'event', 'category'],
                    enable_entity_resolution: false,
                    enable_community: false,
                    graph_llm_id: '',
                };

                form.reset({
                    name: res.name,
                    description: res.description,
                    permission: (res.permission as 'me' | 'team') || 'me',
                    status: res.status === 1,
                    qa_model_id: modelMap.get('qa'),
                    chat_model_id: modelMap.get('chat'),
                    rerank_model_id: modelMap.get('rerank'),
                    rewrite_model_id: modelMap.get('rewrite'),

                    parser_id: (res.parser_id as 'general' | 'resume') || 'general',
                    chunk_token_num: parserConfig.chunk_token_num ?? 512,
                    chunk_overlap_token_num: parserConfig.chunk_overlap_token_num ?? 64,
                    separator: separatorList,
                    layout_recognize: parserConfig.layout_recognize ?? true,
                    qa_num: parserConfig.qa_num ?? 0,
                    qa_llm_id: parserConfig.qa_llm_id || '',
                    pdf_parser: parserConfig.pdf_parser || 'eino',
                    graph_rag: graphConfig,
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

    useEffect(() => {
        if (!user?.user_id) return;
        getUserApiList(user.user_id, { status: 1 })
            .then(res => setModels(res.list || []))
            .catch(err => console.error('Failed to load models:', err));

        listLlmFactories()
            .then(res => setFactories(res.list || []))
            .catch(err => console.error('Failed to load factories:', err));

        listTenantLlm({ model_type: 'LLM', status: 1 })
            .then(res => setTenantLlms(res.list || []))
            .catch(err => console.error('Failed to load tenant llms:', err));
    }, [user?.user_id]);

    const onSubmit = async (data: SettingsFormValues) => {
        if (!knowledgeId) return;
        try {
            const parserConfig: any = {};
            if (data.parser_id === 'general') {
                parserConfig.chunk_token_num = data.chunk_token_num;
                parserConfig.chunk_overlap_token_num = data.chunk_overlap_token_num;
                parserConfig.separator = data.separator; // Save as array
                parserConfig.layout_recognize = data.layout_recognize;
                parserConfig.qa_num = data.qa_num;
                parserConfig.qa_llm_id = data.qa_llm_id;
                parserConfig.pdf_parser = data.pdf_parser;
                parserConfig.graph_rag = data.graph_rag; // Save Graph Config
            } else {
                parserConfig.pdf_parser = data.pdf_parser;
            }

            await updateKnowledgeBase(knowledgeId, {
                name: data.name,
                description: data.description,
                permission: data.permission,
                status: data.status ? 1 : 0,
                qa_model_id: data.qa_model_id,
                chat_model_id: data.chat_model_id,
                rerank_model_id: data.rerank_model_id,
                rewrite_model_id: data.rewrite_model_id,
                parser_id: data.parser_id,
                parser_config: JSON.stringify(parserConfig),
            });
            toast.success('设置更新成功');
            const res = await getKnowledgeDetail(knowledgeId);
            setInfo(res);
            fetchList();
        } catch (error) {
            console.error(error);
            toast.error('更新设置失败');
        }
    };

    // Parse embedding model info
    let displayModelName = '';
    let displayProvider = '';
    let displayIcon = '';
    let displayType = 'Embedding';

    if (info?.embd_id) {
        const parts = info.embd_id.split('@');
        displayModelName = parts[0];
        displayProvider = parts[1] || '';

        // Try to match with existing models to get Icon
        const matched = models.find(m => m.model_name === displayModelName);
        if (matched) {
            displayIcon = matched.icon;
            displayType = matched.model_type;
        } else if (displayProvider) {
            const factory = factories.find(f => f.name === displayProvider);
            if (factory) {
                displayIcon = factory.logo;
            }
        }
    } else if (info?.embedding_model_id) {
        const matched = models.find(m => m.id === info.embedding_model_id);
        if (matched) {
            displayModelName = matched.model_name;
            displayProvider = matched.provider;
            displayIcon = matched.icon;
            displayType = matched.model_type;
        }
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
            </div>
        );
    }

    return (
        <div className="flex flex-col md:flex-row gap-8">
            <aside className="w-full md:w-64 space-y-1">
                <Button variant={activeTab === 'general' ? 'secondary' : 'ghost'} className="w-full justify-start font-normal" onClick={() => setActiveTab('general')}>
                    <Settings className="mr-2 h-4 w-4" /> 通用设置
                </Button>
                <Button variant={activeTab === 'config' ? 'secondary' : 'ghost'} className="w-full justify-start font-normal" onClick={() => setActiveTab('config')}>
                    <FileCog className="mr-2 h-4 w-4" /> 配置
                </Button>
                <Button variant={activeTab === 'graph' ? 'secondary' : 'ghost'} className="w-full justify-start font-normal" onClick={() => setActiveTab('graph')}>
                    <Network className="mr-2 h-4 w-4" /> 知识图谱
                </Button>
            </aside>

            <div className="flex-1 max-w-3xl">
                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                        {activeTab === 'general' && (
                            <>
                                <section className="space-y-4">
                                    <div className="space-y-1">
                                        <h3 className="text-lg font-semibold">基本信息</h3>
                                        <p className="text-sm text-muted-foreground">管理知识库的基本属性</p>
                                    </div>
                                    <div className="space-y-4 p-1">
                                        <FormField control={form.control} name="name" render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>名称</FormLabel>
                                                <FormControl><Input {...field} /></FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )} />
                                        <FormField control={form.control} name="description" render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>描述</FormLabel>
                                                <FormControl><Textarea {...field} className="min-h-[100px] resize-none" /></FormControl>
                                                <FormMessage />
                                            </FormItem>
                                        )} />
                                        <FormField control={form.control} name="status" render={({ field }) => (
                                            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm bg-card">
                                                <div className="space-y-0.5">
                                                    <FormLabel>启用状态</FormLabel>
                                                    <FormDescription>禁用后无法检索</FormDescription>
                                                </div>
                                                <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                                            </FormItem>
                                        )} />
                                        {info && (
                                            <div className="text-xs text-muted-foreground">
                                                创建: {info.created_at}
                                            </div>
                                        )}
                                    </div>
                                </section>

                                {info && info.created_by === user?.user_id && (
                                    <section className="space-y-4 pt-6 border-t">
                                        <h3 className="text-lg font-semibold">权限管理</h3>
                                        <Card className="border-none shadow-none bg-transparent p-0">
                                            <CardContent className="p-0">
                                                <FormField control={form.control} name="permission" render={({ field }) => (
                                                    <FormItem className="space-y-3">
                                                        <FormLabel>可见范围</FormLabel>
                                                        <FormControl>
                                                            <RadioGroup onValueChange={field.onChange} defaultValue={field.value} className="flex flex-col space-y-1">
                                                                <FormItem className="flex items-center space-x-3 space-y-0">
                                                                    <FormControl><RadioGroupItem value="me" /></FormControl>
                                                                    <FormLabel className="font-normal">仅自己</FormLabel>
                                                                </FormItem>
                                                                {/* Example of Team logic if needed */}
                                                                <FormItem className="flex items-center space-x-3 space-y-0">
                                                                    <FormControl><RadioGroupItem value="team" /></FormControl>
                                                                    <FormLabel className="font-normal">团队可见</FormLabel>
                                                                </FormItem>
                                                            </RadioGroup>
                                                        </FormControl>
                                                    </FormItem>
                                                )} />
                                            </CardContent>
                                        </Card>
                                    </section>
                                )}
                            </>
                        )}

                        {activeTab === 'config' && (
                            <div className="space-y-8">
                                <Card>
                                    <CardHeader>
                                        <CardTitle>索引模型</CardTitle>
                                        <CardDescription>创建后不可更改</CardDescription>
                                    </CardHeader>
                                    <CardContent>
                                        {displayModelName ? (
                                            <div className="flex items-center gap-3 p-3 bg-muted/50 rounded-lg border">
                                                {displayIcon ? (
                                                    <img src={displayIcon} alt="provider" className="w-8 h-8 rounded-sm bg-white p-1" />
                                                ) : (
                                                    <div className="w-8 h-8 bg-gray-200 rounded-sm" />
                                                )}
                                                <div>
                                                    <div className="font-medium">{displayModelName}</div>
                                                    <div className="text-xs text-muted-foreground flex items-center gap-1">
                                                        {displayProvider && <span>{displayProvider}</span>}
                                                        <Badge variant="outline" className="text-[10px] h-4 px-1">{displayType}</Badge>
                                                    </div>
                                                </div>
                                            </div>
                                        ) : (
                                            <div className="text-sm text-yellow-600">模型信息加载中或数据异常</div>
                                        )}
                                    </CardContent>
                                </Card>

                                <Card>
                                    <CardHeader>
                                        <CardTitle>解析规则</CardTitle>
                                        <CardDescription>配置文档的解析与切片策略</CardDescription>
                                    </CardHeader>
                                    <CardContent className="space-y-6">
                                        <FormField control={form.control} name="parser_id" render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>解析模式</FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                                                    <SelectContent>
                                                        <SelectItem value="general">General (通用)</SelectItem>
                                                        <SelectItem value="resume">Resume (简历)</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </FormItem>
                                        )} />

                                        {form.watch('parser_id') === 'general' && (
                                            <div className="space-y-4 pl-1">
                                                <div className="grid grid-cols-2 gap-4">
                                                    <FormField control={form.control} name="chunk_token_num" render={({ field }) => (
                                                        <FormItem>
                                                            <div className="flex justify-between">
                                                                <FormLabel>块大小</FormLabel>
                                                                <span className="text-xs font-mono text-muted-foreground">{field.value} tokens</span>
                                                            </div>
                                                            <FormControl>
                                                                <Slider min={128} max={2048} step={1} value={[field.value]} onValueChange={(vals) => field.onChange(vals[0])} />
                                                            </FormControl>
                                                        </FormItem>
                                                    )} />
                                                    <FormField control={form.control} name="chunk_overlap_token_num" render={({ field }) => (
                                                        <FormItem>
                                                            <div className="flex justify-between">
                                                                <FormLabel>重叠大小</FormLabel>
                                                                <span className="text-xs font-mono text-muted-foreground">{field.value} tokens</span>
                                                            </div>
                                                            <FormControl>
                                                                <Slider min={0} max={200} step={1} value={[field.value]} onValueChange={(vals) => field.onChange(vals[0])} />
                                                            </FormControl>
                                                        </FormItem>
                                                    )} />
                                                </div>

                                                <FormField control={form.control} name="separator" render={({ field }) => (
                                                    <FormItem>
                                                        <FormLabel>分隔符</FormLabel>
                                                        <FormControl>
                                                            <div className="flex flex-wrap gap-2 p-2 border rounded-md bg-white min-h-[42px]">
                                                                {field.value.map((sep, idx) => (
                                                                    <Badge key={idx} variant="secondary" className="px-2 py-1 flex items-center gap-1 font-mono text-xs">
                                                                        {sep.replace(/\n/g, '\\n')}
                                                                        <X className="w-3 h-3 cursor-pointer hover:text-destructive" onClick={() => {
                                                                            const newVal = [...field.value];
                                                                            newVal.splice(idx, 1);
                                                                            field.onChange(newVal);
                                                                        }} />
                                                                    </Badge>
                                                                ))}
                                                                <input
                                                                    className="flex-1 outline-none bg-transparent text-sm min-w-[60px]"
                                                                    placeholder={field.value.length === 0 ? "输入字符并回车..." : ""}
                                                                    value={tagInput}
                                                                    onChange={(e) => setTagInput(e.target.value)}
                                                                    onKeyDown={(e) => {
                                                                        if (e.key === 'Enter') {
                                                                            e.preventDefault();
                                                                            if (tagInput) {
                                                                                // Handle escaped characters if user types \n
                                                                                let val = tagInput;
                                                                                if (val === '\\n') val = '\n';
                                                                                if (val === '\\n\\n') val = '\n\n';

                                                                                if (!field.value.includes(val)) {
                                                                                    field.onChange([...field.value, val]);
                                                                                }
                                                                                setTagInput('');
                                                                            }
                                                                        }
                                                                        if (e.key === 'Backspace' && !tagInput && field.value.length > 0) {
                                                                            // remove last
                                                                            const newVal = [...field.value];
                                                                            newVal.pop();
                                                                            field.onChange(newVal);
                                                                        }
                                                                    }}
                                                                />
                                                            </div>
                                                        </FormControl>
                                                        <FormDescription>用于分块的字符序列，优先级从左到右 (支持 \n 转义)</FormDescription>
                                                    </FormItem>
                                                )} />

                                                <FormField control={form.control} name="layout_recognize" render={({ field }) => (
                                                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3">
                                                        <div className="space-y-0.5">
                                                            <FormLabel>版面识别</FormLabel>
                                                            <FormDescription>增强表格与标题的识别效果</FormDescription>
                                                        </div>
                                                        <FormControl><Switch checked={field.value} onCheckedChange={field.onChange} /></FormControl>
                                                    </FormItem>
                                                )} />

                                                <div className="pt-4 border-t">
                                                    <FormField control={form.control} name="qa_num" render={({ field }) => (
                                                        <FormItem>
                                                            <div className="flex justify-between">
                                                                <FormLabel>QA 生成数量 (Pre-generated QA)</FormLabel>
                                                                <span className="text-xs font-mono text-muted-foreground">{field.value}</span>
                                                            </div>
                                                            <FormControl>
                                                                <Slider
                                                                    min={0}
                                                                    max={50}
                                                                    step={1}
                                                                    value={[field.value || 0]}
                                                                    onValueChange={(vals) => field.onChange(vals[0])}
                                                                    disabled={!tenantLlms.length}
                                                                />
                                                            </FormControl>
                                                            <FormDescription>
                                                                每个分片生成的 QA 对数量。
                                                                {tenantLlms.length === 0 && <span className="text-destructive ml-1">未检测到可用生成模型，无法启用。</span>}
                                                            </FormDescription>
                                                        </FormItem>
                                                    )} />

                                                    {(form.watch('qa_num') || 0) > 0 && (
                                                        <div className="mt-4 space-y-4 animate-in fade-in slide-in-from-top-2">
                                                            <Alert className="bg-yellow-50 border-yellow-200 text-yellow-800">
                                                                <Info className="h-4 w-4 text-yellow-600" />
                                                                <AlertTitle>费用提示</AlertTitle>
                                                                <AlertDescription className="text-xs">开启 QA 生成会显著增加 LLM Token 消耗，导致产生额外费用。请确保所选模型额度充足。</AlertDescription>
                                                            </Alert>

                                                            <FormField control={form.control} name="qa_llm_id" render={({ field }) => (
                                                                <FormItem>
                                                                    <FormLabel>生成模型</FormLabel>
                                                                    <Select onValueChange={field.onChange} value={field.value}>
                                                                        <FormControl><SelectTrigger><SelectValue placeholder="选择模型" /></SelectTrigger></FormControl>
                                                                        <SelectContent>
                                                                            {tenantLlms.map((model) => (
                                                                                <SelectItem key={model.id} value={`${model.llm_name}@${model.llm_factory}`}>
                                                                                    {model.llm_name} @ {model.llm_factory}
                                                                                </SelectItem>
                                                                            ))}
                                                                        </SelectContent>
                                                                    </Select>
                                                                    <FormMessage />
                                                                </FormItem>
                                                            )} />
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        )}

                                        <FormField control={form.control} name="pdf_parser" render={({ field }) => (
                                            <FormItem>
                                                <FormLabel>PDF 解析器</FormLabel>
                                                <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                    <FormControl><SelectTrigger><SelectValue /></SelectTrigger></FormControl>
                                                    <SelectContent>
                                                        <SelectItem value="eino">Eino (通用/稳定)</SelectItem>
                                                        <SelectItem value="pdfcpu">pdfcpu (快速)</SelectItem>
                                                        <SelectItem value="deepdoc">DeepDoc (深度OCR)</SelectItem>
                                                    </SelectContent>
                                                </Select>
                                            </FormItem>
                                        )} />
                                    </CardContent>
                                </Card>
                            </div>
                        )}

                        {activeTab === 'graph' && (
                            <GraphSettings />
                        )}

                        <div className="flex justify-end pt-4">
                            <Button type="submit" disabled={form.formState.isSubmitting}>
                                {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                保存设置
                            </Button>
                        </div>
                    </form>
                </Form>
            </div>
        </div>
    );
}
