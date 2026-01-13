import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getDocumentList } from '@/api/document';
import { clearDatasetDocuments } from '@/api/dataset';
import type { KnowledgeDocumentInfo, DocStatus } from '@/types';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  FileText,
  Search,
  Plus,
  MoreHorizontal,
  AlertCircle,
  File,
  Loader2,
  Trash2
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { toast } from 'sonner';

export default function DocumentList() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [list, setList] = useState<KnowledgeDocumentInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [clearDialogOpen, setClearDialogOpen] = useState(false);
  const [clearing, setClearing] = useState(false);

  const fetchDocs = async () => {
    if (!id) return;
    setLoading(true);
    try {
      // Note: Backend currently doesn't support keyword search in this API, 
      // but we prepare the UI for it.
      const res = await getDocumentList(Number(id), {
        page,
        page_size: 10,
      });
      setList(res.list || []);
      setTotal(res.total);
    } catch (error) {
      console.error('Failed to fetch documents', error);
      toast.error('Failed to load documents');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDocs();
  }, [id, page]);

  const handleClearDocuments = async () => {
    if (!id) return;
    setClearing(true);
    try {
      await clearDatasetDocuments(Number(id));
      toast.success('文档已清空');
      setClearDialogOpen(false);
      setPage(1);
      fetchDocs();
    } catch (error) {
      console.error(error);
      toast.error('清空文档失败');
    } finally {
      setClearing(false);
    }
  };

  // Format bytes to readable string
  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  // Status Badge Helper
  const getStatusBadge = (status: DocStatus, errMsg?: string) => {
    const styles = {
      enable: 'bg-green-100 text-green-700 hover:bg-green-100 border-green-200',
      indexing: 'bg-blue-100 text-blue-700 hover:bg-blue-100 border-blue-200',
      pending: 'bg-gray-100 text-gray-700 hover:bg-gray-100 border-gray-200',
      disable: 'bg-gray-50 text-gray-400 hover:bg-gray-50 border-gray-200',
      fail: 'bg-red-50 text-red-600 hover:bg-red-50 border-red-200 cursor-pointer',
    };

    const statusMap: Record<string, string> = {
      enable: '启用',
      indexing: '索引中',
      pending: '等待中',
      disable: '禁用',
      fail: '失败',
    }

    const label = statusMap[status] || status;

    if (status === 'fail' && errMsg) {
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Badge variant="outline" className={cn("font-normal", styles[status])}>
                <AlertCircle size={12} className="mr-1" />
                Error
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs max-w-xs">{errMsg}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      );
    }

    if (status === 'indexing') {
      return (
        <Badge variant="outline" className={cn("font-normal", styles[status])}>
          <Loader2 size={12} className="mr-1 animate-spin" />
          索引中
        </Badge>
      )
    }

    return (
      <Badge variant="outline" className={cn("font-normal", styles[status])}>
        {label}
      </Badge>
    );
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">文档列表</h1>
          <p className="text-sm text-gray-500 mt-1">管理和组织知识库文档。</p>
        </div>
        <Button
          className="bg-blue-600 hover:bg-blue-700"
          onClick={() => navigate(`/knowledge/${id}/dataset/create`)}
        >
          <Plus size={16} className="mr-2" />
          添加文件
        </Button>
      </div>

      <AlertDialog open={clearDialogOpen} onOpenChange={setClearDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确定要清空所有文档吗？</AlertDialogTitle>
            <AlertDialogDescription>
              此操作将永久删除该知识库下所有非索引中 (Indexing) 状态的文档。该操作不可撤销，请谨慎操作。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={clearing}>取消</AlertDialogCancel>
            <AlertDialogAction
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
              onClick={(e) => {
                e.preventDefault();
                handleClearDocuments();
              }}
              disabled={clearing}
            >
              {clearing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              确认清空
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Toolbar */}
      <div className="flex items-center gap-4 bg-white p-4 rounded-xl border border-gray-200 shadow-sm">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={16} />
          <Input
            placeholder="搜索文档..."
            className="pl-9 bg-gray-50 border-gray-200 focus-visible:ring-blue-600/20"
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
          />
        </div>
        <Button
          variant="outline"
          className="text-gray-600 hover:text-red-600 hover:border-red-200 hover:bg-red-50"
          onClick={() => setClearDialogOpen(true)}
        >
          <Trash2 size={16} className="mr-2" />
          清空文档
        </Button>
      </div>

      {/* Table */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <Table>
          <TableHeader className="bg-gray-50/50">
            <TableRow>
              <TableHead className="w-[40%]">名称</TableHead>
              <TableHead>大小</TableHead>
              <TableHead>切片数</TableHead>
              <TableHead>上传时间</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="w-[100px]">启用</TableHead>
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  <TableCell><div className="h-4 bg-gray-100 rounded w-3/4 animate-pulse"></div></TableCell>
                  <TableCell><div className="h-4 bg-gray-100 rounded w-1/2 animate-pulse"></div></TableCell>
                  <TableCell><div className="h-4 bg-gray-100 rounded w-1/2 animate-pulse"></div></TableCell>
                  <TableCell><div className="h-4 bg-gray-100 rounded w-3/4 animate-pulse"></div></TableCell>
                  <TableCell><div className="h-4 bg-gray-100 rounded w-1/2 animate-pulse"></div></TableCell>
                  <TableCell><div className="h-4 bg-gray-100 rounded w-full animate-pulse"></div></TableCell>
                  <TableCell></TableCell>
                </TableRow>
              ))
            ) : list.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="h-32 text-center text-gray-500">
                  暂无文档
                </TableCell>
              </TableRow>
            ) : (
              list.map((doc) => (
                <TableRow
                  key={doc.id}
                  className={cn(
                    "hover:bg-gray-50/50 cursor-pointer",
                    doc.status === 'disable' ? "hover:bg-blue-50/50" : "hover:bg-gray-100/50"
                  )}
                  onClick={() => {
                    if (doc.status === 'disable') {
                      navigate(`/knowledge/${id}/document/${doc.id}/edit`);
                    } else {
                      navigate(`/knowledge/${id}/document/${doc.id}`);
                    }
                  }}
                >
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded bg-gray-100 flex items-center justify-center text-gray-500 flex-shrink-0">
                        {doc.doc_type === 'pdf' ? <FileText size={16} /> : <File size={16} />}
                      </div>
                      <div className="flex flex-col">
                        <span className="font-medium text-gray-900 truncate max-w-[200px] lg:max-w-[300px]" title={doc.doc_name}>
                          {doc.doc_name}
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="text-gray-500 font-mono text-xs">
                    {formatSize(doc.doc_size)}
                  </TableCell>
                  <TableCell className="text-gray-500">
                    <Badge variant="secondary" className="font-normal bg-gray-100 text-gray-600 hover:bg-gray-100">
                      {doc.chunk_count} 个切片
                    </Badge>
                  </TableCell>
                  <TableCell className="text-gray-500 text-sm">
                    {doc.created_at}
                  </TableCell>
                  <TableCell>
                    {getStatusBadge(doc.status, doc.err_msg)}
                  </TableCell>
                  <TableCell>
                    <Switch checked={doc.status === 'enable'} />
                  </TableCell>
                  <TableCell>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-gray-500 hover:text-gray-900">
                      <MoreHorizontal size={16} />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        {/* Pagination (Simple Implementation) */}
        {total > 10 && (
          <div className="flex items-center justify-end p-4 border-t border-gray-200 gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page === 1}
              onClick={() => setPage(p => Math.max(1, p - 1))}
            >
              上一页
            </Button>
            <span className="text-sm text-gray-500">
              第 {page} 页 / 共 {Math.ceil(total / 10)} 页
            </span>
            <Button
              variant="outline"
              size="sm"
              disabled={page * 10 >= total}
              onClick={() => setPage(p => p + 1)}
            >
              下一页
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
