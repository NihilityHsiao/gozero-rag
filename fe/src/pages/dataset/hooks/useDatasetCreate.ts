import { useState } from 'react';
import type { SegmentationSettings } from '@/types';

export const useDatasetCreate = () => {
    const [currentStep, setCurrentStep] = useState(1);
    const [files, setFiles] = useState<{ id: string; name: string }[]>([]);

    // Default settings
    const [settings, setSettings] = useState<SegmentationSettings>({
        separators: ['\n\n', '\n', '。', '！', '？'],
        max_chunk_length: 500,
        chunk_overlap: 50,
        pre_clean_rule: {
            clean_whitespace: true,
            remove_urls_emails: false,
        },
        enable_qa_generation: false,
    });

    const nextStep = () => setCurrentStep(prev => prev + 1);
    const prevStep = () => setCurrentStep(prev => prev - 1);

    return {
        currentStep,
        setCurrentStep,
        nextStep,
        prevStep,
        files,
        setFiles,
        settings,
        setSettings
    };
};
