package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gozero-rag/internal/graphrag/types"

	nebula "github.com/vesoft-inc/nebula-go/v3"
	ngen "github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/zeromicro/go-zero/core/logx"
)

// NebulaGraphModel 定义 NebulaGraph 图数据库操作接口
// 每个方法接收 kbId 参数，内部映射到对应的 Space (命名规则: kg_<kbId>)
type NebulaGraphModel interface {
	// EnsureSpaceAndSchema 确保指定 kbId 的 Space 和 Schema 存在
	EnsureSpaceAndSchema(ctx context.Context, kbId string) error

	// BatchUpsertEntities 批量写入实体到指定 kbId 的 Space
	BatchUpsertEntities(ctx context.Context, kbId string, entities []types.Entity) error

	// BatchInsertRelations 批量写入关系到指定 kbId 的 Space
	BatchInsertRelations(ctx context.Context, kbId string, relations []types.Relation) error

	// GetGraph 获取图谱数据 (支持 limit 限制)
	GetGraph(ctx context.Context, kbId string, limit int) ([]types.Entity, []types.Relation, error)

	// SearchGraphNodes 搜索图谱节点
	SearchGraphNodes(ctx context.Context, kbId string, query string) ([]types.Entity, error)

	// Close 关闭连接池
	Close()
}

// nebulaGraphModel 实现 NebulaGraphModel 接口
type nebulaGraphModel struct {
	pool     *nebula.ConnectionPool
	username string
	password string
}

// NewNebulaGraphModel 创建 NebulaGraph 模型实例
func NewNebulaGraphModel(addresses []string, username, password string) (NebulaGraphModel, error) {
	// 解析地址列表
	hostList := make([]nebula.HostAddress, 0, len(addresses))
	for _, addr := range addresses {
		parts := strings.Split(addr, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid nebula address format: %s, expected host:port", addr)
		}
		port := 0
		if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
			return nil, fmt.Errorf("invalid port in address %s: %w", addr, err)
		}
		hostList = append(hostList, nebula.HostAddress{Host: parts[0], Port: port})
	}

	// 配置连接池
	poolConfig := nebula.GetDefaultConf()
	poolConfig.MaxConnPoolSize = 10
	poolConfig.TimeOut = 30 * time.Second

	// 创建连接池
	pool, err := nebula.NewConnectionPool(hostList, poolConfig, nebula.DefaultLogger{})
	if err != nil {
		return nil, fmt.Errorf("create nebula connection pool failed: %w", err)
	}

	// 验证连接
	session, err := pool.GetSession(username, password)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("get nebula session failed: %w", err)
	}
	session.Release()

	return &nebulaGraphModel{
		pool:     pool,
		username: username,
		password: password,
	}, nil
}

// Close 关闭连接池
func (m *nebulaGraphModel) Close() {
	if m.pool != nil {
		m.pool.Close()
	}
}

// getSpaceName 根据 kbId 生成 Space 名称
func getSpaceName(kbId string) string {
	safeKbId := strings.ReplaceAll(kbId, "-", "_")
	return fmt.Sprintf("kg_%s", safeKbId)
}

