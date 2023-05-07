package proxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/lengdanran/sbl/balancer"
)

var (
	XRealIP       = http.CanonicalHeaderKey("X-Real-IP")
	XProxy        = http.CanonicalHeaderKey("X-Proxy")
	XForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	ReverseProxy  = "Balancer-Reverse-Proxy"
)

// ConnectionTimeout refers to connection timeout for health check
var ConnectionTimeout = 3 * time.Second

// HTTPProxy refers to a reverse proxy in the balancer
type HTTPProxy struct {
	hostMap      map[string]*httputil.ReverseProxy
	lb           balancer.Balancer
	sync.RWMutex // protect alive
	alive        map[string]bool
}

// GetIP get client IP
func GetClientIP(r *http.Request) string {
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	if len(r.Header.Get(XForwardedFor)) != 0 {
		xff := r.Header.Get(XForwardedFor)
		s := strings.Index(xff, ", ")
		if s == -1 {
			s = len(r.Header.Get(XForwardedFor))
		}
		clientIP = xff[:s]
	} else if len(r.Header.Get(XRealIP)) != 0 {
		clientIP = r.Header.Get(XRealIP)
	}

	return clientIP
}

// GetHost get the hostname, looks like IP:Port
func GetHost(url *url.URL) string {
	if _, _, err := net.SplitHostPort(url.Host); err == nil {
		return url.Host
	}
	if url.Scheme == "http" {
		return fmt.Sprintf("%s:%s", url.Host, "80")
	} else if url.Scheme == "https" {
		return fmt.Sprintf("%s:%s", url.Host, "443")
	}
	return url.Host
}

// IsBackendAlive Attempt to establish a tcp connection to determine whether the site is alive
func IsBackendAlive(host string) bool {
	addr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		return false
	}
	resolveAddr := fmt.Sprintf("%s:%d", addr.IP, addr.Port)
	conn, err := net.DialTimeout("tcp", resolveAddr, ConnectionTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// NewHTTPProxy create new reverse proxy with url and balancer algorithm
func NewHTTPProxy(targetHosts []string, algorithm string) (*HTTPProxy, error) {
	hosts := make([]string, 0)
	hostMap := make(map[string]*httputil.ReverseProxy)
	alive := make(map[string]bool)

	for _, targetHost := range targetHosts {
		url, err := url.Parse(targetHost)
		if err != nil {
			log.Printf("url.Parse Error: %s\n", err)
			return nil, err
		}
		// create a reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(url)
		// rewrite director function, add header
		originDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originDirector(req)
			req.Header.Set(XProxy, ReverseProxy)
			req.Header.Set(XRealIP, GetClientIP(req))
		}

		host := GetHost(url)
		alive[host] = true    // mark current host is alive
		hostMap[host] = proxy // add mapping
		hosts = append(hosts, host)
	}
	// build a load balancer instance
	lb, err := balancer.Build(algorithm, hosts)
	if err != nil {
		return nil, err
	}

	return &HTTPProxy{
		hostMap: hostMap,
		lb:      lb,
		alive:   alive,
	}, nil
}

// ServeHTTP implements a proxy to the http server
func (h *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("proxy causes panic :%s", err)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(err.(error).Error()))
		}
	}()

	host, err := h.lb.Balance(GetClientIP(r))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(fmt.Sprintf("balance error: %s", err.Error())))
		return
	}

	h.lb.Inc(host)
	defer h.lb.Done(host)
	h.hostMap[host].ServeHTTP(w, r)
}

// ReadAlive reads the alive status of the site
func (h *HTTPProxy) ReadAlive(url string) bool {
	h.RLock()
	defer h.RUnlock()
	return h.alive[url]
}

// SetAlive sets the alive status to the site
func (h *HTTPProxy) SetAlive(url string, alive bool) {
	h.Lock()
	defer h.Unlock()
	h.alive[url] = alive
}

// HealthCheck enable a health check goroutine for each agent
func (h *HTTPProxy) HealthCheck(interval uint) {
	for host := range h.hostMap {
		go h.healthCheck(host, interval)
	}
}

func (h *HTTPProxy) healthCheck(host string, interval uint) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for range ticker.C {
		if !IsBackendAlive(host) && h.ReadAlive(host) {
			log.Printf("Site unreachable, remove %s from load balancer.", host)

			h.SetAlive(host, false)
			h.lb.Remove(host)
		} else if IsBackendAlive(host) && !h.ReadAlive(host) {
			log.Printf("Site reachable, add %s to load balancer.", host)

			h.SetAlive(host, true)
			h.lb.Add(host)
		}
	}
}
