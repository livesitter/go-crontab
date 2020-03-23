package master

import (
	"context"
	"crontab/go-crontab/common"
	"encoding/json"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

// 单例对象
var (
	JM *JobManager
)

// 任务管理器
type JobManager struct {
	// 客户端
	client *clientv3.Client
	// KV对象
	kv clientv3.KV
	// 租约
	lease clientv3.Lease
}

// 初始化管理器
func InitJobManager() (err error) {

	var (
		client *clientv3.Client
		config clientv3.Config
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   Conf.EtcdEndpoints,
		DialTimeout: time.Duration(Conf.EtcdDialTimeout) * time.Millisecond,
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	// KV和Lease对象
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	JM = &JobManager{
		client: client,
		kv:     kv,
		lease:  lease,
	}

	return nil
}

// 保存任务
func (jm *JobManager) SaveJob(job *common.Job) (oldJob *common.Job, err error) {

	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)

	// 任务的key值
	jobKey = common.JOB_SAVE_DIR + job.Name

	// 序列化job为json
	if jobValue, err = json.Marshal(job); err != nil {
		return nil, err
	}

	// 保存到etcd
	if putResp, err = JM.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return nil, err
	}

	// 如果是更新，返回旧值
	if putResp.PrevKv != nil {

		// 反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}

		oldJob = &oldJobObj
	}

	return
}

// 删除任务
func (jm *JobManager) DeleteJob(name string) (oldJob *common.Job, err error) {

	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	// 任务的key值
	jobKey = common.JOB_SAVE_DIR + name

	// 从etcd中删除
	if delResp, err = JM.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	// 如果之前有被赋值，则将之前的值返回
	if len(delResp.PrevKvs) != 0 {

		// 反序列化
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}

		oldJob = &oldJobObj
	}
	return
}

// 任务列表
func (jm *JobManager) ListJobs() (jobList []*common.Job, err error) {

	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)

	// 任务保存路径
	dirKey = common.JOB_SAVE_DIR

	// 获取所有任务
	if getResp, err = JM.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return nil, err
	}

	// 初始化返回数组
	jobList = make([]*common.Job, 0)

	// 遍历，进行反序列化
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}

	return
}

// 杀掉任务
func (jm *JobManager) KillJob(name string) (err error) {

	// 往/cron/killer目录写入key，worker监听到了就执行强杀

	var (
		killerKey      string
		leaseId        clientv3.LeaseID
		leaseGrantResp *clientv3.LeaseGrantResponse
	)

	// 即将被杀掉的任务的key值
	killerKey = common.JOB_KILL_DIR + name

	// 让worker监听到一次put操作，创建一个租约让其1s后过期
	if leaseGrantResp, err = JM.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	// 租约ID
	leaseId = leaseGrantResp.ID

	// 设置killer标记
	if _, err = JM.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}

	return
}
