import { useEffect, useState, useRef, useCallback } from 'react';
import ForceGraph3D from 'react-force-graph-3d';
import * as THREE from 'three';
import SpriteText from 'three-spritetext';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Search, ZoomIn, ZoomOut, Maximize } from 'lucide-react';
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
            type: "geo",
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
        { source: "孙悟空", target: "唐三藏", description: "师徒/保镖", weight: 10 },
        { source: "猪八戒", target: "唐三藏", description: "师徒", weight: 8 },
        { source: "沙悟净", target: "唐三藏", description: "师徒", weight: 8 },
        { source: "孙悟空", target: "猪八戒", description: "师兄弟/互怼", weight: 9 },
        { source: "猪八戒", target: "沙悟净", description: "师兄弟", weight: 7 },
        { source: "孙悟空", target: "沙悟净", description: "师兄弟", weight: 7 },
        { source: "唐三藏", target: "取经团队", description: "领导者", weight: 10 },
        { source: "孙悟空", target: "取经团队", description: "核心骨干", weight: 10 },

        { source: "牛魔王", target: "孙悟空", description: "昔日结拜兄弟", weight: 6 },
        { source: "牛魔王", target: "铁扇公主", description: "夫妻", weight: 9 },
        { source: "观音菩萨", target: "孙悟空", description: "点化/教导", weight: 8 },
        { source: "观音菩萨", target: "唐三藏", description: "指引", weight: 8 },
        { source: "如来佛祖", target: "孙悟空", description: "压制/收服", weight: 9 },

        { source: "孙悟空", target: "花果山", description: "家乡/根据地", weight: 10 },
        { source: "孙悟空", target: "天庭", description: "任职/反叛", weight: 8 },
        { source: "孙悟空", target: "东海龙宫", description: "强夺兵器", weight: 7 },
        { source: "唐三藏", target: "西天大雷音寺", description: "目的地", weight: 10 },

        { source: "孙悟空", target: "如意金箍棒", description: "持有", weight: 10 },
        { source: "猪八戒", target: "九齿钉耙", description: "持有", weight: 9 },
        { source: "铁扇公主", target: "芭蕉扇", description: "持有", weight: 9 },
        { source: "唐三藏", target: "紧箍咒", description: "使用", weight: 8 },
        { source: "紧箍咒", target: "孙悟空", description: "束缚", weight: 9 },

        { source: "孙悟空", target: "大闹天宫", description: "发起者", weight: 10 },
        { source: "天宫众仙", target: "大闹天宫", description: "镇压", weight: 8 },
        { source: "孙悟空", target: "三打白骨精", description: "主角", weight: 9 },
        { source: "白骨精", target: "三打白骨精", description: "反派", weight: 9 },
        { source: "取经团队", target: "西天取经", description: "执行", weight: 10 },

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
    z?: number; // 3D
}

