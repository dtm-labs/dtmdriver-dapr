package daprdriver

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/proto"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dtm-labs/dtmdriver"
	"github.com/dtm-labs/logger"
	"github.com/go-resty/resty/v2"
)

const (
	DriverName = "dtm-driver-dapr"

	format            = "<schema>://<host>/<dapr-app-id>/<method-name>"
	cDaprEnv          = "DAPR_ENV"
	cAppid            = "dapr-app-id"
	SchemaHTTP        = "daprhttp"
	SchemaProxiedHTTP = "daprphttp"
	SchemaGrpc        = "daprgrpc"
	SchemaProxiedGrpc = "daprpgrpc"
)

type (
	darpDriver struct{}
)

func (z *darpDriver) GetName() string {
	return DriverName
}

func (z *darpDriver) RegisterAddrResolver() {
	dtmdriver.Middlewares.Grpc = append(dtmdriver.Middlewares.Grpc, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		target := cc.Target()
		addr, err := ParseDaprUrl(target)
		if err != nil {
			return err
		}
		if addr.Schema == SchemaProxiedGrpc {
			ctx = metadata.AppendToOutgoingContext(ctx, cAppid, addr.Appid)
		} else if addr.Schema == SchemaGrpc {
			method2 := strings.TrimPrefix(method, "/")
			updateReq := func(r *pb.InvokeServiceRequest) {
				r.Id = addr.Appid
				r.Message.Method = method2
			}
			req2, ok := req.(*pb.InvokeServiceRequest)
			if ok { // if dtm SDK call branch, req is type of *pb.InvokeServiceRequest
				updateReq(req2)
			} else { // if dtm server call branch directly, req is type of []byte
				var req3 pb.InvokeServiceRequest
				err := proto.Unmarshal(req.([]byte), &req3)
				if err == nil {
					updateReq(&req3)
					req, err = proto.Marshal(&req3)
				}
				if err != nil {
					return err
				}
			}
			method = "/dapr.proto.runtime.v1.Dapr/InvokeService"
		}
		logger.Debugf("invoking target: %s, method: %s", target, method)
		return invoker(ctx, method, req, reply, cc, opts...)
	})

	resolver.Register(&proxyBuilder{})
	resolver.Register(&daprBuilder{})

	dtmdriver.Middlewares.HTTP = append(dtmdriver.Middlewares.HTTP, func(c *resty.Client, r *resty.Request) error {
		addr, err := ParseDaprUrl(r.URL)
		if err != nil {
			return err
		}
		old := r.URL
		if addr.Schema == SchemaProxiedHTTP {
			r.SetHeader(cAppid, addr.Appid)
			r.URL = fmt.Sprintf("http://%s/%s", addr.Host, addr.MethodName)
		} else if addr.Schema == SchemaHTTP {
			r.URL = fmt.Sprintf("http://%s/v1.0/invoke/%s/method/%s", addr.Host, addr.Appid, addr.MethodName)
		}
		logger.Debugf("url %s resolved to %s", old, r.URL)
		return nil
	})
}

func (z *darpDriver) RegisterService(target string, endpoint string) error {
	return nil
}

func (z *darpDriver) ParseServerMethod(uri string) (server string, method string, err error) {
	addr, err := ParseDaprUrl(uri)
	if addr.Schema == "" {
		fs := strings.Split(uri, "/")
		return fs[0], "/" + strings.Join(fs[1:], "/"), nil
	}

	return fmt.Sprintf("%s://%s/%s", addr.Schema, addr.Host, addr.Appid), "/" + addr.MethodName, err
}

func init() {
	dtmdriver.Register(&darpDriver{})
}
