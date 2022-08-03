package daprdriver

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/dapr/dapr/pkg/proto/common/v1"
	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dtm-labs/dtmdriver"
	"github.com/dtm-labs/logger"
	"google.golang.org/protobuf/types/known/anypb"
)

type DaprAddr struct {
	Schema     string
	Host       string
	Appid      string
	MethodName string // method name of dapr, or path of url without preceeding '/'
}

func getDaprHost(schema string, host string) string {
	if host != cDaprEnv {
		return host
	}
	if schema == SchemaGrpc || schema == SchemaProxiedGrpc {
		return "localhost:" + os.Getenv("DAPR_GRPC_PORT")
	}
	return "localhost:" + os.Getenv("DAPR_HTTP_PORT")
}

func ParseDaprUrl(uri string) (DaprAddr, error) {
	res := DaprAddr{}
	fs := strings.Split(uri, "/")
	if !strings.HasSuffix(fs[0], ":") {
		return res, nil
	}
	if len(fs) < 4 {
		return res, fmt.Errorf("dapr url format, should be %s but got: %s", format, uri)
	}
	res.Schema = strings.TrimSuffix(fs[0], ":")
	res.Host = getDaprHost(res.Schema, fs[2])
	res.Appid = fs[3]
	if len(fs) > 4 {
		res.MethodName = strings.Join(fs[4:], "/")
	}
	logger.Debugf("uri %s parsed result is %v", uri, res)
	return res, nil
}

func Use() {
	dtmdriver.Use(DriverName)
}

func AddrForProxiedHTTP(appid string, pathAndQuery string) string {
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}
	return fmt.Sprintf("%s://DAPR_ENV/%s%s", SchemaProxiedHTTP, appid, pathAndQuery)
}

func AddrForProxiedGrpc(appid string, pathAndQuery string) string {
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}
	return fmt.Sprintf("%s://DAPR_ENV/%s%s", SchemaProxiedGrpc, appid, pathAndQuery)
}

func AddrForGrpc(appid string, method string) string {
	return fmt.Sprintf("%s://DAPR_ENV/%s/%s", SchemaGrpc, appid, method)
}
func AddrForHTTP(appid string, method string) string {
	return fmt.Sprintf("%s://DAPR_ENV/%s/%s", SchemaHTTP, appid, method)
}

func NewDaprGrpcURL(service string, method string) string {
	return fmt.Sprintf("%s://localhost/%s/dapr.proto.runtime.v1.Dapr/InvokeService/%s", SchemaGrpc, service, method)
}

func PayloadForGrpc(data []byte) *pb.InvokeServiceRequest {
	return &pb.InvokeServiceRequest{
		Id: "dummy-id",
		Message: &v1.InvokeRequest{
			Data:          &anypb.Any{Value: data},
			Method:        "dummy-method",
			ContentType:   "application/json",
			HttpExtension: &v1.HTTPExtension{Verb: v1.HTTPExtension_Verb(3), Querystring: "a=1"},
		},
	}
}

func NewDaprGrpcOldURL(service string, grpcPath string) string {
	return fmt.Sprintf("%s://localhost/%s%s", SchemaProxiedGrpc, service, grpcPath)
}