export default function KnowledgeGraph() {
    const fgRef = useRef<any>(undefined); // 3D Graph ref uses generic or specific ForceGraph3D type
    const [data, setData] = useState<any>({ nodes: [], links: [] });
    // Highlight Set must be efficient
    const [highlightNodes, setHighlightNodes] = useState(new Set<string>());
    const [highlightLinks, setHighlightLinks] = useState(new Set<string>());
    const [hoverNode, setHoverNode] = useState<any | null>(null);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedNode, setSelectedNode] = useState<any | null>(null);

    useEffect(() => {
        // Simulate loading data and preprocess
        setTimeout(() => {
            const data = JSON.parse(JSON.stringify(MOCK_DATA)); // Deep copy to avoid mutation issues

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

            // --- Fan Ren Xiu Xian Zhuan (Mortal's Journey) - Core Only ---
            const extraNodes: any[] = [];
            const extraLinks: any[] = [];

            // Helper
            const addGroup = (count: number, prefix: string, type: string, targetId: string, desc: string, relDesc: string = "隶属") => {
                for (let i = 1; i <= count; i++) {
                    const id = `${prefix} #${i}`;
                    extraNodes.push({ id: id, name: id, type, description: desc, val: Math.floor(Math.random() * 5) + 1, source_id: [] });
                    extraLinks.push({ source: id, target: targetId, description: relDesc, weight: 1 });
                }
            };

            // --- Fan Ren Xiu Xian Zhuan (Mortal's Journey) - Core Only ---
            extraNodes.push(
                { id: "韩立", name: "韩立", type: "person", description: "凡人修仙传主角，心思缜密。", val: 25, source_id: ["fr-001"] },
                { id: "南宫婉", name: "南宫婉", type: "person", description: "韩立的道侣。", val: 18, source_id: ["fr-002"] },
                { id: "厉飞雨", name: "厉飞雨", type: "person", description: "韩立的发小。", val: 15, source_id: ["fr-003"] },
                { id: "掌天瓶", name: "掌天瓶", type: "product", description: "韩立的外挂。", val: 20, source_id: ["fr-001"] },
                { id: "青竹蜂云剑", name: "青竹蜂云剑", type: "product", description: "韩立的本命法宝。", val: 18, source_id: ["fr-004"] },
                { id: "黄枫谷", name: "黄枫谷", type: "organization", description: "越国七大修仙门派之一。", val: 20, source_id: ["fr-005"] },
                { id: "乱星海", name: "乱星海", type: "geo", description: "修仙资源丰富的海外区域。", val: 22, source_id: ["fr-006"] },
                { id: "虚天殿", name: "虚天殿", type: "geo", description: "乱星海第一秘境。", val: 18, source_id: ["fr-007"] }
            );
            extraLinks.push(
                { source: "韩立", target: "南宫婉", description: "道侣", weight: 9 },
                { source: "韩立", target: "厉飞雨", description: "兄弟", weight: 8 },
                { source: "韩立", target: "掌天瓶", description: "持有", weight: 10 },
                { source: "韩立", target: "青竹蜂云剑", description: "本命法宝", weight: 10 },
                { source: "韩立", target: "黄枫谷", description: "入门", weight: 7 },
                { source: "韩立", target: "乱星海", description: "游历", weight: 8 },
                { source: "南宫婉", target: "黄枫谷", description: "邻宗", weight: 5 }
            );

            // Limited procedural additions for Fan Ren
            addGroup(40, "黄枫谷弟子", "person", "黄枫谷", "越国修仙者，资质各异。");
            addGroup(50, "乱星海妖兽", "person", "乱星海", "深海中的强大妖兽，浑身是宝。");
            addGroup(30, "噬金虫", "person", "韩立", "无物不噬的可怕奇虫，韩立的杀手锏。");

            // Re-add West Journey Groups
            addGroup(30, "花果山猴兵", "person", "花果山", "花果山的普通猴子猴孙。");
            addGroup(40, "天兵天将", "organization", "天庭", "镇守天庭的士兵。");
            addGroup(15, "盘丝洞蜘蛛精", "person", "唐三藏", "企图吃唐僧肉的女妖精。");


            // 1. 火影忍者 (Naruto)
            extraNodes.push(
                { id: "漩涡鸣人", name: "漩涡鸣人", type: "person", description: "火影忍者主角，七代火影。", val: 28, source_id: ["naruto-01"] },
                { id: "宇智波佐助", name: "宇智波佐助", type: "person", description: "鸣人的羁绊，支撑的一影。", val: 26, source_id: ["naruto-01"] },
                { id: "木叶村", name: "木叶村", type: "geo", description: "火之国隐村。", val: 25, source_id: ["naruto-geo"] },
                { id: "晓组织", name: "晓组织", type: "organization", description: "S级叛忍组织。", val: 24, source_id: ["naruto-org"] }
            );
            extraLinks.push(
                { source: "漩涡鸣人", target: "木叶村", description: "守护", weight: 10 },
                { source: "宇智波佐助", target: "木叶村", description: "守护", weight: 9 },
                { source: "漩涡鸣人", target: "宇智波佐助", description: "羁绊", weight: 10 }
            );
            addGroup(100, "木叶忍者", "person", "木叶村", "火之意志的继承者。", "守护");
            addGroup(50, "影分身", "person", "漩涡鸣人", "鸣人的实体分身。", "分身");

            // 2. 灵笼 (Ling Cage)
            extraNodes.push(
                { id: "马克", name: "马克", type: "person", description: "猎荒者指挥官。", val: 20, source_id: ["ling-01"] },
                { id: "冉冰", name: "冉冰", type: "person", description: "马克的副官与爱人。", val: 18, source_id: ["ling-01"] },
                { id: "灯塔", name: "灯塔", type: "geo", description: "人类最后的栖息地。", val: 25, source_id: ["ling-geo"] }
            );
            extraLinks.push(
                { source: "马克", target: "灯塔", description: "守护", weight: 9 },
                { source: "马克", target: "冉冰", description: "深爱", weight: 10 }
            );
            addGroup(60, "猎荒者", "person", "马克", "地面采集物资的精锐部队。", "队员");
            addGroup(40, "噬极兽", "person", "灯塔", "地面的恐怖生物。", "威胁");

            // 3. 剑来 (Sword of Coming)
            extraNodes.push(
                { id: "陈平安", name: "陈平安", type: "person", description: "隐官大人，剑气长城。", val: 25, source_id: ["jl-01"] },
                { id: "宁姚", name: "宁姚", type: "person", description: "剑气长城第一人，陈平安媳妇。", val: 24, source_id: ["jl-01"] },
                { id: "剑气长城", name: "剑气长城", type: "geo", description: "抵御妖族的第一线。", val: 28, source_id: ["jl-geo"] }
            );
            extraLinks.push(
                { source: "陈平安", target: "宁姚", description: "道侣", weight: 10 },
                { source: "陈平安", target: "剑气长城", description: "驻守", weight: 10 }
            );
            addGroup(100, "剑修", "person", "剑气长城", "万千剑修，决战城头。", "驻守");
            addGroup(50, "妖族大军", "person", "剑气长城", "蛮荒天下的入侵者。", "攻打");

            // 4. 红楼梦 (Red Dream)
            extraNodes.push(
                { id: "贾宝玉", name: "贾宝玉", type: "person", description: "荣国府衔玉而诞的公子。", val: 20, source_id: ["red-01"] },
                { id: "林黛玉", name: "林黛玉", type: "person", description: "金陵十二钗之首，世外仙姝寂寞林。", val: 20, source_id: ["red-01"] },
                { id: "大观园", name: "大观园", type: "geo", description: "贾府为元妃省亲修建的别墅。", val: 22, source_id: ["red-geo"] }
            );
            extraLinks.push(
                { source: "贾宝玉", target: "林黛玉", description: "木石前盟", weight: 10 },
                { source: "贾宝玉", target: "大观园", description: "居住", weight: 8 }
            );
            addGroup(50, "丫鬟", "person", "大观园", "大观园中的侍女。", "侍奉");
            addGroup(40, "贾氏宗亲", "person", "贾宝玉", "荣宁二府的亲戚。", "亲族");

            // 5. 盗墓笔记 (Tomb Notes)
            extraNodes.push(
                { id: "吴邪", name: "吴邪", type: "person", description: "天真无邪，老九门吴家传人。", val: 22, source_id: ["tomb-01"] },
                { id: "张起灵", name: "张起灵", type: "person", description: "闷油瓶，神秘强大的小哥。", val: 25, source_id: ["tomb-01"] },
                { id: "王胖子", name: "王胖子", type: "person", description: "摸金校尉，铁三角之一。", val: 20, source_id: ["tomb-01"] },
                { id: "七星鲁王宫", name: "七星鲁王宫", type: "geo", description: "战国时期的古墓。", val: 20, source_id: ["tomb-geo"] }
            );
            extraLinks.push(
                { source: "吴邪", target: "张起灵", description: "铁三角", weight: 10 },
                { source: "吴邪", target: "王胖子", description: "铁三角", weight: 10 },
                { source: "吴邪", target: "七星鲁王宫", description: "探险", weight: 9 }
            );
            addGroup(50, "尸蹩", "person", "七星鲁王宫", "墓中的危险生物。", "栖息");
            addGroup(30, "粽子", "person", "七星鲁王宫", "起尸的古尸。", "守护");

            // 6. 大王饶命 (Spare Me)
            extraNodes.push(
                { id: "吕树", name: "吕树", type: "person", description: "可以将负面情绪值转化为修行资源的毒舌少年。", val: 24, source_id: ["spare-01"] },
                { id: "吕小鱼", name: "吕小鱼", type: "person", description: "吕树的妹妹，御兽能力者。", val: 22, source_id: ["spare-01"] },
                { id: "天罗地网", name: "天罗地网", type: "organization", description: "华夏修行者组织。", val: 25, source_id: ["spare-org"] }
            );
            extraLinks.push(
                { source: "吕树", target: "吕小鱼", description: "兄妹/相依为命", weight: 10 },
                { source: "吕树", target: "天罗地网", description: "第九天罗", weight: 9 }
            );
            addGroup(100, "道元班学生", "person", "天罗地网", "觉醒资质的学生。", "学员");

            // 7. 诡秘之主 (LOTM)
            extraNodes.push(
                { id: "克莱恩", name: "克莱恩", type: "person", description: "愚者先生，来自源堡的穿越者。", val: 26, source_id: ["lotm-01"] },
                { id: "奥黛丽", name: "奥黛丽", type: "person", description: "正义小姐，贝克兰德最耀眼的宝石。", val: 20, source_id: ["lotm-01"] },
                { id: "塔罗会", name: "塔罗会", type: "organization", description: "隐秘组织，以此交换情报和资源。", val: 24, source_id: ["lotm-org"] },
                { id: "廷根市", name: "廷根市", type: "geo", description: "克莱恩穿越初始之地。", val: 18, source_id: ["lotm-geo"] }
            );
            extraLinks.push(
                { source: "克莱恩", target: "塔罗会", description: "召集人", weight: 10 },
                { source: "奥黛丽", target: "塔罗会", description: "成员", weight: 9 },
                { source: "克莱恩", target: "廷根市", description: "居住", weight: 8 }
            );
            addGroup(60, "值夜者", "person", "廷根市", "黑夜女神的守护力量。", "守护");
            addGroup(50, "非凡者", "person", "塔罗会", "掌握非凡力量的人。", "交易");

            data.nodes = [...data.nodes, ...extraNodes];
            data.links = [...data.links, ...extraLinks];

            setData(data);
        }, 500);
    }, []);

    const handleNodeClick = useCallback((node: any) => {
        setSelectedNode(node);
        if (fgRef.current) {
            // Aim at node from distance dist
            const dist = 60;
            const distRatio = 1 + dist / Math.hypot(node.x, node.y, node.z);

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
        sprite.color = 'white';
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
                sphere.material.color.set('#555');
                sprite.visible = false;
            } else {
                // Active
                const originalColor = TYPE_COLORS[node.type] || TYPE_COLORS.default;
                sphere.material.opacity = 1;
                sphere.material.color.set(originalColor);
                sprite.visible = true;

                if (isRelevant) {
                    sphere.material.emissive.set(originalColor);
                    sphere.material.emissiveIntensity = 0.5;
                } else {
                    sphere.material.emissiveIntensity = 0;
                }
            }
        });

    }, [hoverNode, selectedNode, highlightNodes, data.nodes]);


    const handleSearch = () => {
        if (!searchQuery) return;
        const node = data.nodes.find((n: GraphNode) => n.name.includes(searchQuery));
        if (node) {
            handleNodeClick(node); // Fly to node
        }
    }

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
            <Card className="flex-1 relative overflow-hidden bg-[#000011] border border-gray-800 shadow-sm rounded-xl">
                <div className="absolute top-4 left-4 z-10 flex flex-col gap-3 w-auto min-w-[300px] pointer-events-none">
                    <div className="pointer-events-auto backdrop-blur-sm bg-white/10 border border-white/20 shadow-lg rounded-xl p-1 flex gap-1 items-center">
                        <Input
                            placeholder="搜索节点..."
                            className="border-0 bg-transparent focus-visible:ring-0 text-white placeholder:text-gray-400 h-9 w-64"
                            value={searchQuery}
                            onChange={e => setSearchQuery(e.target.value)}
                            onKeyDown={e => e.key === 'Enter' && handleSearch()}
                        />
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-300 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleSearch}>
                            <Search size={18} />
                        </Button>
                        <Separator orientation="vertical" className="h-6 bg-white/20 mx-1" />
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-300 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleZoomIn} title="放大">
                            <ZoomIn size={18} />
                        </Button>
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-300 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleZoomOut} title="缩小">
                            <ZoomOut size={18} />
                        </Button>
                        <Button size="icon" variant="ghost" className="h-9 w-9 text-gray-300 hover:text-white hover:bg-white/10 rounded-lg transition-colors" onClick={handleZoomToFit} title="全览">
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
                        return highlightLinks.has(idStr) ? '#55afff' : '#ffffff';
                    }}
                    linkWidth={link => {
                        // @ts-ignore
                        const idStr = `${link.source.id}-${link.target.id}`;
                        return highlightLinks.has(idStr) ? 2 : 0.5;
                    }}
                    linkOpacity={0.5}
                    backgroundColor="#000011"
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
                                {selectedNode?.source_id.map((id: string, index: number) => (
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
