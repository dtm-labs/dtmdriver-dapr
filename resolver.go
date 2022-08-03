package daprdriver

import (
	"github.com/dtm-labs/logger"
	"google.golang.org/grpc/resolver"
)

type nopResolver struct {
	cc   resolver.ClientConn
	host string
}

func (r *nopResolver) Close() {
}

func (r *nopResolver) ResolveNow(options resolver.ResolveNowOptions) {
	logger.Debugf("resolve now using %s", r.host)
}
