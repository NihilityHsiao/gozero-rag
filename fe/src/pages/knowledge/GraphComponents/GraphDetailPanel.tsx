import { X, FileText, Share2, Sparkles, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import type { GraphNode, GraphLink } from '@/types/graph';

interface GraphDetailPanelProps {
    node: GraphNode | null;
    links: GraphLink[];
    onClose: () => void;
    onNavigateToNode?: (id: string) => void;
    typeColors: Record<string, string>;
}

export function GraphDetailPanel({ node, links, onClose, onNavigateToNode, typeColors }: GraphDetailPanelProps) {
    if (!node) return null;

    const color = typeColors[node.type] || typeColors.default;

    // Filter relevant links
    // @ts-ignore
    const nodeLinks = links.filter(l => (l.source.id || l.source) === node.id || (l.target.id || l.target) === node.id);

    return (
        <div className="absolute top-4 right-4 bottom-4 w-96 bg-slate-900/95 backdrop-blur-xl border border-slate-700/50 rounded-2xl shadow-2xl flex flex-col z-30 transition-transform duration-300 ring-1 ring-white/10">
            {/* Header */}
            <div className="p-5 border-b border-slate-700/50 flex justify-between items-start bg-slate-800/30 rounded-t-2xl">
                <div className="flex items-start gap-4">
                    <div
                        className="h-12 w-12 rounded-xl flex items-center justify-center shrink-0 shadow-lg"
                        style={{ backgroundColor: `${color}20`, color: color }}
                    >
                        <Share2 size={24} />
                    </div>
                    <div>
                        <h2 className="text-xl font-bold text-white leading-tight break-words">{node.name}</h2>
                        <div className="flex gap-2 mt-2">
                            <Badge variant="outline" className="bg-slate-800/50 border-slate-600 text-slate-300">
                                {node.type}
                            </Badge>
                            <Badge variant="outline" className="bg-indigo-500/10 border-indigo-500/30 text-indigo-300">
                                Val: {node.val}
                            </Badge>
                        </div>
                    </div>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    className="text-slate-400 hover:text-white hover:bg-white/10 rounded-full h-8 w-8"
                    onClick={onClose}
                >
                    <X size={18} />
                </Button>
            </div>

            {/* Content */}
            <ScrollArea className="flex-1 p-5">
                <div className="space-y-6">
                    {/* AI Summary Section */}
                    <section>
                        <div className="flex items-center gap-2 mb-3">
                            <Sparkles size={16} className="text-purple-400" />
                            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">AI Summary</h3>
                        </div>
                        <div className="bg-slate-800/50 p-4 rounded-xl border border-slate-700/50 ring-1 ring-white/5">
                            <p className="text-sm text-slate-300 leading-relaxed">
                                {node.description || "No description available for this node."}
                            </p>
                        </div>
                    </section>

                    {/* Source Documents */}
                    <section>
                        <div className="flex items-center gap-2 mb-3">
                            <FileText size={16} className="text-blue-400" />
                            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">Source Documents</h3>
                        </div>
                        <div className="space-y-2">
                            {node.source_id && node.source_id.length > 0 ? (
                                node.source_id.map((id, idx) => (
                                    <div key={idx} className="group flex items-center p-3 rounded-lg border border-slate-700/50 bg-slate-800/20 hover:bg-slate-800/80 hover:border-indigo-500/30 transition-all cursor-pointer">
                                        <div className="h-8 w-8 rounded bg-red-500/10 text-red-500 flex items-center justify-center mr-3 group-hover:scale-110 transition-transform">
                                            <FileText size={16} />
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <p className="text-sm font-medium text-slate-200 truncate group-hover:text-indigo-400 transition-colors">
                                                {id}
                                            </p>
                                        </div>
                                        <ExternalLink size={14} className="text-slate-500 opacity-0 group-hover:opacity-100 transition-opacity" />
                                    </div>
                                ))
                            ) : (
                                <p className="text-sm text-slate-500 italic">No source documents linked.</p>
                            )}
                        </div>
                    </section>

                    {/* Relationships */}
                    <section>
                        <div className="flex items-center gap-2 mb-3">
                            <Share2 size={16} className="text-emerald-400" />
                            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400">Relationships</h3>
                        </div>
                        <div className="bg-slate-800/20 rounded-xl border border-slate-700/50 overflow-hidden divide-y divide-slate-700/50">
                            {nodeLinks.map((link, idx) => {
                                // @ts-ignore
                                const isSource = (link.source.id || link.source) === node.id;
                                // @ts-ignore
                                const otherNode = isSource ? link.target : link.source;
                                const otherNodeId = otherNode.id || otherNode;
                                const otherNodeName = otherNode.name || otherNodeId;

                                return (
                                    <div
                                        key={idx}
                                        className="p-3 hover:bg-slate-800/60 transition-colors cursor-pointer group"
                                        onClick={() => onNavigateToNode?.(otherNodeId)}
                                    >
                                        <div className="flex justify-between items-center mb-1">
                                            <h4 className="text-sm font-semibold text-slate-300 group-hover:text-indigo-400 transition-colors">
                                                {otherNodeName}
                                            </h4>
                                            <Badge variant="secondary" className="text-[10px] bg-slate-800 text-slate-400 border-none px-1.5 h-5">
                                                {isSource ? 'OUTGOING' : 'INCOMING'}
                                            </Badge>
                                        </div>
                                        <p className="text-xs text-slate-500 flex items-center gap-1">
                                            <span className="font-mono text-slate-600">--[{link.description || 'RELATED'}]--&gt;</span>
                                        </p>
                                    </div>
                                );
                            })}
                            {nodeLinks.length === 0 && (
                                <div className="p-4 text-center text-sm text-slate-500">No connections</div>
                            )}
                        </div>
                    </section>
                </div>
            </ScrollArea>

            {/* Footer Actions */}
            <div className="p-4 border-t border-slate-700/50 bg-slate-800/30 rounded-b-2xl">
                <Button className="w-full bg-indigo-600 hover:bg-indigo-700 text-white shadow-lg shadow-indigo-500/20">
                    Expand Analysis
                </Button>
            </div>
        </div>
    );
}
