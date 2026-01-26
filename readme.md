# Go-Zero RAG 知识库系统

基于 Go-Zero 和 React 构建的企业级高性能 RAG（检索增强生成）知识库系统。

## 📖 项目介绍

本项目帮企业构建私有化知识库，通过 **RAG (检索增强生成)** 架构解决大模型在垂直领域的幻觉与知识缺失问题。
最新版本集成 **GraphRAG (图增强检索)**，通过 **High-Dimensional Visualization (高维可视化)** 技术，直观呈现复杂的知识拓扑结构。

核心理念：**Engineering-First (工程优先)**。关注系统的可扩展性、检索召回率与交互性能。

### 核心特性

- **多模态解析**: 支持 PDF, TXT, Word, Markdown 等非结构化数据的流水线处理（清洗、切片、向量化）。
- **GraphRAG (图增强检索)**: 
    - **高性能可视化**: 基于 **WebGL**与 **Force-Directed Algorithms (力导向算法)**，支持前端渲染万级节点的大规模知识图谱。
    - **智能提取**: 集成 LLM 的 Loop-Gleaning 策略，自动化提取非结构化文档中的实体（Entity）与关系（Relation）。
    - **专业图存储**: 采用 **Nebula Graph** 分布式图数据库存储实体与关系数据，支持多跳查询与复杂图算法分析。
- **混合检索架构**: 
    - **多路并行检索 (Parallel Multi-Path Retrieval)**: 同时执行 **Elasticsearch** (向量检索/BM25) 与 **Nebula Graph** (图遍历) 查询。
    - **Hybrid Search**: 结合 **Dense Vector (稠密向量)** 与 **Sparse Retrieval (稀疏检索)**，兼顾语义理解与关键词匹配。
    - **Rerank**:  引入重排序模型对多路召回结果进行统一打分排序。
- **多租户架构 (Multi-Tenancy)**: 在逻辑层面实现了完善的多租户数据隔离机制，支持多团队/多用户环境下的数据安全与隐私保护。
- **可视化分析**: 提供悬浮搜索、数据仪表盘及详情面板，支持对图谱数据的下钻分析。
- **灵活编排**: 基于字节跳动 **Eino** 框架实现 DAG (有向无环图) 编排，灵活定义复杂的 RAG 检索流与数据处理流。

## 🛠 技术栈

### 后端 (Backend)
- **核心框架**: [Go-Zero](https://go-zero.dev/) (Web/RPC, Middleware)
- **大模型编排**: [ByteDance Eino](https://github.com/cloudwego/eino) (Graph/Chain, Node, Edge)
- **数据库**: 
    - **MySQL 8.0+** (业务元数据)
    - **Redis** (缓存与会话)
    - **Elasticsearch** (向量检索引擎：负责文档切片 Vectors 存储与倒排索引)
    - **Nebula Graph** (图数据库：负责存储知识图谱实体与关系)
- **对象存储**: MinIO
- **消息队列**: Kafka (异步任务解耦：文档解析、图谱生成)

### 前端 (Frontend)
- **核心框架**: [React 18](https://react.dev/) (Vite)
- **语言**: TypeScript
- **UI 组件库**: 
    - [Shadcn/ui](https://ui.shadcn.com/) (基础组件)
    - [Tailwind CSS](https://tailwindcss.com/) (样式引擎)
    - [Lucide React](https://lucide.dev/) (图标)
- **数据可视化 (3D Graph)**:
    - **react-force-graph-3d**: 力导向图引擎
    - **Three.js**: 3D 渲染引擎 (自定义 Shader 实现星球光晕与粒子星云)
    - **three-spritetext**: 3D 文本标签
- **状态管理**: Zustand
- **表单管理**: React Hook Form + Zod

## 📂 项目结构

```text
├── restful/          # HTTP API 服务 (Go-Zero Gateway)
│   └── rag/          # 主业务服务
├── consumer/         # 异步消息消费者 (Workers)
│   ├── document_index/ # 文档切片与向量化消费者
│   └── graph_extract/  # 知识图谱提取消费者
├── internal/         # 核心业务逻辑与共享代码
│   ├── model/        # 数据库模型 (MySQL, ES)
│   ├── mq/           # 消息队列定义
│   └── graphrag/     # GraphRAG 核心算法实现
├── script/           # 数据库初始化脚本 (MySQL, ES, Docker)
└── fe/               # 前端项目 (React)
    ├── src/pages/knowledge/  # 知识库核心页面
    │   ├── KnowledgeGraph.tsx # 3D 宇宙图谱主入口
    │   └── GraphComponents/   # 图谱专用 UI 组件 (Search, Stats, Panel)
```

## 🚀 快速开始

### 1. 环境要求
- **Go**: 1.21+
- **Node.js**: 18+
- **Docker & Docker Compose**
- **Goctl**: `go install github.com/zeromicro/go-zero/tools/goctl@latest`

### 2. 基础设施搭建
使用 Docker Compose 启动所有依赖服务（MySQL, Redis, MinIO, Kafka, Elasticsearch）：

```bash
docker-compose up -d
```

请确保所有容器状态健康 (Healthy) 后再继续。

### 2.1 启动 Nebula Graph (图数据库)

如果启用 GraphRAG 功能，需额外部署 Nebula Graph 集群：

```bash
cd script/nubalagraph
docker-compose up -d
```
> Nebula Graph Studio (可视化控制台) 地址: `http://localhost:7001`
> 默认账号: `root` / `nebula`

### 3. 后端配置与启动 (多服务)

#### 3.1 配置 Nebula Graph 连接

在启动之前，需在以下配置文件中添加 Nebula Graph 连接信息：

- **API 服务**: `restful/rag/etc/rag.yaml`
- **文档索引消费者**: `consumer/document_index/etc/conf.yaml`
- **图谱提取消费者**: `consumer/graph_extract/etc/conf.yaml`

配置示例：

```yaml
NebulaConf:
  Host: "127.0.0.1:9669" # Graphd 服务地址
  User: "root"
  Pwd:  "nebula"
  SpaceName: "rag_space" # 自动创建的图空间名称
```

#### 3.2 启动 API 服务
```bash
# 1. 安装依赖
go mod tidy

# 2. 运行 RESTful API
cd restful/rag
go run rag.go
# 服务默认运行在 8888 端口
```

#### 3.2 启动消费者 (Workers)
建议新开终端窗口启动异步任务消费者：

```bash
# 1. 启动文档索引消费者 (处理文档解析、向量化)
cd consumer/document_index
go run main.go

# 2. 启动图谱提取消费者 (可选, 需开启 GraphRAG 功能)
cd consumer/graph_extract
go run main.go
```

### 4. 前端启动

1. **安装依赖**:
    ```bash
    cd fe
    npm install
    ```

2. **启动开发服务器**:
    ```bash
    npm run dev
    ```
    前端页面通常访问地址为 `http://localhost:5173`。

## 💻 代码生成指南

> [!IMPORTANT]
> **请勿直接使用 `goctl` 生成 API/Model 代码。** 请务必使用预定义的 Make 命令以确保一致性。

- **生成 API 代码**: `make gen-api` (修改 `restful/rag/rag.api` 后)
- **生成 MySQL Model**: `make gen-model` (修改 `script/mysql/*.sql` 后)
- **生成 API 文档**: `make gen-doc`

## 📸 运行截图
![知识图谱](images/image-kg.png)
![知识图谱2](images/image-kg2.png)
![alt text](images/image.png)
![alt text](images/image-1.png)
![alt text](images/image-2.png)
![alt text](images/image-3.png)
![alt text](images/image-4.png)