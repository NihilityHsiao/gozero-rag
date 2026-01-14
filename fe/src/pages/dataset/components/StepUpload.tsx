import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { Upload, FileText, X, CheckCircle, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { uploadFiles } from '@/api/dataset';
import { toast } from 'sonner';
import { useParams } from 'react-router-dom';



interface StepUploadProps {
    onNext: () => void;
    setFiles: (files: { id: string; name: string }[]) => void;
}

interface UploadFile extends File {
    preview?: string;
}

export default function StepUpload({ onNext, setFiles }: StepUploadProps) {
    const { id: knowledgeId } = useParams<{ id: string }>();
    const [files, setLocalFiles] = useState<UploadFile[]>([]);
    const [uploading, setUploading] = useState(false);
    const [progress, setProgress] = useState(0);
    const [uploadedIds, setUploadedIds] = useState<string[]>([]);

    const onDrop = useCallback((acceptedFiles: File[]) => {
        setLocalFiles(prev => [...prev, ...acceptedFiles]);
    }, []);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,
        accept: {
            'application/pdf': ['.pdf'],
            'text/plain': ['.txt', '.md'],
            'application/msword': ['.doc'],
            'application/vnd.openxmlformats-officedocument.wordprocessingml.document': ['.docx']
        }
    });

    const removeFile = (index: number) => {
        setLocalFiles(prev => prev.filter((_, i) => i !== index));
    };

    const handleUpload = async () => {
        if (files.length === 0 || !knowledgeId) return;

        setUploading(true);
        setProgress(0);
        try {
            const resp: any = await uploadFiles(knowledgeId, files, (p: number) => setProgress(p));

            // axios 拦截器已经解包了 data，所以直接访问 resp.file_ids
            if (!resp || !resp.file_ids) {
                throw new Error("服务器响应无效");
            }

            const ids = resp.file_ids;
            setUploadedIds(ids);

            // Map IDs to file names (assuming order is preserved)
            const uploadedFiles = ids.map((id: string, index: number) => ({
                id,
                name: files[index]?.name || `Document ${index + 1}`
            }));
            setFiles(uploadedFiles);

            toast.success('文件上传成功，已加入解析队列');
        } catch (error) {
            console.error(error);
            toast.error('文件上传失败');
        } finally {
            setUploading(false);
        }
    };

    return (
        <div className="space-y-6">
            <div
                {...getRootProps()}
                className={cn(
                    "border-2 border-dashed rounded-xl p-10 text-center cursor-pointer transition-colors bg-white hover:bg-gray-50",
                    isDragActive ? "border-blue-500 bg-blue-50" : "border-gray-200"
                )}
            >
                <input {...getInputProps()} />
                <div className="flex flex-col items-center gap-2 text-gray-500">
                    <div className="w-12 h-12 rounded-full bg-gray-100 flex items-center justify-center mb-2">
                        <Upload className="w-6 h-6 text-gray-600" />
                    </div>
                    <p className="font-medium text-gray-900">点击 or 拖拽文件上传</p>
                    <p className="text-xs text-gray-400">PDF, MD, TXT, DOCX (Max 15MB)</p>
                    <p className="text-xs text-gray-400 mt-1">新文件将使用当前知识库的默认解析规则进行处理</p>
                </div>
            </div>


            {files.length > 0 && (
                <div className="space-y-3">
                    <h3 className="text-sm font-medium text-gray-700">已选文件 ({files.length})</h3>
                    <div className="max-h-[400px] overflow-y-auto pr-2 custom-scrollbar border border-gray-100 rounded-lg">
                        <div className="grid gap-3 p-2">
                            {files.map((file, i) => (
                                <div key={i} className="flex items-center justify-between p-3 bg-white border border-gray-100 rounded-lg shadow-sm">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 flex items-center justify-center bg-blue-50 rounded-lg text-blue-600">
                                            <FileText size={16} />
                                        </div>
                                        <div className="flex flex-col">
                                            <span className="text-sm font-medium text-gray-700 truncate max-w-[300px]">{file.name}</span>
                                            <span className="text-xs text-gray-400">{(file.size / 1024).toFixed(1)} KB</span>
                                        </div>
                                    </div>
                                    {!uploadedIds.length && (
                                        <button onClick={() => removeFile(i)} className="text-gray-400 hover:text-red-500">
                                            <X size={16} />
                                        </button>
                                    )}
                                    {uploadedIds.length > 0 && (
                                        <div className="text-green-500 flex items-center gap-1 text-xs">
                                            <CheckCircle size={14} /> 已上传
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            {uploading && (
                <div className="space-y-1">
                    <div className="flex justify-between text-xs text-gray-500">
                        <span>上传中...</span>
                        <span>{progress}%</span>
                    </div>
                    <div className="h-2 w-full bg-gray-100 rounded-full overflow-hidden">
                        <div
                            className="h-full bg-blue-600 transition-all duration-300"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                </div>
            )}

            <div className="flex justify-end sticky bottom-0 bg-gray-50 z-20 py-4 mt-4 border-t border-gray-100">
                {!uploadedIds.length ? (
                    <Button
                        onClick={handleUpload}
                        disabled={files.length === 0 || uploading}
                        className="w-[120px]"
                    >
                        {uploading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                        {uploading ? '上传中' : '开始上传'}
                    </Button>
                ) : (
                    <Button onClick={onNext} className="w-[120px]">
                        下一步
                    </Button>
                )}
            </div>
        </div>
    );
}
