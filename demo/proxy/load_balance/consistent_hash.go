package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

func main() {
	csb := NewConsistentHashBalance(10, nil)
	csb.Add("127.0.0.1:2003") //0
	csb.Add("127.0.0.1:2004") //1
	csb.Add("127.0.0.1:2005") //2
	csb.Add("127.0.0.1:2006") //3
	csb.Add("127.0.0.1:2007") //4

	//url hash
	fmt.Println(csb.Get("http://127.0.0.1:2002/base/getinfo"))
	fmt.Println(csb.Get("http://127.0.0.1:2002/base/error"))
	fmt.Println(csb.Get("http://127.0.0.1:2002/base/getinfo"))
	fmt.Println(csb.Get("http://127.0.0.1:2002/base/changepwd"))

	//ip hash
	fmt.Println(csb.Get("127.0.0.1"))
	fmt.Println(csb.Get("192.168.0.1"))
	fmt.Println(csb.Get("127.0.0.1"))
}

// 一致性hash算法负载均衡

type Hash func(data []byte) uint32

type UInt32Slice []uint32

func (s UInt32Slice) Len() int {
	return len(s)
}

func (s UInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s UInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ConsistentHashBalance struct {
	mux      sync.RWMutex // 读写锁
	hash     Hash
	replicas int               // 复制因子
	keys     UInt32Slice       // 已排序的节点hash切片
	hashMap  map[uint32]string // 节点哈希和key的map, key hash value port

	// 观察主体,暂未用到
	// conf LoadBalanceConf
}

func NewConsistentHashBalance(replicas int, fn Hash) *ConsistentHashBalance {
	m := &ConsistentHashBalance{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[uint32]string),
	}
	if m.hash == nil {
		// 最多32位,保证是一个2^32-1环
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty 验证是否为空
func (c *ConsistentHashBalance) IsEmpty() bool {
	return len(c.keys) == 0
}

// Add 方法用来添加缓存节点，参数为节点key，比如使用IP
func (c *ConsistentHashBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("param len 1 at least")
	}
	addr := params[0]
	c.mux.Lock()
	defer c.mux.Unlock()
	// 结合复制因子计算所有虚拟节点的hash值,并存入m.keys中,同时在m.hashMap中保存哈希值和key的映射
	for i := 0; i < c.replicas; i++ {
		hash := c.hash([]byte(strconv.Itoa(i) + addr))
		c.keys = append(c.keys, hash)
		c.hashMap[hash] = addr
	}
	// 对所有虚拟节点的哈希值进行排序,方便之后进行二分查找
	sort.Sort(c.keys)
	return nil
}

// Get 方法根据给定的对象获取最靠近它的那个节点
func (c *ConsistentHashBalance) Get(key string) (string, error) {
	if c.IsEmpty() {
		return "", errors.New("node is empty")
	}
	hash := c.hash([]byte(key))

	// 二分查找获取最优节点,第一个服务器hash值大于数据hash值的就是最优服务器节点
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hash
	})

	// 如果查找结果大于服务器节点哈希数组的最大索引,表示此时该对象哈希值位于最后一个节点之后,那么放入第一个节点中
	if idx == len(c.keys) {
		idx = 0
	}
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.hashMap[c.keys[idx]], nil
}

func (c *ConsistentHashBalance) Update() {
	// 后续接入
}
