package daprdriver

import (
	"google.golang.org/grpc/resolver"
)

type daprBuilder struct{}

func (d *daprBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (
	resolver.Resolver, error) {
	if err := cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{{Addr: getDaprHost(SchemaGrpc, target.URL.Host)}},
	}); err != nil {
		return nil, err
	}

	return &nopResolver{cc: cc}, nil
}

func (d *daprBuilder) Scheme() string {
	return SchemaGrpc
}
