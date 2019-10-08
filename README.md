## Bruteservice

### Installation
```
go get -u github.com/vodafon/bruteservice
```

### Usage

```
wget https://raw.githubusercontent.com/vodafon/bruteservice/master/example/services.json
wget https://raw.githubusercontent.com/vodafon/bruteservice/master/example/wordlist.txt

bruteservice -company yahoo -services ./services.json -wordlist ./wordlist.txt

GET https://yahoo.atlassian.net
GET https://gitlab.com/yahoo
GET https://github.com/yahoo
GET https://hub.docker.com/v2/orgs/yahoo/
GET https://yahoo0.atlassian.net
GET https://github.com/yahoo0
GET https://github.com/yahoo-1
GET https://github.com/yahoo1
GET https://github.com/yahoo3
GET https://github.com/yahoo7
GET https://github.com/yahoo-com
GET https://github.com/yahoocom
GET https://github.com/githubyahoo
GET https://github.com/yahooprojects
```


### Arguments

```
bruteservice -h

Usage of bruteservice:
  -company string
        company
  -procs int
        concurrency (default 6)
  -services string
        services config
  -v int
        verbose level (default 1)
  -wordlist string
        path to wordlist
```
