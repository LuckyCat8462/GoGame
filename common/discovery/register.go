package discovery

import (
	"common/config"
	"common/logs"
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Register grpc服务注册到etcd
// 原理：创建一个租约，grpc服务注册到etcd 绑定租约
// 过了租约时间，etcd就会删除grpc服务信息
// 实现心跳(完成续租)	，如果etcd没有 就重新注册
type Register struct {
	etcdCli     *clientv3.Client                        //etcd连接
	leaseId     clientv3.LeaseID                        //租约id
	DialTimeout int                                     //超时时间
	ttl         int                                     //租约时间
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse //心跳,利用channel
	info        Server                                  //注册的server信息
	closeCh     chan struct{}
}

// NewRegister 注册，维持过程
func NewRegister() *Register {
	return &Register{
		DialTimeout: 3,
	}
}

func (r *Register) Close() {
	r.closeCh <- struct{}{}
}
func (r *Register) Register(conf config.EtcdConf) error {
	//注册信息
	info := Server{
		Name:    conf.Register.Name,
		Addr:    conf.Register.Addr,
		Weight:  conf.Register.Weight,
		Version: conf.Register.Version,
		Ttl:     conf.Register.Ttl,
	}
	////建立etcd的连接
	var err error
	//1、etcd客户端
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Addrs, //端点，填写地址addrs,其是一个数组，多个时可以拼接
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return err
	}
	//2、注册相关info
	r.info = info
	//3、开始注册
	if err = r.register(); err != nil {
		return err
	}
	r.closeCh = make(chan struct{})
	//放入协程中 根据心跳的结果 做相应的操作
	go r.watcher()
	return nil
}

// 注册：创建租约、心跳检测、绑定租约
func (r *Register) register() error {
	//1. 创建租约
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(r.DialTimeout))
	defer cancel()
	var err error
	if err = r.createLease(ctx, r.info.Ttl); err != nil {
		fmt.Println("register.go：etcd租约error")
		return err
	}
	//2. 心跳检测。更新-》若没有则注册
	if r.keepAliveCh, err = r.keepAlive(); err != nil {
		fmt.Println("register.go：etcd心跳检测error")
		return err
	}
	//3. 绑定租约
	data, _ := json.Marshal(r.info)
	//etcd基于key value，所以此时也要给它一个key
	return r.bindLease(ctx, r.info.BuildRegisterKey(), string(data))
}

// createLease ttl秒
func (r *Register) createLease(ctx context.Context, ttl int64) error {
	grant, err := r.etcdCli.Grant(ctx, ttl)
	if err != nil {
		logs.Error("createLease failed,err:%v", err)
		return err
	}
	r.leaseId = grant.ID
	return nil
}

// keepAlive 心跳检测
func (r *Register) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	//心跳 要求是一个长连接 如果做了超时 长连接就断掉了 不要设置超时
	//就是一直不停的发消息 保持租约 续租
	keepAliveResponses, err := r.etcdCli.KeepAlive(context.Background(), r.leaseId)
	if err != nil {
		logs.Error("keepAlive failed,err:%v", err)
		return keepAliveResponses, err
	}
	return keepAliveResponses, nil
}

// 将info信息转换为json,使用后再重新转换回原格式
func (r *Register) bindLease(ctx context.Context, key, value string) error {
	//绑定就是：put动作
	_, err := r.etcdCli.Put(ctx, key, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		logs.Error("bindLease failed,err:%v", err)
		return err
	}
	logs.Info("register service success,key=%s", key)
	return nil
}

// watcher 续约功能。若没有，新注册；到期，是否续约；close 注销
func (r *Register) watcher() {
	//租约到期了 是不是需要去检查是否自动注册
	ticker := time.NewTicker(time.Duration(r.info.Ttl) * time.Second)
	for {
		select {
		case <-r.closeCh: //情况三：检测到了注销操作
			if err := r.unregister(); err != nil { //删除key,value
				logs.Error("close and unregister failed,err:%v", err)
			}
			//根据租约id执行租约撤销
			if _, err := r.etcdCli.Revoke(context.Background(), r.leaseId); err != nil {
				logs.Error("close and Revoke lease failed,err:%v", err)
			}
			if r.etcdCli != nil {
				r.etcdCli.Close()
			}
			logs.Info("unregister etcd...")
		case <-r.keepAliveCh: //情况二：续约；拿到心跳结果，进行续约
			//logs.Info("%v", res)
			//if res != nil {
			//	if err := r.register(); err != nil {
			//		logs.Error("keepAliveCh register failed,err:%v", err)
			//	}
			//}
		case <-ticker.C: //情况一：没有心跳检测返回，则进行相应注册操作
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					logs.Error("ticker register failed,err:%v", err)
				}
			}
		}
	}
}

func (r *Register) unregister() error {
	//删除操作
	_, err := r.etcdCli.Delete(context.Background(), r.info.BuildRegisterKey())
	return err
}
