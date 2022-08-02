package daprdriver

import (
	"os"

	"google.golang.org/grpc/resolver"
)

type daprBuilder struct{}

func (d *daprBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (
	resolver.Resolver, error) {
	if err := cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{resolver.Address{Addr: os.Getenv("DAPR_GRPC_HOST") + ":" + os.Getenv("DAPR_GRPC_PORT")}},
	}); err != nil {
		return nil, err
	}

	return &nopResolver{cc: cc}, nil
}

func (d *daprBuilder) Scheme() string {
	return sGrpc
}
