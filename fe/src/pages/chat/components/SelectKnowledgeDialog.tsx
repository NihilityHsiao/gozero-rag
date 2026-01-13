import { useState, useEffect } from 'react';
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { ScrollArea } from "@/components/ui/scroll-area";
import { getKnowledgeList } from '@/api/knowledge';
import type { KnowledgeBaseInfo } from '@/types';
import { Database, Search } from 'lucide-react';
import { Input } from '@/components/ui/input';

interface SelectKnowledgeDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    currentSelection: number[]; // IDs
    onConfirm: (selected: KnowledgeBaseInfo[]) => void;
}

export default function SelectKnowledgeDialog({
    open,
    onOpenChange,
    currentSelection = [],
    onConfirm
}: SelectKnowledgeDialogProps) {
    const [list, setList] = useState<KnowledgeBaseInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [selectedIds, setSelectedIds] = useState<number[]>([]);
    const [search, setSearch] = useState('');

    useEffect(() => {
        if (open) {
            fetchList();
            setSelectedIds(currentSelection);
        }
    }, [open]);

    const fetchList = async () => {
        setLoading(true);
        try {
            const res = await getKnowledgeList({ page: 1, page_size: 100 });
            setList(res.list || []);
        } catch (e) {
            console.error(e);
        } finally {
            setLoading(false);
        }
    };

    const handleToggle = (id: number) => {
        setSelectedIds(prev => {
            if (prev.includes(id)) {
                return prev.filter(x => x !== id);
            } else {
                return [...prev, id];
            }
        });
    };

    const handleConfirm = () => {
        const selectedItems = list.filter(item => selectedIds.includes(item.id));
        onConfirm(selectedItems);
        onOpenChange(false);
    };

    const filteredList = list.filter(item =>
        item.name.toLowerCase().includes(search.toLowerCase())
    );

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>添加知识库 (Context)</DialogTitle>
                </DialogHeader>

                <div className="space-y-4 py-2">
                    <div className="relative">
                        <Search className="absolute left-2 top-2.5 h-4 w-4 text-gray-400" />
                        <Input
                            placeholder="搜索知识库..."
                            className="pl-8"
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                        />
                    </div>

                    <div className="border rounded-md h-[300px] overflow-hidden">
                        <ScrollArea className="h-full">
                            <div className="p-1">
                                {loading ? (
                                    <div className="p-4 text-center text-sm text-gray-500">加载中...</div>
                                ) : filteredList.length === 0 ? (
                                    <div className="p-4 text-center text-sm text-gray-500">未找到相关知识库</div>
                                ) : (
                                    filteredList.map(item => (
                                        <div
                                            key={item.id}
                                            className="flex items-start space-x-3 p-3 hover:bg-gray-50 rounded-md cursor-pointer transition-colors"
                                            onClick={() => handleToggle(item.id)}
                                        >
                                            <Checkbox
                                                checked={selectedIds.includes(item.id)}
                                                onCheckedChange={() => handleToggle(item.id)}
                                            />
                                            <div className="flex-1 space-y-1">
                                                <div className="flex items-center gap-2">
                                                    <Database className="w-4 h-4 text-blue-500" />
                                                    <span className="text-sm font-medium">{item.name}</span>
                                                </div>
                                                {item.description && (
                                                    <p className="text-xs text-gray-500 line-clamp-1">{item.description}</p>
                                                )}
                                            </div>
                                        </div>
                                    ))
                                )}
                            </div>
                        </ScrollArea>
                    </div>

                    <div className="flex items-center justify-between">
                        <div className="text-xs text-gray-500">
                            已选: {selectedIds.length}
                        </div>
                        <div className="flex gap-2">
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 text-xs"
                                onClick={() => setSelectedIds(filteredList.map(i => i.id))}
                            >
                                全选
                            </Button>
                            <Button
                                variant="ghost"
                                size="sm"
                                className="h-6 text-xs"
                                onClick={() => {
                                    // Also keep selected items that are NOT in the current filtered list?
                                    // "Invert" usually applies to the current view.
                                    // Let's implement invert for current filtered list only, but preserve others if needed.
                                    // Simpler approach: Invert the visible list status.

                                    const visibleIds = new Set(filteredList.map(i => i.id));
                                    const invertedIds = filteredList
                                        .filter(item => !selectedIds.includes(item.id))
                                        .map(item => item.id);

                                    // Merge: (Original - Visible) + InvertedVisible
                                    const otherIds = selectedIds.filter(id => !visibleIds.has(id));
                                    setSelectedIds([...otherIds, ...invertedIds]);
                                }}
                            >
                                反选
                            </Button>
                        </div>
                    </div>
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>取消</Button>
                    <Button onClick={handleConfirm}>确认</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
