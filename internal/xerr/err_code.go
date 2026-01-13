package xerr

// OK 成功返回
var message map[uint32]string

const (
	OK            uint32 = 200
	Unauthorized  uint32 = 401
	BadRequest    uint32 = 400
	InternalError uint32 = 500

	// ServerCommonError 全局错误码
	ServerCommonError uint32 = 100001

	FileInternalError         uint32 = 101001 // 文件操作内部错误
	FileNotFoundError         uint32 = 101002 // 文件不存在
	FileGetInfoError          uint32 = 101003 // 获取文件信息失败
	FileInvalidFormatError    uint32 = 101004 // 文件格式无效
	FileReadNoPermissionError uint32 = 101005 // 无文件读取权限

	/**(前3位代表业务,后三位代表具体功能)**/

	// UserLoginError 用户相关错误码 (2xx)

	UserLoginError           uint32 = 200001 // 登录失败
	UserNotFoundError        uint32 = 200002 // 用户不存在
	UserAlreadyExistError    uint32 = 200003 // 用户已存在
	UserPasswordError        uint32 = 200006 // 密码错误
	UserRegisterError        uint32 = 200007 // 注册失败
	UserSessionNotFoundError uint32 = 200008 // 会话不存在
	UserSessionInvalidError  uint32 = 200009 // 会话参数非法
	UserAnswerEmptyError     uint32 = 200010 // 答案不能为空

	// ResumeUploadError 简历上传相关错误码 (3xx)
	ResumeUploadError            uint32 = 300001 // 简历上传失败
	ResumeTooLargeError          uint32 = 300002 // 简历文件大小超过限制
	ResumeFileNotFoundError      uint32 = 300003 // 未找到简历文件
	ResumeFileExtNotSupportError uint32 = 300004 // 不支持的文件格式
	ResumeNotExistError          uint32 = 300005 // 简历不存在

	// UserApiError 用户API配置相关错误码 (4xx)
	UserApiModelNameExistError   uint32 = 400001 // 模型名称已存在
	UserApiInvalidModelTypeError uint32 = 400002 // 无效的模型类型
	UserApiNotFoundError         uint32 = 400003 // API配置不存在
	UserApiAddError              uint32 = 400004 // 添加API配置失败

	// KnowledgeError 知识库相关错误码 (5xx)
	KnowledgeBaseNotFoundError uint32 = 500001 // 知识库不存在
	KnowledgeDocUploadError    uint32 = 500002 // 文档上传失败
	KnowledgeDocTypeNotSupport uint32 = 500003 // 不支持的文档类型
	KnowledgeDocTooLargeError  uint32 = 500004 // 文档大小超过限制
	KnowledgeDocNotFoundError  uint32 = 500005 // 文档不存在
	KnowledgeDocSaveError      uint32 = 500006 // 文档保存失败

	// VectorStoreError 向量存储相关错误码 (6xx)
	VectorStoreError           uint32 = 600001 // 内部错误
	VectorStoreConnError       uint32 = 600002 // 连接失败
	VectorStoreCollectionError uint32 = 600003 // 集合操作失败
	VectorStoreInsertError     uint32 = 600004 // 插入失败
	VectorStoreSearchError     uint32 = 600005 // 检索失败
)

func init() {
	message = make(map[uint32]string)
	message[OK] = "SUCCESS"
	message[Unauthorized] = "未授权，请先登录"
	message[ServerCommonError] = "服务器开小差啦,稍后再来试一试"
	message[BadRequest] = "请求参数错误"

	// 文件操作相关错误消息
	message[FileInternalError] = "文件操作内部错误"
	message[FileNotFoundError] = "文件不存在"
	message[FileGetInfoError] = "获取文件信息失败"
	message[FileInvalidFormatError] = "文件格式无效"
	message[FileReadNoPermissionError] = "无文件读取权限"

	// 用户相关错误消息
	message[UserLoginError] = "登录失败，请检查用户名和密码"
	message[UserNotFoundError] = "用户不存在"
	message[UserAlreadyExistError] = "用户已存在"
	message[UserPasswordError] = "用户名或密码错误"
	message[UserRegisterError] = "注册失败,请检查用户名、邮箱、密码是否符合要求"
	message[UserSessionNotFoundError] = "会话不存在"
	message[UserSessionInvalidError] = "会话参数非法"
	message[UserAnswerEmptyError] = "答案不能为空"

	// 简历上传相关错误消息
	message[ResumeUploadError] = "简历上传失败,请检查文件格式是否正确"
	message[ResumeTooLargeError] = "简历文件大小超过限制,请上传小于20MB的文件"
	message[ResumeFileNotFoundError] = "未找到简历文件"
	message[ResumeFileExtNotSupportError] = "不支持的文件格式,请上传pdf,docx,doc文件"
	message[ResumeNotExistError] = "简历不存在"

	// 用户API配置相关错误消息
	message[UserApiModelNameExistError] = "模型名称已存在"
	message[UserApiInvalidModelTypeError] = "无效的模型类型,仅支持:embedding,chat,qa,rewrite,rerank"
	message[UserApiNotFoundError] = "API配置不存在"
	message[UserApiAddError] = "添加API配置失败"

	// 知识库相关错误消息
	message[KnowledgeBaseNotFoundError] = "知识库不存在"
	message[KnowledgeDocUploadError] = "文档上传失败"
	message[KnowledgeDocTypeNotSupport] = "不支持的文档类型,仅支持pdf/txt/docx/md"
	message[KnowledgeDocTooLargeError] = "文档大小超过限制,请上传小于50MB的文件"
	message[KnowledgeDocNotFoundError] = "文档不存在"
	message[KnowledgeDocSaveError] = "文档保存失败"

	// 向量存储相关错误消息 (6xx)
	message[VectorStoreError] = "向量存储内部错误"
	message[VectorStoreConnError] = "向量数据库连接失败"
	message[VectorStoreCollectionError] = "向量集合操作失败"
	message[VectorStoreInsertError] = "向量数据插入失败"
	message[VectorStoreSearchError] = "向量检索失败"
}
