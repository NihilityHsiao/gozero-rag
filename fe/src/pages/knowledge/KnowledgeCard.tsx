import { useNavigate } from 'react-router-dom';
import type { KnowledgeBaseInfo } from '@/types';
import { BookOpen, MoreHorizontal } from 'lucide-react';

interface KnowledgeCardProps {
  info: KnowledgeBaseInfo;
}

export default function KnowledgeCard({ info }: KnowledgeCardProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/knowledge/${info.id}`);
  };

  const handleMoreClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    // TODO: Show menu
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
        <button
          onClick={handleMoreClick}
          className="p-1 -mr-2 -mt-2 text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <MoreHorizontal size={20} />
        </button>
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
      </div>
    </div>
  );
}
