package trace

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/singleflight"
)

// funcNameCache 缓存函数名称，避免重复调用 runtime.FuncForPC
var funcNameCache = make(map[uintptr]string)
var funcNameCacheMutex sync.RWMutex
var singleFlight singleflight.Group

// Span 用于表示追踪的一个节点
type Span struct {
	SpanID       int
	ParentSpanID int
	Attributes   map[string]interface{}
}

// TraceCtx 表示一个完整的追踪上下文
type TraceCtx struct {
	context.Context
	level     zapcore.Level
	logger    *zap.Logger
	TraceID   string
	Span      *Span
	StartTime time.Time
}

// 初始化 zap.Logger
func newLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stack"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return logger
}

var logger = newLogger()

func Init(l *zap.Logger) {
	logger = l
}

// getFuncName 获取调用 StartTrace 的函数名称，使用缓存来优化性能
func getFuncName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	funcNameCacheMutex.RLock()
	name, found := funcNameCache[pc]
	funcNameCacheMutex.RUnlock()
	if found {
		return name
	}

	singleFlight.Do(fmt.Sprint(pc), func() (interface{}, error) {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			name = fn.Name()
		} else {
			name = "unknown"
		}
		funcNameCacheMutex.Lock()
		funcNameCache[pc] = name
		funcNameCacheMutex.Unlock()
		return nil, nil
	})
	return name
}

func New(cxt context.Context, logger *zap.Logger) *TraceCtx {
	traceCtx := &TraceCtx{
		logger:    logger,
		level:     zapcore.InfoLevel,
		Context:   ctx,
		TraceID:   fmt.Sprintf("trace-%d", time.Now().UnixNano()),
		StartTime: time.Now(),
		Span: &Span{
			SpanID: 1,
			Attributes: map[string]interface{}{
				"FuncName": funcName,
			},
		},
	}
	return traceCtx

}

// StartTrace 返回 TraceCtx，且实现了 context.Context 接口，能自动管理 Trace 和 Span
func StartTrace(ctx context.Context) *TraceCtx {
	funcName := getFuncName()
	// 如果 ctx 中没有 TraceCtx，则创建一个新的 TraceCtx
	lastCtx, ok := ctx.Value("traceKey").(*TraceCtx)
	if !ok {
		return New(ctx, logger)
	}
	// 基于父 TraceCtx 创建新的 TraceCtx
	traceCtx := &TraceCtx{
		Context:   ctx,
		level:     zapcore.InfoLevel,
		logger:    lastCtx.Logger,
		TraceID:   lastCtx.TraceID,
		StartTime: time.Now(),
		Span: &Span{
			ParentSpanID: lastCtx.Span.SpanID,
			SpanID:       lastCtx.Span.SpanID + 1,
			Attributes: map[string]interface{}{
				"FuncName": funcName,
			},
		},
	}
	return traceCtx
}

// AddAttribute 向 Span 添加自定义属性
func (traceCtx *TraceCtx) AddAttribute(key string, value interface{}) *TraceCtx {
	if traceCtx.Span != nil {
		traceCtx.Span.Attributes[key] = value
	}
	return traceCtx
}

// EndTrace 输出 span 的日志，带有自定义属性（合并到一条日志）
func (traceCtx *TraceCtx) EndTrace() {
	if traceCtx.Logger == nil {
		return
	}

	span := traceCtx.Span
	duration := time.Since(traceCtx.StartTime)
	// 合并自定义属性到日志输出字段
	fields := []zap.Field{
		zap.String("traceId", traceCtx.TraceID),
		zap.Time("startTime", traceCtx.StartTime),
		zap.Int("spanId", span.SpanID),
		zap.Int("parentSpanId", span.ParentSpanID),
		zap.Duration("duration", duration),
	}

	// 自定义属性
	for key, value := range span.Attributes {
		fields = append(fields, zap.Any(key, value))
	}

	// 输出日志
	traceCtx.Logger.Check(traceCtx.level, "Trace").Write(fields...)
}

// 实现 context.Context 接口中的 Value 方法
func (traceCtx *TraceCtx) Value(key interface{}) interface{} {
	if key == "traceKey" {
		return traceCtx
	}
	return traceCtx.Context.Value(key)
}

func (traceCtx *TraceCtx) SetLevel(lv zapcore.Level) *TraceCtx {
	traceCtx.level = lv
	return traceCtx
}
