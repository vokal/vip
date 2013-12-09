package peer

import (
	"flag"
	"fmt"
	"github.com/golang/groupcache"
	"launchpad.net/goamz/ec2"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

var (
	cacheport *string = flag.String("cacheport", "9001", "target port")
	rpcport   *string = flag.String("rpcport", "7001", "target port")
)

type Args interface{}

type CachePool interface {
	Listen() error
	Port() string
	SetContext(func(r *http.Request) groupcache.Context)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type DebugCachePool struct {
	*groupcache.HTTPPool
}

type EC2CachePool struct {
	*groupcache.HTTPPool
	*ec2.EC2
}

func getSystemAddrs() []string {
	addrs, _ := net.InterfaceAddrs()
	result := make([]string, len(addrs))

	for i, v := range addrs {
		result[i] = strings.Split(v.String(), "/")[0]
	}

	return result
}

func localAddress(addr string) bool {
	for _, a := range getSystemAddrs() {
		if a == addr {
			return true
		}
	}

	return false
}

func getLocalIP(ec2conn *ec2.EC2) string {
	f := ec2.NewFilter()
	f.Add("tag:server-type", "image-proxy")

	resp, err := ec2conn.Instances(nil, f)
	if err != nil {
		return ""
	}

	for _, reserv := range resp.Reservations {
		for _, instance := range reserv.Instances {
			if localAddress(instance.PrivateIPAddress) {
				return instance.PrivateIPAddress
			}
		}
	}

	return ""
}

func DebugPool() *DebugCachePool {
	peers := &DebugCachePool{
		groupcache.NewHTTPPool("http://localhost:9001"),
	}

	peers.Set("http://localhost:9001")

	return peers
}

func (p *DebugCachePool) Listen() error {
	return nil
}

func (p *DebugCachePool) Port() string {
	return "9001"
}

func (p *DebugCachePool) SetContext(f func(r *http.Request) groupcache.Context) {
	p.Context = f
}

func Pool(ec2conn *ec2.EC2) *EC2CachePool {
	localip := getLocalIP(ec2conn)

	peers := &EC2CachePool{
		groupcache.NewHTTPPool(fmt.Sprintf("http://%s:%s", localip, *cacheport)),
		ec2conn,
	}

	peerAddrs, err := peers.discoverPeers()
	if err != nil {
		log.Fatal("discovery:", err)
	}

	log.Println("Alerting peers:", peerAddrs)

	// Inform each peer that they should also discover peers
	for _, peer := range peerAddrs {
		client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", peer, *rpcport))
		if err != nil {
			log.Println("error dialing:", err)
			continue
		}

		var reply int
		err = client.Call("EC2CachePool.RefreshPeers", new(Args), &reply)
		if err != nil {
			log.Fatal("refresh rpc failure:", err)
		}

		log.Println("Refreshed peers on:", peer)
	}

	return peers
}

func (p *EC2CachePool) discoverPeers() ([]string, error) {
	f := ec2.NewFilter()
	f.Add("tag:server-type", "image-proxy")

	resp, err := p.Instances(nil, f)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, reserv := range resp.Reservations {
		for _, instance := range reserv.Instances {
			if instance.State.Name == "running" {
				ips = append(ips, instance.PrivateIPAddress)
			}
		}
	}

	peerIPs := make([]string, len(ips))
	_ = copy(peerIPs, ips)

	for i := range peerIPs {
		peerIPs[i] = fmt.Sprintf("http://%s:%s", peerIPs[i], *cacheport)
	}

	log.Println("Setting peers:", peerIPs)
	p.Set(peerIPs...)

	return ips, nil
}

func (p *EC2CachePool) RefreshPeers(a *Args, r *int) error {
	log.Println("Asked to discover peers")
	_, err := p.discoverPeers()
	return err
}

func (p *EC2CachePool) Listen() error {
	rpc.Register(p)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", *rpcport))
	if err != nil {
		return err
	}

	return http.Serve(l, nil)
}

func (p *EC2CachePool) Port() string {
	return *cacheport
}

func (p *EC2CachePool) SetContext(f func(r *http.Request) groupcache.Context) {
	p.Context = f
}
