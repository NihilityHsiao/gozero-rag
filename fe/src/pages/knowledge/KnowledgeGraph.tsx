import { useEffect, useState, useRef, useCallback } from 'react';
import ForceGraph3D from 'react-force-graph-3d';
import * as THREE from 'three';
import SpriteText from 'three-spritetext';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Search, ZoomIn, ZoomOut, Maximize, Loader2 } from 'lucide-react';
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

import { useParams } from 'react-router-dom';
import { getKnowledgeGraph, searchKnowledgeGraph } from '@/api/knowledge';
import type { GraphNode, GraphLink } from '@/types';
import { toast } from 'sonner';

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

export default function KnowledgeGraph() {
    const { id: kbId } = useParams();
    const fgRef = useRef<any>(undefined);
    const [data, setData] = useState<{ nodes: GraphNode[]; links: GraphLink[] }>({ nodes: [], links: [] });
    // Highlight Set must be efficient
    const [highlightNodes, setHighlightNodes] = useState(new Set<string>());
    const [highlightLinks, setHighlightLinks] = useState(new Set<string>());
    const [hoverNode, setHoverNode] = useState<any | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedNode, setSelectedNode] = useState<any | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchData = useCallback(async () => {
        if (!kbId) return;
        setLoading(true);
        try {
            const res = await getKnowledgeGraph(kbId);
            // Process nodes to ensure val exists or calculate degrees if needed
            const degrees: Record<string, number> = {};
            res.links.forEach((l: any) => {
                // l.source is string initially
                const s = typeof l.source === 'object' ? l.source.id : l.source;
                const t = typeof l.target === 'object' ? l.target.id : l.target;
                degrees[s] = (degrees[s] || 0) + 1;
                degrees[t] = (degrees[t] || 0) + 1;
            });
            res.nodes.forEach((n: any) => {
                n.val = n.val || (degrees[n.id] * 2) || 1;
            });
            setData(res);
        } catch (error) {
            console.error("Failed to fetch graph:", error);
            toast.error("获取知识图谱失败");
        } finally {
            setLoading(false);
        }
    }, [kbId]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const handleNodeClick = useCallback((node: any) => {
        setSelectedNode(node);

        if (fgRef.current) {
            // Aim at node from outside it
            const distance = 40;
            const distRatio = 1 + distance / Math.hypot(node.x, node.y, node.z);

            fgRef.current.cameraPosition(
                { x: node.x * distRatio, y: node.y * distRatio, z: node.z * distRatio }, // new position
                node, // lookAt ({ x, y, z })
                1000  // ms transition duration
            );
        }
    }, []);

    const handleNodeHover = (node: any | null) => {
        if ((!node && !hoverNode) || (node && hoverNode && node.id === hoverNode.id)) return;

        setHoverNode(node || null);
        const newHighlightNodes = new Set<string>();
        const newHighlightLinks = new Set<string>();

        if (node) {
            newHighlightNodes.add(node.id);
            data.links.forEach((link: any) => {
                const isDirectLink = link.source.id === node.id || link.target.id === node.id;
                if (isDirectLink) {
                    newHighlightLinks.add(`${link.source.id}-${link.target.id}`);
                    const neighborId = link.source.id === node.id ? link.target.id : link.source.id;
                    newHighlightNodes.add(neighborId);
                }
            });
        }
        setHighlightNodes(newHighlightNodes);
        setHighlightLinks(newHighlightLinks);
    };

    // Node Object Generator
    const nodeThreeObject = useCallback((node: any) => {
        const group = new THREE.Group();

        // 1. The Sphere
        const radius = Math.sqrt(node.val) * 1.5;
        const color = TYPE_COLORS[node.type] || TYPE_COLORS.default;

        const geometry = new THREE.SphereGeometry(radius);
        const material = new THREE.MeshLambertMaterial({
            color: color,
            transparent: true,
            opacity: 0.9
        });
        const sphere = new THREE.Mesh(geometry, material);
        group.add(sphere);

        // 2. The Text Label (SpriteText)
        const sprite = new SpriteText(node.name);
        sprite.color = 'rgba(255, 255, 255, 0.9)'; // Bright white for contrast against deep space
        sprite.textHeight = 4;
        sprite.position.set(0, radius + 4, 0);
        group.add(sprite);

        // Save reference for updates
        // @ts-ignore
        node.__threeObj = group;
        // @ts-ignore
        node.__sphere = sphere;
        // @ts-ignore
        node.__sprite = sprite;

        return group;
    }, []);

    // Frame Update for Focus Mode (Performance Critical)
    useEffect(() => {
        if (!data.nodes.length) return;

        data.nodes.forEach((node: any) => {
            const sphere = node.__sphere;
            const sprite = node.__sprite;
            if (!sphere || !sprite) return;

            const isHovered = hoverNode === node;
            const isSelected = selectedNode?.id === node.id;
            const isNeighbor = highlightNodes.has(node.id);
            const isRelevant = isHovered || isSelected || isNeighbor;
            const hasFocus = (hoverNode || selectedNode) !== null;

            if (hasFocus && !isRelevant) {
                // Dim
                sphere.material.opacity = 0.1;
                sphere.material.color.set('#cbd5e1'); // Light gray dim
                sprite.visible = false;
            } else {
                // Active
                const originalColor = TYPE_COLORS[node.type] || TYPE_COLORS.default;
                sphere.material.opacity = 0.9;
                sphere.material.color.set(originalColor);
                sprite.visible = true;

                if (isRelevant) {
                    sphere.material.emissive.set(originalColor);
                    sphere.material.emissiveIntensity = 0.6; // Increased glow
                } else {
                    sphere.material.emissiveIntensity = 0;
                }
            }
        });

    }, [hoverNode, selectedNode, highlightNodes, data.nodes]);

    // Starfield Effect
    useEffect(() => {
        if (!fgRef.current) return;

        const scene = fgRef.current.scene();

        // Create Stars
        const starsGeometry = new THREE.BufferGeometry();
        const starsCount = 1500;
        const posArray = new Float32Array(starsCount * 3);

        for (let i = 0; i < starsCount * 3; i++) {
            // Spread stars far away
            posArray[i] = (Math.random() - 0.5) * 2000;
        }

        starsGeometry.setAttribute('position', new THREE.BufferAttribute(posArray, 3));

        // Star Material
        const starsMaterial = new THREE.PointsMaterial({
            size: 2,
            color: 0x4f90ff, // Light Blue tint
            transparent: true,
            opacity: 0.8,
            blending: THREE.AdditiveBlending,
            sizeAttenuation: true
        });

        const starField = new THREE.Points(starsGeometry, starsMaterial);
        scene.add(starField);

        return () => {
            scene.remove(starField);
            starsGeometry.dispose();
            starsMaterial.dispose();
        };

    }, []);


    const handleSearch = async () => {
        if (!kbId) return;
        if (!searchQuery.trim()) {
            fetchData(); // Reset to full graph
            return;
        }

        try {
            const res = await searchKnowledgeGraph(kbId, { q: searchQuery });
            // Similar processing if needed
            const degrees: Record<string, number> = {};
            res.links.forEach((l: any) => {
                const s = typeof l.source === 'object' ? l.source.id : l.source;
                const t = typeof l.target === 'object' ? l.target.id : l.target;
                degrees[s] = (degrees[s] || 0) + 1;
                degrees[t] = (degrees[t] || 0) + 1;
            });
            res.nodes.forEach((n: any) => {
                n.val = n.val || (degrees[n.id] * 2) || 1;
            });
            setData(res);
            toast.success(`Found ${res.nodes.length} nodes`);
        } catch (error) {
            console.error("Search failed:", error);
            toast.error("搜索失败");
        }
    };

    const handleZoomIn = () => {
        if (fgRef.current) {
            const currentPos = fgRef.current.cameraPosition();
            const target = fgRef.current.controls().target;

            // Vector from Target to Camera
            const v = {
                x: currentPos.x - target.x,
                y: currentPos.y - target.y,
                z: currentPos.z - target.z
            };

            // Zoom In: multiple by < 1
            fgRef.current.cameraPosition(
                { x: target.x + v.x * 0.6, y: target.y + v.y * 0.6, z: target.z + v.z * 0.6 },
                target, // lookAt
                400
            );
        }
    };

    const handleZoomOut = () => {
        if (fgRef.current) {
            const currentPos = fgRef.current.cameraPosition();
            const target = fgRef.current.controls().target;

            // Vector from Target to Camera
            const v = {
                x: currentPos.x - target.x,
                y: currentPos.y - target.y,
                z: currentPos.z - target.z
            };

            // Zoom Out: multiply by > 1
            fgRef.current.cameraPosition(
                { x: target.x + v.x * 1.4, y: target.y + v.y * 1.4, z: target.z + v.z * 1.4 },
                target, // lookAt
                400
            );
        }
    };

    const handleZoomToFit = () => {
        if (fgRef.current) {
            fgRef.current.zoomToFit(1000, 50);
        }
    };

    return (
        <div className="flex h-[calc(100vh-140px)] gap-4">
            <Card className="flex-1 relative overflow-hidden border border-gray-800 shadow-xl rounded-xl"
                style={{
                    background: 'radial-gradient(circle at center, #0B1121 0%, #000000 100%)' // Deep Blue to Pure Black
                }}
            >
                {/* Optional: Vignette overlay for extra depth */}
                <div className="absolute inset-0 bg-[radial-gradient(transparent_0%,#000000_100%)] opacity-60 pointer-events-none" />

                {loading && (
                    <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-black/50 backdrop-blur-sm">
                        <Loader2 className="h-10 w-10 text-blue-500 animate-spin mb-4" />
                        <p className="text-gray-300">正在构建知识图谱...</p>
                    </div>
                )}

                {!loading && data.nodes.length === 0 && (
                    <div className="absolute inset-0 z-40 flex flex-col items-center justify-center pointer-events-none">
                        <div className="bg-black/40 backdrop-blur-md p-6 rounded-xl border border-white/10 text-center pointer-events-auto max-w-md">
                            <div className="text-gray-400 mb-2">暂无图谱数据</div>
                            <p className="text-sm text-gray-500 mb-4">请确保已上传并解析了相关文档 (NebulaGraph extraction required)</p>
                            <Button variant="outline" size="sm" onClick={fetchData}>
                                重新加载
                            </Button>
                        </div>
                    </div>
                )}

                <div className="absolute top-4 left-4 z-10 flex flex-col gap-3 w-auto min-w-[300px] pointer-events-none">
                    <div className="pointer-events-auto backdrop-blur-md bg-black/40 border border-white/10 shadow-lg rounded-xl p-1 flex gap-1 items-center">
                        <Input
                            placeholder="搜索节点..."
                            className="border-0 bg-transparent focus-visible:ring-0 text-white placeholder:text-gray-400 h-9 w-64"
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && handleSearch()}
                        />
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleSearch}>
                            <Search size={18} />
                        </Button>
                        <Separator orientation="vertical" className="h-6 bg-white/10 mx-1" />
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleZoomIn} title="放大">
                            <ZoomIn size={18} />
                        </Button>
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleZoomOut} title="缩小">
                            <ZoomOut size={18} />
                        </Button>
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleZoomToFit} title="全览">
                            <Maximize size={18} />
                        </Button>
                    </div>
                </div>

                <ForceGraph3D
                    ref={fgRef}
                    graphData={data}
                    nodeLabel="name"
                    nodeThreeObject={nodeThreeObject}
                    onNodeClick={handleNodeClick}
                    onNodeHover={handleNodeHover}
                    linkColor={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        // Highlight: Cyan Blue, Normal: Faint Blue
                        return highlightLinks.has(idStr) ? '#00e5ff' : 'rgba(100, 149, 237, 0.2)';
                    }}
                    linkWidth={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        return highlightLinks.has(idStr) ? 2 : 0.5;
                    }}
                    linkOpacity={0.5}
                    backgroundColor="rgba(0,0,0,0)" // Transparent to let radial gradient show
                    controlType="orbit"
                />
            </Card>

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
                                {selectedNode?.source_id?.map((id: string, index: number) => (
                                    <div key={index} className="flex items-center gap-2 p-2 rounded border border-gray-100 bg-white hover:bg-gray-50 cursor-pointer text-xs font-mono text-gray-600">
                                        <span className="w-2 h-2 rounded-full bg-blue-400"></span>
                                        {id}
                                    </div>
                                ))}
                                {(!selectedNode?.source_id || selectedNode?.source_id.length === 0) && (
                                    <span className="text-gray-400 text-sm">无关联文档</span>
                                )}
                            </div>
                        </div>

                        <Separator />

                        <div>
                            <h4 className="text-sm font-medium text-gray-500 mb-2">关联关系</h4>
                            {/* Find links connected to this node */}
                            <div className="space-y-2">
                                {data.links.filter((l: any) => l.source.id === selectedNode?.id || l.target.id === selectedNode?.id).map((l: any, idx: number) => {
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
