import { Activity } from 'lucide-react';

interface GraphStatsProps {
    nodeCount: number;
    linkCount: number;
}

export function GraphStats({ nodeCount, linkCount }: GraphStatsProps) {
    return (
        <div className="absolute top-24 left-6 z-20 w-64 bg-slate-900/80 backdrop-blur-md border border-slate-700/50 rounded-xl shadow-2xl text-white overflow-hidden ring-1 ring-white/5">
            <div className="px-4 py-3 border-b border-slate-700/50 flex justify-between items-center bg-slate-800/50">
                <h3 className="text-sm font-semibold text-slate-200 flex items-center gap-2">
                    <Activity size={14} className="text-indigo-400" />
                    Graph Statistics
                </h3>
                <span className="flex h-2 w-2 relative">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                </span>
            </div>
            <div className="p-4 space-y-4">
                <div className="grid grid-cols-2 gap-4">
                    <div className="bg-slate-800/40 p-2 rounded-lg border border-slate-700/50">
                        <p className="text-xs text-slate-400 mb-1">Total Nodes</p>
                        <p className="text-xl font-bold text-indigo-400">{nodeCount.toLocaleString()}</p>
                    </div>
                    <div className="bg-slate-800/40 p-2 rounded-lg border border-slate-700/50">
                        <p className="text-xs text-slate-400 mb-1">Total Edges</p>
                        <p className="text-xl font-bold text-purple-400">{linkCount.toLocaleString()}</p>
                    </div>
                </div>
            </div>
        </div>
    );
}
