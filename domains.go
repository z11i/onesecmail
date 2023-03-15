package onesecmail

import "sync"

var Domains = map[string]struct{}{
	"1secmail.com": {},
	"1secmail.org": {},
	"1secmail.net": {},
	"bheps.com":    {},
	"dcctb.com":    {},
	"kzccv.com":    {},
	"qiott.com":    {},
	"wuuvo.com":    {},
}

var domainsMu sync.Mutex
