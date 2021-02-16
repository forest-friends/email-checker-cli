package services

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"

	"github.com/forest-friends/email-checker-cli/pkg/utils"
)

const (
	CONNECTION_TIMEOUT = 200 * time.Millisecond
)

var (
	DefaultDialer    = net.Dialer{Timeout: CONNECTION_TIMEOUT}
	successFileMutex sync.Mutex

	emailMutex sync.Mutex
	mxMutex    sync.Mutex
)

func CheckSlow(ch chan string, mxCache *lru.Cache, hostCache *lru.Cache, successFile *os.File, badFile *os.File, badFileMutex *sync.Mutex) {
	for {
		select {
		case email := <-ch:
			_, host := utils.SplitEmail(email)
			if _, ok := hostCache.Get(host); ok {
				successFileMutex.Lock()
				successFile.WriteString(fmt.Sprintln(email))
				successFileMutex.Unlock()
				continue
			}

			mxs, err := utils.СheckMX(host, mxCache, &mxMutex)
			if err != nil {
				badFileMutex.Lock()
				badFile.WriteString(fmt.Sprintln(email))
				badFileMutex.Unlock()
				continue
			}

			isFinded := false
			for _, mx := range mxs {
				smtpConn, err := utils.MakeSMTPConnection(host, mx, 25)
				if err != nil {
					continue
				}

				if err = smtpConn.Hello(host); err != nil {
					badFileMutex.Lock()
					badFile.WriteString(fmt.Sprintln(email))
					badFileMutex.Lock()
					break
				}

				hostCache.Add(host, true)
				successFileMutex.Lock()
				successFile.WriteString(fmt.Sprintln(email))
				successFileMutex.Unlock()
				isFinded = true
				break
			}

			if !isFinded && err == nil {
				badFileMutex.Lock()
				badFile.WriteString(fmt.Sprintln(email))
				badFileMutex.Unlock()
			}
		}
	}
}

func CheckStrict(ch chan string, mxCache *lru.Cache, emailCache *lru.Cache, successFile *os.File, badFile *os.File, badFileMutex *sync.Mutex) {
	for {
		select {
		case email := <-ch:
			emailMutex.Lock()
			_, ok := emailCache.Get(email)
			emailMutex.Unlock()
			if ok {
				successFileMutex.Lock()
				successFile.WriteString(fmt.Sprintln(email))
				successFileMutex.Unlock()
				continue
			}

			_, host := utils.SplitEmail(email)

			mxs, err := utils.СheckMX(host, mxCache, &mxMutex)
			if err != nil {
				badFileMutex.Lock()
				badFile.WriteString(fmt.Sprintln(email))
				badFileMutex.Unlock()
				continue
			}

			isFinded := false
			for _, mx := range mxs {
				smtpConn, err := utils.MakeSMTPConnection(host, mx, 25)
				if err != nil {
					continue
				}

				if err = smtpConn.Hello(host); err != nil {
					badFileMutex.Lock()
					badFile.WriteString(fmt.Sprintln(email))
					badFileMutex.Unlock()
					break
				}

				if err = smtpConn.Mail(email); err != nil {
					badFileMutex.Lock()
					badFile.WriteString(fmt.Sprintln(email))
					badFileMutex.Unlock()
					break
				}
				if err = smtpConn.Rcpt(email); err != nil {
					badFileMutex.Lock()
					badFile.WriteString(fmt.Sprintln(email))
					badFileMutex.Unlock()
					break
				}

				emailMutex.Lock()
				emailCache.Add(email, true)
				emailMutex.Unlock()
				successFileMutex.Lock()
				successFile.WriteString(fmt.Sprintln(email))
				successFileMutex.Unlock()
				isFinded = true
				break
			}

			if !isFinded && err == nil {
				badFileMutex.Lock()
				badFile.WriteString(fmt.Sprintln(email))
				badFileMutex.Unlock()
			}
		}
	}
}
