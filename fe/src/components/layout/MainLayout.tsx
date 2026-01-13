import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { Book, LayoutDashboard, Settings, Wrench, MessageSquare } from 'lucide-react';
import { cn } from '@/lib/utils';

export default function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <div className="flex h-screen w-full bg-gray-50">
      {/* Sidebar */}
      <aside className="w-16 flex-shrink-0 flex flex-col items-center py-4 bg-white border-r border-gray-200 z-10">
        <div className="mb-6" onClick={() => navigate('/settings')}>
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center text-white font-bold cursor-pointer hover:opacity-90 transition-opacity">D</div>
        </div>
        <nav className="flex flex-col gap-4 w-full px-2">
          <NavItem
            icon={<LayoutDashboard size={20} />}
            label="探索"
            onClick={() => navigate('/')}
            active={location.pathname === '/' || location.pathname === '/explore'}
          />
          <NavItem
            icon={<Wrench size={20} />}
            label="工作室"
            onClick={() => navigate('/studio')}
            active={location.pathname.startsWith('/studio')}
          />
          <NavItem
            icon={<Book size={20} />}
            label="知识库"
            onClick={() => navigate('/knowledge')}
            active={location.pathname.startsWith('/knowledge')}
          />
          <NavItem
            icon={<MessageSquare size={20} />}
            label="聊天"
            onClick={() => navigate('/chat')}
            active={location.pathname.startsWith('/chat')}
          />
          <NavItem
            icon={<Settings size={20} />}
            label="工具"
            onClick={() => navigate('/tools')}
            active={location.pathname.startsWith('/tools')}
          />
        </nav>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden relative">
        <Outlet />
      </main>
    </div>
  );
}

function NavItem({ icon, label, active, onClick }: { icon: React.ReactNode; label: string; active?: boolean; onClick?: () => void }) {
  return (
    <div className="group relative flex justify-center">
      <button
        onClick={onClick}
        className={cn(
          "p-2.5 rounded-xl transition-all duration-200",
          active
            ? "bg-blue-50 text-blue-600 shadow-sm"
            : "text-gray-500 hover:bg-gray-100 hover:text-gray-900"
        )}>
        {icon}
      </button>
      {/* Tooltip */}
      <div className="absolute left-14 top-1/2 -translate-y-1/2 px-2 py-1 bg-gray-900 text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50">
        {label}
      </div>
    </div>
  )
}
