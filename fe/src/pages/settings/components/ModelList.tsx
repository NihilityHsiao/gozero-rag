import { useState, useEffect } from 'react';
import { Plus, Trash2, MoreHorizontal, Check, X, RefreshCw } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';

import { listTenantLlmGrouped, deleteTenantLlm } from '@/api/llm';
import type { TenantLlmGroupByFactory, TenantLlmInfo } from '@/types';
import AddModelDialog from './AddModelDialog';

export default function ModelList() {
    const [loading, setLoading] = useState(false);
    const [groupedList, setGroupedList] = useState<TenantLlmGroupByFactory[]>([]);
    const [dialogOpen, setDialogOpen] = useState(false);
    const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
    const [deleteTarget, setDeleteTarget] = useState<TenantLlmInfo | null>(null);

    const fetchModels = async () => {
        setLoading(true);
        try {
            const res = await listTenantLlmGrouped();
            setGroupedList(res.list || []);
        } catch (error) {
            console.error(error);
            toast.error('获取模型列表失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchModels();
    }, []);

    const handleDeleteClick = (item: TenantLlmInfo) => {
        setDeleteTarget(item);
        setDeleteConfirmOpen(true);
    };

    const handleDeleteConfirm = async () => {
        if (!deleteTarget) return;

        try {
            await deleteTenantLlm(deleteTarget.id);
            toast.success('删除成功');
            fetchModels();
        } catch (error) {
            console.error(error);
            toast.error('删除失败');
        } finally {
            setDeleteConfirmOpen(false);
            setDeleteTarget(null);
        }
    };

    const getModelTypeBadgeColor = (type: string) => {
        const colors: Record<string, string> = {
            'LLM': 'bg-blue-100 text-blue-700 border-blue-200',
            'Embedding': 'bg-purple-100 text-purple-700 border-purple-200',
            'Rerank': 'bg-green-100 text-green-700 border-green-200',
            'ASR': 'bg-yellow-100 text-yellow-700 border-yellow-200',
            'TTS': 'bg-orange-100 text-orange-700 border-orange-200',
            'Image2Text': 'bg-pink-100 text-pink-700 border-pink-200',
            'Text2Image': 'bg-indigo-100 text-indigo-700 border-indigo-200',
            'Video': 'bg-red-100 text-red-700 border-red-200',
        };
        return colors[type] || 'bg-gray-100 text-gray-700 border-gray-200';
    };

    return (
        <div className="space-y-6">
            {/* 头部 */}
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-xl font-semibold tracking-tight">模型配置</h2>
                    <p className="text-sm text-gray-500">管理您的 LLM 模型接口配置</p>
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm" onClick={fetchModels} disabled={loading}>
                        <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        刷新
                    </Button>
                    <Button onClick={() => setDialogOpen(true)}>
                        <Plus className="mr-2 h-4 w-4" />
                        添加模型
                    </Button>
                </div>
            </div>

            {/* 模型列表 (按厂商分组) */}
            {loading ? (
                <div className="flex items-center justify-center h-48 text-gray-500">
                    加载中...
                </div>
            ) : groupedList.length === 0 ? (
                <Card className="border-dashed">
                    <CardContent className="flex flex-col items-center justify-center py-12">
                        <div className="text-4xl mb-4">🤖</div>
                        <p className="text-gray-500 text-center mb-4">
                            暂无模型配置<br />
                            点击上方【添加模型】开始配置
                        </p>
                    </CardContent>
                </Card>
            ) : (
                <div className="space-y-4">
                    {groupedList.map((group) => (
                        <Card key={group.llm_factory} className="overflow-hidden">
                            <CardHeader className="bg-gradient-to-r from-gray-50 to-white border-b py-4">
                                <div className="flex items-center justify-between">
                                    <div className="flex items-center gap-3">
                                        {group.factory_logo ? (
                                            <img
                                                src={group.factory_logo}
                                                alt={group.llm_factory}
                                                className="w-8 h-8 rounded"
                                            />
                                        ) : (
                                            <div className="w-8 h-8 rounded bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-sm">
                                                {group.llm_factory.charAt(0)}
                                            </div>
                                        )}
                                        <CardTitle className="text-base font-medium">
                                            {group.llm_factory}
                                        </CardTitle>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Badge variant="outline" className="text-xs text-green-600 border-green-200 bg-green-50">
                                            <Check className="w-3 h-3 mr-1" />
                                            API 已配置
                                        </Badge>
                                        {group.api_base && (
                                            <span className="text-xs text-gray-400">{group.api_base}</span>
                                        )}
                                    </div>
                                </div>
                            </CardHeader>
                            <CardContent className="p-0">
                                <div className="divide-y">
                                    {group.models.map((model) => (
                                        <div
                                            key={model.id}
                                            className="flex items-center justify-between px-4 py-3 hover:bg-gray-50 transition-colors"
                                        >
                                            <div className="flex items-center gap-4">
                                                <Badge
                                                    variant="outline"
                                                    className={`text-xs font-normal min-w-[80px] justify-center ${getModelTypeBadgeColor(model.model_type)}`}
                                                >
                                                    {model.model_type}
                                                </Badge>
                                                <span className="font-mono text-sm text-gray-700">
                                                    {model.llm_name}
                                                </span>
                                            </div>
                                            <div className="flex items-center gap-3">
                                                {model.status === 1 ? (
                                                    <span className="flex items-center text-xs text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                                                        <Check className="mr-1 h-3 w-3" /> 启用
                                                    </span>
                                                ) : (
                                                    <span className="flex items-center text-xs text-red-500 bg-red-50 px-2 py-0.5 rounded-full">
                                                        <X className="mr-1 h-3 w-3" /> 禁用
                                                    </span>
                                                )}
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button variant="ghost" className="h-8 w-8 p-0">
                                                            <span className="sr-only">打开菜单</span>
                                                            <MoreHorizontal className="h-4 w-4" />
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent align="end">
                                                        <DropdownMenuItem disabled>
                                                            编辑 (待实现)
                                                        </DropdownMenuItem>
                                                        <DropdownMenuSeparator />
                                                        <DropdownMenuItem
                                                            className="text-red-600 focus:text-red-600 cursor-pointer"
                                                            onClick={() => handleDeleteClick(model)}
                                                        >
                                                            <Trash2 className="mr-2 h-4 w-4" />
                                                            删除配置
                                                        </DropdownMenuItem>
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            {/* 添加模型对话框 */}
            <AddModelDialog
                open={dialogOpen}
                onOpenChange={setDialogOpen}
                onSuccess={fetchModels}
            />

            {/* 删除确认对话框 */}
            <AlertDialog open={deleteConfirmOpen} onOpenChange={setDeleteConfirmOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认删除</AlertDialogTitle>
                        <AlertDialogDescription>
                            确定要删除模型 <strong>{deleteTarget?.llm_name}</strong> 吗？此操作无法撤销。
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleDeleteConfirm}
                            className="bg-red-600 hover:bg-red-700"
                        >
                            删除
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
