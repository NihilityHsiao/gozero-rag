import React from 'react';
import type { KnowledgeDocumentInfo } from '@/types';

interface DocInfoSidebarProps {
    doc?: KnowledgeDocumentInfo;
}

const DocInfoSidebar: React.FC<DocInfoSidebarProps> = ({ doc }) => {
    if (!doc) {
        return <div className="p-4 text-sm text-gray-500">加载文档信息...</div>;
    }

    const formatSize = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <div className="w-80 border-l bg-gray-50/50 h-full flex flex-col">
            <div className="p-4 font-medium border-b">文档详情</div>
            <div className="flex-1 overflow-y-auto">
                <div className="p-4 space-y-6">
                    {/* Basic Info */}
                    <div className="space-y-3">
                        <h4 className="text-sm font-semibold text-gray-900">基本信息</h4>
                        <div className="space-y-2 text-sm">
                            <div className="flex justify-between">
                                <span className="text-gray-500">名称</span>
                                <span className="font-medium truncate max-w-[150px]" title={doc.doc_name}>{doc.doc_name}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-500">ID</span>
                                <span className="font-mono text-xs text-gray-600 truncate max-w-[150px]" title={String(doc.id)}>{doc.id}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-500">类型</span>
                                <span className="uppercase">{doc.doc_type}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-500">大小</span>
                                <span>{formatSize(doc.doc_size)}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-500">创建时间</span>
                                <span className="text-xs">{doc.created_at}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-500">更新时间</span>
                                <span className="text-xs">{doc.updated_at}</span>
                            </div>
                        </div>
                    </div>

                    <hr className="border-gray-200" />

                    {/* Parser Config */}
                    <div className="space-y-3">
                        <h4 className="text-sm font-semibold text-gray-900">解析配置</h4>
                        {doc.parser_config ? (
                            <div className="space-y-2 text-sm">
                                <div className="flex justify-between">
                                    <span className="text-gray-500">最大分块长度</span>
                                    <span>{doc.parser_config.max_chunk_length}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-gray-500">重叠</span>
                                    <span>{doc.parser_config.chunk_overlap}</span>
                                </div>
                                <div className="flex justify-between">
                                    <span className="text-gray-500">QA 生成</span>
                                    <span>{doc.parser_config.enable_qa_generation ? '已开启' : '已禁用'}</span>
                                </div>
                                <div className="flex flex-col gap-1">
                                    <span className="text-gray-500">分隔符</span>
                                    <div className="flex flex-wrap gap-1">
                                        {doc.parser_config.separators.map((sep, idx) => (
                                            <code key={idx} className="bg-gray-200 px-1 rounded text-xs">
                                                {sep === '\n' ? '\\n' : sep}
                                            </code>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className="text-sm text-gray-400">无解析配置</div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default DocInfoSidebar;
