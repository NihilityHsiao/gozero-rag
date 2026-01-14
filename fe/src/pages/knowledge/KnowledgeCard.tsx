import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import type { KnowledgeBaseInfo } from '@/types';
import { BookOpen, MoreHorizontal, Eye, Users, Lock } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import { updateKnowledgeBasePermission } from '@/api/knowledge';
import { toast } from 'sonner';
import { useAuthStore } from '@/store/useAuthStore';
import { useKnowledgeStore } from '@/store/useKnowledgeStore';

interface KnowledgeCardProps {
  info: KnowledgeBaseInfo;
}

export default function KnowledgeCard({ info }: KnowledgeCardProps) {
  const navigate = useNavigate();
  const { userInfo } = useAuthStore();
  const { fetchList, page } = useKnowledgeStore();
  const [updating, setUpdating] = useState(false);

  const isOwner = info.created_by === userInfo?.user_id;

  const handleClick = () => {
    navigate(`/knowledge/${info.id}`);
  };

  const handleUpdatePermission = async (permission: string) => {
    if (updating) return;

    setUpdating(true);
    try {
      await updateKnowledgeBasePermission(info.id, permission);
      toast.success(permission === 'me' ? '已设为仅自己可见' : '已设为团队可见');
      // 刷新列表
      fetchList(page);
    } catch (error: any) {
      toast.error(error?.msg || '更新权限失败');
    } finally {
      setUpdating(false);
    }
  };

  return (
    <div
      onClick={handleClick}
      className="group relative flex flex-col h-[180px] p-5 bg-white border border-gray-200 rounded-xl hover:shadow-lg hover:-translate-y-1 hover:border-blue-200 hover:ring-1 hover:ring-blue-600 transition-all duration-200 cursor-pointer"
    >
      <div className="flex justify-between items-start mb-2">
        <div className="w-10 h-10 rounded-lg bg-blue-50 flex items-center justify-center text-xl shrink-0">
          📚
        </div>

        {/* 下拉菜单 (仅 Owner 可见) */}
        {isOwner && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                onClick={(e) => e.stopPropagation()}
                className="p-1 -mr-2 -mt-2 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-all opacity-0 group-hover:opacity-100"
              >
                <MoreHorizontal size={20} />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48" onClick={(e) => e.stopPropagation()}>
              <div className="px-2 py-1.5 text-xs font-medium text-gray-500">可见性设置</div>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                disabled={info.permission === 'me' || updating}
                onClick={(e) => {
                  e.stopPropagation();
                  handleUpdatePermission('me');
                }}
                className="cursor-pointer"
              >
                <Lock className="mr-2 h-4 w-4" />
                <span>仅自己可见</span>
                {info.permission === 'me' && (
                  <Eye className="ml-auto h-4 w-4 text-blue-600" />
                )}
              </DropdownMenuItem>
              <DropdownMenuItem
                disabled={info.permission === 'team' || updating}
                onClick={(e) => {
                  e.stopPropagation();
                  handleUpdatePermission('team');
                }}
                className="cursor-pointer"
              >
                <Users className="mr-2 h-4 w-4" />
                <span>团队可见</span>
                {info.permission === 'team' && (
                  <Eye className="ml-auto h-4 w-4 text-blue-600" />
                )}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>

      <div className="flex-1 min-h-0">
        <h3 className="text-base font-bold text-gray-900 truncate pr-2" title={info.name}>
          {info.name}
        </h3>
        <p className="mt-1 text-sm text-gray-500 line-clamp-2 leading-relaxed h-10">
          {info.description || '暂无描述'}
        </p>
      </div>

      <div className="mt-3 flex items-center text-xs text-gray-400 gap-3">
        <div className="flex items-center gap-1">
          <BookOpen size={14} />
          <span>0 文档</span>
        </div>
        <span className="w-1 h-1 rounded-full bg-gray-300"></span>
        <span>{info.updated_at ? new Date(info.updated_at).toLocaleDateString() : '刚刚'}</span>
        {/* 显示权限标识 */}
        {info.permission === 'me' ? (
          <div className="ml-auto flex items-center gap-1 text-xs text-gray-400">
            <Lock size={12} />
            <span>私有</span>
          </div>
        ) : (
          <div className="ml-auto flex items-center gap-1 text-xs text-blue-600">
            <Users size={12} />
            <span>团队</span>
          </div>
        )}
      </div>
    </div>
  );
}
