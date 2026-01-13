import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Eye, Loader2, ArrowLeft, ArrowRight } from 'lucide-react';
import type { SegmentationSettings, PreviewDocResult } from '@/types';
import SegmentationForm from './SegmentationForm';
import ChunkPreview from './ChunkPreview';
import { previewChunks, processFiles } from '@/api/dataset';
import { toast } from 'sonner';
import { useParams, useNavigate } from 'react-router-dom';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"

// Zod Schema
const segmentationSchema = z.object({
    separators: z.array(z.string()).min(1, "At least one separator is required"),
    max_chunk_length: z.number().min(1).max(4000),
    chunk_overlap: z.number().min(0).max(500),
    pre_clean_rule: z.object({
        clean_whitespace: z.boolean(),
        remove_urls_emails: z.boolean(),
    }),
    enable_qa_generation: z.boolean(),
});

interface StepProcessProps {
    files: { id: string; name: string }[];
    onBack: () => void;
}

export default function StepProcess({ files, onBack }: StepProcessProps) {
    const { id: knowledgeId } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const [previewLoading, setPreviewLoading] = useState(false);
    const [processLoading, setProcessLoading] = useState(false);
    const [previewResults, setPreviewResults] = useState<PreviewDocResult[]>([]);

    // Default to first file if available
    const [selectedDocId, setSelectedDocId] = useState<string>(files.length > 0 ? files[0].id : '');

    const form = useForm<SegmentationSettings>({
        resolver: zodResolver(segmentationSchema),
        defaultValues: {
            separators: ['\n\n', '\n', '。', '！', '？'],
            max_chunk_length: 500,
            chunk_overlap: 50,
            pre_clean_rule: {
                clean_whitespace: true,
                remove_urls_emails: false,
            },
            enable_qa_generation: false,
        }
    });

    // Reset preview when doc selection changes (optional, or auto-preview)
    useEffect(() => {
        setPreviewResults([]);
    }, [selectedDocId]);

    const handlePreview = async () => {
        if (!knowledgeId || !selectedDocId) {
            toast.error('请先选择一个文档');
            return;
        }
        const settings = form.getValues();
        setPreviewLoading(true);
        try {
            const res = await previewChunks(Number(knowledgeId), { doc_id: selectedDocId, settings });
            // Wrap in array for ChunkPreview if it expects multiple, but we only have one now
            setPreviewResults([res]);
        } catch (error) {
            console.error(error);
            toast.error('预览生成失败');
        } finally {
            setPreviewLoading(false);
        }
    };

    const handleProcess = async (data: SegmentationSettings) => {
        if (!knowledgeId) return;
        setProcessLoading(true);
        try {
            await processFiles(Number(knowledgeId), { file_ids: files.map(f => f.id), settings: data });
            toast.success('文档正在处理中');
            navigate(`/knowledge/${knowledgeId}/documents`);
        } catch (error) {
            console.error(error);
            toast.error('处理请求失败');
        } finally {
            setProcessLoading(false);
        }
    };

    return (
        <div className="flex h-[calc(100vh-200px)] gap-6">
            {/* Left: Settings */}
            <div className="w-1/3 flex flex-col h-full bg-white border border-gray-200 rounded-xl shadow-sm p-6 overflow-y-auto">
                <div className="flex-1">
                    <SegmentationForm form={form} />
                </div>

                <div className="pt-6 border-t border-gray-100 flex gap-3">
                    <Button
                        variant="outline"
                        className="flex-1"
                        onClick={handlePreview}
                        disabled={previewLoading || !selectedDocId}
                    >
                        {previewLoading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Eye className="w-4 h-4 mr-2" />}
                        预览
                    </Button>
                    <Button
                        variant="ghost"
                        onClick={() => {
                            form.reset();
                            setPreviewResults([]);
                        }}
                    >
                        重置
                    </Button>
                </div>
            </div>

            {/* Right: Preview */}
            <div className="w-2/3 h-full bg-gray-50 border border-gray-200 rounded-xl p-6 overflow-hidden flex flex-col gap-4">
                {/* Header Row: Document Selector */}
                <div className="flex items-center justify-between">
                    <div className="w-[300px] bg-white rounded-md">
                        <Select value={selectedDocId} onValueChange={setSelectedDocId}>
                            <SelectTrigger>
                                <SelectValue placeholder="选择要预览的文档" />
                            </SelectTrigger>
                            <SelectContent>
                                {files.map((file) => (
                                    <SelectItem key={file.id} value={file.id}>
                                        {file.name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                    {/* Chunk Count Display */}
                    {previewResults.length > 0 && (
                        <div className="text-sm text-gray-500">
                            {previewResults[0].total_chunks} 个切片
                        </div>
                    )}
                </div>

                <div className="flex-1 overflow-hidden">
                    <ChunkPreview previewResults={previewResults} loading={previewLoading} />
                </div>
            </div>

            {/* Floating Action Bar */}
            <div className="fixed bottom-0 left-[250px] right-0 bg-white border-t p-4 flex items-center justify-between px-8 z-50">
                <Button variant="ghost" onClick={onBack}>
                    <ArrowLeft className="w-4 h-4 mr-2" />
                    上一步
                </Button>
                <div className="text-gray-500 text-sm">
                    已选 {files.length} 个文件
                </div>
                <Button onClick={form.handleSubmit(handleProcess)} disabled={processLoading}>
                    {processLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                    保存并处理
                    <ArrowRight className="w-4 h-4 ml-2" />
                </Button>
            </div>
        </div>
    );
}
