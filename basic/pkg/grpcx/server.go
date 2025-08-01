// package grpcx
//
// import (
//
//	logger2 "compus_blog/basic/pkg/logger"
//	"compus_blog/basic/pkg/netx"
//	"context"
//	clientv3 "go.etcd.io/etcd/client/v3"
//	"go.etcd.io/etcd/client/v3/naming/endpoints"
//	"google.golang.org/grpc"
//	"net"
//	"strconv"
//	"time"
//
// )
//
//	type Server struct {
//		*grpc.Server
//		Port int
//		// ETCD 服务注册租约 TTL
//		EtcdTTL     int64
//		EtcdClient  *clientv3.Client
//		etcdManager endpoints.Manager
//		etcdKey     string
//		cancel      func()
//		Name        string
//		L           logger2.LoggerV1
//	}
//
// // Serve 启动服务器并且阻塞
//
//	func (s *Server) Serve() error {
//		// 初始化一个控制整个过程的 ctx
//		// 你也可以考虑让外面传进来，这样的话就是 main 函数自己去控制了
//		ctx, cancel := context.WithCancel(context.Background())
//		s.cancel = cancel
//		port := strconv.Itoa(s.Port)
//		l, err := net.Listen("tcp", ":"+port)
//		if err != nil {
//			return err
//		}
//		// 要先确保启动成功，再注册服务
//		err = s.register(ctx, port)
//		if err != nil {
//			return err
//		}
//		return s.Server.Serve(l)
//	}
//
//	func (s *Server) register(ctx context.Context, port string) error {
//		cli := s.EtcdClient
//		serviceName := "service/" + s.Name
//		em, err := endpoints.NewManager(cli,
//			serviceName)
//		if err != nil {
//			return err
//		}
//		s.etcdManager = em
//		ip := netx.GetOutboundIP()
//		s.etcdKey = serviceName + "/" + ip
//		addr := ip + ":" + port
//		leaseResp, err := cli.Grant(ctx, s.EtcdTTL)
//		// 开启续约
//		ch, err := cli.KeepAlive(ctx, leaseResp.ID)
//		if err != nil {
//			return err
//		}
//		go func() {
//			// 可以预期，当我们的 cancel 被调用的时候，就会退出这个循环
//			for chResp := range ch {
//				s.L.Debug("续约：", logger2.String("resp", chResp.String()))
//			}
//		}()
//		// metadata 我们这里没啥要提供的
//		return em.AddEndpoint(ctx, s.etcdKey,
//			endpoints.Endpoint{Addr: addr}, clientv3.WithLease(leaseResp.ID))
//	}
//
//	func (s *Server) Close() error {
//		s.cancel()
//		if s.etcdManager != nil {
//			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//			defer cancel()
//			err := s.etcdManager.DeleteEndpoint(ctx, s.etcdKey)
//			if err != nil {
//				return err
//			}
//		}
//		err := s.EtcdClient.Close()
//		if err != nil {
//			return err
//		}
//		s.Server.GracefulStop()
//		return nil
//	}
package grpcx

import (
	"compus_blog/basic/pkg/logger"
	"compus_blog/basic/pkg/netx"
	"context"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	EtcdAddr string
	Port     int
	Name     string
	client   *etcdv3.Client
	kaCancel func()
	L        logger.LoggerV1
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 我们要在这里完成注册
	err = s.register()
	if err != nil {
		return err
	}
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.NewFromURL(s.EtcdAddr)
	if err != nil {
		return err
	}
	s.client = client
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期
	var ttl int64 = 5
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连你
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	go func() {
		//require.NoError(t, err1)
		for kaResp := range ch {
			// 记录日志
			s.L.Debug(kaResp.String())
		}
	}()
	return err
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.client != nil {
		// 依赖注入，你就不要关
		return s.client.Close()
	}
	s.GracefulStop()
	return nil
}