// EnsureSpaceAndSchema 确保指定 kbId 的 Space 和 Schema 存在
func (m *nebulaGraphModel) EnsureSpaceAndSchema(ctx context.Context, kbId string) error {
	spaceName := getSpaceName(kbId)

	session, err := m.pool.GetSession(m.username, m.password)
	if err != nil {
		return fmt.Errorf("get session failed: %w", err)
	}
	defer session.Release()

	// 1. 创建 Space (如果不存在)
	createSpaceNgql := fmt.Sprintf(`
		CREATE SPACE IF NOT EXISTS %s (
			partition_num = 10,
			replica_factor = 1,
			vid_type = FIXED_STRING(128)
		);
	`, spaceName)

	if _, err := session.Execute(createSpaceNgql); err != nil {
		return fmt.Errorf("create space failed: %w", err)
	}

	// 等待 Space 创建完成 (Nebula 是异步创建)
	time.Sleep(3 * time.Second)

	// 2. 切换到 Space
	useSpaceNgql := fmt.Sprintf("USE %s;", spaceName)
	if _, err := session.Execute(useSpaceNgql); err != nil {
		return fmt.Errorf("use space failed: %w", err)
	}

	// 3. 创建 Tag: entity
	createEntityTagNgql := `
		CREATE TAG IF NOT EXISTS entity (
			name string,
			type string,
			description string,
			embedding string,
			source_ids string
		);
	`
	if _, err := session.Execute(createEntityTagNgql); err != nil {
		return fmt.Errorf("create entity tag failed: %w", err)
	}

	// 4. 创建 EdgeType: relates_to
	createEdgeNgql := `
		CREATE EDGE IF NOT EXISTS relates_to (
			type string,
			description string,
			weight double,
			source_ids string
		);
	`
	if _, err := session.Execute(createEdgeNgql); err != nil {
		return fmt.Errorf("create edge type failed: %w", err)
	}

	// 5. 创建 Index (MATCH查询需要)
	createIndexNgql := "CREATE TAG INDEX IF NOT EXISTS entity_index ON entity(name(64));"
	if _, err := session.Execute(createIndexNgql); err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	// 等待 Index 生效 (需要重建索引)
	// REBUILD TAG INDEX entity_index; (通常首次创建不需要重建，但如果是后续添加需重建)
	// 这里简单等待
	time.Sleep(3 * time.Second)

	logx.Infof("ensured nebula space and schema for kb: %s (space: %s)", kbId, spaceName)
	return nil
}

// BatchUpsertEntities 批量写入实体到指定 kbId 的 Space
func (m *nebulaGraphModel) BatchUpsertEntities(ctx context.Context, kbId string, entities []types.Entity) error {
	if len(entities) == 0 {
		return nil
	}

	spaceName := getSpaceName(kbId)

	session, err := m.pool.GetSession(m.username, m.password)
	if err != nil {
		return fmt.Errorf("get session failed: %w", err)
	}
	defer session.Release()

	// 切换到 Space
	useSpaceNgql := fmt.Sprintf("USE %s;", spaceName)
	if _, err := session.Execute(useSpaceNgql); err != nil {
		return fmt.Errorf("use space failed: %w", err)
	}

	// 批量 UPSERT 实体
	successCount := 0
	for _, e := range entities {
		vid := escapeVid(e.Name)
		sourceIdsStr := strings.Join(e.SourceId, ",")

		// 将 embedding 序列化为 JSON 字符串
		embeddingStr := ""
		if len(e.Embedding) > 0 {
			embBytes, _ := json.Marshal(e.Embedding)
			embeddingStr = string(embBytes)
		}

		// 使用 UPSERT 实现新增或更新
		ngql := fmt.Sprintf(`
			UPSERT VERTEX ON entity "%s"
			SET name = "%s",
				type = "%s",
				description = "%s",
				embedding = "%s",
				source_ids = "%s";
		`, vid, escapeString(e.Name), escapeString(e.Type), escapeString(e.Description), escapeString(embeddingStr), escapeString(sourceIdsStr))

		result, err := session.Execute(ngql)
		if err != nil {
			logx.Errorf("upsert entity %s failed: %v", e.Name, err)
			continue
		}
		if !result.IsSucceed() {
			logx.Errorf("upsert entity %s failed: %s", e.Name, result.GetErrorMsg())
			continue
		}
		successCount++
	}

	logx.Infof("upserted %d/%d entities to space %s", successCount, len(entities), spaceName)
	return nil
}

