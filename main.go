package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vodafon/verlog"
)

var (
	flagServices = flag.String("services", "", "services config")
	flagCompany  = flag.String("company", "", "company")
	flagWordlist = flag.String("wordlist", "", "path to wordlist")
	flagProcs    = flag.Int("procs", 6, "concurrency")
	flagV        = flag.Int("v", 1, "verbose level")

	log *verlog.Logger
)

type Request struct {
	ServiceName string
	URL         string
	Method      string
	Analysis    *Analysis
}

type Service struct {
	Name     string    `json:"name"`
	Method   string    `json:"method"`
	URL      string    `json:"url"`
	Analysis *Analysis `json:"analysis"`
}

func main() {
	flag.Parse()
	if *flagServices == "" || *flagCompany == "" || *flagWordlist == "" || *flagV < 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log = verlog.New(*flagV)

	services := []Service{}
	mustLoadJSON(*flagServices, &services)
	for i := 0; i < len(services); i++ {
		services[i].Analysis.Compile()
	}
	wordFile, err := os.Open(*flagWordlist)
	if err != nil {
		log.Fatal(err)
	}

	if len(services) == 0 {
		log.Fatalf("services not found in file")
	}

	requestC := make(chan Request)

	wg := sync.WaitGroup{}
	for i := 0; i < *flagProcs; i++ {
		wg.Add(1)

		go func() {
			for request := range requestC {
				processRequest(request)
			}
			wg.Done()
		}()
	}

	permutations(requestC, services, "")

	sc := bufio.NewScanner(wordFile)
	for sc.Scan() {
		permutations(requestC, services, sc.Text())
	}

	close(requestC)
	wg.Wait()
}

func httpClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: time.Second,
			DualStack: true,
		}).DialContext,
	}
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
func processRequest(request Request) {
	client := httpClient()
	req, err := http.NewRequest(request.Method, request.URL, nil)
	if err != nil {
		log.Printf(1, "NewRequest error for %s %s: %v", request.Method, request.URL, err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf(1, "Do request error for %s %s: %v", request.Method, request.URL, err)
		return
	}
	defer resp.Body.Close()
	match, err := request.Analysis.Analyze(resp)
	if err != nil {
		log.Printf(1, "Analize error for %s %s: %v", request.Method, request.URL, err)
		return
	}
	if match {
		fmt.Printf("%s %s\n", request.Method, request.URL)
	}
}

func permutations(requestC chan Request, services []Service, word string) {
	for _, service := range services {
		servicePermutations(requestC, service, word)
	}
}

func servicePermutations(requestC chan Request, service Service, word string) {
	word = strings.TrimSpace(word)
	if word == "" {
		requestC <- requestService(service, *flagCompany, "", "")
		return
	}
	requestC <- requestService(service, *flagCompany, word, "")
	requestC <- requestService(service, word, *flagCompany, "")
	requestC <- requestService(service, word, "-", *flagCompany)
	requestC <- requestService(service, *flagCompany, "-", word)
}

func requestService(service Service, part1, part2, part3 string) Request {
	cmp := fmt.Sprintf("%s%s%s", part1, part2, part3)
	return Request{
		ServiceName: service.Name,
		Method:      service.Method,
		URL:         strings.ReplaceAll(service.URL, "COMPANY", cmp),
		Analysis:    service.Analysis,
	}
}

func mustLoadJSON(path string, v interface{}) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(file, v)
	if err != nil {
		log.Fatal(err)
	}
}
