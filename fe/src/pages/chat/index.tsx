import { useState } from 'react';
import ChatSidebar from './components/ChatSidebar';
import ChatWindow from './components/ChatWindow';
import ConfigPanel from './components/ConfigPanel';
import { Button } from '@/components/ui/button';
import { PanelRightOpen, PanelRightClose } from 'lucide-react';

export default function ChatPage() {
    const [showConfig, setShowConfig] = useState(true);

    return (
        <div className="flex h-[calc(100vh-64px)] w-full overflow-hidden bg-white">
            {/* Sidebar */}
            <ChatSidebar />

            {/* Main Chat Area */}
            <div className="flex-1 flex flex-col min-w-0 relative">
                {/* Header / Toolbar (Optional) */}
                <div className="absolute top-4 right-4 z-10">
                    <Button
                        variant="ghost"
                        size="icon"
                        className="bg-white/80 backdrop-blur shadow-sm border"
                        onClick={() => setShowConfig(!showConfig)}
                    >
                        {showConfig ? <PanelRightClose className="w-4 h-4" /> : <PanelRightOpen className="w-4 h-4" />}
                    </Button>
                </div>

                <ChatWindow />
            </div>

            {/* Right Config Panel */}
            {showConfig && (
                <ConfigPanel />
            )}
        </div>
    );
}
