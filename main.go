package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	var proxy = "138.68.45.192:80"
	var google = "https://google.com/ncr"
	proxys, err := get_proxy_from_url("https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	os.Unsetenv("https_proxy")
	os.Unsetenv("http_proxy")

	var wg sync.WaitGroup
	for _, proxy = range proxys {
		wg.Add(1)
		func(server string) {
			defer wg.Done()
			use, err := check(google, proxy, time.Second*time.Duration(2))
			if err != nil {
				fmt.Printf("%11s TIMEOUT", server)
			}
			fmt.Printf("%-25s %4.2fms\n", server, float64(use)/float64(int64(time.Millisecond)))
		}(proxy)
	}
	wg.Wait()
}

func get_proxy_from_url(link string) (proxys []string, err error) {
	res, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	buffer, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	proxys = strings.Split(string(buffer), "\n")
	return proxys, err
}

type ProxyType = int

func check(link string, proxy string, timeout time.Duration) (time.Duration, error) {
	os.Setenv("HTTP_PROXY", fmt.Sprintf("http://%s", proxy))
	var start = time.Now()
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := client.Get(link)
	if err != nil {
		return 0, err
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36")
	defer req.Body.Close()
	io.Copy(io.Discard, req.Body)
	return time.Now().Sub(start), nil
}
