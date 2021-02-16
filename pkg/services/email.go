package services

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang/groupcache/lru"

	"github.com/forest-friends/email-checker-cli/pkg/utils"
)

const (
	CONNECTION_TIMEOUT = 200 * time.Millisecond
)

var (
	DefaultDialer = net.Dialer{Timeout: CONNECTION_TIMEOUT}
)

func CheckSlow(ch chan string, mxCache *lru.Cache, hostCache *lru.Cache, successFile *os.File, badFile *os.File) {
	for {
		select {
		case email := <-ch:
			_, host := utils.SplitEmail(email)
			if _, ok := hostCache.Get(host); ok {
				successFile.WriteString(fmt.Sprintln(email))
				continue
			}

			mxs, err := utils.СheckMX(host, mxCache)
			if err != nil {
				badFile.WriteString(fmt.Sprintln(email))
				continue
			}

			for _, mx := range mxs {
				smtpConn, err := utils.MakeSMTPConnection(host, mx, 25)
				if err != nil {
					continue
				}

				if err = smtpConn.Hello(host); err != nil {
					badFile.WriteString(fmt.Sprintln(email))
					break
				}

				hostCache.Add(host, true)
				successFile.WriteString(fmt.Sprintln(email))
				break
			}

			badFile.WriteString(fmt.Sprintln(email))
		}
	}
}

func CheckStrict(ch chan string, mxCache *lru.Cache, emailCache *lru.Cache, successFile *os.File, badFile *os.File) {
	for {
		select {
		case email := <-ch:
			_, host := utils.SplitEmail(email)

			mxs, err := utils.СheckMX(host, mxCache)
			if err != nil {
				badFile.WriteString(fmt.Sprintln(email))
				continue
			}

			for _, mx := range mxs {
				smtpConn, err := utils.MakeSMTPConnection(host, mx, 25)
				if err != nil {
					continue
				}

				if err = smtpConn.Hello(host); err != nil {
					badFile.WriteString(fmt.Sprintln(email))
					break
				}

				if err = smtpConn.Mail(email); err != nil {
					badFile.WriteString(fmt.Sprintln(email))
					break
				}
				if err = smtpConn.Rcpt(email); err != nil {
					badFile.WriteString(fmt.Sprintln(email))
					break
				}

				emailCache.Add(email, true)
				successFile.WriteString(fmt.Sprintln(email))
				break
			}

			badFile.WriteString(fmt.Sprintln(email))
		}
	}
}