// BatchInsertRelations 批量写入关系到指定 kbId 的 Space
func (m *nebulaGraphModel) BatchInsertRelations(ctx context.Context, kbId string, relations []types.Relation) error {
	if len(relations) == 0 {
		return nil
	}

	spaceName := getSpaceName(kbId)

	session, err := m.pool.GetSession(m.username, m.password)
	if err != nil {
		return fmt.Errorf("get session failed: %w", err)
	}
	defer session.Release()

	// 切换到 Space
	useSpaceNgql := fmt.Sprintf("USE %s;", spaceName)
	if _, err := session.Execute(useSpaceNgql); err != nil {
		return fmt.Errorf("use space failed: %w", err)
	}

	// 批量 INSERT 边 (边使用 INSERT 覆盖，不用 UPSERT)
	successCount := 0
	for _, r := range relations {
		srcVid := escapeVid(r.SrcId)
		dstVid := escapeVid(r.DstId)
		sourceIdsStr := strings.Join(r.SourceId, ",")

		ngql := fmt.Sprintf(`
			INSERT EDGE relates_to(type, description, weight, source_ids) VALUES 
			"%s"->"%s":("%s", "%s", %f, "%s");
		`, srcVid, dstVid, escapeString(r.Type), escapeString(r.Description), r.Weight, escapeString(sourceIdsStr))

		result, err := session.Execute(ngql)
		if err != nil {
			logx.Errorf("insert relation %s->%s failed: %v", r.SrcId, r.DstId, err)
			continue
		}
		if !result.IsSucceed() {
			logx.Errorf("insert relation %s->%s failed: %s", r.SrcId, r.DstId, result.GetErrorMsg())
			continue
		}
		successCount++
	}

	logx.Infof("inserted %d/%d relations to space %s", successCount, len(relations), spaceName)
	return nil
}

// GetGraph 获取图谱数据 (支持 limit 限制)
func (m *nebulaGraphModel) GetGraph(ctx context.Context, kbId string, limit int) ([]types.Entity, []types.Relation, error) {
	spaceName := getSpaceName(kbId)

	session, err := m.pool.GetSession(m.username, m.password)
	if err != nil {
		return nil, nil, fmt.Errorf("get session failed: %w", err)
	}
	defer session.Release()

	// 切换 Space
	if _, err := session.Execute(fmt.Sprintf("USE %s;", spaceName)); err != nil {
		return nil, nil, fmt.Errorf("use space failed: %w", err)
	}

	// 执行 nGQL 查询
	// MATCH (v:entity)-[e:relates_to]->(v2:entity) RETURN properties(v), properties(e), properties(v2) LIMIT {limit}
	ngql := fmt.Sprintf("MATCH (v:entity)-[e:relates_to]->(v2:entity) RETURN properties(v) AS `src`, properties(e) AS `edge`, properties(v2) AS `dst` LIMIT %d;", limit)

	resultSet, err := session.Execute(ngql)
	if err != nil {
		return nil, nil, fmt.Errorf("execute ngql failed: %w", err)
	}
	if !resultSet.IsSucceed() {
		// 如果 Space 不存在或还没有数据，可能报错，视为为空
		if strings.Contains(resultSet.GetErrorMsg(), "SpaceNotFound") {
			return []types.Entity{}, []types.Relation{}, nil
		}
		return nil, nil, fmt.Errorf("query graph failed: %s", resultSet.GetErrorMsg())
	}

	// 解析结果
	entitiesMap := make(map[string]types.Entity)
	relations := make([]types.Relation, 0)

	rows := resultSet.GetRows()
	for _, row := range rows {
		// row.Values 是 []*ngen.Value
		if len(row.Values) < 3 {
			continue
		}
		srcVal := row.Values[0]
		edgeVal := row.Values[1]
		dstVal := row.Values[2]

		// 解析实体
		var srcKvs, dstKvs map[string]*ngen.Value
		if srcVal.GetMVal() != nil {
			srcKvs = srcVal.GetMVal().Kvs
		}
		if dstVal.GetMVal() != nil {
			dstKvs = dstVal.GetMVal().Kvs
		}

		srcEntity := parseEntityFromMap(srcKvs)
		dstEntity := parseEntityFromMap(dstKvs)

		if srcEntity.Name != "" {
			entitiesMap[srcEntity.Name] = srcEntity
		}
		if dstEntity.Name != "" {
			entitiesMap[dstEntity.Name] = dstEntity
		}

		// 解析关系
		var edgeKvs map[string]*ngen.Value
		if edgeVal.GetMVal() != nil {
			edgeKvs = edgeVal.GetMVal().Kvs
		}

		rel := parseRelationFromMap(edgeKvs)
		rel.SrcId = srcEntity.Name
		rel.DstId = dstEntity.Name
		relations = append(relations, rel)
	}

	entities := make([]types.Entity, 0, len(entitiesMap))
	for _, e := range entitiesMap {
		entities = append(entities, e)
	}

	return entities, relations, nil
}

