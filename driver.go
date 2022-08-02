package daprdriver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/protobuf/proto"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dtm-labs/dtmdriver"
	"github.com/go-resty/resty/v2"
)

const (
	DriverName = "dtm-driver-dapr"

	format   = "<schema>://<host>/<dapr-app-id>/<method-name>"
	cDaprEnv = "DAPR_ENV"
	cAppid   = "dapr-app-id"
	// daprh://localhost/v1.0/invoke/[dapr-app-id]/method
	sHTTP = "daprhttp"
	// daprho://localhost/[dapr-app-id]/[oldpath]
	spHTTP = "daprphttp"
	// daprg://localhost/[dapr-app-id]/[method]/dapr.proto.runtime.v1.Dapr/InvokeService
	sGrpc = "daprgrpc"
	// daprgo://localhost/[dapr-app-id]/[oldpath]
	spGrpc = "daprpgrpc"
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
		if addr.Schema == spGrpc {
			ctx = metadata.AppendToOutgoingContext(ctx, cAppid, addr.Appid)
		} else if addr.Schema == sGrpc {
			fmt.Printf("target: %s, method: %s\n", target, method)
			method2 := strings.TrimPrefix(method, "/")
			updateReq := func(r *pb.InvokeServiceRequest) {
				r.Id = addr.Appid
				r.Message.Method = method2
			}
			req2, ok := req.(*pb.InvokeServiceRequest)
			if !ok {
				var req3 pb.InvokeServiceRequest
				err := proto.Unmarshal(req.([]byte), &req3)
				if err != nil {
					return err
				}
				updateReq(&req3)
				req, err = proto.Marshal(&req3)
				if err != nil {
					return err
				}
			} else {
				updateReq(req2)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
	resolver.Register(&proxyBuilder{})
	resolver.Register(&daprBuilder{})

	dtmdriver.Middlewares.HTTP = append(dtmdriver.Middlewares.HTTP, func(c *resty.Client, r *resty.Request) error {
		addr, err := ParseDaprUrl(r.URL)
		if err != nil {
			return err
		}
		if addr.Schema == spHTTP {
			r.SetHeader(cAppid, addr.Appid)
			r.URL = fmt.Sprintf("http://%s/%s", addr.Host, addr.MethodName)
		} else if addr.Schema == sHTTP {
			r.URL = fmt.Sprintf("http://%s/v1.0/invoke/%s/%s", addr.Host, addr.Appid, addr.MethodName)
		}
		return nil
	})
}

func (z *darpDriver) RegisterService(target string, endpoint string) error {
	return nil
}

func (z *darpDriver) ParseServerMethod(uri string) (server string, method string, err error) {
	fs := strings.Split(uri, "/")
	if len(fs) < 5 {
		return "", "", fmt.Errorf("dapr url format, should be %s but got: %s", format, uri)
	}
	schema := fs[0]
	host := fs[2]
	appid := fs[3]

	if host == cDaprEnv {
		host = os.Getenv("DAPR_HTTP_HOST") + ":" + os.Getenv("DAPR_HTTP_PORT")
	}

	return fmt.Sprintf("%s://%s/%s", schema, host, appid), "/" + strings.Join(fs[4:], "/"), nil
}

func init() {
	dtmdriver.Register(&darpDriver{})
}
