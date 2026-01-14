func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, data *{{.upperStartCamelObject}}) (sql.Result,error) {
    // --------------------------------------------------------------------------------
    // [AUTO-FILL] 自动填充时间 (NOT NULL 版本)
    // --------------------------------------------------------------------------------
    now := time.Now()
    nowUnix := now.UnixMilli()
    // 直接赋值 int64，不需要 sql.NullInt64
    data.CreatedTime = nowUnix
    data.UpdatedTime = nowUnix

    // 直接赋值 time.Time，不需要 sql.NullTime
    data.CreatedDate = now
    data.UpdatedDate = now
    // --------------------------------------------------------------------------------

	{{if .withCache}}{{.keys}}
    ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values ({{.expression}})", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, {{.expressionValues}})
	}, {{.keyValues}}){{else}}query := fmt.Sprintf("insert into %s (%s) values ({{.expression}})", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
    ret,err:=m.conn.ExecCtx(ctx, query, {{.expressionValues}}){{end}}
	return ret,err
}
