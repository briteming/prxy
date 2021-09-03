# prxy

```
go get -u github.com/hellojukay/prxyc
```

# usage
```
hellojukay@local prxy (main) $ prxyc -h
Usage of prxyc:
  -from-url string
        input from url (default "https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt")
  -ignore-timeout
        ingore timeout proxy server (default true)
  -input string
        input file with proxy line by line
  -proxy string
        the proxy type [sock4,sock5,http] (default "http")
  -t int
        alias to  --thread (default 2)
  -thread int
        Worker thread number (default 2)
  -url string
        the website you want to check (default "https://google.com")
```