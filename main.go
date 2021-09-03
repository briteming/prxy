package main

import (
	"flag"
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
	var thread = flag.Int("thread", 2, "Worker thread number")
	flag.IntVar(thread, "t", 2, "alias to  --thread")
	var url = flag.String("url", "https://google.com", "the website you want to check")
	var proxy_type = flag.String("proxy", "http", "the proxy type [sock4,sock5,http]")
	var input = flag.String("input", "", "input file with proxy line by line")
	var from_url = flag.String("from-url", "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt", "input from url")
	var ignore_time = flag.Bool("ignore-timeout", true, "ingore timeout proxy server")

	flag.Parse()

	if *thread < 2 {
		fmt.Fprint(os.Stderr, "thread number too small")
		os.Exit(1)
	}
	if !(*proxy_type == "http" || *proxy_type == "socks4" || *proxy_type == "socks5") {
		fmt.Fprintf(os.Stderr, "unknown proxy type %s", *proxy_type)
		os.Exit(1)
	}
	if *input != "" && *from_url != "" {
		fmt.Fprint(os.Stderr, "flag '--input' conflict with '--from-url'")
	}

	var proxys []string
	var err error
	if *from_url != "" {
		reader, err := from_remote(*from_url)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		proxys, err = read_proxys(reader)
	} else {
		reader, err := from_file(*input)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		proxys, err = read_proxys(reader)
	}
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	os.Unsetenv("https_proxy")
	os.Unsetenv("http_proxy")

	var limit = make(chan int, *thread)
	var wg sync.WaitGroup
	var lock sync.RWMutex
	for _, proxy := range proxys {
		wg.Add(1)
		limit <- 1
		go func(server string) {
			defer wg.Done()
			defer func() {
				<-limit
			}()
			use, err := with_proxy(*proxy_type, server, *url, time.Second*time.Duration(2), check)
			if err != nil {
				if !*ignore_time {
					lock.Lock()
					fmt.Fprint(os.Stderr, with_color(fmt.Sprintf("%-25s TIMEOUT\n", server), RED))
					lock.Unlock()
				}
				return
			}
			var t = float64(use) / float64(int64(time.Millisecond))
			var msg = fmt.Sprintf("%-25s %4.2fms\n", server, t)

			if t > 300.0 {
				msg = with_color(msg, YELLOW)
			} else {
				msg = with_color(msg, GREEN)
			}
			lock.Lock()
			fmt.Print(msg)
			lock.Unlock()
		}(proxy)
	}
	wg.Wait()
	close(limit)
}

func from_remote(link string) (io.ReadCloser, error) {
	http.DefaultClient.Timeout = time.Duration(2) * time.Second
	res, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func from_file(file string) (io.ReadCloser, error) {
	return os.Open(file)
}

func read_proxys(input io.ReadCloser) (proxys []string, err error) {
	defer input.Close()
	buffer, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, err
	}
	proxys = strings.Split(string(buffer), "\n")
	return proxys, err
}

type ProxyType = int
type Color = int

const (
	RED Color = iota
	GREEN
	YELLOW
)

func with_proxy(proxy_type string, proxy string, url string, timeout time.Duration, check func(string, time.Duration) (time.Duration, error)) (time.Duration, error) {
	if proxy_type == "http" {
		os.Setenv("ALL_PROXY", fmt.Sprintf("http://%s", proxy))
	}
	if proxy_type == "socks4" {
		os.Setenv("ALL_PROXY", fmt.Sprintf("socks://%s", proxy))
	}
	if proxy_type == "socks5" {
		os.Setenv("ALL_PROXY", fmt.Sprintf("socks5://%s", proxy))
	}
	return check(url, timeout)
}

func check(link string, timeout time.Duration) (time.Duration, error) {
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
	_, err = io.Copy(io.Discard, req.Body)
	return time.Now().Sub(start), err
}

func with_color(str string, c Color) string {
	if c == RED {
		return fmt.Sprintf("\033[0;31;1m%s\033[0m", str)
	}
	if c == GREEN {
		return fmt.Sprintf("\033[0;32;1m%s\033[0m", str)
	}
	if c == YELLOW {
		return fmt.Sprintf("\033[0;33;1m%s\033[0m", str)
	}
	return str
}
