/*
 *
 * Copyright (c) 2021 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package main

import (
	"fmt"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"gozero-rag/internal/graphrag/types"
	"gozero-rag/internal/slicex"
	"strings"
)

const (
	address = "192.168.0.6"
	// The default port of NebulaGraph 2.x is 9669.
	// 3699 is only for testing.
	port     = 49174
	username = "root"
	password = "nebula"
	useHTTP2 = false
)

var (
	entities = []types.Entity{
		{
			Name:        "唐三藏",
			Type:        "person",
			Description: "唐朝高僧，取经团队领袖，俗姓陈，法号玄奘。",
			SourceId:    []string{"chunk-abc123"},
		},
		{
			Name:        "花果山",
			Type:        "geo",
			Description: "东胜神洲傲来国的一座名山，是孙悟空的出生地和根据地。",
			SourceId:    []string{"chunk-ghi789"},
		},
		{
			Name:        "孙悟空",
			Type:        "person",
			Description: "法号行者，齐天大圣，拥有七十二变和筋斗云。",
			SourceId:    []string{"chunk-wukong001"},
		},
		{
			Name:        "如意金箍棒",
			Type:        "weapon", // 或 item
			Description: "大禹治水留下的定海神针，重一万三千五百斤。",
			SourceId:    []string{"chunk-weapon001"},
		},
		{
			Name:        "紧箍咒",
			Type:        "spell", // 或 concept
			Description: "观音菩萨传授给唐僧的法术，用于约束孙悟空的野性。",
			SourceId:    []string{"chunk-spell001"},
		},
		{
			Name:        "女儿国",
			Type:        "geo",
			Description: "西梁女国，国内全是女子，唐僧师徒曾在此遭遇情劫。",
			SourceId:    []string{"chunk-geo002"},
		},
	}

	relations = []types.Relation{
		// 1. 师徒关系：唐三藏 -> 孙悟空
		{
			SrcId:       "唐三藏",
			DstId:       "孙悟空",
			Description: "师徒关系，唐三藏将孙悟空从五行山下救出并收为大徒弟。",
			Weight:      1.0,
			SourceId:    []string{"chunk-abc123", "chunk-wukong001"},
		},
		// 2. 归属地关系：孙悟空 -> 花果山
		{
			SrcId:       "孙悟空",
			DstId:       "花果山",
			Description: "出生地与根据地，孙悟空生于花果山顶的仙石，并在此称美猴王。",
			Weight:      0.9,
			SourceId:    []string{"chunk-wukong001", "chunk-ghi789"},
		},
		// 3. 持有关系：孙悟空 -> 如意金箍棒
		{
			SrcId:       "孙悟空",
			DstId:       "如意金箍棒",
			Description: "持有与使用，金箍棒是孙悟空的专属兵器，听从其指令变化。",
			Weight:      1.0,
			SourceId:    []string{"chunk-wukong001", "chunk-weapon001"},
		},
		// 4. 施咒/使用关系：唐三藏 -> 紧箍咒
		{
			SrcId:       "唐三藏",
			DstId:       "紧箍咒",
			Description: "使用者，唐三藏通过念诵咒语来激活紧箍咒。",
			Weight:      0.8,
			SourceId:    []string{"chunk-abc123", "chunk-spell001"},
		},
		// 5. 约束关系：紧箍咒 -> 孙悟空
		{
			SrcId:       "紧箍咒",
			DstId:       "孙悟空",
			Description: "约束与控制，戴在孙悟空头上，念咒时会使其头痛欲裂。",
			Weight:      1.0,
			SourceId:    []string{"chunk-spell001", "chunk-wukong001"},
		},
		// 6. 途径/经历关系：唐三藏 -> 女儿国
		{
			SrcId:       "唐三藏",
			DstId:       "女儿国",
			Description: "途径经历，唐僧师徒取经途中经过此地，国王欲招唐僧为夫。",
			Weight:      0.7,
			SourceId:    []string{"chunk-abc123", "chunk-geo002"},
		},
		// 7. 物品来源（补充）：如意金箍棒 -> 花果山 (实际上是东海，但在逻辑上若无东海entity，可暂不写或关联到拥有者)
		// 这里补充一个反向关系，体现双向图谱的丰富性
		{
			SrcId:       "孙悟空",
			DstId:       "唐三藏",
			Description: "保卫与护送，孙悟空负责保护唐三藏西天取经。",
			Weight:      1.0,
			SourceId:    []string{"chunk-wukong001"},
		},
	}
)

// Initialize logger
var log = nebula.DefaultLogger{}

// 一个知识库对应一个space
func setupSpace(session *nebula.Session, kbId string) error {
	kbId = strings.ReplaceAll(kbId, "-", "_")

	spaceName := "kg_" + kbId

	createSchema := fmt.Sprintf("CREATE SPACE IF NOT EXISTS %s (partition_num = 10, vid_type = FIXED_STRING(64));", spaceName) +
		fmt.Sprintf("USE %s;", spaceName) +
		`CREATE TAG IF NOT EXISTS entity (
		name string COMMENT "实体名称",
		type string COMMENT "实体类型:person, geo, event, product, organization等",
		description string NULL COMMENT "实体描述",
		source_ids string COMMENT "英文逗号分隔的 chunk-id 列表",
	);
	
	CREATE TAG INDEX IF NOT EXISTS entity_type_idx ON entity(type(32));
	CREATE TAG INDEX IF NOT EXISTS entity_name_idx ON entity(name(64));

	CREATE EDGE IF NOT EXISTS relates_to (
    	description string NULL COMMENT "LLM提取的关系描述", 
    	weight double,
    	source_ids string COMMENT "英文逗号分隔的 chunk-id 列表"
	);
`
	res, err := session.Execute(createSchema)
	if err != nil {
		return err
	}

	if !res.IsSucceed() {
		return fmt.Errorf("create schema err:%v", res.GetErrorMsg())
	}
	return nil
}

func insertEntities(session *nebula.Session, entities []types.Entity) error {

	insertSchema := "INSERT VERTEX entity(name, type, description, source_ids) VALUES "

	sub := slicex.Into(entities, func(e types.Entity) string {
		return fmt.Sprintf("'%s': ('%s', '%s', '%s', '%s')", e.Name, e.Name, e.Type, e.Description, strings.Join(e.SourceId, ","))
	})

	insertSchema += strings.Join(sub, ",")

	res, err := session.Execute(insertSchema)
	if err != nil {
		return err
	}

	if !res.IsSucceed() {
		return fmt.Errorf("insert schema err:%v", res.GetErrorMsg())
	}

	return nil
}

func insertEdge(session *nebula.Session, edges []types.Relation) error {
	insertSchema := "INSERT EDGE relates_to(description, weight, source_ids) VALUES "

	sub := slicex.Into(edges, func(e types.Relation) string {
		return fmt.Sprintf("'%s' -> '%s': ('%s', %f, '%s')", e.SrcId, e.DstId, e.Description, e.Weight, strings.Join(e.SourceId, ","))

	})
	insertSchema += strings.Join(sub, ",")

	res, err := session.Execute(insertSchema)
	if err != nil {
		return err
	}

	if !res.IsSucceed() {
		return fmt.Errorf("insert edge err:%v", res.GetErrorMsg())
	}
	return nil
}

func main() {
	hostAddress := nebula.HostAddress{Host: address, Port: port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()
	testPoolConfig.UseHTTP2 = useHTTP2

	// Initialize connection pool
	pool, err := nebula.NewConnectionPool(hostList, testPoolConfig, log)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to initialize the connection pool, host: %s, port: %d, %s", address, port, err.Error()))
	}
	// Close all connections in the pool
	defer pool.Close()
	session, err := pool.GetSession(username, password)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to create a new session from connection pool, username: %s, password: %s, %s",
			username, password, err.Error()))
	}
	// Release session and return connection back to connection pool
	defer session.Release()

	kbId := "019be53c-f6f7-79d7-9957-4631606e228b"
	err = setupSpace(session, kbId)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to setup space, %s", err.Error()))
	}

	// 创建 tag

	err = insertEntities(session, entities)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to insert entities, %s", err.Error()))
	}
	err = insertEdge(session, relations)
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to insert edges, %s", err.Error()))
	}

	// Create session and send query in go routine

	//var wg sync.WaitGroup
	//wg.Add(1)
	//go func(wg *sync.WaitGroup) {
	//	defer wg.Done()
	//	// Create session
	//
	//	// Method used to check execution response
	//	checkResultSet := func(prefix string, res *nebula.ResultSet) {
	//		if !res.IsSucceed() {
	//			log.Fatal(fmt.Sprintf("%s, ErrorCode: %v, ErrorMsg: %s", prefix, res.GetErrorCode(), res.GetErrorMsg()))
	//		}
	//	}
	//	{
	//		createSchema := "CREATE SPACE IF NOT EXISTS example_space(vid_type=FIXED_STRING(20)); " +
	//			"USE example_space;" +
	//			"CREATE TAG IF NOT EXISTS person(name string, age int);" +
	//			"CREATE EDGE IF NOT EXISTS like(likeness double)"
	//
	//		// Execute a query
	//		resultSet, err := session.Execute(createSchema)
	//		if err != nil {
	//			fmt.Print(err.Error())
	//			return
	//		}
	//		checkResultSet(createSchema, resultSet)
	//	}
	//	time.Sleep(5 * time.Second)
	//	{
	//		insertVertexes := "INSERT VERTEX person(name, age) VALUES " +
	//			"'Bob':('Bob', 10), " +
	//			"'Lily':('Lily', 9), " +
	//			"'Tom':('Tom', 10), " +
	//			"'Jerry':('Jerry', 13), " +
	//			"'John':('John', 11);"
	//
	//		// Insert multiple vertexes
	//		resultSet, err := session.Execute(insertVertexes)
	//		if err != nil {
	//			fmt.Print(err.Error())
	//			return
	//		}
	//		checkResultSet(insertVertexes, resultSet)
	//	}
	//	{
	//		// Insert multiple edges
	//		insertEdges := "INSERT EDGE like(likeness) VALUES " +
	//			"'Bob'->'Lily':(80.0), " +
	//			"'Bob'->'Tom':(70.0), " +
	//			"'Lily'->'Jerry':(84.0), " +
	//			"'Tom'->'Jerry':(68.3), " +
	//			"'Bob'->'John':(97.2);"
	//
	//		resultSet, err := session.Execute(insertEdges)
	//		if err != nil {
	//			fmt.Print(err.Error())
	//			return
	//		}
	//		checkResultSet(insertEdges, resultSet)
	//	}
	//	// Extract data from the resultSet
	//	{
	//		query := "GO FROM 'Bob' OVER like YIELD $^.person.name, $^.person.age, like.likeness"
	//		// Send query
	//		resultSet, err := session.Execute(query)
	//		if err != nil {
	//			fmt.Print(err.Error())
	//			return
	//		}
	//		checkResultSet(query, resultSet)
	//
	//		// Get all column names from the resultSet
	//		colNames := resultSet.GetColNames()
	//		fmt.Printf("column names: %s\n", strings.Join(colNames, ", "))
	//
	//		// Get a row from resultSet
	//		record, err := resultSet.GetRowValuesByIndex(0)
	//		if err != nil {
	//			log.Error(err.Error())
	//		}
	//		// Print whole row
	//		fmt.Printf("row elements: %s\n", record.String())
	//		// Get a value in the row by column index
	//		valueWrapper, err := record.GetValueByIndex(0)
	//		if err != nil {
	//			log.Error(err.Error())
	//		}
	//		// Get type of the value
	//		fmt.Printf("valueWrapper type: %s \n", valueWrapper.GetType())
	//		// Check if valueWrapper is a string type
	//		if valueWrapper.IsString() {
	//			// Convert valueWrapper to a string value
	//			v1Str, err := valueWrapper.AsString()
	//			if err != nil {
	//				log.Error(err.Error())
	//			}
	//			fmt.Printf("Result of ValueWrapper.AsString(): %s\n", v1Str)
	//		}
	//		// Print ValueWrapper using String()
	//		fmt.Printf("Print using ValueWrapper.String(): %s", valueWrapper.String())
	//	}
	//	//// Drop space
	//	//{
	//	//	query := "DROP SPACE IF EXISTS example_space"
	//	//	// Send query
	//	//	resultSet, err := session.Execute(query)
	//	//	if err != nil {
	//	//		fmt.Print(err.Error())
	//	//		return
	//	//	}
	//	//	checkResultSet(query, resultSet)
	//	//}
	//}(&wg)
	//wg.Wait()
	//
	fmt.Print("\n")
	log.Info("Nebula Go Client Goroutines Example Finished")
}
