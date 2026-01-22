
/**
 * 预处理 Markdown 文本，解决 Obsidian 等编辑器与标准 CommonMark/GFM 的渲染差异
 */
export const preprocessMarkdown = (text: string): string => {
    if (!text) return '';

    let processed = text;

    // 1. 解决 Setext Heading 问题 (即 --- 被解析为上一行的 H2 标题)
    // 场景:
    // Some text
    // ---
    // 预期: Some text <hr>
    // 实际: <h2>Some text</h2>
    // 方案: 在 --- 前插入换行
    processed = processed.replace(/^([^\n]+)\n\s*{-{3,}}\s*$/gm, '$1\n\n---');

    // 2. 解决列表紧邻文本/加粗导致无法跳出列表的问题 (Obsidian 允许紧邻，标准 MD 视为列表内换行)
    // 场景:
    // - item 1
    // **Next Title**
    // 实际: <li>item 1 <br> <strong>Next Title</strong></li>
    // 方案: 在列表项和非列表项(尤其是加粗/标题)之间插入换行
    // 注意：不敢太激进，只针对 **Bold** 或 # Header 这种情况尝试修复
    processed = processed.replace(/(\n\s*[-*]\s+[^\n]+)\n(\s*(?:\*\*|#))/g, '$1\n\n$2');

    return processed;
};
