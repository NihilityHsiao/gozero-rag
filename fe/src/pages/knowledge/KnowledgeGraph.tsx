import { useEffect, useState, useRef, useCallback } from 'react';
import ForceGraph2D, { type ForceGraphMethods } from 'react-force-graph-2d';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Search, ZoomIn, ZoomOut, Maximize2 } from 'lucide-react';
import { Input } from '@/components/ui/input';
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
} from "@/components/ui/sheet";
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';

// Mock Data
const MOCK_DATA = {
    nodes: [
        {
            id: "孙悟空",
            name: "孙悟空",
            type: "person",
            description: "法号行者，唐僧的大徒弟，会七十二变、筋斗云。曾大闹天宫，后被压在五行山下，最终护送唐僧西天取经。",
            val: 20,
            source_id: ["018d96f3-1001-7000-8000-000000000001", "018d96f3-1001-7000-8000-000000000002", "018d96f3-1001-7000-8000-000000000003"]
        },
        {
            id: "唐三藏",
            name: "唐三藏",
            type: "person",
            description: "前世为如来佛祖弟子金蝉子，转世为唐朝高僧。也就是唐僧，性格慈悲但有时迂腐，是取经团队的核心领导者。",
            val: 15,
            source_id: ["018d96f3-2001-7000-8000-000000000001", "018d96f3-2001-7000-8000-000000000002"]
        },
        {
            id: "猪八戒",
            name: "猪八戒",
            type: "person",
            description: "法号悟能，原是天蓬元帅，因调戏嫦娥被贬下凡。性格贪吃好色但也憨厚有力，是团队中的开心果。",
            val: 12,
            source_id: ["018d96f3-3001-7000-8000-000000000001"]
        },
        {
            id: "沙悟净",
            name: "沙悟净",
            type: "person",
            description: "原是卷帘大将，因打碎琉璃盏被贬流沙河。性格沉默寡言，任劳任怨，负责挑担和照顾师父。",
            val: 10,
            source_id: ["018d96f3-4001-7000-8000-000000000001"]
        },
        {
            id: "白骨精",
            name: "白骨精",
            type: "person",
            description: "又称白骨夫人，本是白虎岭上的一具化为白骨的女尸，采天地灵气受日月精华变幻成人型，擅长变化。",
            val: 10,
            source_id: ["018d96f3-5001-7000-8000-000000000001"]
        },
        {
            id: "牛魔王",
            name: "牛魔王",
            type: "person",
            description: "自号平天大圣，翠云山和积雷山的主人，孙悟空的结拜大哥，实力强大的妖王。",
            val: 12,
            source_id: ["018d96f3-6001-7000-8000-000000000001"]
        },
        {
            id: "铁扇公主",
            name: "铁扇公主",
            type: "person",
            description: "又叫罗刹女，牛魔王的妻子，持有芭蕉扇，掌管火焰山一带的气候。",
            val: 10,
            source_id: ["018d96f3-7001-7000-8000-000000000001"]
        },
        {
            id: "花果山",
            name: "花果山",
            type: "geo",
            description: "位于东胜神洲傲来国，是孙悟空的出生地和老家，也是水帘洞的所在地。",
            val: 15,
            source_id: ["018d96f3-8001-7000-8000-000000000001"]
        },
        {
            id: "天庭",
            name: "天庭",
            type: "organization",
            description: "掌管三界的最高行政机构，由玉皇大帝统治，拥有众多天兵天将。",
            val: 18,
            source_id: ["018d96f3-9001-7000-8000-000000000001"]
        },
        {
            id: "大闹天宫",
            name: "大闹天宫",
            type: "event",
            description: "孙悟空因不满天庭官职低微及未被邀请参加蟠桃会，反出天庭，与十万天兵天将大战的历史性事件。",
            val: 12,
            source_id: ["018d96f3-0001-7000-8000-000000000001"]
        }
    ],
    links: [
        { source: "孙悟空", target: "唐三藏", description: "师徒关系", weight: 9.5 },
        { source: "猪八戒", target: "唐三藏", description: "师徒关系", weight: 8.0 },
        { source: "沙悟净", target: "唐三藏", description: "师徒关系", weight: 8.0 },
        { source: "孙悟空", target: "猪八戒", description: "师兄弟", weight: 7.5 },
        { source: "孙悟空", target: "花果山", description: "归属", weight: 10.0 },
        { source: "孙悟空", target: "大闹天宫", description: "发起者", weight: 9.0 },
        { source: "大闹天宫", target: "天庭", description: "冲突", weight: 6.0 },
        { source: "孙悟空", target: "白骨精", description: "敌对", weight: 5.0 },
        { source: "牛魔王", target: "孙悟空", description: "结拜兄弟", weight: 6.5 },
        { source: "牛魔王", target: "铁扇公主", description: "夫妻", weight: 8.5 }
    ]
};

