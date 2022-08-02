package daprdriver

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/dapr/dapr/pkg/proto/common/v1"
	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dtm-labs/dtmdriver"
	"google.golang.org/protobuf/types/known/anypb"
)

type DaprAddr struct {
	Schema     string
	Host       string
	Appid      string
	MethodName string // method name of dapr, or path of url without preceeding '/'
}

func ParseDaprUrl(uri string) (DaprAddr, error) {
	res := DaprAddr{}
	fs := strings.Split(uri, "/")
	if len(fs) < 4 || !strings.HasSuffix(fs[0], ":") {
		return res, fmt.Errorf("dapr url format, should be %s but got: %s", format, uri)
	}
	res.Schema = strings.TrimSuffix(fs[0], ":")
	res.Host = fs[2]
	if res.Host == cDaprEnv {
		res.Host = os.Getenv("DAPR_HTTP_HOST") + ":" + os.Getenv("DAPR_HTTP_PORT")
	}
	res.Appid = fs[3]
	if len(fs) > 4 {
		res.MethodName = strings.Join(fs[4:], "/")
	}
	return res, nil
}

func Use() {
	dtmdriver.Use(DriverName)
}

func AddrForProxiedHTTP(appid string, pathAndQuery string) string {
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}
	return fmt.Sprintf("%s://DAPR_ENV/%s%s", spHTTP, appid, pathAndQuery)
}

func AddrForProxiedGrpc(appid string, pathAndQuery string) string {
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}
	return fmt.Sprintf("%s://DAPR_ENV/%s%s", spGrpc, appid, pathAndQuery)
}

func NewDaprGrpcURL(service string, method string) string {
	return fmt.Sprintf("%s://localhost/%s/dapr.proto.runtime.v1.Dapr/InvokeService/%s", sGrpc, service, method)
}

func NewDaprGrpcPayload(data []byte) *pb.InvokeServiceRequest {
	return &pb.InvokeServiceRequest{
		Message: &v1.InvokeRequest{
			Data:        &anypb.Any{Value: data},
			ContentType: "application/json",
		},
	}
}

func NewDaprGrpcOldURL(service string, grpcPath string) string {
	return fmt.Sprintf("%s://localhost/%s%s", spGrpc, service, grpcPath)
}
