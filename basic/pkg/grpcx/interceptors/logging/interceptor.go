package logging

import (
	"compus_blog/basic/pkg/grpcx/interceptors"
	logger2 "compus_blog/basic/pkg/logger"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
)

type InterceptorBuilder struct {
	// 如果你要非常通用
	l logger2.LoggerV1
	//fn func(msg string, fields...logger.Field)
	interceptors.Builder

	reqBody  bool
	respBody bool
}

func (i *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		//start := time.Now()
		//event := "normal"
		//defer func() {
		//	// 照着抄
		//}()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		var event = "normal"
		defer func() {
			// 执行时间
			duration := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}
			fields := []logger2.Field{
				logger2.Int64("cost", duration.Milliseconds()),
				logger2.String("type", "unary"),
				logger2.String("method", info.FullMethod),
				logger2.String("event", event),
				// 这一个部分，是需要你的客户端配合的,
				// 你需要知道是哪一个业务调用过来的
				// 是哪个业务的哪个节点过来的
				logger2.String("peer", i.PeerName(ctx)),
				logger2.String("peer_ip", i.PeerIP(ctx)),
			}
			if err != nil {
				st, _ := status.FromError(err)
				fields = append(fields, logger2.String("code",
					st.Code().String()),
					logger2.String("code_msg", st.Message()))
			}

			i.l.Info("RPC请求", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
