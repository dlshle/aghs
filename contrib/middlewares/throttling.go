package middlewares

import (
	"github.com/dlshle/aghs/constant"
	"github.com/dlshle/aghs/server"
	throttle "github.com/dlshle/aghs/utils"
	"net/http"
	"strconv"
	"time"
)

const (
	ThrottleWindowNextKey   = "X-Next-Window"
	ThrottleWindowRemainKey = "X-Hit-Remain"
)

type ThrottleCriteria interface {
	Hit(ctx server.MiddlewareContext) (throttle.Record, error)
}

type IpThrottleCriteria struct {
	controller throttle.Controller
	limit      int
}

func newIpThrottleCriteria(limit int) IpThrottleCriteria {
	return IpThrottleCriteria{
		controller: throttle.NewThrottleController(),
		limit:      limit,
	}
}

func (c IpThrottleCriteria) Hit(ctx server.MiddlewareContext) (throttle.Record, error) {
	return c.controller.Hit(ctx.Request().RemoteAddress(), c.limit, time.Minute)
}

func ThrottlingMiddleware(criteria ThrottleCriteria) server.Middleware {
	return func(ctx server.MiddlewareContext) {
		rec, err := criteria.Hit(ctx)
		defer func() {
			hitRemains := rec.Limit - rec.HitsUnderWindow
			if hitRemains < 0 {
				hitRemains = 0
			}
			ctx.Response().SetHeader(ThrottleWindowNextKey, rec.WindowExpiration.Format(constant.ISOTimeFormat))
			ctx.Response().SetHeader(ThrottleWindowRemainKey, strconv.Itoa(hitRemains))
		}()
		if err != nil {
			ctx.Response().SetCode(http.StatusTooManyRequests)
			ctx.Response().SetPayload(nil)
			return
		}
		ctx.Next()
	}
}

func NewIpThrottlingMiddleware(limit int) server.Middleware {
	return ThrottlingMiddleware(newIpThrottleCriteria(limit))
}
