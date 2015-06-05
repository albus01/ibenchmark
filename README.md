# iBenchmark
A benckmark that can generate http(s)'s query by short and long connection

###一、工具介绍
iBenchmark使用Go语言研发，为测试HTTPS Server的QPS、CPS性能指标而设计。最初版本只能测试HTTPS短连接，即CPS指标。囊括了ab、wrk的特性，支持HTTP以及HTTPS的长连接、短连接，可测试HTTPS、HTTP的QPS、CPS性能指标。
###二、工具使用
使用帮助：

> Usage: iBenchmark [options]  

> -H="Host: baike.baidu.com": request header  

> -k=false: send request after handshake on the keep-alive connection  

> -c=1: concurrency

> -h=false: show help

> -r=0: total requests per connection

> -s="TLS_RSA_WITH_RC4_128_SHA": cipher suite  

> -t=0: timelimit (msec) 

> -u="https://0.0.0.0:28080/": server url  

> cihper suite: <br />
> TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA <br />
> TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA TLS_RSA_WITH_AES_256_CBC_SHA <br />
> TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA <br />
> TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 <br />
> TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 TLS_FALLBACK_SCSV <br />
> TLS_RSA_WITH_3DES_EDE_CBC_SHA TLS_ECDHE_ECDSA_WITH_RC4_128_SHA <br />
> TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA <br />
> TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA TLS_RSA_WITH_RC4_128_SHA <br />
> TLS_RSA_WITH_AES_128_CBC_SHA TLS_ECDHE_RSA_WITH_RC4_128_SHA <br />

参数说明：

- -k TCP(TLS)握手完成后保持长连接并在此长连接上发送r(r为参数-r)个Query请求，每个concurrency保持一个长连接。默认：关
- -c 并发数。若使用-k，则-c的数量即为长连接的数量。默认：1
- -t 每个连接持续的时间，持续时间内将持续发送请求，覆盖-r选项。默认：0
- -r 每个连接上的请求数。注意：此参数与-t冲突，若指定了-t，则此选项无效。默认：0
- -s 指定TLS握手时的chipersuite(加密套件)。默认：TLS_RSA_WITH_RC4_128_SHA
- -u 服务的url。组成 protocol://hostname:port/path。protocol有http、https。默认：https://0.0.0.0:28080/，此时protocol：https，hostname:0.0.0.0，port：28080，path：/
注：protocol要与port匹配。若要向HTTP发送压力，则需要将protocol改为http，port改为对应的端口，例如80。
- -h 帮助。
- -H 指定request Header头。使用方式 -H '["Host:baike.baidu.com","Connection:Keep-alive"]' 注：只有在连接上发送query请求时此参数才有效(即添加-k参数) 。注意格式：中括号外用单引号括起来，中括号内每个元素使用双引号"括起来，如果元素大于1个，元素间使用逗号隔开。不按此格式书写的-H将解析失败。

###三、使用案例
e.g. HTTPS QPS

> $./iBenchmark -c 2 -r 10 -u https://127.0.0.1:8800/shaheng.html -k -H '["Host:baike.baidu.com"]'  

> Server Software:nginx/1.4.1  

> Server Port:8800 Request 

> Headers: 

>  Host:baike.baidu.com 

> 
> Document Path:/shaheng.html 

> Document Length:131 

> Concurrency:2 

> Time Duration:36ms 

> Avg Time Taken:3ms 

> Complete Requests:20 

> Failed Request:0 

> Request Per Second:555 

> Connections Per Second:0 

> Non2XXCode:0 


#####由于使用了-k参数，故2个并发建立了两个长连接，每个长连接发送10个Request请求。此为一个QPS测试的案例。
####输出说明：
Server Software为请求的后端服务器</br>
Server Hostname为请求的Server HostName</br>
Server Port 请求的Server端口</br>
Document Path 请求的路径</br>
Document Length 服务端返回的文档的长度</br>
Concurrency 并发数</br>
Time Duration 从启动到结束经历的时间</br>
Avg Time Taken 每个请求的平均响应时间</br>
Complete Request 完成的请求数</br>
Failed Request 失败的请求数</br>
Request Per Second 每秒请求量 QPS (测试CPS时，此数据为0)</br>
Connections Per Second 每秒连接量 CPS(打开-k 测试QPS时为0)</br>
Non2XXCode 不是200~299之间的HTTP 状态码</br>

e.g. HTTPS CPS

> ./iBenchmark -c 2 -t 5000 -u https://127.0.0.1:8800/shaheng.html -H '["Host:baike.baidu.com"]'  

> Server Software: 

> Server Hostname:127.0.0.1 

> Server Port:8800   

> Request Headers: ["Host:baike.baidu.com"]  

> Document Path:/shaheng.html  

> Concurrency:2  

> Complete Requests:1917 

> Failed Request:0  

> Request Per Second:0  

> Connections Per Second:383  

> Non2XXCode:0

此案例没有使用-k，没有发送query，Header头部也没有解析(因为是没有意义的),都为短连接。-t 5000运行了5000ms。-c 2 两个并发，每个并发持续建立连接、关闭连接，不发送query。为CPS的性能测试。
