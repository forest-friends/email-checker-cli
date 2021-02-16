package utils

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/golang/groupcache/lru"
)

const (
	EmailSeparator     = '@'
	CONNECTION_TIMEOUT = 1000 * time.Millisecond
)

var (
	DefaultDialer = net.Dialer{Timeout: CONNECTION_TIMEOUT}
)

// pre-validation must be success
func SplitEmail(email string) (account, host string) {
	i := strings.LastIndexByte(email, EmailSeparator)
	return email[:i], email[i+1:]
}

func СheckMX(host string, hostCache *lru.Cache) ([]*net.MX, error) {
	var err error
	var mxList []*net.MX
	mxs, ok := hostCache.Get(host)
	if !ok {
		mxList, err = net.LookupMX(host)
		if err != nil {
			return nil, fmt.Errorf("can't check host MX records")
		}
		if len(mxList) == 0 {
			return nil, fmt.Errorf("host not contain MX records")
		}
	} else {
		mxList, ok = mxs.([]*net.MX)
		if !ok {
			return nil, fmt.Errorf("can't parse mxs")
		}
	}

	return mxList, nil
}

func MakeSMTPConnection(host string, mx *net.MX, port int) (*smtp.Client, error) {
	ctxWithTimeout, _ := context.WithTimeout(context.Background(), CONNECTION_TIMEOUT)
	conn, err := DefaultDialer.DialContext(ctxWithTimeout, "tcp", fmt.Sprintf("%s:%d", mx.Host, port))

	if err != nil {
		return nil, err
	}

	smtpConn, err := smtp.NewClient(conn, mx.Host)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("can't connect to SMTP server")
	}

	return smtpConn, nil
}
