package schedule

import (
	"net/url"
	"regexp"
)

var domainAll = map[string]*DomainIdentify{
	"335v.net": {
		ExactHosts: []string{
			"335v.net",
		},
		Regular: []string{"*.335v.net"},
	},
}

// quickly identify the domain

var (
	ExactHosts  = map[string]string{}
	RegexpHosts = map[string]string{}
)

type DomainIdentify struct {
	Domain     string
	ExactHosts []string
	// Fuzzy     []string
	Regular []string
}

func (s *Scheduler) initIdentify() {
	// 将正方向转换为逆方向
	for domain, identify := range domainAll {
		identify.Domain = domain
		for _, host := range identify.ExactHosts {
			ExactHosts[host] = domain
		}
		for _, host := range identify.Regular {
			RegexpHosts[host] = domain
		}
	}
}

// 获取
func (s *Scheduler) obtainDomain(u *url.URL) string {
	uh := u.Hostname()

	// 精确匹配
	if v, ok := ExactHosts[uh]; ok {
		return v
	}

	// 正则匹配
	for p, d := range RegexpHosts {
		match, err := regexp.MatchString(p, uh)
		if err != nil {
			return uh
		}
		if match {
			return d
		}
	}

	// 其他

	return uh
}
