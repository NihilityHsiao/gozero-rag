import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2, ChevronLeft, ChevronRight } from 'lucide-react';
import { toast } from 'sonner';

import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
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
import { Button } from '@/components/ui/button';

import { listLlmFactories, addTenantLlm } from '@/api/llm';
import type { LlmFactoryInfo, ModelConfig } from '@/types';

interface AddModelDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSuccess: () => void;
}

// 表单验证 schema (Step 2)
const configFormSchema = z.object({
    api_key: z.string().min(1, 'API Key 不能为空'),
    api_base: z.string().url('请输入有效的 URL').min(1, 'API Base 不能为空'),
});

type ConfigFormValues = z.infer<typeof configFormSchema>;

export default function AddModelDialog({ open, onOpenChange, onSuccess }: AddModelDialogProps) {
    // Step 1: 选择厂商, Step 2: 配置模型
    const [step, setStep] = useState<1 | 2>(1);
    const [factories, setFactories] = useState<LlmFactoryInfo[]>([]);
    const [loadingFactories, setLoadingFactories] = useState(false);
    const [selectedFactory, setSelectedFactory] = useState<LlmFactoryInfo | null>(null);
    const [modelInputs, setModelInputs] = useState<Record<string, string>>({});
    const [submitting, setSubmitting] = useState(false);

    const form = useForm<ConfigFormValues>({
        resolver: zodResolver(configFormSchema),
        defaultValues: {
            api_key: '',
            api_base: '',
        },
    });

    // 获取厂商列表
    const fetchFactories = async () => {
        setLoadingFactories(true);
        try {
            const res = await listLlmFactories();
            setFactories(res.list || []);
        } catch (error) {
            console.error(error);
            toast.error('获取厂商列表失败');
        } finally {
            setLoadingFactories(false);
        }
    };

    useEffect(() => {
        if (open && step === 1) {
            fetchFactories();
        }
    }, [open, step]);

    // 重置状态
    const resetState = () => {
        setStep(1);
        setSelectedFactory(null);
        setModelInputs({});
        form.reset();
    };

    // 关闭对话框
    const handleClose = () => {
        onOpenChange(false);
        setTimeout(resetState, 300);
    };

    // 选择厂商
    const handleSelectFactory = (factory: LlmFactoryInfo) => {
        setSelectedFactory(factory);
        // 根据厂商设置默认的 API Base
        const defaultApiBase = getDefaultApiBase(factory.name);
        form.setValue('api_base', defaultApiBase);
        // 初始化模型输入
        const inputs: Record<string, string> = {};
        factory.tag_list.forEach(tag => {
            inputs[tag] = '';
        });
        setModelInputs(inputs);
        setStep(2);
    };

    // 获取默认 API Base
    const getDefaultApiBase = (factoryName: string): string => {
        const defaults: Record<string, string> = {
            'SiliconFlow': 'https://api.siliconflow.cn/v1',
            'OpenAI': 'https://api.openai.com/v1',
            'Anthropic': 'https://api.anthropic.com',
            'DeepSeek': 'https://api.deepseek.com',
            'Moonshot': 'https://api.moonshot.cn/v1',
            'Zhipu': 'https://open.bigmodel.cn/api/paas/v4',
        };
        return defaults[factoryName] || '';
    };

    // 更新模型输入
    const handleModelInputChange = (type: string, value: string) => {
        setModelInputs(prev => ({
            ...prev,
            [type]: value,
        }));
    };

    // 提交表单
    const onSubmit = async (values: ConfigFormValues) => {
        if (!selectedFactory) return;

        // 过滤出有值的模型
        const models: ModelConfig[] = Object.entries(modelInputs)
            .filter(([_, llmName]) => llmName.trim() !== '')
            .map(([modelType, llmName]) => ({
                model_type: modelType,
                llm_name: llmName.trim(),
            }));

        if (models.length === 0) {
            toast.error('请至少填写一个模型名称');
            return;
        }

        setSubmitting(true);
        try {
            const res = await addTenantLlm({
                llm_factory: selectedFactory.name,
                api_key: values.api_key,
                api_base: values.api_base,
                models,
            });

            if (res.success_count > 0) {
                toast.success(`成功添加 ${res.success_count} 个模型`);
            }
            if (res.failed_count > 0) {
                toast.warning(`${res.failed_count} 个模型添加失败: ${res.failed_models.join(', ')}`);
            }

            onSuccess();
            handleClose();
        } catch (error: any) {
            console.error(error);
            const msg = error?.message || '添加失败';
            toast.error(msg);
        } finally {
            setSubmitting(false);
        }
    };

    // 获取模型类型的中文名称
    const getModelTypeLabel = (type: string): string => {
        const labels: Record<string, string> = {
            'LLM': '对话模型 (Chat)',
            'Embedding': '嵌入模型 (Embedding)',
            'Rerank': '重排序模型 (Rerank)',
            'ASR': '语音识别 (ASR)',
            'TTS': '语音合成 (TTS)',
            'Image2Text': '图片转文字',
            'Text2Image': '文字转图片',
            'Video': '视频生成',
        };
        return labels[type] || type;
    };

    // 获取模型类型的 placeholder
    const getModelTypePlaceholder = (type: string): string => {
        const placeholders: Record<string, string> = {
            'LLM': '例如: deepseek-ai/DeepSeek-V3',
            'Embedding': '例如: BAAI/bge-m3',
            'Rerank': '例如: BAAI/bge-reranker-v2-m3',
            'ASR': '例如: FunAudioLLM/SenseVoiceSmall',
            'TTS': '例如: fishaudio/fish-speech-1.5',
            'Image2Text': '例如: OpenGVLab/InternVL2-8B',
            'Text2Image': '例如: stabilityai/stable-diffusion-3',
            'Video': '例如: Lightricks/LTX-Video',
        };
        return placeholders[type] || '输入模型名称';
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>
                        {step === 1 ? '选择模型厂商' : `配置 ${selectedFactory?.name} 模型`}
                    </DialogTitle>
                    <DialogDescription>
                        {step === 1
                            ? '选择您要使用的 LLM 厂商，系统会自动配置 API 地址'
                            : '填写 API 凭证和您需要的模型名称，留空的模型类型将被跳过'}
                    </DialogDescription>
                </DialogHeader>

                {/* Step 1: 选择厂商 */}
                {step === 1 && (
                    <div className="py-4">
                        {loadingFactories ? (
                            <div className="flex items-center justify-center py-12 text-gray-500">
                                <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                                加载厂商列表...
                            </div>
                        ) : (
                            <div className="grid grid-cols-3 gap-3">
                                {factories.map((factory) => (
                                    <button
                                        key={factory.name}
                                        onClick={() => handleSelectFactory(factory)}
                                        className="flex flex-col items-center p-4 border rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors group"
                                    >
                                        {factory.logo ? (
                                            <img
                                                src={factory.logo}
                                                alt={factory.name}
                                                className="w-10 h-10 rounded mb-2"
                                            />
                                        ) : (
                                            <div className="w-10 h-10 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-lg mb-2">
                                                {factory.name.charAt(0)}
                                            </div>
                                        )}
                                        <span className="text-sm font-medium text-gray-700 group-hover:text-blue-600">
                                            {factory.name}
                                        </span>
                                        <span className="text-xs text-gray-400 mt-1">
                                            {factory.tag_list.length} 种模型类型
                                        </span>
                                    </button>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {/* Step 2: 配置模型 */}
                {step === 2 && selectedFactory && (
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 py-2">
                            {/* 厂商信息 */}
                            <div className="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                                {selectedFactory.logo ? (
                                    <img
                                        src={selectedFactory.logo}
                                        alt={selectedFactory.name}
                                        className="w-8 h-8 rounded"
                                    />
                                ) : (
                                    <div className="w-8 h-8 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-sm">
                                        {selectedFactory.name.charAt(0)}
                                    </div>
                                )}
                                <div>
                                    <div className="font-medium">{selectedFactory.name}</div>
                                    <div className="text-xs text-gray-500">
                                        支持: {selectedFactory.tag_list.join(', ')}
                                    </div>
                                </div>
                            </div>

                            {/* API 配置 */}
                            <div className="grid grid-cols-2 gap-4">
                                <FormField
                                    control={form.control}
                                    name="api_key"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>API Key *</FormLabel>
                                            <FormControl>
                                                <Input type="password" placeholder="sk-..." {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="api_base"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>API Base *</FormLabel>
                                            <FormControl>
                                                <Input placeholder="https://api.example.com/v1" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            {/* 模型配置 */}
                            <div className="border-t pt-4 mt-4">
                                <h4 className="text-sm font-medium mb-3">模型配置 (填写需要的模型，留空将跳过)</h4>
                                <div className="space-y-3">
                                    {selectedFactory.tag_list.map((type) => (
                                        <div key={type} className="flex items-center gap-3">
                                            <div className="w-28 text-sm text-gray-600 shrink-0">
                                                {getModelTypeLabel(type)}
                                            </div>
                                            <Input
                                                value={modelInputs[type] || ''}
                                                onChange={(e) => handleModelInputChange(type, e.target.value)}
                                                placeholder={getModelTypePlaceholder(type)}
                                                className="flex-1"
                                            />
                                        </div>
                                    ))}
                                </div>
                                <FormDescription className="mt-3 text-xs">
                                    模型名称必须与厂商提供的 Model ID 完全一致
                                </FormDescription>
                            </div>

                            <DialogFooter className="pt-4 flex justify-between">
                                <Button
                                    type="button"
                                    variant="ghost"
                                    onClick={() => setStep(1)}
                                >
                                    <ChevronLeft className="mr-1 h-4 w-4" />
                                    返回选择厂商
                                </Button>
                                <div className="flex gap-2">
                                    <Button type="button" variant="outline" onClick={handleClose}>
                                        取消
                                    </Button>
                                    <Button type="submit" disabled={submitting}>
                                        {submitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                        添加模型
                                        <ChevronRight className="ml-1 h-4 w-4" />
                                    </Button>
                                </div>
                            </DialogFooter>
                        </form>
                    </Form>
                )}
            </DialogContent>
        </Dialog>
    );
}
