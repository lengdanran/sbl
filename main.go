package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lengdanran/sbl/proxy"
	"log"
	"net/http"
	"strconv"
)

const CONF_FILE = "./conf/config.yaml" // filename of configuration.

// readConf read the configuration from CONF_FILE
func readConf() *Config {
	conf, err := ReadConfig(CONF_FILE)
	if err != nil {
		log.Fatalf("read config error: %s", err)
		return nil
	}
	err = conf.Validation()
	if err != nil {
		log.Fatalf("verify config error: %s", err)
		return nil
	}
	conf.Print()
	return conf
}

func maxAllowedMiddleware(n uint) mux.MiddlewareFunc {
	sem := make(chan struct{}, n)
	acquire := func() { sem <- struct{}{} }
	release := func() { <-sem }

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acquire()
			defer release()
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	// 1. read configuration
	config := readConf()
	if config == nil {
		log.Printf("Read configuration from %s failed, exit....", CONF_FILE)
		return
	}
	// 2. make routers for locations
	router := mux.NewRouter()
	for _, l := range config.Location {
		httpProxy, err := proxy.NewHTTPProxy(l.ProxyPass, l.BalanceMode, l.ProxyPassWeight)
		if err != nil {
			log.Fatalf("create proxy error: %s", err)
		}
		if httpProxy == nil {
			log.Printf("Init httpProxy failed....Skip this location %v\n", l)
			continue
		}
		// start health check
		if config.HealthCheck {
			httpProxy.HealthCheck(config.HealthCheckInterval)
		}
		router.Handle(l.Pattern, httpProxy)
	}
	if config.MaxAllowed > 0 {
		router.Use(maxAllowedMiddleware(config.MaxAllowed))
	}
	svr := http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: router,
	}

	// 3. listen and serve
	log.Printf("Serve Schema = %s\n", config.Schema)
	if config.Schema == "http" {
		err := svr.ListenAndServe()
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	} else if config.Schema == "https" {
		err := svr.ListenAndServeTLS(config.SSLCertificate, config.SSLCertificateKey)
		if err != nil {
			log.Fatalf("listen and serve error: %s", err)
		}
	}

}

type WeightRoundRobinBalance struct {
	curIndex int
	rss      []*WeightNode
}

type WeightNode struct {
	weight          int    // 配置的权重，即在配置文件或初始化时约定好的每个节点的权重
	currentWeight   int    // 节点当前权重，会一直变化
	effectiveWeight int    // 有效权重，初始值为weight, 通讯过程中发现节点异常，则-1 ，之后再次选取本节点，调用成功一次则+1，直达恢复到weight 。 用于健康检查，处理异常节点，降低其权重。
	addr            string // 服务器addr
}

/**
 * @Author: yang
 * @Description：添加服务
 * @Date: 2021/4/7 15:36
 */
func (r *WeightRoundRobinBalance) Add(params ...string) error {
	if len(params) != 2 {
		return errors.New("params len need 2")
	}
	// @Todo 获取值
	addr := params[0]
	parInt, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		return err
	}
	node := &WeightNode{
		weight:          int(parInt),
		effectiveWeight: int(parInt), // 初始化時有效权重 = 配置权重值
		currentWeight:   int(parInt), // 初始化時当前权重 = 配置权重值
		addr:            addr,
	}
	r.rss = append(r.rss, node)
	return nil
}

/**
 * @Author: yang
 * @Description：轮询获取服务
 * @Date: 2021/4/7 15:36
 */
func (r *WeightRoundRobinBalance) Next() string {
	// @Todo 没有服务
	if len(r.rss) == 0 {
		return ""
	}
	fmt.Printf("currentWeight = ")
	for _, node := range r.rss {
		fmt.Printf("%d ", node.currentWeight)
	}
	fmt.Printf("\neffectiveWeight = ")
	for _, node := range r.rss {
		fmt.Printf("%d ", node.effectiveWeight)
	}
	fmt.Printf("\n")
	totalWeight := 0
	var maxWeightNode *WeightNode
	for key, node := range r.rss {
		// @Todo 计算当前状态下所有节点的effectiveWeight之和totalWeight
		totalWeight += node.effectiveWeight

		// @Todo 计算currentWeight
		node.currentWeight += node.effectiveWeight

		// @Todo 寻找权重最大的
		if maxWeightNode == nil || maxWeightNode.currentWeight < node.currentWeight {
			maxWeightNode = node
			r.curIndex = key
		}
	}

	// @Todo 更新选中节点的currentWeight
	maxWeightNode.currentWeight -= totalWeight

	// @Todo 返回addr
	return maxWeightNode.addr
}

/**
 * @Author: yang
 * @Description：测试
 * @Date: 2021/4/7 15:36
 */
func main2() {
	rb := new(WeightRoundRobinBalance)
	rb.Add("127.0.0.1:80", "2")
	rb.Add("127.0.0.1:81", "1")
	rb.Add("127.0.0.1:82", "1")

	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
	fmt.Println(rb.Next())
}
