import { Activity } from 'lucide-react';

interface GraphStatsProps {
    nodeCount: number;
    linkCount: number;
}

export function GraphStats({ nodeCount, linkCount }: GraphStatsProps) {
    return (
        <div className="absolute bottom-6 left-1/2 transform -translate-x-1/2 z-20 flex items-center">
            <div className="bg-slate-900/80 backdrop-blur-md border border-slate-700/50 rounded-full px-4 py-1.5 flex items-center gap-4 text-xs font-mono text-slate-400 shadow-lg ring-1 ring-white/5">
                <div className="flex items-center gap-2">
                    <span className="relative flex h-2 w-2">
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                        <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                    </span>
                    <span className="font-medium text-slate-300">Live Graph</span>
                </div>

                <div className="h-3 w-px bg-slate-700"></div>

                <div className="flex items-center gap-1.5">
                    <span>Nodes:</span>
                    <span className="text-indigo-400 font-bold">{nodeCount.toLocaleString()}</span>
                </div>

                <div className="h-3 w-px bg-slate-700"></div>

                <div className="flex items-center gap-1.5">
                    <span>Edges:</span>
                    <span className="text-purple-400 font-bold">{linkCount.toLocaleString()}</span>
                </div>

                <div className="h-3 w-px bg-slate-700"></div>

                <div className="flex items-center gap-1.5 opacity-70">
                    <Activity size={12} />
                    <span>FPS: 60</span>
                </div>
            </div>
        </div>
    );
}
