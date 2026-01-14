import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Eye, Loader2, ChevronLeft, Save } from 'lucide-react';
import type { SegmentationSettings, PreviewDocResult } from '@/types';
import SegmentationForm from './components/SegmentationForm';
import ChunkPreview from './components/ChunkPreview';
import { previewChunks, saveChunkSettings } from '@/api/dataset';
import { toast } from 'sonner';
import { useParams, useNavigate } from 'react-router-dom';

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

// 默认值
const defaultSettings: SegmentationSettings = {
    separators: ['\n\n', '\n', '。', '！', '？'],
    max_chunk_length: 500,
    chunk_overlap: 50,
    pre_clean_rule: {
        clean_whitespace: true,
        remove_urls_emails: false,
    },
    enable_qa_generation: false,
};

export default function DocumentEditPage() {
    const { id: knowledgeId, doc_id: docId } = useParams<{ id: string; doc_id: string }>();
    const navigate = useNavigate();
    const [previewLoading, setPreviewLoading] = useState(false);
    const [saveLoading, setSaveLoading] = useState(false);
    const [previewResults, setPreviewResults] = useState<PreviewDocResult[]>([]);

    // TODO: 如果后端有获取单个文档详情的 API，可以在此加载 parser_config
    // 目前假设用户从 DocumentList 跳转过来，暂时使用默认值
    // 可以通过 location state 传递 doc 信息

    const form = useForm<SegmentationSettings>({
        resolver: zodResolver(segmentationSchema),
        defaultValues: defaultSettings
    });

    const handlePreview = async () => {
        if (!knowledgeId || !docId) {
            toast.error('缺少文档信息');
            return;
        }
        const settings = form.getValues();
        setPreviewLoading(true);
        try {
            const res = await previewChunks(knowledgeId, { doc_id: docId, settings });
            setPreviewResults([res]);
        } catch (error) {
            console.error(error);
            toast.error('预览生成失败');
        } finally {
            setPreviewLoading(false);
        }
    };

    const handleSave = async (data: SegmentationSettings) => {
        if (!knowledgeId || !docId) return;
        setSaveLoading(true);
        try {
            await saveChunkSettings(knowledgeId, { file_ids: [docId], settings: data });
            toast.success('分段配置已保存');
            navigate(`/knowledge/${knowledgeId}/documents`);
        } catch (error) {
            console.error(error);
            toast.error('保存失败');
        } finally {
            setSaveLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">
            {/* Header */}
            <header className="h-16 bg-white border-b border-gray-200 px-6 flex items-center justify-between sticky top-0 z-40">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="icon" onClick={() => navigate(`/knowledge/${knowledgeId}/documents`)}>
                        <ChevronLeft className="w-5 h-5" />
                    </Button>
                    <div className="flex flex-col">
                        <h1 className="text-base font-semibold text-gray-900">编辑文档配置</h1>
                        <span className="text-xs text-gray-500">文档 ID: {docId}</span>
                    </div>
                </div>
            </header>

            {/* Content */}
            <main className="flex-1 p-8 max-w-5xl mx-auto w-full mb-16">
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
                                disabled={previewLoading}
                            >
                                {previewLoading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Eye className="w-4 h-4 mr-2" />}
                                预览
                            </Button>
                            <Button
                                variant="ghost"
                                onClick={() => {
                                    form.reset(defaultSettings);
                                    setPreviewResults([]);
                                }}
                            >
                                重置
                            </Button>
                        </div>
                    </div>

                    {/* Right: Preview */}
                    <div className="w-2/3 h-full bg-gray-50 border border-gray-200 rounded-xl p-6 overflow-hidden flex flex-col gap-4">
                        <div className="flex items-center justify-between">
                            <div className="text-sm font-medium text-gray-700">预览</div>
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
                </div>

                {/* Floating Action Bar */}
                <div className="fixed bottom-0 left-[250px] right-0 bg-white border-t p-4 flex items-center justify-between px-8 z-50">
                    <Button variant="ghost" onClick={() => navigate(`/knowledge/${knowledgeId}/documents`)}>
                        <ChevronLeft className="w-4 h-4 mr-2" />
                        返回文档列表
                    </Button>
                    <Button onClick={form.handleSubmit(handleSave)} disabled={saveLoading}>
                        {saveLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                        <Save className="w-4 h-4 mr-2" />
                        保存配置
                    </Button>
                </div>
            </main>
        </div>
    );
}