// Node Colors by Type (Dify-like Light Theme Palette)
const TYPE_COLORS: Record<string, string> = {
    person: '#296DFF',      // Brighter Blue (Primary)
    geo: '#00BFA5',         // Teal
    organization: '#6840FA',// Purple
    event: '#F59E0B',       // Amber
    product: '#10B981',     // Green
    concept: '#F43F5E',     // Rose
    default: '#94A3B8'      // Slate 400
};

interface GraphNode {
    id: string;
    name: string;
    type: string;
    description: string;
    val: number;
    source_id: string[];
    x?: number;
    y?: number;
}



export default function KnowledgeGraph() {
    const fgRef = useRef<ForceGraphMethods | undefined>(undefined);
    const [data, setData] = useState({ nodes: [], links: [] });
    const [highlightNodes, setHighlightNodes] = useState(new Set<string>());
    const [highlightLinks, setHighlightLinks] = useState(new Set<string>()); // Use string representation "source-target"
    const [hoverNode, setHoverNode] = useState<GraphNode | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);

    useEffect(() => {
        // Simulate loading data
        setTimeout(() => {
            // @ts-ignore
            setData(MOCK_DATA);
        }, 500);
    }, []);

    const handleNodeClick = useCallback((node: GraphNode) => {
        setSelectedNode(node);
        if (fgRef.current) {
            fgRef.current.centerAt(node.x, node.y, 1000);
            fgRef.current.zoom(4, 2000);
        }
    }, []);

    const handleNodeHover = (node: GraphNode | null) => {
        setHoverNode(node || null);
        const newHighlightNodes = new Set<string>();
        const newHighlightLinks = new Set<string>();

        if (node) {
            newHighlightNodes.add(node.id);
            data.links.forEach((link: any) => {
                if (link.source.id === node.id || link.target.id === node.id) {
                    newHighlightLinks.add(`${link.source.id}-${link.target.id}`); // Keep direction agnostic if undirected, but assume src-tgt
                    newHighlightNodes.add(link.source.id);
                    newHighlightNodes.add(link.target.id);
                }
            });
        }

        setHighlightNodes(newHighlightNodes);
        setHighlightLinks(newHighlightLinks);
    };

    const handleSearch = () => {
        if (!searchQuery) return;
        const node = data.nodes.find((n: GraphNode) => n.name.includes(searchQuery));
        if (node) {
            const gNode = node as GraphNode; // Cast safely
            if (fgRef.current && gNode.x !== undefined && gNode.y !== undefined) {
                fgRef.current.centerAt(gNode.x, gNode.y, 1000);
                fgRef.current.zoom(4, 2000);
                handleNodeClick(gNode);
            }
        }
    }

    const paintNode = useCallback((node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
        const isHover = highlightNodes.has(node.id) || node === hoverNode;
        const isOtherHover = hoverNode && !isHover;

        // Opacity
        ctx.globalAlpha = isOtherHover ? 0.2 : 1; // Slightly more visible for non-hovered

        const label = node.name;
        const fontSize = 12 / globalScale;
        ctx.font = `${fontSize}px Sans-Serif`;

        const color = TYPE_COLORS[node.type] || TYPE_COLORS.default;

        // Draw Circle
        ctx.beginPath();
        const r = Math.sqrt(node.val) * 2; // Size based on val
        ctx.arc(node.x, node.y, r, 0, 2 * Math.PI, false);
        ctx.fillStyle = color;
        ctx.fill();

        // Draw Ring if selected or hovered
        if (selectedNode?.id === node.id || isHover) {
            ctx.beginPath();
            ctx.arc(node.x, node.y, r + 2, 0, 2 * Math.PI, false);
            ctx.strokeStyle = '#296DFF'; // Blue highlight ring
            ctx.lineWidth = 2 / globalScale;
            ctx.stroke();
        }

        // Draw Label
        if (globalScale > 1.5 || isHover || selectedNode?.id === node.id) {
            // Label Background (Optional for readability)
            // const textWidth = ctx.measureText(label).width;
            // ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
            // ctx.fillRect(node.x - textWidth / 2 - 2, node.y + r + 2, textWidth + 4, fontSize + 2);

            ctx.fillStyle = '#334155'; // Dark slate text
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(label, node.x, node.y + r + fontSize);
        }
    }, [highlightNodes, hoverNode, selectedNode]);

    return (
        <div className="flex h-[calc(100vh-140px)] gap-4">
            <Card className="flex-1 relative overflow-hidden bg-[#F9FAFB] border border-gray-100 shadow-sm rounded-xl">
                {/* Toolbar */}
                <div className="absolute top-4 left-4 z-10 flex flex-col gap-2 w-64">
                    <div className="flex gap-2">
                        <Input
                            placeholder="搜索节点..."
                            className="bg-white/90 backdrop-blur border-gray-200 text-gray-900 placeholder:text-gray-400 shadow-sm"
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && handleSearch()}
                        />
                        <Button size="icon" variant="secondary" className="bg-white border border-gray-200 shadow-sm text-gray-600 hover:bg-gray-50" onClick={handleSearch}>
                            <Search size={16} />
                        </Button>
                    </div>

                    <div className="flex gap-2">
                        <Button size="icon" variant="outline" className="bg-white/90 backdrop-blur border-gray-200 text-gray-600 hover:bg-gray-50 shadow-sm" onClick={() => fgRef.current?.zoomToFit(400)}>
                            <Maximize2 size={16} />
                        </Button>
                        <Button size="icon" variant="outline" className="bg-white/90 backdrop-blur border-gray-200 text-gray-600 hover:bg-gray-50 shadow-sm" onClick={() => fgRef.current?.zoom(fgRef.current.zoom() * 1.2, 400)}>
                            <ZoomIn size={16} />
                        </Button>
                        <Button size="icon" variant="outline" className="bg-white/90 backdrop-blur border-gray-200 text-gray-600 hover:bg-gray-50 shadow-sm" onClick={() => fgRef.current?.zoom(fgRef.current.zoom() / 1.2, 400)}>
                            <ZoomOut size={16} />
                        </Button>
                    </div>
                </div>

                {/* Graph */}
                <ForceGraph2D
                    ref={fgRef}
                    graphData={data}
                    nodeLabel="name"
                    nodeRelSize={6}
                    nodeCanvasObject={paintNode}
                    onNodeClick={(node) => handleNodeClick(node as GraphNode)}
                    onNodeHover={(node) => handleNodeHover(node ? (node as GraphNode) : null)}
                    linkColor={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        // Highlight: Blue; Normal: Light Gray
                        return highlightLinks.has(idStr) ? '#296DFF' : '#E2E8F0';
                    }}
                    linkWidth={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        return highlightLinks.has(idStr) ? 2 : 1;
                    }}
                    backgroundColor="#F9FAFB" // Light Gray Background
                    cooldownTicks={100}
                />
            </Card>

            {/* Details Sheet/Drawer */}
            <Sheet open={!!selectedNode} onOpenChange={(open) => !open && setSelectedNode(null)}>
                <SheetContent className="w-[400px] sm:w-[540px] overflow-y-auto">
                    <SheetHeader>
                        <SheetTitle className="flex items-center gap-2 text-xl">
                            <div className="w-4 h-4 rounded-full" style={{ backgroundColor: TYPE_COLORS[selectedNode?.type || 'default'] }}></div>
                            {selectedNode?.name}
                        </SheetTitle>
                        <SheetDescription>
                            <div className="flex gap-2 mt-2">
                                <Badge variant="outline">{selectedNode?.type}</Badge>
                                <Badge variant="secondary">Value: {selectedNode?.val}</Badge>
                            </div>
                        </SheetDescription>
                    </SheetHeader>

                    <div className="mt-6 space-y-6">
                        <div>
                            <h4 className="text-sm font-medium text-gray-500 mb-2">描述</h4>
                            <p className="text-gray-900 leading-relaxed bg-gray-50 p-3 rounded-lg text-sm">
                                {selectedNode?.description}
                            </p>
                        </div>

                        <Separator />

                        <div>
                            <h4 className="text-sm font-medium text-gray-500 mb-2">来源文档</h4>
                            <div className="flex flex-col gap-2">
                                {selectedNode?.source_id.map((id, index) => (
                                    <div key={index} className="flex items-center gap-2 p-2 rounded border border-gray-100 bg-white hover:bg-gray-50 cursor-pointer text-xs font-mono text-gray-600">
                                        <span className="w-2 h-2 rounded-full bg-blue-400"></span>
                                        {id}
                                    </div>
                                ))}
                                {selectedNode?.source_id.length === 0 && (
                                    <span className="text-gray-400 text-sm">无关联文档</span>
                                )}
                            </div>
                        </div>

                        <Separator />

                        <div>
                            <h4 className="text-sm font-medium text-gray-500 mb-2">关联关系</h4>
                            {/* Find links connected to this node */}
                            <div className="space-y-2">
                                {data.links.filter((l: any) => l.source.id === selectedNode?.id || l.target.id === selectedNode?.id).map((l: any, idx) => {
                                    const isSource = l.source.id === selectedNode?.id;
                                    const otherNode = isSource ? l.target : l.source;
                                    return (
                                        <div key={idx} className="flex items-center justify-between p-2 rounded bg-gray-50 text-sm">
                                            <span className="text-gray-600 w-1/3 truncate text-right">{isSource ? 'This' : otherNode.name}</span>
                                            <span className="px-2 text-xs text-gray-400">--- {l.description} ({l.weight}) ---&gt;</span>
                                            <span className="text-gray-900 w-1/3 truncate font-medium">{isSource ? otherNode.name : 'This'}</span>
                                        </div>
                                    )
                                })}
                            </div>
                        </div>

                    </div>
                </SheetContent>
            </Sheet>
        </div>
    );
}
