import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { User, Server, ArrowLeft, Users } from 'lucide-react';
import { Button, buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';
// MainLayout is imported but not used in layout directly, kept if needed later or remove.
// import MainLayout from '@/components/layout/MainLayout';

interface SidebarNavProps extends React.HTMLAttributes<HTMLElement> {
    items: {
        href: string;
        title: string;
        icon: React.ReactNode;
    }[];
}

function SidebarNav({ className, items, ...props }: SidebarNavProps) {
    return (
        <nav
            className={cn(
                "flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1",
                className
            )}
            {...props}
        >
            {items.map((item) => (
                <NavLink
                    key={item.href}
                    to={item.href}
                    className={({ isActive }) => cn(
                        buttonVariants({ variant: "ghost" }),
                        isActive
                            ? "bg-muted hover:bg-muted"
                            : "hover:bg-transparent hover:underline",
                        "justify-start"
                    )}
                >
                    {item.icon}
                    <span className="ml-2">{item.title}</span>
                </NavLink>
            ))}
        </nav>
    );
}

const sidebarNavItems = [
    {
        title: "用户信息",
        href: "/settings/profile",
        icon: <User className="w-4 h-4" />,
    },
    {
        title: "模型配置",
        href: "/settings/provider",
        icon: <Server className="w-4 h-4" />,
    },
    {
        title: "团队",
        href: "/settings/team",
        icon: <Users className="w-4 h-4" />,
    },
];


export default function SettingsLayout() {
    const navigate = useNavigate();

    return (
        <div className="h-screen overflow-auto">
            <div className="space-y-6 p-10 pb-16 md:block">
                <div className="flex flex-col gap-2">
                    <div>
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => navigate('/')}
                            className="-ml-3 gap-1 text-muted-foreground hover:text-foreground px-3"
                        >
                            <ArrowLeft className="w-4 h-4" />
                            返回首页
                        </Button>
                    </div>
                    <div className="space-y-0.5">
                        <h2 className="text-2xl font-bold tracking-tight">设置</h2>
                        <p className="text-muted-foreground">
                            管理您的账户设置和模型配置。
                        </p>
                    </div>
                </div>

                <div className="shrink-0 bg-border h-[1px] w-full my-6" />
                <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
                    <aside className="-mx-4 lg:w-1/5">
                        <SidebarNav items={sidebarNavItems} />
                    </aside>
                    <div className="flex-1 lg:max-w-5xl">
                        <Outlet />
                    </div>
                </div>
            </div>
        </div>
    );
}
