import { Outlet, NavLink, useParams, useLocation, Link } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { FileText, Settings, TestTube, ChevronRight, Database, Share2 } from 'lucide-react';
import { useKnowledgeStore } from '@/store/useKnowledgeStore';
import { useEffect, useState } from 'react';
import type { KnowledgeBaseInfo } from '@/types';

export default function KnowledgeDetailLayout() {
  const { id } = useParams<{ id: string }>();
  const location = useLocation();
  const { list } = useKnowledgeStore();
  const [currentKnowledge, setCurrentKnowledge] = useState<KnowledgeBaseInfo | null>(null);

  // Find current knowledge base info from store or fetch it (simplified here)
  useEffect(() => {
    if (id && list.length > 0) {
      const found = list.find(k => k.id === id);
      if (found) setCurrentKnowledge(found);
    }
  }, [id, list]);

  const navItems = [
    {
      title: '文档列表',
      icon: FileText,
      href: `/knowledge/${id}/documents`,
      active: location.pathname.includes('/documents'),
    },
    {
      title: '召回测试',
      icon: TestTube,
      href: `/knowledge/${id}/retrieve`,
      active: location.pathname.includes('/retrieve'),
    },
    {
      title: '知识图谱',
      icon: Share2,
      href: `/knowledge/${id}/graph`,
      active: location.pathname.includes('/graph'),
    },
    {
      title: '设置',
      icon: Settings,
      href: `/knowledge/${id}/settings`,
      active: location.pathname.includes('/settings'),
    },
  ];

  return (
    <div className="flex h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="w-[250px] bg-white border-r border-gray-200 flex flex-col flex-shrink-0">
        <div className="h-16 flex items-center px-6 border-b border-gray-100">
          <Link to="/knowledge" className="flex items-center gap-2 text-blue-600 font-semibold hover:opacity-80 transition-opacity">
            <div className="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center">
              <Database size={18} />
            </div>
            <span className="truncate max-w-[160px]" title={currentKnowledge?.name || '加载中...'}>
              {currentKnowledge?.name || '知识库'}
            </span>
          </Link>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.href}
              to={item.href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors",
                item.active
                  ? "bg-blue-50 text-blue-600"
                  : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
              )}
            >
              <item.icon size={18} />
              {item.title}
            </NavLink>
          ))}
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Header */}
        <header className="h-16 bg-white border-b border-gray-200 px-8 flex items-center justify-between flex-shrink-0">
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <Link to="/knowledge" className="hover:text-gray-700 cursor-pointer transition-colors">知识库</Link>
            <ChevronRight size={14} />
            <span className="font-medium text-gray-900">{currentKnowledge?.name || '...'}</span>
          </div>
        </header>

        {/* Page Content */}
        {/* Page Content */}
        {location.pathname.includes('/graph') ? (
          <div className="flex-1 overflow-hidden relative">
            <Outlet />
          </div>
        ) : (
          <div className="flex-1 overflow-auto p-8">
            <div className="max-w-6xl mx-auto">
              <Outlet />
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
