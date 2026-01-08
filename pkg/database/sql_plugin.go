package database

import (
	"time"

	"gin-app-start/internal/common"
	"gin-app-start/pkg/timeutil"
	"gin-app-start/pkg/trace"

	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

const (
	callBackBeforeName = "core:before"
	callBackAfterName  = "core:after"
	startTime          = "_start_time"
)

type TracePlugin struct{}

func (op *TracePlugin) Name() string {
	return "tracePlugin"
}

// 这个函数是GORM数据库操作追踪插件的初始化函数，负责为所有数据库操作注册前后拦截器，实现SQL执行追踪和性能监控
func (op *TracePlugin) Initialize(db *gorm.DB) (err error) {
	// 开始前
	_ = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before)       // 创建操作前
	_ = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before)                // 查询操作前
	_ = db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before)       // 删除操作前
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, before) // 更新操作前
	_ = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before)                    // 行级操作前
	_ = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before)                    // 原始SQL操作前

	// 结束后
	_ = db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after) // 创建操作后
	_ = db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after)   // 查询操作后
	_ = db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after) // 删除操作后
	_ = db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after) // 更新操作后
	_ = db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)             // 行级操作后
	_ = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)             // 原始SQL操作后
	return
}

// 确保实现了完整的gorm.Plugin插件接口
var _ gorm.Plugin = &TracePlugin{}

func before(db *gorm.DB) {
	// 记录SQL执行的开始时间戳
	db.InstanceSet(startTime, time.Now())
}

func after(db *gorm.DB) {
	// 1. 获取上下文和开始时间
	_ctx := db.Statement.Context
	ctx, ok := _ctx.(common.StdContext)
	if !ok {
		return
	}

	_ts, isExist := db.InstanceGet(startTime)
	if !isExist {
		return
	}

	ts, ok := _ts.(time.Time)
	if !ok {
		return
	}

	// 2. 构建SQL追踪信息
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)

	sqlInfo := new(trace.SQL)
	sqlInfo.Timestamp = timeutil.CSTLayoutString() // 中国标准时间
	sqlInfo.SQL = sql                              // 完整的SQL语句
	sqlInfo.Stack = utils.FileWithLineNum()        // 文件地址和行号
	sqlInfo.Rows = db.Statement.RowsAffected       // 受影响的行数
	sqlInfo.CostSeconds = time.Since(ts).Seconds() // 执行耗时（秒）

	// 3. 追加到上下文的SQL追踪列表
	if ctx.Trace != nil {
		ctx.Trace.AppendSQL(sqlInfo)
	}
}
