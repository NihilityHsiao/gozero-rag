import { useEffect, useState, useRef, useCallback } from 'react';
import ForceGraph3D from 'react-force-graph-3d';
import * as THREE from 'three';
import SpriteText from 'three-spritetext';
import { Loader2 } from 'lucide-react';
import { useParams, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import { getKnowledgeGraph, searchKnowledgeGraph } from '@/api/knowledge';
import type { GraphNode, GraphLink } from '@/types/graph';

// Import New Components
import { FloatingSearch } from './GraphComponents/FloatingSearch';
import { GraphStats } from './GraphComponents/GraphStats';
import { GraphControls } from './GraphComponents/GraphControls';
import { GraphDetailPanel } from './GraphComponents/GraphDetailPanel';
import { Button } from '@/components/ui/button';

// Node Colors by Type (Dify-like Light Theme Palette) - Reused
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
    const navigate = useNavigate();
    const fgRef = useRef<any>(undefined);

    // State
    const [data, setData] = useState<{ nodes: GraphNode[]; links: GraphLink[] }>({ nodes: [], links: [] });
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
            // Process nodes (calculate degrees if needed for sizing)
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

    // Graph Interaction Handlers
    const handleNodeClick = useCallback((node: any) => {
        setSelectedNode(node);

        if (fgRef.current) {
            // Aim at node from outside it
            const distance = 60;
            const distRatio = 1 + distance / Math.hypot(node.x, node.y, node.z);

            fgRef.current.cameraPosition(
                { x: node.x * distRatio, y: node.y * distRatio, z: node.z * distRatio }, // new position
                node, // lookAt ({ x, y, z })
                1500  // ms transition duration
            );
        }
    }, []);

    const handleNodeFocus = useCallback(() => {
        if (selectedNode && fgRef.current) {
            const distance = 40;
            const distRatio = 1 + distance / Math.hypot(selectedNode.x, selectedNode.y, selectedNode.z);

            fgRef.current.cameraPosition(
                { x: selectedNode.x * distRatio, y: selectedNode.y * distRatio, z: selectedNode.z * distRatio },
                selectedNode,
                1000
            );
        }
    }, [selectedNode]);

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

    // Node Object Generator (Enhanced Material)
    const nodeThreeObject = useCallback((node: any) => {
        const group = new THREE.Group();

        // 1. The Sphere
        const radius = Math.sqrt(node.val) * 1.5;
        const color = TYPE_COLORS[node.type] || TYPE_COLORS.default;

        const geometry = new THREE.SphereGeometry(radius);
        const material = new THREE.MeshLambertMaterial({
            color: color,
            transparent: true,
            opacity: 0.9,
            emissive: color,
            emissiveIntensity: 0.2
        });
        const sphere = new THREE.Mesh(geometry, material);
        group.add(sphere);

        // 2. The Text Label (SpriteText)
        const sprite = new SpriteText(node.name);
        sprite.color = 'rgba(255, 255, 255, 0.95)';
        sprite.textHeight = 4 + (node.val * 0.2); // Scale text slightly with value
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

    // Frame Update for Focus/Highlight Mode
    useEffect(() => {
        if (!data.nodes.length) return;

        data.nodes.forEach((node: any) => {
            const sphere = node.__sphere;
            const sprite = node.__sprite;
            if (!sphere || !sprite) return; // Wait for initialization

            const isHovered = hoverNode === node;
            const isSelected = selectedNode?.id === node.id;
            const isNeighbor = highlightNodes.has(node.id);
            const isRelevant = isHovered || isSelected || isNeighbor;
            const hasFocus = (hoverNode || selectedNode) !== null;

            if (hasFocus && !isRelevant) {
                // Dim irrelevant nodes
                sphere.material.opacity = 0.1;
                sphere.material.emissiveIntensity = 0;
                sprite.material.opacity = 0.2; // Fade text but keep visible
            } else {
                // Active nodes
                const originalColor = TYPE_COLORS[node.type] || TYPE_COLORS.default;
                sphere.material.opacity = 0.9;
                sphere.material.color.set(originalColor);

                // Highlight Effect
                if (isRelevant) {
                    sphere.material.emissive.set(originalColor);
                    sphere.material.emissiveIntensity = isSelected ? 0.8 : 0.5;
                    sprite.material.opacity = 1;
                } else {
                    sphere.material.emissiveIntensity = 0.2;
                    sprite.material.opacity = 0.9;
                }
            }
        });

    }, [hoverNode, selectedNode, highlightNodes, data.nodes]);

    // Starfield Background Effect
    useEffect(() => {
        if (!fgRef.current) return;
        const scene = fgRef.current.scene();

        // Create Stars
        const starsGeometry = new THREE.BufferGeometry();
        const starsCount = 2000;
        const posArray = new Float32Array(starsCount * 3);

        for (let i = 0; i < starsCount * 3; i++) {
            posArray[i] = (Math.random() - 0.5) * 3000;
        }

        starsGeometry.setAttribute('position', new THREE.BufferAttribute(posArray, 3));

        // Star Material
        const starsMaterial = new THREE.PointsMaterial({
            size: 1.5,
            color: 0x6366f1, // Indigo tint
            transparent: true,
            opacity: 0.6,
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
            fetchData();
            return;
        }

        try {
            const res = await searchKnowledgeGraph(kbId, { q: searchQuery });
            // Reuse processing logic
            // Note: In real app, consider extracting this duplicate logic
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

    // Zoom Controls
    const handleZoomIn = () => {
        if (fgRef.current) {
            const currentPos = fgRef.current.cameraPosition();
            const target = fgRef.current.controls().target;
            const v = { x: currentPos.x - target.x, y: currentPos.y - target.y, z: currentPos.z - target.z };
            fgRef.current.cameraPosition(
                { x: target.x + v.x * 0.6, y: target.y + v.y * 0.6, z: target.z + v.z * 0.6 },
                target, 400
            );
        }
    };

    const handleZoomOut = () => {
        if (fgRef.current) {
            const currentPos = fgRef.current.cameraPosition();
            const target = fgRef.current.controls().target;
            const v = { x: currentPos.x - target.x, y: currentPos.y - target.y, z: currentPos.z - target.z };
            fgRef.current.cameraPosition(
                { x: target.x + v.x * 1.4, y: target.y + v.y * 1.4, z: target.z + v.z * 1.4 },
                target, 400
            );
        }
    };

    const handleZoomToFit = () => {
        if (fgRef.current) fgRef.current.zoomToFit(1000, 50);
    };

    return (
        <div className="relative w-full h-[calc(100vh-64px)] overflow-hidden bg-black">
            {/* Deep Space Background Gradient */}
            <div
                className="absolute inset-0 pointer-events-none z-0"
                style={{
                    background: 'radial-gradient(circle at center, #1e1b4b 0%, #020617 100%)' // Indigo-950 to Slate-950
                }}
            />

            {/* Loading Overlay */}
            {loading && (
                <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-black/50 backdrop-blur-sm text-white">
                    <Loader2 className="h-10 w-10 text-indigo-500 animate-spin mb-4" />
                    <p className="text-slate-300 tracking-wider font-light">Loading Neural Network...</p>
                </div>
            )}

            {/* Empty State */}
            {!loading && data.nodes.length === 0 && (
                <div className="absolute inset-0 z-40 flex flex-col items-center justify-center pointer-events-none">
                    <div className="bg-slate-900/60 backdrop-blur-md p-8 rounded-2xl border border-slate-700/50 text-center pointer-events-auto max-w-md shadow-2xl">
                        <div className="text-indigo-400 mb-3 text-lg font-medium">No Knowledge Graph Data</div>
                        <p className="text-slate-400 mb-6 font-light">No entities found. Ensure documents are parsed with extraction enabled.</p>
                        <Button
                            variant="outline"
                            className="border-indigo-500/50 text-indigo-300 hover:text-white hover:bg-indigo-500/20"
                            onClick={fetchData}
                        >
                            Reload Graph
                        </Button>
                    </div>
                </div>
            )}

            {/* Floating UI Layers */}
            <FloatingSearch
                searchQuery={searchQuery}
                onSearchChange={setSearchQuery}
                onSearchEnter={handleSearch}
            />

            <GraphStats
                nodeCount={data.nodes.length}
                linkCount={data.links.length}
            />

            <GraphControls
                onZoomIn={handleZoomIn}
                onZoomOut={handleZoomOut}
                onZoomToFit={handleZoomToFit}
                onFocus={handleNodeFocus}
            />

            <GraphDetailPanel
                node={selectedNode}
                links={data.links}
                typeColors={TYPE_COLORS}
                onClose={() => setSelectedNode(null)}
                onNavigateToNode={(nodeId) => {
                    // Find the node object in data
                    const target = data.nodes.find(n => n.id === nodeId);
                    if (target) handleNodeClick(target);
                }}
            />

            {/* 3D Force Graph */}
            <ForceGraph3D
                ref={fgRef}
                graphData={data}
                nodeLabel="name"
                nodeThreeObject={nodeThreeObject}
                onNodeClick={handleNodeClick}
                onNodeHover={handleNodeHover}
                backgroundColor="rgba(0,0,0,0)"
                showNavInfo={false}
                linkColor={link => {
                    // @ts-ignore
                    const idStr = `${link.source.id}-${link.target.id}`;
                    // Highlight: Bright Cyan, Normal: Faint Indigo
                    return highlightLinks.has(idStr) ? '#22d3ee' : 'rgba(99, 102, 241, 0.15)';
                }}
                linkWidth={link => {
                    // @ts-ignore
                    const idStr = `${link.source.id}-${link.target.id}`;
                    return highlightLinks.has(idStr) ? 2 : 0.5;
                }}
                linkDirectionalParticles={2}
                linkDirectionalParticleSpeed={0.005}
                linkDirectionalParticleWidth={link => {
                    // @ts-ignore
                    const idStr = `${link.source.id}-${link.target.id}`;
                    return highlightLinks.has(idStr) ? 3 : 1;
                }}
                linkOpacity={0.3}
                controlType="orbit"
            />
        </div>
    );
}
