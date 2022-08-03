package daprdriver

import (
	"github.com/dtm-labs/logger"
	"google.golang.org/grpc/resolver"
)

type proxyBuilder struct{}

func (d *proxyBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (
	resolver.Resolver, error) {
	host := getDaprHost(SchemaProxiedGrpc, target.URL.Host)
	logger.Infof("dapr resolver build host is: %s", host)
	if err := cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{{Addr: host}},
	}); err != nil {
		return nil, err
	}

	return &nopResolver{cc: cc, host: host}, nil
}

func (d *proxyBuilder) Scheme() string {
	return SchemaProxiedGrpc
}
