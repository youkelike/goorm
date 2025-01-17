package prometheus

import (
	"context"
	"time"

	"gitee.com/youkelike/orm"
	"github.com/prometheus/client_golang/prometheus"
)

type MiddlewareBuilder struct {
	NameSpace string
	Subsystem string
	Name      string
	Help      string
}

func (m MiddlewareBuilder) Build() orm.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.NameSpace,
		Subsystem: m.Subsystem,
		Name:      m.Name,
		Help:      m.Help,
		Objectives: map[float64]float64{
			0.50:  0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})

	prometheus.MustRegister(vector)

	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				vector.WithLabelValues(qc.Type, qc.Model.TableName).Observe(float64(time.Since(startTime).Microseconds()))
			}()
			return next(ctx, qc)
		}
	}
}
