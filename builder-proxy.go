package daprdriver

import (
	"log"

	"google.golang.org/grpc/resolver"
)

type proxyBuilder struct{}

func (d *proxyBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (
	resolver.Resolver, error) {
	log.Printf("in builder: host is: %s", getDaprHost(SchemaProxiedGrpc, target.URL.Host))
	if err := cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{{Addr: getDaprHost(SchemaProxiedGrpc, target.URL.Host)}},
	}); err != nil {
		return nil, err
	}

	return &nopResolver{cc: cc}, nil
}

func (d *proxyBuilder) Scheme() string {
	return SchemaProxiedGrpc
}
