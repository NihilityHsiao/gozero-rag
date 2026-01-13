import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { Loader2 } from 'lucide-react';
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
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Checkbox } from '@/components/ui/checkbox';
import { addUserApi } from '@/api/user_model';

const formSchema = z.object({
    model_type: z.enum(['chat', 'embedding', 'qa', 'rewrite', 'rerank']),
    config_name: z.string().min(1, '配置名称不能为空'),
    model_name: z.string().min(1, '模型名称不能为空'),
    base_url: z.string().url('必须是有效的URL').min(1, 'Base URL不能为空'),
    api_key: z.string().min(1, 'API Key不能为空'),
    max_tokens: z.number().optional(),
    temperature: z.number().optional(),
    is_default: z.boolean(),
});

type FormValues = z.infer<typeof formSchema>;

interface AddModelDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSuccess: () => void;
}

export default function AddModelDialog({ open, onOpenChange, onSuccess }: AddModelDialogProps) {
    const [showAdvanced, setShowAdvanced] = useState(false);

    // Default values must match the schema type completely
    const form = useForm<FormValues>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            model_type: 'chat',
            config_name: '',
            model_name: '',
            base_url: 'https://api.siliconflow.cn/v1',
            api_key: '',
            max_tokens: 2048,
            temperature: 0.7,
            is_default: false,
        },
    });

    const onSubmit = async (values: FormValues) => {
        try {
            await addUserApi({
                ...values,
                max_tokens: values.max_tokens ?? 2048,
                temperature: values.temperature ?? 0.7,
                status: 1,
                is_default: values.is_default ? 1 : 0,
            });
            toast.success('模型添加成功');
            onSuccess();
            onOpenChange(false);
            form.reset();
        } catch (error: any) {
            console.error(error);
            const msg = error?.message || '模型添加失败';
            toast.error(msg);
        }
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>添加模型配置</DialogTitle>
                    <DialogDescription>
                        配置新的 LLM 模型接口，用于 Knowledge Base 的各个环节。
                    </DialogDescription>
                </DialogHeader>

                <Form {...form}>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <FormField
                            control={form.control}
                            name="model_type"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>模型类型</FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue placeholder="选择模型类型" />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="chat">Chat (对话)</SelectItem>
                                            <SelectItem value="embedding">Embedding (向量)</SelectItem>
                                            <SelectItem value="qa">QA (问答生成)</SelectItem>
                                            <SelectItem value="rerank">Rerank (重排序)</SelectItem>
                                            <SelectItem value="rewrite">Rewrite (重写)</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="config_name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>配置名称</FormLabel>
                                    <FormControl>
                                        <Input placeholder="例如: DeepSeek v3 Chat" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="model_name"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>模型名称 (Model ID)</FormLabel>
                                    <FormControl>
                                        <Input placeholder="例如: deepseek-ai/DeepSeek-V3" {...field} />
                                    </FormControl>
                                    <FormDescription className="text-xs">
                                        必须与服务商提供的 Model ID 完全一致
                                    </FormDescription>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="base_url"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Base URL</FormLabel>
                                    <FormControl>
                                        <Input placeholder="https://api.example.com/v1" {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <FormField
                            control={form.control}
                            name="api_key"
                            render={({ field }) => (
                                <FormItem>
                                    <FormLabel>API Key</FormLabel>
                                    <FormControl>
                                        <Input type="password" placeholder="sk-..." {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )}
                        />

                        <div className="pt-2">
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                className="w-full text-xs text-gray-500"
                                onClick={() => setShowAdvanced(!showAdvanced)}
                            >
                                {showAdvanced ? '隐藏高级设置' : '显示高级设置 (Max Tokens, Temperature)'}
                            </Button>
                        </div>

                        {showAdvanced && (
                            <div className="grid grid-cols-2 gap-4 p-4 bg-gray-50 rounded-md">
                                <FormField
                                    control={form.control}
                                    name="max_tokens"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className="text-xs">Max Tokens</FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="number"
                                                    {...field}
                                                    value={field.value || 2048}
                                                    onChange={e => field.onChange(e.target.value ? parseInt(e.target.value) : undefined)}
                                                    className="h-8 text-xs"
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="temperature"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel className="text-xs">Temperature</FormLabel>
                                            <FormControl>
                                                <Input
                                                    type="number"
                                                    step="0.1"
                                                    {...field}
                                                    value={field.value || 0.7}
                                                    onChange={e => field.onChange(e.target.value ? parseFloat(e.target.value) : undefined)}
                                                    className="h-8 text-xs"
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>
                        )}

                        <FormField
                            control={form.control}
                            name="is_default"
                            render={({ field }) => (
                                <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                                    <FormControl>
                                        <Checkbox
                                            checked={field.value}
                                            onCheckedChange={field.onChange}
                                        />
                                    </FormControl>
                                    <div className="space-y-1 leading-none">
                                        <FormLabel>
                                            设为默认模型
                                        </FormLabel>
                                        <FormDescription>
                                            该类型的任务将优先使用此模型
                                        </FormDescription>
                                    </div>
                                </FormItem>
                            )}
                        />

                        <DialogFooter className="pt-4">
                            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                                取消
                            </Button>
                            <Button type="submit" disabled={form.formState.isSubmitting}>
                                {form.formState.isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                确定
                            </Button>
                        </DialogFooter>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
}
