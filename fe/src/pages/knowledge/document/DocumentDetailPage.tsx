import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { ArrowLeft, Search, Trash2 } from 'lucide-react';
import { getDocDetail, getDocumentChunks } from '@/api/document';
import type { KnowledgeDocumentInfo, KnowledgeDocumentChunkInfo } from '@/types';
import DocInfoSidebar from './DocInfoSidebar';
import ChunkList from './ChunkList';
import EditChunkDrawer from './EditChunkDrawer';
import { toast } from 'sonner';

const DocumentDetailPage: React.FC = () => {
    const { id: kbId, docId } = useParams<{ id: string; docId: string }>();
    const navigate = useNavigate();

    const [doc, setDoc] = useState<KnowledgeDocumentInfo | undefined>(undefined);
    const [chunks, setChunks] = useState<KnowledgeDocumentChunkInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [, setLoadingDoc] = useState(false);
    const [total, setTotal] = useState(0);

    // Filters
    const [page, setPage] = useState(1);
    const [keyword, setKeyword] = useState('');
    const [searchTerm, setSearchTerm] = useState('');

    // Selection
    const [selectedIds, setSelectedIds] = useState<string[]>([]);

    // Drawer
    const [isDrawerOpen, setIsDrawerOpen] = useState(false);
    const [editingChunk, setEditingChunk] = useState<KnowledgeDocumentChunkInfo | null>(null);

    useEffect(() => {
        if (kbId && docId) {
            fetchDocDetail();
            fetchChunks();
        }
    }, [kbId, docId]);

    useEffect(() => {
        fetchChunks();
    }, [page, searchTerm]);

    const fetchDocDetail = async () => {
        if (!kbId || !docId) return;
        setLoadingDoc(true);
        try {
            const res = await getDocDetail(docId);
            setDoc(res);
        } catch (error) {
            toast.error("加载文档详情失败");
        } finally {
            setLoadingDoc(false);
        }
    };

    const fetchChunks = async () => {
        if (!kbId || !docId) return;
        setLoading(true);
        try {
            const res = await getDocumentChunks({
                knowledge_base_id: kbId,
                document_id: docId,
                page,
                page_size: 20,
                keyword: searchTerm
            });
            setChunks(res.list || []);
            setTotal(res.total || 0);
        } catch (error) {
            toast.error("加载切片失败");
        } finally {
            setLoading(false);
        }
    };

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        setPage(1);
        setSearchTerm(keyword);
    };

    const handleSelectChunk = (id: string, checked: boolean) => {
        if (checked) {
            setSelectedIds(prev => [...prev, id]);
        } else {
            setSelectedIds(prev => prev.filter(item => item !== id));
        }
    };

    const handleSelectAll = (checked: boolean) => {
        if (checked) {
            setSelectedIds(chunks.map(c => c.id));
        } else {
            setSelectedIds([]);
        }
    };

    const handleBulkDelete = () => {
        toast.info(`删除 ${selectedIds.length} 个切片的功能即将推出。`);
    };

    const handleEditChunk = (chunk: KnowledgeDocumentChunkInfo) => {
        setEditingChunk(chunk);
        setIsDrawerOpen(true);
    };

    const handleSaveChunk = async (id: string, newText: string) => {
        // TODO: Implement backend API for updating chunk
        console.log("Updating chunk", id, newText);
        // For now we just update local state to reflect change immediately in UI
        setChunks(prev => prev.map(c => c.id === id ? { ...c, chunk_text: newText } : c));
    };

    return (
        <div className="flex h-screen bg-gray-50 overflow-hidden">
            {/* Main Content Area */}
            <div className="flex-1 flex flex-col min-w-0">
                {/* Header */}
                <header className="h-16 border-b bg-white px-6 flex items-center justify-between shrink-0">
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
                            <ArrowLeft className="h-5 w-5" />
                        </Button>
                        <div>
                            <h1 className="text-lg font-semibold text-gray-900 truncate max-w-md">
                                {doc?.doc_name || '加载中...'}
                            </h1>
                            <div className="text-xs text-gray-500 flex gap-2">
                                <span>{total} 个切片</span>
                            </div>
                        </div>
                    </div>

                    <form onSubmit={handleSearch} className="flex gap-2 w-96">
                        <div className="relative flex-1">
                            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-gray-500" />
                            <Input
                                placeholder="搜索切片..."
                                className="pl-9"
                                value={keyword}
                                onChange={(e) => setKeyword(e.target.value)}
                            />
                        </div>
                    </form>
                </header>

                {/* Content & List */}
                <main className="flex-1 overflow-y-auto p-6">
                    <div className="max-w-4xl mx-auto space-y-4">
                        {/* Bulk Actions Bar */}
                        {selectedIds.length > 0 && (
                            <div className="sticky top-0 z-20 bg-white border rounded-lg p-2 shadow-sm flex items-center justify-between animate-in fade-in slide-in-from-top-2">
                                <div className="flex items-center gap-4 px-2">
                                    <span className="text-sm font-medium">已选 {selectedIds.length} 项</span>
                                    <Button variant="ghost" size="sm" onClick={() => handleSelectAll(false)}>取消选择</Button>
                                </div>
                                <div className="flex gap-2">
                                    <Button variant="destructive" size="sm" onClick={handleBulkDelete}>
                                        <Trash2 className="h-4 w-4 mr-2" /> 删除
                                    </Button>
                                </div>
                            </div>
                        )}

                        <ChunkList
                            chunks={chunks}
                            loading={loading}
                            selectedIds={selectedIds}
                            total={total}
                            page={page}
                            pageSize={20}
                            onPageChange={setPage}
                            onSelectChunk={handleSelectChunk}
                            onEdit={handleEditChunk}
                        />
                    </div>
                </main>
            </div>

            {/* Right Sidebar - Document Info */}
            <DocInfoSidebar doc={doc} />

            {/* Edit Drawer */}
            <EditChunkDrawer
                chunk={editingChunk}
                isOpen={isDrawerOpen}
                onClose={() => setIsDrawerOpen(false)}
                onSave={handleSaveChunk}
            />
        </div>
    );
};

export default DocumentDetailPage;
