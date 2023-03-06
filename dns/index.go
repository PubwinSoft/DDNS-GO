package dns

import (
	"log"
	"time"

	"github.com/jeessy2/ddns-go/v4/config"
	"github.com/jeessy2/ddns-go/v4/util"
)

// DNS interface
type DNS interface {
	Init(conf *config.Config, cache [2]*util.IpCache)
	// 添加或更新IPv4/IPv6记录
	AddUpdateDomainRecords() (domains config.Domains)
}

var Ipcache = [][2]util.IpCache{}

// RunTimer 定时运行
func RunTimer(firstDelay time.Duration, delay time.Duration) {
	log.Printf("第一次运行将等待 %d 秒后运行 (等待网络)", int(firstDelay.Seconds()))
	time.Sleep(firstDelay)
	for {
		RunOnce()
		time.Sleep(delay)
	}
}

// RunOnce RunOnce
func RunOnce() {
	confa, err := config.GetConfigCache()
	if err != nil {
		return
	}
	if util.ForceCompare || len(Ipcache) != len(confa.Dnsconfig) {
		Ipcache = [][2]util.IpCache{}
		for range confa.Dnsconfig {
			Ipcache = append(Ipcache, [2]util.IpCache{{}, {}})
		}
	}

	for i, conf := range confa.Dnsconfig {
		var dnsSelected DNS
		switch conf.DNS.Name {
		case "alidns":
			dnsSelected = &Alidns{}
		case "dnspod":
			dnsSelected = &Dnspod{}
		case "cloudflare":
			dnsSelected = &Cloudflare{}
		case "huaweicloud":
			dnsSelected = &Huaweicloud{}
		case "callback":
			dnsSelected = &Callback{}
		case "baiducloud":
			dnsSelected = &BaiduCloud{}
		case "porkbun":
			dnsSelected = &Porkbun{}
		case "godaddy":
			dnsSelected = &GoDaddyDNS{}
		case "googledomain":
			dnsSelected = &GoogleDomain{}
		default:
			dnsSelected = &Alidns{}
		}
		dnsSelected.Init(&conf, [2]*util.IpCache{&Ipcache[i][0], &Ipcache[i][1]})
		domains := dnsSelected.AddUpdateDomainRecords()
		config.ExecWebhook(&domains, &confa)
	}
	util.ForceCompare = false
}
