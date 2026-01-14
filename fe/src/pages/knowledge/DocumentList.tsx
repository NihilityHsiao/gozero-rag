import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getDocumentList, batchParseDocument } from '@/api/document';
import { clearDatasetDocuments } from '@/api/dataset';
import type { KnowledgeDocumentInfo, RunStatus } from '@/types';
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
import { Checkbox } from '@/components/ui/checkbox';
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
  Trash2,
  Play
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

  // Batch Selection State
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [parsing, setParsing] = useState(false);

  const fetchDocs = async () => {
    if (!id) return;
    setLoading(true);
    try {
      // Note: Backend currently doesn't support keyword search in this API, 
      // but we prepare the UI for it.
      const res = await getDocumentList(id, {
        page,
        page_size: 10,
      });
      setList(res.list || []);
      setTotal(res.total);
      // Clear selection on page change or refresh
      setSelectedIds(new Set());
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
      await clearDatasetDocuments(id);
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

  const handleBatchParse = async () => {
    if (!id || selectedIds.size === 0) return;
    setParsing(true);
    try {
      await batchParseDocument(id, Array.from(selectedIds));
      toast.success('Batch parse task submitted');
      // Optimistic update
      setList(prev => prev.map(doc => {
        if (selectedIds.has(doc.id)) {
          return { ...doc, run_status: 'pending' };
        }
        return doc;
      }));
      setSelectedIds(new Set()); // Clear selection
    } catch (error) {
      console.error('Batch parse failed', error);
      toast.error('Failed to submit batch parse task');
    } finally {
      setParsing(false);
    }
  };

  const handleParseOne = async (e: React.MouseEvent, docId: string) => {
    e.stopPropagation();
    if (!id) return;
    try {
      await batchParseDocument(id, [docId]);
      toast.success('Parse task submitted');
      setList(prev => prev.map(doc => {
        if (doc.id === docId) {
          return { ...doc, run_status: 'pending' };
        }
        return doc;
      }));
    } catch (error) {
      console.error('Parse failed', error);
      toast.error('Failed to submit parse task');
    }
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === list.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(list.map(d => d.id)));
    }
  };

  const toggleSelectOne = (docId: string) => {
    const newSet = new Set(selectedIds);
    if (newSet.has(docId)) {
      newSet.delete(docId);
    } else {
      newSet.add(docId);
    }
    setSelectedIds(newSet);
  };

  // Format bytes to readable string
  const formatSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  // 格式化时间戳（毫秒）为可读字符串
  const formatTime = (timestamp: number) => {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  // Status Badge Helper - 使用 run_status
  const getStatusBadge = (runStatus: RunStatus, progressMsg?: string) => {
    const styles = {
      success: 'bg-green-100 text-green-700 hover:bg-green-100 border-green-200',
      indexing: 'bg-blue-100 text-blue-700 hover:bg-blue-100 border-blue-200',
      pending: 'bg-gray-100 text-gray-700 hover:bg-gray-100 border-gray-200',
      paused: 'bg-yellow-50 text-yellow-700 hover:bg-yellow-50 border-yellow-200',
      failed: 'bg-red-50 text-red-600 hover:bg-red-50 border-red-200 cursor-pointer',
      canceled: 'bg-gray-50 text-gray-400 hover:bg-gray-50 border-gray-200',
    };

    const statusMap: Record<string, string> = {
      success: '已完成',
      indexing: '索引中',
      pending: '等待中',
      paused: '暂停',
      failed: '失败',
      canceled: '已取消',
    }

    const label = statusMap[runStatus] || runStatus;

    if (runStatus === 'failed' && progressMsg) {
      return (
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Badge variant="outline" className={cn("font-normal", styles[runStatus])}>
                <AlertCircle size={12} className="mr-1" />
                {label}
              </Badge>
            </TooltipTrigger>
            <TooltipContent>
              <p className="text-xs max-w-xs">{progressMsg}</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      );
    }

    if (runStatus === 'indexing') {
      return (
        <Badge variant="outline" className={cn("font-normal", styles[runStatus])}>
          <Loader2 size={12} className="mr-1 animate-spin" />
          {label}
        </Badge>
      )
    }

    return (
      <Badge variant="outline" className={cn("font-normal", styles[runStatus])}>
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
        <div className="flex gap-2">
          {selectedIds.size > 0 && (
            <Button
              variant="secondary"
              onClick={handleBatchParse}
              disabled={parsing}
            >
              {parsing ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play size={16} className="mr-2" />}
              批量解析 ({selectedIds.size})
            </Button>
          )}
          <Button
            className="bg-blue-600 hover:bg-blue-700"
            onClick={() => navigate(`/knowledge/${id}/dataset/create`)}
          >
            <Plus size={16} className="mr-2" />
            添加文件
          </Button>
        </div>
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
              <TableHead className="w-[40px]">
                <Checkbox
                  checked={list.length > 0 && selectedIds.size === list.length}
                  onCheckedChange={toggleSelectAll}
                />
              </TableHead>
              <TableHead className="w-[35%]">名称</TableHead>
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
                  <TableCell><div className="h-4 w-4 bg-gray-100 rounded animate-pulse"></div></TableCell>
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
                <TableCell colSpan={8} className="h-32 text-center text-gray-500">
                  暂无文档
                </TableCell>
              </TableRow>
            ) : (
              list.map((doc) => (
                <TableRow
                  key={doc.id}
                  className={cn(
                    "hover:bg-gray-50/50 cursor-pointer",
                    doc.status === 0 ? "hover:bg-blue-50/50" : "hover:bg-gray-100/50",
                    selectedIds.has(doc.id) && "bg-blue-50/30"
                  )}
                  onClick={() => {
                    // Default: View details. Need to be careful not to conflict with checkbox
                    if (doc.status === 0) {
                      navigate(`/knowledge/${id}/document/${doc.id}/edit`);
                    } else {
                      navigate(`/knowledge/${id}/document/${doc.id}`);
                    }
                  }}
                >
                  <TableCell onClick={(e) => { e.stopPropagation(); }}>
                    <Checkbox
                      checked={selectedIds.has(doc.id)}
                      onCheckedChange={() => toggleSelectOne(doc.id)}
                    />
                  </TableCell>
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
                      {doc.chunk_num} 个切片
                    </Badge>
                  </TableCell>
                  <TableCell className="text-gray-500 text-sm">
                    {formatTime(doc.created_time)}
                  </TableCell>
                  <TableCell>
                    {getStatusBadge(doc.run_status, doc.progress_msg)}
                  </TableCell>
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <Switch checked={doc.status === 1} />
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      {/* Parse Button */}
                      {(doc.run_status !== 'indexing' && doc.run_status !== 'pending') && (
                        <TooltipProvider>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8 text-gray-500 hover:text-blue-600"
                                onClick={(e) => handleParseOne(e, doc.id)}
                              >
                                <Play size={16} />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>Run Parsing</TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      )}

                      {/* More Button */}
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-gray-500 hover:text-gray-900">
                        <MoreHorizontal size={16} />
                      </Button>
                    </div>
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
