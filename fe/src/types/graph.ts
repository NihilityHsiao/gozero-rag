export interface GraphNode {
    id: string;
    name: string;
    type: string;
    description: string;
    val: number;
    source_id: string[];
    // Frontend specific properties for 3D graph
    x?: number;
    y?: number;
    z?: number;
    color?: string;
}

export interface GraphLink {
    source: string | GraphNode; // Can be ID string or Node object after processing
    target: string | GraphNode;
    description: string;
    weight: number;
}

export interface GraphDetailResp {
    nodes: GraphNode[];
    links: GraphLink[];
}

export interface GraphReq {
    kb_id: string;
    limit?: number;
}

export interface GraphSearchReq {
    kb_id: string;
    query: string;
}
