package dns

import (
	"bytes"
	"ddns-go/config"
	"ddns-go/util"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// https://cloud.baidu.com/doc/BCD/s/4jwvymhs7

const (
	baiduEndpoint = "https://bcd.baidubce.com"
)

type BaiduCloud struct {
	DNSConfig config.DNSConfig
	Domains   config.Domains
	TTL       int
}

// BaiduRecord 单条解析记录
type BaiduRecord struct {
	RecordId uint   `json:"recordId"`
	Domain   string `json:"domain"`
	View     string `json:"view"`
	Rdtype   string `json:"rdtype"`
	TTL      int    `json:"ttl"`
	Rdata    string `json:"rdata"`
	ZoneName string `json:"zoneName"`
	Status   string `json:"status"`
}

// BaiduRecordsResp 获取解析列表拿到的结果
type BaiduRecordsResp struct {
	TotalCount int           `json:"totalCount"`
	Result     []BaiduRecord `json:"result"`
}

// BaiduListRequest 获取解析列表请求的body json
type BaiduListRequest struct {
	Domain   string `json:"domain"`
	PageNum  int    `json:"pageNum"`
	PageSize int    `json:"pageSize"`
}

// BaiduModifyRequest 修改解析请求的body json
type BaiduModifyRequest struct {
	RecordId uint   `json:"recordId"`
	Domain   string `json:"domain"`
	View     string `json:"view"`
	RdType   string `json:"rdType"`
	TTL      int    `json:"ttl"`
	Rdata    string `json:"rdata"`
	ZoneName string `json:"zoneName"`
}

// BaiduCreateRequest 创建新解析请求的body json
type BaiduCreateRequest struct {
	Domain   string `json:"domain"`
	RdType   string `json:"rdType"`
	TTL      int    `json:"ttl"`
	Rdata    string `json:"rdata"`
	ZoneName string `json:"zoneName"`
}

func (baidu *BaiduCloud) Init(conf *config.Config) {
	baidu.DNSConfig = conf.DNS
	baidu.Domains.GetNewIp(conf)
	if conf.TTL == "" {
		// 默认300s
		baidu.TTL = 300
	} else {
		ttl, err := strconv.Atoi(conf.TTL)
		if err != nil {
			baidu.TTL = 300
		} else {
			baidu.TTL = ttl
		}
	}
}

// AddUpdateDomainRecords 添加或更新IPv4/IPv6记录
func (baidu *BaiduCloud) AddUpdateDomainRecords() config.Domains {
	baidu.addUpdateDomainRecords("A")
	baidu.addUpdateDomainRecords("AAAA")
	return baidu.Domains
}

func (baidu *BaiduCloud) addUpdateDomainRecords(recordType string) {
	ipAddr, domains := baidu.Domains.GetNewIpResult(recordType)
	if ipAddr == "" {
		return
	}

	for _, domain := range domains {
		var records BaiduRecordsResp

		requestBody := BaiduListRequest{
			Domain:   domain.DomainName,
			PageNum:  1,
			PageSize: 1000,
		}

		err := baidu.request("POST", baiduEndpoint+"/v1/domain/resolve/list", requestBody, &records)
		if err != nil {
			return
		}

		find := false
		for _, record := range records.Result {
			if record.Domain == domain.SubDomain {
				//存在就去更新
				baidu.modify(record, domain, ipAddr)
				find = true
				break
			}
		}
		if !find {
			//没找到，去创建
			baidu.create(domain, recordType, ipAddr)
		}
	}
}

//create 创建新的解析
func (baidu *BaiduCloud) create(domain *config.Domain, recordType string, ipAddr string) {
	var baiduCreateRequest = BaiduCreateRequest{
		Domain:   domain.GetSubDomain(), //处理一下@
		RdType:   recordType,
		TTL:      baidu.TTL,
		Rdata:    ipAddr,
		ZoneName: domain.DomainName,
	}
	var result BaiduRecordsResp

	err := baidu.request("POST", baiduEndpoint+"/v1/domain/resolve/add", baiduCreateRequest, &result)
	if err == nil {
		log.Printf("新增域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
	} else {
		log.Printf("新增域名解析 %s 失败！", domain)
		domain.UpdateStatus = config.UpdatedFailed
	}
}

//modify 更新解析
func (baidu *BaiduCloud) modify(record BaiduRecord, domain *config.Domain, ipAddr string) {
	//没有变化直接跳过
	if record.Rdata == ipAddr {
		log.Printf("你的IP %s 没有变化, 域名 %s", ipAddr, domain)
		return
	}
	var baiduModifyRequest = BaiduModifyRequest{
		RecordId: record.RecordId,
		Domain:   record.Domain,
		View:     record.View,
		RdType:   record.Rdtype,
		TTL:      record.TTL,
		Rdata:    ipAddr,
		ZoneName: record.ZoneName,
	}
	var result BaiduRecordsResp

	err := baidu.request("POST", baiduEndpoint+"/v1/domain/resolve/edit", baiduModifyRequest, &result)
	if err == nil {
		log.Printf("更新域名解析 %s 成功！IP: %s", domain, ipAddr)
		domain.UpdateStatus = config.UpdatedSuccess
	} else {
		log.Printf("更新域名解析 %s 失败！", domain)
		domain.UpdateStatus = config.UpdatedFailed
	}
}

// request 统一请求接口
func (baidu *BaiduCloud) request(method string, url string, data interface{}, result interface{}) (err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, err = json.Marshal(data)
		if err != nil {
			fmt.Println("sdfsdfsdf", err)
		}
	}

	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(jsonStr),
	)

	if err != nil {
		log.Println("http.NewRequest失败. Error: ", err)
		return
	}

	util.BaiduSigner(baidu.DNSConfig.ID, baidu.DNSConfig.Secret, req)

	client := util.CreateHTTPClient()
	resp, err := client.Do(req)
	err = util.GetHTTPResponse(resp, url, err, result)

	return
}
