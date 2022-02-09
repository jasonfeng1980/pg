package finder

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/clientv3"

	"github.com/jasonfeng1980/pg/ecode"
	"github.com/jasonfeng1980/pg/micro/endpoint"
	"github.com/jasonfeng1980/pg/util"
	"time"
)

var etcdInstances = make(map[string]*etcdClient)

// 活动etcd实例
func NewEtcd(ctx context.Context, etcdDns string) (*etcdClient, error){
	// 有缓存就不重新链接
	if _, ok := etcdInstances[etcdDns]; ok {
		return etcdInstances[etcdDns], nil
	}

	m, e:= util.MapFromDns(etcdDns)
	if e!= nil {
		return nil, e
	}

	//创建etcd连接
	// 客户端配置
	util.Info(m.GetListString("hostList"))
	host := m.GetListString("hostList", []string{"127.0.0.1:2379"})
	util.Info("创建etcd连接", "host", host)
	config := clientv3.Config{
		Endpoints: host,
	}
	config.Username = m.GetStr("user", "")
	config.Password = m.GetStr("password", "")
	config.DialTimeout = m.GetTimeDuration("params.DialTimeout", 3) * time.Second
	config.DialTimeout = m.GetTimeDuration("params.KeepAlive", 3) * time.Second

	// 建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}
	// 判断下client的链接状态
	timeoutCtx, cancel := context.WithTimeout(ctx, 2 * time.Second)
	defer cancel()
	if _, err =client.Status(timeoutCtx, config.Endpoints[0]); err !=nil {
		return nil, err
	}


	etcdInstances[etcdDns] = &etcdClient{
		EtcdRetryTimes: 3,
		EtcdRetryTimeout: time.Second * 3,
		ctx:	ctx,
		client: client,
		kv:     clientv3.NewKV(client),
		leaser: clientv3.NewLease(client),
		watcher: make(map[string]clientv3.Watcher),
		cache:	 make(map[string][]string),
		Balancer: Balancer{
			F: Random,
		},
		Retry: Retry{
			ctx: ctx,
			Timeout: m.GetTimeDuration("params.RetryTimeout", 30) * time.Second,
			RetryTimes: m.GetInt("params.RetryTimes", 1),
		},
	}
	return etcdInstances[etcdDns],  nil
}

type etcdClient struct {
	EtcdRetryTimes int
	EtcdRetryTimeout  time.Duration
	ctx context.Context
	client  *clientv3.Client
	kv    clientv3.KV
	leaser   clientv3.Lease
	watcher  map[string]clientv3.Watcher
	hbch 	 map[string]<-chan *clientv3.LeaseKeepAliveResponse
	cache    map[string][]string

	Balancer Balancer
	Retry    Retry
}

// 设置分发的方法
func (c *etcdClient) SetBalancer(f BalancerFunc) {
	c.Balancer.F = f
}

// 监听指定的前缀
func (c *etcdClient) WatchPrefix(prefix string) {
	if _, ok := c.watcher[prefix]; !ok {
		c.watcher[prefix] = clientv3.NewWatcher(c.client)
	}
	newCtx, cancal := context.WithCancel(c.ctx)

	for  {
		watchChan := c.watcher[prefix].Watch(newCtx, prefix, clientv3.WithPrefix(), clientv3.WithRev(0))
		select {
		case <-newCtx.Done():
			cancal()
			return
		case resp := <-watchChan:
			// 任何情况都清除缓存
			delete(c.cache, prefix)
			if resp.Canceled {
				cancal()
				return
			}
		}
	}
}

// 经过缓存，查询指定前缀的值
func (c *etcdClient) Get(prefix string) (srvList []string,  err error) {
	// 1. 有缓存，就直接返回
	if v, ok := c.cache[prefix]; ok {
		return v, nil
	}

	// 2. 获取数据
	srvList, err = c.GetEntries(prefix)
	// 3. 写入缓存
	c.cache[prefix] = srvList

	// 4. 开启监控
	go c.WatchPrefix(prefix)

	return
}


// 获取指定前缀的值，无缓存，直接查询etcd
func (c *etcdClient) GetEntries(key string) ([]string, error) {
	resp, err := c.kv.Get(c.ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	srvList := make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		srvList[i] = string(kv.Value)
	}

	return srvList, nil
}

// 注册K-V
func (c *etcdClient)Register(k string, v string) error{
	s := Server{
		k,
		v,
	}
	// 1. 检查server
	if s.Key == "" {
		return ecode.EtcdMissServerK.Error()
	}
	if s.Value == "" {
		return ecode.EtcdMissServerV.Error()
	}
	// 2. 申请一个10秒的租约
	grantResp, err := c.leaser.Grant(c.ctx, 10)
	if err != nil {
		return err
	}
	leaseID := grantResp.ID
	// 3. 写入数据到etcd
	_, err = c.kv.Put(
		c.ctx,
		s.Key,
		s.Value,
		clientv3.WithLease(leaseID),
	)
	if err != nil {
		return err
	}

	// 4. 自动续租
	hbch, err := c.leaser.KeepAlive(c.ctx, leaseID)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case r := <-hbch:
				if r == nil {	// 租约失效,就退出
					return
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()
	// 删除缓存
	//c.cache = make(map[string][]string)

	return nil
}

func (c *etcdClient) Unregister(name string) error {
	defer c.close()

	if name == "" {
		return ecode.EtcdMissServerK.Error()
	}
	if _, err := c.client.Delete(c.ctx, name); err != nil {
		return err
	}

	return nil
}

func (c *etcdClient) close() {
	if c.leaser != nil {
		c.leaser.Close()
	}
	if c.watcher != nil {
		for _, v := range c.watcher{
			v.Close()
		}
	}
}

type Server struct {
	Key		string
	Value	string
}

type DoEptFunc  func(scheme string, instanceAddr string)(endpoint.Endpoint, error)

func (e *etcdClient)Endpoint(serverName string, scheme string, eptF DoEptFunc) (endpoint.Endpoint, error) {
	var (
		ept          endpoint.Endpoint
		err          error
	)
	// 获取服务列表
	prefix  := fmt.Sprintf("/%s/%s/", serverName, scheme)
	srvList, err := e.GetEntries(prefix)
	if err != nil {
		return ept, err
	}
	if len(srvList) == 0 {
		return ept, ecode.EtcdEmptySrv.Error(prefix)
	}
	
	return e.Retry.Endpoint(eptF, e.Balancer, srvList, scheme), nil
}

