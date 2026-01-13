gen-model:
	goctl model mysql ddl --src ./script/mysql/user.sql --dir ./internal/model/user --style go_zero -c
	goctl model mysql ddl --src ./script/mysql/user_api.sql --dir ./internal/model/user_api --style go_zero -c
	goctl model mysql ddl --src ./script/mysql/knowledge.sql --dir ./internal/model/knowledge --style go_zero
	goctl model mysql ddl --src ./script/mysql/knowledge_retrieval_log.sql --dir ./internal/model/knowledge_retrieval_log --style go_zero
	goctl model mysql ddl --src ./script/mysql/chat_message.sql --dir ./internal/model/chat_message --style go_zero
	goctl model mysql ddl --src ./script/mysql/chat_conversation.sql --dir ./internal/model/chat_conversation --style go_zero -c
gen-api:
	goctl api go --api ./restful/rag/rag.api --dir ./restful/rag --style go_zero

gen-doc:
	goctl api swagger --api ./restful/rag/rag.api --dir ./docs/swagger --filename rag