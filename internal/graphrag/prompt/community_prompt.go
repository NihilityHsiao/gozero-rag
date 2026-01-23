package prompt

const COMMUNITY_REPORT = `
你是一名 AI 助手，负责协助人类分析师进行通用信息发现。信息发现是指在网络中识别和评估与特定实体（例如组织和个人）相关的相关信息的过程。

# 目标
根据属于某个社区的实体列表及其相互关系和可选的相关声明，撰写一份关于该社区的综合报告。该报告将用于向决策者通报与该社区相关的信息及其潜在影响。报告的内容包括该社区关键实体的概览、其法律合规性、技术能力、声誉以及值得注意的声明。

# 报告结构

报告应包含以下部分：

- 标题 (TITLE)：代表其关键实体的社区名称——标题应简短但具体。如果在可能的情况下，请在标题中包含具有代表性的命名实体。
- 摘要 (SUMMARY)：关于社区整体结构、实体间相互关系以及与实体相关的重要信息的执行摘要。
- 影响严重程度评分 (IMPACT SEVERITY RATING)：一个 0-10 之间的浮点数评分，代表社区内实体所带来的“影响”严重程度。“影响”是指一个社区的评分重要性。
- 评分解释 (RATING EXPLANATION)：用一句话解释给出该影响严重程度评分的原因。
- 详细发现 (DETAILED FINDINGS)：关于该社区的 5-10 个关键见解列表。每个见解应包含一个简短的总结，后跟多段解释性文本，并根据下方的“依据规则”进行数据支撑。内容要全面。

将输出结果返回为格式良好的 JSON 格式字符串，格式如下（使用与“文本”内容相同的语言）：
{
	"title": <报告标题>,
	"summary": <执行摘要>,
	"rating": <影响严重程度评分>,
	"rating_explanation": <评分解释>,
	"findings": [
		{
			"summary":<见解_1_总结>,
			"explanation": <见解_1_解释>
		},
		{
			"summary":<见解_2_总结>,
			"explanation": <见解_2_解释>
		}
	]
}

# 依据规则 (Grounding Rules)

由数据支持的观点应按如下方式列出其数据引用：

"这是一个由多个数据引用支持的示例句子 [Data: <数据集名称> (记录id); <数据集名称> (记录id)]。"

不要在单个引用中列出超过 5 个记录 ID。相反，列出前 5 个最相关的记录 ID，并添加 "+more" 以表示还有更多。

例如：
"X 先生是 Y 公司的所有者，并受到多项不当行为的指控 [Data: Reports (1), Entities (5, 7); Relationships (23); Claims (7, 2, 34, 64, 46, +more)]。"

其中 1, 5, 7, 23, 2, 34, 46, 和 64 代表相关数据记录的 id（不是索引）。

不要包含未提供支持证据的信息。

# Example Input
-----------
Text:

-Entities-

id,entity,description
5,VERDANT OASIS PLAZA,Verdant Oasis Plaza is the location of the Unity March
6,HARMONY ASSEMBLY,Harmony Assembly is an organization that is holding a march at Verdant Oasis Plaza

-Relationships-

id,source,target,description
37,VERDANT OASIS PLAZA,UNITY MARCH,Verdant Oasis Plaza is the location of the Unity March
38,VERDANT OASIS PLAZA,HARMONY ASSEMBLY,Harmony Assembly is holding a march at Verdant Oasis Plaza
39,VERDANT OASIS PLAZA,UNITY MARCH,The Unity March is taking place at Verdant Oasis Plaza
40,VERDANT OASIS PLAZA,TRIBUNE SPOTLIGHT,Tribune Spotlight is reporting on the Unity march taking place at Verdant Oasis Plaza
41,VERDANT OASIS PLAZA,BAILEY ASADI,Bailey Asadi is speaking at Verdant Oasis Plaza about the march
43,HARMONY ASSEMBLY,UNITY MARCH,Harmony Assembly is organizing the Unity March

Output:
{
    "title": "Verdant Oasis Plaza and Unity March",
    "summary": "The community revolves around the Verdant Oasis Plaza, which is the location of the Unity March. The plaza has relationships with the Harmony Assembly, Unity March, and Tribune Spotlight, all of which are associated with the march event.",
    "rating": 5.0,
    "rating_explanation": "The impact severity rating is moderate due to the potential for unrest or conflict during the Unity March.",
    "findings": [
        {
            "summary": "Verdant Oasis Plaza as the central location",
            "explanation": "Verdant Oasis Plaza is the central entity in this community, serving as the location for the Unity March. This plaza is the common link between all other entities, suggesting its significance in the community. The plaza's association with the march could potentially lead to issues such as public disorder or conflict, depending on the nature of the march and the reactions it provokes. [Data: Entities (5), Relationships (37, 38, 39, 40, 41,+more)]"
        },
        {
            "summary": "Harmony Assembly's role in the community",
            "explanation": "Harmony Assembly is another key entity in this community, being the organizer of the march at Verdant Oasis Plaza. The nature of Harmony Assembly and its march could be a potential source of threat, depending on their objectives and the reactions they provoke. The relationship between Harmony Assembly and the plaza is crucial in understanding the dynamics of this community. [Data: Entities(6), Relationships (38, 43)]"
        },
        {
            "summary": "Unity March as a significant event",
            "explanation": "The Unity March is a significant event taking place at Verdant Oasis Plaza. This event is a key factor in the community's dynamics and could be a potential source of threat, depending on the nature of the march and the reactions it provokes. The relationship between the march and the plaza is crucial in understanding the dynamics of this community. [Data: Relationships (39)]"
        },
        {
            "summary": "Role of Tribune Spotlight",
            "explanation": "Tribune Spotlight is reporting on the Unity March taking place in Verdant Oasis Plaza. This suggests that the event has attracted media attention, which could amplify its impact on the community. The role of Tribune Spotlight could be significant in shaping public perception of the event and the entities involved. [Data: Relationships (40)]"
        }
    ]
}
# 真实数据

请使用以下文本作为你的回答依据。不要在回答中编造任何内容。

文本 (Text):

-实体 (Entities)-
{entity_df}

-关系 (Relationships)-
{relation_df}

报告应包含以下部分：

- 标题 (TITLE)：代表其关键实体的社区名称——标题应简短但具体。如果在可能的情况下，请在标题中包含具有代表性的命名实体。
- 摘要 (SUMMARY)：关于社区整体结构、实体间相互关系以及与实体相关的重要信息的执行摘要。
- 影响严重程度评分 (IMPACT SEVERITY RATING)：一个 0-10 之间的浮点数评分，代表社区内实体所带来的“影响”严重程度。“影响”是指一个社区的评分重要性。
- 评分解释 (RATING EXPLANATION)：用一句话解释给出该影响严重程度评分的原因。
- 详细发现 (DETAILED FINDINGS)：关于该社区的 5-10 个关键见解列表。每个见解应包含一个简短的总结，后跟多段解释性文本，并根据下方的“依据规则”进行数据支撑。内容要全面。

将输出结果返回为格式良好的 JSON 格式字符串，格式如下（使用与“文本”内容相同的语言）：
    {
        "title": <报告标题>,
        "summary": <执行摘要>,
        "rating": <影响严重程度评分>,
        "rating_explanation": <评分解释>,
        "findings": [
            {
                "summary":<见解_1_总结>,
                "explanation": <见解_1_解释>
            },
            {
                "summary":<见解_2_总结>,
                "explanation": <见解_2_解释>
            }
        ]
    }

# 依据规则 (Grounding Rules)

由数据支持的观点应按如下方式列出其数据引用：

"这是一个由多个数据引用支持的示例句子 [Data: <数据集名称> (记录id); <数据集名称> (记录id)]。"

不要在单个引用中列出超过 5 个记录 ID。相反，列出前 5 个最相关的记录 ID，并添加 "+more" 以表示还有更多。

例如：
"X 先生是 Y 公司的所有者，并受到多项不当行为的指控 [Data: Reports (1), Entities (5, 7); Relationships (23); Claims (7, 2, 34, 64, 46, +more)]。"

其中 1, 5, 7, 23, 2, 34, 46, 和 64 代表相关数据记录的 id（不是索引）。

不要包含未提供支持证据的信息。

输出 (Output):
`
