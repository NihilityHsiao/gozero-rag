import { useEffect, useState } from 'react';
import { Plus, Trash2, MoreHorizontal, Check, X, Copy, Star } from 'lucide-react';
import { toast } from 'sonner';

import { Button } from '@/components/ui/button';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
    DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "@/components/ui/tooltip";

import { useAuthStore } from '@/store/useAuthStore';
import { getUserApiList, deleteUserApi, setUserModelDefault } from '@/api/user_model';
import type { UserApiInfo } from '@/types';
import AddModelDialog from './AddModelDialog';

export default function ModelList() {
    const { userInfo } = useAuthStore();
    const [loading, setLoading] = useState(false);
    const [list, setList] = useState<UserApiInfo[]>([]);
    const [dialogOpen, setDialogOpen] = useState(false);

    const fetchModels = async () => {
        if (!userInfo?.user_id) return;
        setLoading(true);
        try {
            const res = await getUserApiList(userInfo.user_id, { page: 1, page_size: 100 });
            setList(res.list || []);
        } catch (error) {
            console.error(error);
            toast.error('获取模型列表失败');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchModels();
    }, [userInfo?.user_id]);

    const handleDelete = async (id: number) => {
        if (!window.confirm('确定要删除这个模型配置吗？此操作无法撤销。')) return;

        try {
            await deleteUserApi(id);
            toast.success('删除成功');
            fetchModels();
        } catch (error) {
            console.error(error);
            toast.error('删除失败');
        }
    };

    const handleSetDefault = async (item: UserApiInfo) => {
        if (!userInfo?.user_id) return;
        try {
            await setUserModelDefault({
                user_id: userInfo.user_id,
                model_id: item.id,
                model_type: item.model_type
            });
            toast.success(`已将 ${item.config_name} 设为 ${item.model_type} 默认模型`);
            fetchModels();
        } catch (error) {
            console.error(error);
            toast.error('设置默认模型失败');
        }
    };

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text);
        toast.success('已复制到剪贴板');
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h2 className="text-xl font-semibold tracking-tight">模型配置</h2>
                    <p className="text-sm text-gray-500">管理您的 LLM 模型接口配置</p>
                </div>
                <Button onClick={() => setDialogOpen(true)}>
                    <Plus className="mr-2 h-4 w-4" />
                    添加模型
                </Button>
            </div>

            <div className="rounded-md border bg-white shadow-sm">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[200px]">配置名称</TableHead>
                            <TableHead className="w-[100px]">类型</TableHead>
                            <TableHead className="w-[180px]">模型 ID</TableHead>
                            <TableHead className="max-w-[300px]">Base URL</TableHead>
                            <TableHead className="w-[100px]">状态</TableHead>
                            <TableHead className="w-[80px] text-right">操作</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center text-gray-500">
                                    加载中...
                                </TableCell>
                            </TableRow>
                        ) : list.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center text-gray-500">
                                    暂无模型配置，请点击右上角添加
                                </TableCell>
                            </TableRow>
                        ) : (
                            list.map((item) => (
                                <TableRow key={item.id}>
                                    <TableCell className="font-medium">
                                        <div className="flex flex-col">
                                            <div className="flex items-center gap-2">
                                                <span className="truncate max-w-[180px]" title={item.config_name}>
                                                    {item.config_name}
                                                </span>
                                                {item.is_default === 1 && (
                                                    <Badge variant="outline" className="text-[10px] px-1 py-0 h-4 border-blue-200 text-blue-600 bg-blue-50 shrink-0">
                                                        默认
                                                    </Badge>
                                                )}
                                            </div>
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <Badge variant="secondary" className="capitalize font-normal text-xs">
                                            {item.model_type}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>
                                        <TooltipProvider>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <div className="max-w-[160px] truncate text-xs font-mono bg-gray-50 px-2 py-1 rounded border cursor-pointer hover:bg-gray-100 transition-colors">
                                                        {item.model_name}
                                                    </div>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p className="font-mono">{item.model_name}</p>
                                                    <p className="text-xs text-muted-foreground mt-1">点击复制 (未实现)</p>
                                                </TooltipContent>
                                            </Tooltip>
                                        </TooltipProvider>
                                    </TableCell>
                                    <TableCell>
                                        <TooltipProvider>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <div className="max-w-[280px] truncate text-xs font-mono text-gray-500">
                                                        {item.base_url}
                                                    </div>
                                                </TooltipTrigger>
                                                <TooltipContent>
                                                    <p className="font-mono">{item.base_url}</p>
                                                </TooltipContent>
                                            </Tooltip>
                                        </TooltipProvider>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex items-center gap-2">
                                            {item.status === 1 ? (
                                                <span className="flex items-center text-xs text-green-600 bg-green-50 px-2 py-0.5 rounded-full">
                                                    <Check className="mr-1 h-3 w-3" /> 启用
                                                </span>
                                            ) : (
                                                <span className="flex items-center text-xs text-red-500 bg-red-50 px-2 py-0.5 rounded-full">
                                                    <X className="mr-1 h-3 w-3" /> 禁用
                                                </span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <DropdownMenu>
                                            <DropdownMenuTrigger asChild>
                                                <Button variant="ghost" className="h-8 w-8 p-0">
                                                    <span className="sr-only">Open menu</span>
                                                    <MoreHorizontal className="h-4 w-4" />
                                                </Button>
                                            </DropdownMenuTrigger>
                                            <DropdownMenuContent align="end">
                                                {item.is_default !== 1 && (
                                                    <DropdownMenuItem onClick={() => handleSetDefault(item)}>
                                                        <Star className="mr-2 h-4 w-4" />
                                                        设为默认
                                                    </DropdownMenuItem>
                                                )}
                                                {item.is_default !== 1 && <DropdownMenuSeparator />}
                                                <DropdownMenuItem
                                                    onClick={() => copyToClipboard(item.api_key)}
                                                >
                                                    <Copy className="mr-2 h-4 w-4" />
                                                    复制 API Key
                                                </DropdownMenuItem>
                                                <DropdownMenuItem
                                                    className="text-red-600 focus:text-red-600 cursor-pointer"
                                                    onClick={() => handleDelete(item.id)}
                                                >
                                                    <Trash2 className="mr-2 h-4 w-4" />
                                                    删除配置
                                                </DropdownMenuItem>
                                            </DropdownMenuContent>
                                        </DropdownMenu>
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>

            <AddModelDialog
                open={dialogOpen}
                onOpenChange={setDialogOpen}
                onSuccess={fetchModels}
            />
        </div>
    );
}
