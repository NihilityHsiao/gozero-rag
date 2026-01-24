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
        // --- Person (人物) ---
        {
            id: "孙悟空",
            name: "孙悟空",
            type: "person",
            description: "法号行者，唐僧的大徒弟，会七十二变、筋斗云。曾大闹天宫，后被压在五行山下，最终护送唐僧西天取经。",
            val: 25,
            source_id: ["doc-001", "doc-002", "doc-003"]
        },
        {
            id: "唐三藏",
            name: "唐三藏",
            type: "person",
            description: "前世为如来佛祖弟子金蝉子，转世为唐朝高僧。也就是唐僧，性格慈悲但有时迂腐，是取经团队的核心领导者。",
            val: 20,
            source_id: ["doc-001", "doc-004"]
        },
        {
            id: "猪八戒",
            name: "猪八戒",
            type: "person",
            description: "法号悟能，原是天蓬元帅，因调戏嫦娥被贬下凡。性格贪吃好色但也憨厚有力，是团队中的开心果。",
            val: 15,
            source_id: ["doc-001", "doc-005"]
        },
        {
            id: "沙悟净",
            name: "沙悟净",
            type: "person",
            description: "原是卷帘大将，因打碎琉璃盏被贬流沙河。性格沉默寡言，任劳任怨，负责挑担和照顾师父。",
            val: 12,
            source_id: ["doc-001", "doc-006"]
        },
        {
            id: "白骨精",
            name: "白骨精",
            type: "person",
            description: "又称白骨夫人，本是白虎岭上的一具化为白骨的女尸，采天地灵气受日月精华变幻成人型，擅长变化。",
            val: 10,
            source_id: ["doc-007"]
        },
        {
            id: "牛魔王",
            name: "牛魔王",
            type: "person",
            description: "自号平天大圣，翠云山和积雷山的主人，孙悟空的结拜大哥，实力强大的妖王。",
            val: 14,
            source_id: ["doc-008"]
        },
        {
            id: "铁扇公主",
            name: "铁扇公主",
            type: "person",
            description: "又叫罗刹女，牛魔王的妻子，持有芭蕉扇，掌管火焰山一带的气候。",
            val: 12,
            source_id: ["doc-008"]
        },
        {
            id: "如来佛祖",
            name: "如来佛祖",
            type: "person",
            description: "西方极乐世界释迦牟尼尊者，法力无边，是取经行动的最终决策者。",
            val: 22,
            source_id: ["doc-009"]
        },
        {
            id: "观音菩萨",
            name: "观音菩萨",
            type: "person",
            description: "大慈大悲救苦救难观世音菩萨，取经项目的实际执行总监，多次帮助师徒四人度过难关。",
            val: 20,
            source_id: ["doc-009", "doc-001"]
        },

        // --- Geo (地点) ---
        {
            id: "花果山",
            name: "花果山",
            type: "geo",
            description: "位于东胜神洲傲来国，是孙悟空的出生地和老家，也是水帘洞的所在地。",
            val: 15,
            source_id: ["doc-002"]
        },
        {
            id: "天庭",
            name: "天庭",
            type: "geo", // Geo or Organization, kept as Geo for location context
            description: "掌管三界的最高行政区域，位于天界三十三层天之上。",
            val: 18,
            source_id: ["doc-003"]
        },
        {
            id: "西天大雷音寺",
            name: "西天大雷音寺",
            type: "geo",
            description: "位于西牛贺洲灵山，是如来佛祖说法的道场，取经的终点。",
            val: 18,
            source_id: ["doc-009"]
        },
        {
            id: "东海龙宫",
            name: "东海龙宫",
            type: "geo",
            description: "东海龙王的宫殿，孙悟空曾在此索取兵器。",
            val: 12,
            source_id: ["doc-002"]
        },

        // --- Organization (组织) ---
        {
            id: "天宫众仙",
            name: "天宫众仙",
            type: "organization",
            description: "由玉皇大帝统领的仙官体系，包括托塔天王、太上老君等。",
            val: 16,
            source_id: ["doc-003"]
        },
        {
            id: "取经团队",
            name: "取经团队",
            type: "organization",
            description: "为了前往西天求取真经而组成的四人一马特别行动小组。",
            val: 20,
            source_id: ["doc-001"]
        },
        {
            id: "灵山佛界",
            name: "灵山佛界",
            type: "organization",
            description: "以如来佛祖为首的佛教组织体系。",
            val: 18,
            source_id: ["doc-009"]
        },

        // --- Event (事件) ---
        {
            id: "大闹天宫",
            name: "大闹天宫",
            type: "event",
            description: "孙悟空因不满天庭官职低微及未被邀请参加蟠桃会，反出天庭，与十万天兵天将大战的历史性事件。",
            val: 18,
            source_id: ["doc-003"]
        },
        {
            id: "三打白骨精",
            name: "三打白骨精",
            type: "event",
            description: "白骨精三次变身欺骗唐僧，均被孙悟空识破并打死，导致唐僧误会并将悟空逐出师门的悲剧事件。",
            val: 14,
            source_id: ["doc-007"]
        },
        {
            id: "西天取经",
            name: "西天取经",
            type: "event",
            description: "唐僧师徒四人历经九九八十一难，前往西天大雷音寺求取真经的伟大征程。",
            val: 25,
            source_id: ["doc-001"]
        },

        // --- Concept/Category (概念) ---
        {
            id: "佛法",
            name: "佛法",
            type: "concept",
            description: "佛教的教义，普度众生，劝人向善。",
            val: 15,
            source_id: ["doc-009"]
        },
        {
            id: "七十二变",
            name: "七十二变",
            type: "concept",
            description: "地煞七十二术，孙悟空的高级法术技能。",
            val: 12,
            source_id: ["doc-002"]
        },
        {
            id: "长生不老",
            name: "长生不老",
            type: "concept",
            description: "修行者追求的终极目标，也是众多妖精想吃唐僧肉的原因。",
            val: 14,
            source_id: ["doc-007"]
        },

        // --- Product (物品/法宝) ---
        {
            id: "如意金箍棒",
            name: "如意金箍棒",
            type: "product",
            description: "原是太上老君冶炼的神铁，后被大禹借走治水，珍藏于东海龙宫，最终成为孙悟空的兵器。",
            val: 16,
            source_id: ["doc-002"]
        },
        {
            id: "九齿钉耙",
            name: "九齿钉耙",
            type: "product",
            description: "又名上宝沁金琶，太上老君用神冰铁亲自锤炼，借五方五帝、六丁六甲之力锻造而成，猪八戒的武器。",
            val: 12,
            source_id: ["doc-005"]
        },
        {
            id: "紧箍咒",
            name: "紧箍咒",
            type: "product",
            description: "观音菩萨赐给唐僧用于管教孙悟空的法宝，也就是“定心真言”。",
            val: 14,
            source_id: ["doc-001"]
        },
        {
            id: "芭蕉扇",
            name: "芭蕉扇",
            type: "product",
            description: "铁扇公主的宝物，能扇灭火焰山的八百里火焰。",
            val: 12,
            source_id: ["doc-008"]
        }
    ],
    links: [
        // 核心人物关系
        { source: "孙悟空", target: "唐三藏", description: "师徒/保镖", weight: 10 },
        { source: "猪八戒", target: "唐三藏", description: "师徒", weight: 8 },
        { source: "沙悟净", target: "唐三藏", description: "师徒", weight: 8 },
        { source: "孙悟空", target: "猪八戒", description: "师兄弟/互怼", weight: 9 },
        { source: "猪八戒", target: "沙悟净", description: "师兄弟", weight: 7 },
        { source: "孙悟空", target: "沙悟净", description: "师兄弟", weight: 7 },
        { source: "唐三藏", target: "取经团队", description: "领导者", weight: 10 },
        { source: "孙悟空", target: "取经团队", description: "核心骨干", weight: 10 },

        // 亲缘/社交关系
        { source: "牛魔王", target: "孙悟空", description: "昔日结拜兄弟", weight: 6 },
        { source: "牛魔王", target: "铁扇公主", description: "夫妻", weight: 9 },
        { source: "观音菩萨", target: "孙悟空", description: "点化/教导", weight: 8 },
        { source: "观音菩萨", target: "唐三藏", description: "指引", weight: 8 },
        { source: "如来佛祖", target: "孙悟空", description: "压制/收服", weight: 9 },

        // 地点关联
        { source: "孙悟空", target: "花果山", description: "家乡/根据地", weight: 10 },
        { source: "孙悟空", target: "天庭", description: "任职/反叛", weight: 8 },
        { source: "孙悟空", target: "东海龙宫", description: "强夺兵器", weight: 7 },
        { source: "唐三藏", target: "西天大雷音寺", description: "目的地", weight: 10 },

        // 物品关联
        { source: "孙悟空", target: "如意金箍棒", description: "持有", weight: 10 },
        { source: "猪八戒", target: "九齿钉耙", description: "持有", weight: 9 },
        { source: "铁扇公主", target: "芭蕉扇", description: "持有", weight: 9 },
        { source: "唐三藏", target: "紧箍咒", description: "使用", weight: 8 },
        { source: "紧箍咒", target: "孙悟空", description: "束缚", weight: 9 },

        // 事件关联
        { source: "孙悟空", target: "大闹天宫", description: "发起者", weight: 10 },
        { source: "天宫众仙", target: "大闹天宫", description: "镇压", weight: 8 },
        { source: "孙悟空", target: "三打白骨精", description: "主角", weight: 9 },
        { source: "白骨精", target: "三打白骨精", description: "反派", weight: 9 },
        { source: "取经团队", target: "西天取经", description: "执行", weight: 10 },

        // 概念/组织关联
        { source: "唐三藏", target: "佛法", description: "信仰", weight: 9 },
        { source: "如来佛祖", target: "灵山佛界", description: "统治", weight: 10 },
        { source: "灵山佛界", target: "佛法", description: "传承", weight: 10 },
        { source: "孙悟空", target: "七十二变", description: "技能", weight: 9 },
        { source: "白骨精", target: "长生不老", description: "欲望", weight: 8 }
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
    const [highlightLinks, setHighlightLinks] = useState(new Set<string>());
    const [hoverNode, setHoverNode] = useState<GraphNode | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedNode, setSelectedNode] = useState<GraphNode | null>(null);

    useEffect(() => {
        // Simulate loading data and preprocess
        setTimeout(() => {
            const data = MOCK_DATA as any;

            // Calculate degrees
            const degrees: Record<string, number> = {};
            data.nodes.forEach((n: any) => degrees[n.id] = 0);
            data.links.forEach((l: any) => {
                degrees[l.source] = (degrees[l.source] || 0) + 1;
                degrees[l.target] = (degrees[l.target] || 0) + 1;
            });


            // Assign degree to val for sizing if not present, or use as factor
            data.nodes.forEach((n: any) => {
                n.val = n.val || (degrees[n.id] * 2) || 1;
            });

            setData(data);
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
                    newHighlightLinks.add(`${link.source.id}-${link.target.id}`);
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
            const gNode = node as GraphNode;
            if (fgRef.current && gNode.x !== undefined && gNode.y !== undefined) {
                fgRef.current.centerAt(gNode.x, gNode.y, 1000);
                fgRef.current.zoom(4, 2000);
                handleNodeClick(gNode);
            }
        }
    }

    const paintNode = useCallback((node: any, ctx: CanvasRenderingContext2D, globalScale: number) => {
        const isHover = highlightNodes.has(node.id) || node === hoverNode;
        const isSelected = selectedNode?.id === node.id;
        const isRelevant = isHover || isSelected;
        const isBackground = (hoverNode || selectedNode) && !isRelevant; // Focus Mode active but node not relevant

        // --- Focus Mode Logic ---
        // If we are in focus mode (some node is hovered/selected), non-relevant nodes fade out and turn grayscale
        if (isBackground) {
            ctx.globalAlpha = 0.2;
            ctx.filter = 'grayscale(100%)';
        } else {
            ctx.globalAlpha = 1;
            ctx.filter = 'none';
        }

        const label = node.name;
        const fontSize = 12 / globalScale;
        ctx.font = `${fontSize}px Sans-Serif`;

        const color = TYPE_COLORS[node.type] || TYPE_COLORS.default;

        // Dynamic Radius with Breathing Effect for Active Node
        const baseRadius = Math.sqrt(node.val) * 2;
        let r = isBackground ? baseRadius * 0.5 : baseRadius; // Shrink background nodes

        if (isSelected || (isHover && !isBackground)) {
            // Breathing animation calculation (simulated with time if possible, or static burst)
            // For simple canvas inside paintNode without continuous loop, we typically use a static larger size or rely on engine tick
            // To fake breathing, we might need a time param, but for strict prop compliance we stick to highlighting size
            r = baseRadius * 1.2;
        }

        // Draw Glow (Shadow) for Relevant Nodes
        if (isRelevant) {
            ctx.shadowColor = color;
            ctx.shadowBlur = 15; // Neon glow
        } else {
            ctx.shadowBlur = 0;
        }

        // Draw Circle
        ctx.beginPath();
        ctx.arc(node.x, node.y, r, 0, 2 * Math.PI, false);
        ctx.fillStyle = color;
        ctx.fill();

        // Reset Shadow for subsequent draws (text/rings)
        ctx.shadowBlur = 0;

        // Draw Ring if selected (Inner White Ring)
        if (isRelevant) {
            ctx.beginPath();
            ctx.arc(node.x, node.y, r, 0, 2 * Math.PI, false);
            ctx.strokeStyle = 'rgba(255, 255, 255, 0.5)';
            ctx.lineWidth = 1 / globalScale;
            ctx.stroke();
        }

        // Draw Label: Only for relevant nodes or high zoom, OR large nodes
        // In focus mode, hide background labels to reduce noise
        const shouldDrawLabel = isRelevant || (globalScale > 1.5 && !isBackground) || (node.val > 15 && !isBackground);

        if (shouldDrawLabel) {
            ctx.fillStyle = isBackground ? '#94A3B8' : '#334155'; // Faded text for background
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(label, node.x, node.y + r + fontSize);
        }

        // Reset Filter
        ctx.filter = 'none';

    }, [highlightNodes, hoverNode, selectedNode]);

    return (
        <div className="flex h-[calc(100vh-140px)] gap-4">
            <Card className="flex-1 relative overflow-hidden bg-[#F9FAFB] border border-gray-100 shadow-sm rounded-xl">
                {/* Toolbar (HUD Style) */}
                <div className="absolute top-4 left-4 z-10 flex flex-col gap-3 w-72 pointer-events-none">
                    <div className="pointer-events-auto backdrop-blur-sm bg-white/70 border border-white/50 shadow-[0_8px_32px_rgba(0,0,0,0.05)] rounded-xl p-1 flex gap-2">
                        <Input
                            placeholder="搜索节点..."
                            className="border-0 bg-transparent focus-visible:ring-0 text-slate-800 placeholder:text-slate-400 h-9"
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && handleSearch()}
                        />
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-slate-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors" onClick={handleSearch}>
                            <Search size={18} />
                        </Button>
                    </div>

                    <div className="pointer-events-auto backdrop-blur-md bg-white/70 border border-white/50 shadow-[0_8px_32px_rgba(0,0,0,0.05)] rounded-xl p-1.5 flex gap-1 w-fit">
                        <Button size="icon" variant="ghost" className="h-8 w-8 text-slate-600 hover:bg-slate-100/50 rounded-lg" onClick={() => fgRef.current?.zoomToFit(400)}>
                            <Maximize2 size={16} />
                        </Button>
                        <Button size="icon" variant="ghost" className="h-8 w-8 text-slate-600 hover:bg-slate-100/50 rounded-lg" onClick={() => fgRef.current?.zoom(fgRef.current.zoom() * 1.2, 400)}>
                            <ZoomIn size={16} />
                        </Button>
                        <Button size="icon" variant="ghost" className="h-8 w-8 text-slate-600 hover:bg-slate-100/50 rounded-lg" onClick={() => fgRef.current?.zoom(fgRef.current.zoom() / 1.2, 400)}>
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
                    linkDirectionalParticles={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        return highlightLinks.has(idStr) ? 4 : 0; // 4 particles for highlight links
                    }}
                    linkDirectionalParticleWidth={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        return highlightLinks.has(idStr) ? 4 : 0;
                    }}
                    linkDirectionalParticleSpeed={0.005}
                    backgroundColor="#F9FAFB" // Light Gray Background
                    cooldownTicks={100}
                    onEngineStop={() => fgRef.current?.zoomToFit(400)}
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
