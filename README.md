# iBenchmark
A benckmark that can generate http(s)'s query by short and long connection
iBenchmark使用Go语言研发，为测试HTTPS Server的QPS、CPS性能指标而设计。最初版本只能测试HTTPS短连接，即CPS指标。囊括了ab、wrk的特性，支持HTTP以及HTTPS的长连接、短连接，可测试HTTPS、HTTP的QPS、CPS性能指标。

#Usage

> Usage: iBenchmark [options]  

> -B="": request Body,empty default

> -H="": request Headers,empty default

> -c=1: concurrency:the worker's number,1 default

> -h=false: show help

> -k=false: keep the connections every worker established alive,false default

> -m="GET": HTTP Method,GET default

> -o=false: print response body

> -r=0: total requests per connection,0 default

> -s="TLS_RSA_WITH_RC4_128_SHA": cipher suite,TLS_RSA_WITH_RC4_128_SHA default

> -t=0: timelimit (msec),0 default

> -u="https://0.0.0.0:28080/": server url

> -w=false: send request after handshake connection,false default

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

- -B 自定义request的Body。默认：空
- -H 自定义Request Header头。使用'header1:v1;header2:v2'的格式，以';'分隔，单引号括起来。默认：空
- -c 并发数。每个并发会建立一个transport连接。
- -h 帮助。
- -k keep-Alive。如果使用该项，建立的transport建立都为长连接。如果不使用，则全为短连接。
- -m HTTP Method。自定义request的Method方法，默认为GET。
- -o 打印response的payload，默认不打印。
- -r 每个连接上的请求数。注意：此参数与-t冲突，若指定了-t，则此选项无效。默认：0
- -s 指定TLS握手时的chipersuite(加密套件)。默认：TLS_RSA_WITH_RC4_128_SHA
- -t 每个连接持续的时间，持续时间内将持续发送请求，覆盖-r选项。默认：0
- -u 服务的url。组成 protocol://hostname:port/path。protocol有http、https。默认：https://0.0.0.0:28080/，此时protocol：https，hostname:0.0.0.0，port：28080，path：/ 
注：protocol要与port匹配。若要向HTTP发送压力，则需要将protocol改为http，port改为对应的端口，例如80。
- -w withRequet。当transport建完成时，是否发送query。和-t -r一起使用，发送的query数量由这两个参数决定。

#Example
e.g. HTTPS QPS

> $go run iBenchmark -c 2 -r 10 -u https://www.baidu.com:443/index.html -k -w -H "Host:baike.baidu.com"  

> Server Software:bfe/1.0.8.2  

> Server Port:443 

> Request Headers: 

>  Host:baike.baidu.com 

> 
> Document Path:/index.html 

> Document Length:443 

> Concurrency:2 

> Time Duration:36ms 

> Avg Time Taken:3ms 

> Complete Requests:20 

> Failed Request:0 

> Request Per Second:3 

> Connections Per Second:0 

> Non2XXCode:0 


#####由于使用了-k参数，故2个并发建立了两个长连接，-w 每个长连接发送10个Request请求。此为一个QPS测试的案例。
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

> go run iBenchmark -c 2 -t 5000 -u https://www.baidu.com:443/index.html -w -H "Host:baike.baidu.com"  

> Server Software: 

> Server Hostname:www.baidu.com 

> Server Port:443   

> Request Headers: 
>   Host:baike.baidu.com  

> Document Path:/index.html  

> Document Length:0

> Concurrency:2  

> Time Duration:3826ms

> Avg Time Taken:356ms

> Complete Requests:1917 

> Failed Request:0  

> Request Per Second:0  

> Connections Per Second:24  

> Non2XXCode:0

此案例没有使用-k，即都为短连接。-t 5000运行了5000ms，在此期间一直建立连接。-w ，在每个连接上发送一个请求。-c 2 两个并发。
为CPS的性能测试。

#License
   Copyright 2015 Albus <albus@shaheng.me>.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
