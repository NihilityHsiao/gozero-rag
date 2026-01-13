import type { PreviewDocResult } from '@/types';
import { FileText, AlertCircle } from 'lucide-react';
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

interface ChunkPreviewProps {
    previewResults: PreviewDocResult[];
    loading?: boolean;
}

export default function ChunkPreview({ previewResults, loading }: ChunkPreviewProps) {
    // Note: Selection is handled in parent StepProcess. 
    // We only expect one result in the array for single-doc preview, or handle list if needed.
    // For now, based on requirement, we just show the first one (active one).
    const activeDoc = previewResults.length > 0 ? previewResults[0] : null;

    if (loading) {
        return (
            <div className="h-full flex flex-col items-center justify-center text-gray-400 gap-4">
                <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
                <p className="text-sm font-medium">生成预览中...</p>
            </div>
        );
    }

    if (!activeDoc) {
        return (
            <div className="h-full flex flex-col items-center justify-center text-gray-400 gap-4">
                <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center">
                    <FileText className="w-8 h-8 opacity-50" />
                </div>
                <p>点击“预览”查看结果</p>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full overflow-hidden">
            {/* Content Area */}
            <div className="flex-1 overflow-y-auto min-h-0 space-y-4 pr-2 pt-4">
                {activeDoc.error_msg && (
                    <Alert variant="destructive">
                        <AlertCircle className="h-4 w-4" />
                        <AlertTitle>错误</AlertTitle>
                        <AlertDescription>{activeDoc.error_msg}</AlertDescription>
                    </Alert>
                )}

                {activeDoc.chunks.map((chunk, i) => (
                    <div
                        key={i}
                        className="bg-white p-4 rounded-lg border border-gray-200 shadow-sm text-sm hover:border-blue-300 transition-colors group relative"
                    >
                        <div className="absolute top-3 right-3 text-xs text-gray-400 bg-gray-50 px-2 py-0.5 rounded">
                            {chunk.length} 字符
                        </div>
                        <div className="mb-2">
                            <span className="inline-flex items-center justify-center min-w-[24px] h-6 px-1.5 text-xs font-semibold text-blue-600 bg-blue-50 rounded">
                                #{chunk.index}
                            </span>
                        </div>
                        <p className="text-gray-700 leading-relaxed whitespace-pre-wrap font-mono text-xs break-words">
                            {chunk.content}
                        </p>
                    </div>
                ))}
            </div>
        </div>
    );
}
