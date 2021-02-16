package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/golang/groupcache/lru"

	"github.com/forest-friends/email-checker-cli/pkg/services"
	"github.com/forest-friends/email-checker-cli/pkg/utils"
)

const ()

func main() {
	inputFilePath := flag.String("f", "./input.txt", "Input file path")
	successFilePath := flag.String("s", "./success.txt", "Output file for success email")
	badFilePath := flag.String("b", "./bad.txt", "Output file for fail email")
	hosts := flag.String("h", "gmail.com,yahoo.com,hotmail.com,aol.com,hotmail.co.uk,hotmail.fr,msn.com,yahoo.fr,wanadoo.fr,orange.fr,comcast.net,yahoo.co.uk,yahoo.com.br,yahoo.co.in,live.com,yandex.ru,outlook.com,mail.ru,hotmail.it,rambler.ru,googlemail.com", "Verified email hosts")
	mxHostsCacheSize := flag.Int("ms", 1000, "MX hosts cache size")
	hostsCacheSize := flag.Int("hs", 1000, "Hosts cache size")
	emailCacheSize := flag.Int("es", 1000, "Email cache size")
	slowGoroutines := flag.Int("sp", 10, "Slow goroutines")
	strictGoroutines := flag.Int("sg", 10, "Strict goroutines")
	flag.Parse()

	inputFile, err := os.Open(*inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	successFile, err := os.Create(*successFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer successFile.Close()

	badFile, err := os.Create(*badFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer badFile.Close()

	mxHostCache := lru.New(*mxHostsCacheSize)
	hostCache := lru.New(*hostsCacheSize)
	emailCache := lru.New(*emailCacheSize)
	validator := validator.New()
	slowChannel := make(chan string, 100)
	strictChannel := make(chan string, 100)
	var badMutex sync.Mutex

	for i := 1; i < *slowGoroutines; i++ {
		go services.CheckSlow(slowChannel, mxHostCache, hostCache, successFile, badFile, &badMutex)
	}
	for i := 1; i < *strictGoroutines; i++ {
		go services.CheckStrict(strictChannel, mxHostCache, emailCache, successFile, badFile, &badMutex)
	}

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		email := scanner.Text()

		err = validator.Var(email, "required,email")
		if err != nil {
			badMutex.Lock()
			badFile.WriteString(fmt.Sprintln(email))
			badMutex.Unlock()
			continue
		}

		_, host := utils.SplitEmail(email)
		if strings.Contains(*hosts, host) {
			strictChannel <- email
		} else {
			slowChannel <- email
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
