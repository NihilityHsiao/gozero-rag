import { useEffect } from 'react';
import { Plus, MessageSquare, Search, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useChatStore } from '@/store/useChatStore';
import { useNavigate, useParams } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { formatDistanceToNow } from 'date-fns';
import { zhCN } from 'date-fns/locale';
import { getConversationList, deleteConversation } from '@/api/chat';

export default function ChatSidebar() {
    const navigate = useNavigate();
    const { conversationId } = useParams();
    const { conversations, setCurrentConversationId, setConversations } = useChatStore();

    const fetchConversations = async () => {
        try {
            const res = await getConversationList({ page: 1, page_size: 50 });
            setConversations(res.list);
        } catch (e) {
            console.error('Failed to fetch conversations', e);
        }
    };

    useEffect(() => {
        fetchConversations();
    }, []);

    const handleNewChat = () => {
        setCurrentConversationId(null);
        navigate('/chat');
    };

    const handleSelectChat = (id: string) => {
        setCurrentConversationId(id);
        navigate(`/chat/${id}`);
    };

    const handleDeleteChat = async (id: string, e: React.MouseEvent) => {
        e.stopPropagation();
        if (!confirm('确认删除该对话?')) return;
        try {
            await deleteConversation(id);
            if (id === conversationId) {
                handleNewChat();
            }
            await fetchConversations();
        } catch (error) {
            console.error('Delete failed', error);
        }
    };

    return (
        <div className="w-[260px] flex flex-col border-r bg-gray-50/50 h-full">
            <div className="p-4 space-y-4">
                <Button
                    className="w-full justify-start gap-2 bg-blue-600 hover:bg-blue-700 text-white shadow-sm"
                    onClick={handleNewChat}
                >
                    <Plus className="w-4 h-4" />
                    新对话
                </Button>

                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-gray-400" />
                    <Input placeholder="搜索历史记录..." className="pl-8 bg-white" />
                </div>
            </div>

            <ScrollArea className="flex-1 px-3">
                <div className="space-y-1 pb-4">
                    {conversations.map((conv) => (
                        <div
                            key={conv.id}
                            onClick={() => handleSelectChat(conv.id)}
                            className={cn(
                                "group flex items-center gap-3 px-3 py-3 rounded-lg text-sm transition-colors cursor-pointer",
                                conversationId === conv.id
                                    ? "bg-white text-blue-600 shadow-sm border border-gray-100"
                                    : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                            )}
                        >
                            <MessageSquare className="w-4 h-4 shrink-0" />
                            <div className="flex-1 overflow-hidden">
                                <div className="truncate font-medium">{conv.title}</div>
                                <div className="text-xs text-gray-400 mt-0.5">
                                    {conv.updated_at ? formatDistanceToNow(new Date(conv.updated_at), { addSuffix: true, locale: zhCN }) : ''}
                                </div>
                            </div>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 opacity-0 group-hover:opacity-100 shrink-0 text-gray-400 hover:text-red-500"
                                onClick={(e) => handleDeleteChat(conv.id, e)}
                            >
                                <Trash2 className="w-3 h-3" />
                            </Button>
                        </div>
                    ))}
                </div>
            </ScrollArea>
        </div>
    );
}
