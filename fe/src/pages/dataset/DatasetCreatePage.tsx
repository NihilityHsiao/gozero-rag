import { useNavigate, useParams } from 'react-router-dom';
import { useDatasetCreate } from './hooks/useDatasetCreate';
import StepUpload from './components/StepUpload';
import StepProcess from './components/StepProcess';
import { Button } from '@/components/ui/button';
import { ChevronLeft } from 'lucide-react';
import { cn } from '@/lib/utils';

export default function DatasetCreatePage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const {
        currentStep,
        files,
        setFiles,
        nextStep,
        prevStep
    } = useDatasetCreate();

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">
            {/* Header */}
            <header className="h-16 bg-white border-b border-gray-200 px-6 flex items-center justify-between sticky top-0 z-40">
                <div className="flex items-center gap-4">
                    <Button variant="ghost" size="icon" onClick={() => navigate(`/knowledge/${id}/documents`)}>
                        <ChevronLeft className="w-5 h-5" />
                    </Button>
                    <div className="flex flex-col">
                        <h1 className="text-base font-semibold text-gray-900">添加文件</h1>
                        <span className="text-xs text-gray-500">知识库 / 上传文档</span>
                    </div>
                </div>

                {/* Stepper Indicator */}
                <div className="flex items-center gap-2">
                    <div className={cn(
                        "flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium transition-colors",
                        currentStep === 1 ? "bg-blue-600 text-white" : "bg-green-500 text-white"
                    )}>
                        {currentStep > 1 ? "✓" : "1"}
                    </div>
                    <span className="text-gray-300">——</span>
                    <div className={cn(
                        "flex items-center justify-center w-8 h-8 rounded-full text-sm font-medium transition-colors",
                        currentStep === 2 ? "bg-blue-600 text-white" : "bg-gray-200 text-gray-500"
                    )}>
                        2
                    </div>
                </div>

                <div className="w-[100px]" /> {/* Spacer for centering */}
            </header>

            {/* Content */}
            <main className="flex-1 p-8 max-w-5xl mx-auto w-full mb-16">
                {currentStep === 1 && (
                    <div className="max-w-3xl mx-auto">
                        <div className="mb-8">
                            <h2 className="text-2xl font-bold text-gray-900">上传文件</h2>
                            <p className="text-gray-500 mt-2">上传文档至知识库。支持格式：PDF, TXT, MD, DOCX。</p>
                        </div>
                        <StepUpload onNext={nextStep} setFiles={setFiles} />
                    </div>
                )}

                {currentStep === 2 && (
                    <StepProcess files={files} onBack={prevStep} />
                )}
            </main>
        </div>
    );
}