// SearchGraphNodes 搜索图谱节点
func (m *nebulaGraphModel) SearchGraphNodes(ctx context.Context, kbId string, query string) ([]types.Entity, error) {
	spaceName := getSpaceName(kbId)

	session, err := m.pool.GetSession(m.username, m.password)
	if err != nil {
		return nil, fmt.Errorf("get session failed: %w", err)
	}
	defer session.Release()

	// 切换 Space
	if _, err := session.Execute(fmt.Sprintf("USE %s;", spaceName)); err != nil {
		return nil, fmt.Errorf("use space failed: %w", err)
	}

	// 模糊查询: MATCH (v:entity) WHERE v.entity.name CONTAINS "query" RETURN properties(v)
	// 模糊查询: MATCH (v:entity) WHERE v.entity.name CONTAINS "query" RETURN properties(v)
	ngql := fmt.Sprintf("MATCH (v:entity) WHERE v.entity.name CONTAINS \"%s\" RETURN properties(v) AS node LIMIT 20;", escapeString(query))

	resultSet, err := session.Execute(ngql)
	if err != nil {
		return nil, fmt.Errorf("execute ngql failed: %w", err)
	}
	if !resultSet.IsSucceed() {
		return nil, fmt.Errorf("search nodes failed: %s", resultSet.GetErrorMsg())
	}

	entities := make([]types.Entity, 0)
	rows := resultSet.GetRows()
	for _, row := range rows {
		if len(row.Values) < 1 {
			continue
		}
		val := row.Values[0]
		var kvs map[string]*ngen.Value
		if val.GetMVal() != nil {
			kvs = val.GetMVal().Kvs
		}
		entity := parseEntityFromMap(kvs)
		if entity.Name != "" {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

func parseEntityFromMap(m map[string]*ngen.Value) types.Entity {
	e := types.Entity{}
	if m == nil {
		return e
	}
	// Thrift 生成的 map key 是 string (因为属性名是 string)
	if val, ok := m["name"]; ok && len(val.SVal) > 0 {
		e.Name = string(val.SVal)
	}
	if val, ok := m["type"]; ok && len(val.SVal) > 0 {
		e.Type = string(val.SVal)
	}
	if val, ok := m["description"]; ok && len(val.SVal) > 0 {
		e.Description = string(val.SVal)
	}
	if val, ok := m["source_ids"]; ok && len(val.SVal) > 0 {
		// 分割逗号
		e.SourceId = strings.Split(string(val.SVal), ",")
	}
	return e
}

func parseRelationFromMap(m map[string]*ngen.Value) types.Relation {
	r := types.Relation{}
	if m == nil {
		return r
	}
	if val, ok := m["type"]; ok && len(val.SVal) > 0 {
		r.Type = string(val.SVal)
	}
	if val, ok := m["description"]; ok && len(val.SVal) > 0 {
		r.Description = string(val.SVal)
	}
	if val, ok := m["weight"]; ok {
		if val.FVal != nil {
			r.Weight = *val.FVal
		} else if val.IVal != nil {
			r.Weight = float64(*val.IVal)
		}
	}
	if val, ok := m["source_ids"]; ok && len(val.SVal) > 0 {
		r.SourceId = strings.Split(string(val.SVal), ",")
	}
	return r
}

// escapeVid 转义 VID 中的特殊字符
func escapeVid(vid string) string {
	vid = strings.ReplaceAll(vid, `\`, `\\`)
	vid = strings.ReplaceAll(vid, `"`, `\"`)
	return vid
}

// escapeString 转义字符串中的特殊字符
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
