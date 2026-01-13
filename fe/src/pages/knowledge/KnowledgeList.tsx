import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useInView } from 'react-intersection-observer';
import { useKnowledgeStore } from '@/store/useKnowledgeStore';
import KnowledgeCard from './KnowledgeCard';
import { Plus, Search, ExternalLink, Loader2 } from 'lucide-react';

export default function KnowledgeList() {
  const navigate = useNavigate();
  const { list, loading, fetchList, page, hasMore, reset } = useKnowledgeStore();
  const { ref, inView } = useInView({
    threshold: 0,
    rootMargin: '100px',
  });

  // Reset store on unmount or initial load
  useEffect(() => {
    reset();
    fetchList(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Infinite scroll logic
  useEffect(() => {
    if (inView && hasMore && !loading && list.length > 0) {
      fetchList(page + 1);
    }
  }, [inView, hasMore, loading, list.length, page, fetchList]);

  const handleCreate = () => {
    navigate('/knowledge/create');
  };

  return (
    <div className="flex flex-col h-full bg-gray-50">
      {/* Header Section */}
      <header className="flex-shrink-0 h-16 bg-white border-b border-gray-200 px-6 lg:px-8 flex items-center justify-between sticky top-0 z-10">
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <span className="font-semibold text-gray-900 text-lg">知识库</span>
        </div>

        <div className="flex items-center gap-3">
          <div className="relative hidden md:block">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" size={16} />
            <input
              type="text"
              placeholder="搜索"
              className="h-9 w-64 pl-9 pr-4 bg-gray-100 border-none rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-600/20 transition-all placeholder:text-gray-400"
            />
          </div>
          <button className="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors shadow-sm">
            <ExternalLink size={16} />
            <span className="hidden sm:inline">API 文档</span>
          </button>
          <button
            onClick={handleCreate}
            className="flex items-center gap-2 px-3 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors shadow-sm"
          >
            <Plus size={16} />
            <span className="hidden sm:inline">创建知识库</span>
          </button>
        </div>
      </header>

      {/* Content Section */}
      <div className="flex-1 overflow-auto p-6 lg:p-8">
        <div className="max-w-[1600px] mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-5">
            {/* Create New Card */}
            <button
              onClick={handleCreate}
              className="group flex flex-col items-center justify-center h-[180px] bg-gray-50 border border-dashed border-gray-300 rounded-xl hover:border-blue-500 hover:bg-blue-50/30 transition-all duration-200 cursor-pointer"
            >
              <div className="w-10 h-10 rounded-lg bg-gray-200 flex items-center justify-center text-gray-500 group-hover:bg-blue-100 group-hover:text-blue-600 transition-colors">
                <Plus size={24} />
              </div>
              <span className="mt-3 text-sm font-medium text-gray-600 group-hover:text-blue-600">创建知识库</span>
            </button>

            {/* Data List */}
            {list.map((item) => (
              <KnowledgeCard key={item.id} info={item} />
            ))}

            {/* Initial Loading State */}
            {loading && list.length === 0 && (
              Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="h-[180px] bg-white rounded-xl border border-gray-200 animate-pulse p-5">
                  <div className="w-10 h-10 bg-gray-200 rounded-lg mb-4"></div>
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                </div>
              ))
            )}
          </div>

          {/* Infinite Scroll Trigger & Status */}
          <div ref={ref} className="mt-8 flex justify-center h-10">
            {loading && list.length > 0 && (
              <div className="flex items-center gap-2 text-gray-500">
                <Loader2 className="animate-spin" size={20} />
                <span className="text-sm">加载中...</span>
              </div>
            )}
            {!hasMore && list.length > 0 && (
              <span className="text-sm text-gray-400">没有更多数据了</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
