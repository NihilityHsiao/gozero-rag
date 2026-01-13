import type { UseFormReturn } from 'react-hook-form';
import { Settings2, AlignLeft, HelpCircle } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type { SegmentationSettings } from '@/types';
import { useState } from 'react';
import { useAuthStore } from '@/store/useAuthStore';
import { getUserApiList } from '@/api/user_model';
import { useNavigate } from 'react-router-dom';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Controller } from 'react-hook-form';

// Custom Slider controlled component
const CustomSlider = ({ label, name, min, max, control, suffix = '' }: any) => {
    return (
        <Controller
            name={name}
            control={control}
            render={({ field }) => (
                <div className="space-y-3">
                    <div className="flex justify-between items-center">
                        <Label>{label}</Label>
                        <div className="flex items-center gap-2 bg-gray-100 px-2 py-1 rounded-md">
                            <input
                                type="number"
                                className="bg-transparent w-16 text-right text-sm font-medium focus:outline-none"
                                min={min}
                                max={max}
                                value={field.value}
                                onChange={(e) => field.onChange(Number(e.target.value))}
                            />
                            <span className="text-sm text-gray-500">{suffix}</span>
                        </div>
                    </div>
                    <input
                        type="range"
                        min={min}
                        max={max}
                        step={1}
                        value={field.value}
                        onChange={(e) => field.onChange(Number(e.target.value))}
                        className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-blue-600"
                    />
                </div>
            )}
        />
    );
};

interface SegmentationFormProps {
    form: UseFormReturn<SegmentationSettings>;
}

export default function SegmentationForm({ form }: SegmentationFormProps) {
    const { register, control, setValue } = form;
    const [alertOpen, setAlertOpen] = useState(false);
    const { userInfo } = useAuthStore();
    const navigate = useNavigate();

    const handleQAChange = async (checked: boolean) => {
        if (!checked) {
            setValue('enable_qa_generation', false);
            return;
        }

        if (userInfo?.user_id) {
            try {
                const res = await getUserApiList(userInfo.user_id, { model_type: 'qa', status: 1 });
                // The request interceptor unwraps the response, so res is already the data object
                if (res.list && res.list.length > 0) {
                    setValue('enable_qa_generation', true);
                } else {
                    setAlertOpen(true);
                    setValue('enable_qa_generation', false);
                }
            } catch (error) {
                console.error("Failed to check QA model:", error);
                // Fallback: allow enabling but maybe warn? For now, we block if check fails or errors?
                // Let's safe fail to showing alert if we can't verify.
                setAlertOpen(true);
                setValue('enable_qa_generation', false);
            }
        }
    };

    return (
        <Card className="border-0 shadow-none bg-transparent">
            <CardHeader className="px-0 pt-0">
                <CardTitle className="flex items-center gap-2 text-lg">
                    <Settings2 className="w-5 h-5" />
                    Segmentation Settings
                </CardTitle>
            </CardHeader>
            <CardContent className="px-0 space-y-6">

                {/* Separators (Tag Input) */}
                <div className="space-y-2">
                    <Label>Separators</Label>
                    <Controller
                        name="separators"
                        control={control}
                        render={({ field }) => (
                            <div className="space-y-2">
                                <div className="flex flex-wrap gap-2 min-h-[38px] p-2 border border-gray-200 rounded-md bg-white">
                                    {field.value.map((sep: string, index: number) => (
                                        <span
                                            key={index}
                                            className="inline-flex items-center gap-1 px-2 py-1 text-sm bg-blue-100 text-blue-800 rounded-md"
                                        >
                                            <code className="font-mono">{sep.replace(/\n/g, '\\n').replace(/\t/g, '\\t')}</code>
                                            <button
                                                type="button"
                                                onClick={() => {
                                                    const newSeps = field.value.filter((_: string, i: number) => i !== index);
                                                    field.onChange(newSeps);
                                                }}
                                                className="text-blue-600 hover:text-blue-800"
                                            >
                                                ×
                                            </button>
                                        </span>
                                    ))}
                                </div>
                                <div className="relative">
                                    <div className="absolute left-3 top-2.5 text-gray-400">
                                        <AlignLeft size={16} />
                                    </div>
                                    <Input
                                        className="pl-9"
                                        placeholder="输入分隔符后按 Enter 添加 (如: \\n, ##, ---)"
                                        onKeyDown={(e) => {
                                            if (e.key === 'Enter') {
                                                e.preventDefault();
                                                const input = e.currentTarget;
                                                const value = input.value.trim()
                                                    .replace(/\\n/g, '\n')
                                                    .replace(/\\t/g, '\t');
                                                if (value && !field.value.includes(value)) {
                                                    field.onChange([...field.value, value]);
                                                    input.value = '';
                                                }
                                            }
                                        }}
                                    />
                                </div>
                            </div>
                        )}
                    />
                    <p className="text-xs text-gray-500">支持多个分隔符，按 Enter 逐个添加。</p>
                </div>

                {/* Max Length */}
                <CustomSlider
                    label="最大分块长度"
                    name="max_chunk_length"
                    control={control}
                    min={1}
                    max={4000}
                    suffix=" 字符"
                />

                {/* Overlap */}
                <CustomSlider
                    label="分块重叠"
                    name="chunk_overlap"
                    control={control}
                    min={0}
                    max={500}
                    suffix=" 字符"
                />

                <div className="space-y-4 pt-4 border-t border-gray-100">
                    {/* QA Generation Toggle */}
                    <Controller
                        name="enable_qa_generation"
                        control={control}
                        render={({ field }) => (
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <Label htmlFor="qa-gen" className="cursor-pointer">开启 QA 生成</Label>
                                    <TooltipProvider>
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <HelpCircle className="w-4 h-4 text-gray-400 cursor-help" />
                                            </TooltipTrigger>
                                            <TooltipContent className="max-w-[300px]">
                                                <p>开启 QA 生成会在后台利用大模型为每个分块生成问答对，虽然能显著提高检索精度，但会消耗大量 LLM Token，请按需开启。</p>
                                            </TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                </div>
                                <Switch
                                    id="qa-gen"
                                    checked={field.value}
                                    onCheckedChange={handleQAChange}
                                />
                            </div>
                        )}
                    />

                    <div className="flex items-center justify-between">
                        <Label className="cursor-pointer" htmlFor="clean-ws">清理空白字符</Label>
                        <input
                            id="clean-ws"
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
                            {...register('pre_clean_rule.clean_whitespace')}
                        />
                    </div>
                    <div className="flex items-center justify-between">
                        <Label className="cursor-pointer" htmlFor="rm-urls">移除 URL 和邮箱</Label>
                        <input
                            id="rm-urls"
                            type="checkbox"
                            className="w-4 h-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
                            {...register('pre_clean_rule.remove_urls_emails')}
                        />
                    </div>
                </div>

                <AlertDialog open={alertOpen} onOpenChange={setAlertOpen}>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>未配置 QA 模型</AlertDialogTitle>
                            <AlertDialogDescription>
                                开启 QA 生成需要配置对应的 QA 模型。检测到您尚未配置或未启用默认 QA 模型。
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                            <AlertDialogCancel>取消</AlertDialogCancel>
                            <AlertDialogAction onClick={() => navigate('/settings/provider')}>
                                去配置
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog>
            </CardContent>
        </Card>
    );
}
