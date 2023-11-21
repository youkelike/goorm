package slowquery

import (
	"context"
	"log"
	"time"

	"gitee.com/youkelike/orm"
)

type MiddlewareBuilder struct {
	threshold time.Duration
	logFunc   func(query string, args []any)
}

func NewMiddlerwareBuild(threshold time.Duration) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		threshold: threshold,
		logFunc: func(query string, args []any) {
			log.Printf("sql: %s, args: %v", query, args)
		},
	}
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				if duration < m.threshold {
					return
				}
				q, err := qc.Builder.Build()
				if err == nil {
					m.logFunc(q.SQL, q.Args)
				}
			}()
			return next(ctx, qc)
		}
	}
}
