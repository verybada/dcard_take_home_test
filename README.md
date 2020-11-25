# dcard_take_home_test

結構上一共分為三部分 handler, middlware, ratelimiter

* handler
    * 只負責讀取X-RATE-LIMIT-LIMIT及X-RATE-LIMIT-REMAINING header來顯示目前已使用的rate
* middleware
    * 使用ratelimiter來決定是否放行request，若已達到上限，則會在此層返回。
* ratelimiter
    * 採用fixed window算法並搭配redis實作，利用redis atomic的特性來避免concurrency access造成的race condition。
    * 每個window的key為`<ip>-<start_time>`，並設定TTL來自動回收過期的window。


## 環境
* golang 1.15
* golangci-lint 1.33 (https://golangci-lint.run/usage/install/#local-installation)
* docker

### 執行步驟
1. 執行lint, unittest及compile
    `make build` 
2. 啟動測試所需的redis server
    `make start_redis`
3. 啟動api server
    `build/bin/api_server`
    * 預設為8080，若衝突可透過 -host `":XXXX"`調整
