# iBenchmark
iBenchmark is a benchmark send queries to a web application which include the short connections and long connections.So you can use it to test the web application's both QPS (queries per second) and CPS (connections per second).

It also support the SPDY.

#Install
Simple as it takes to type the following command:

> ./Install.sh

#Usage

> Usage: iBench [options]  

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

> -M=8: max cores used,8 default

> -S=false: turn to SPDY

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

#Example
e.g. HTTPS QPS

> $./iBench -c 2 -r 10 -u https://www.baidu.com/ -k -H "Host:baike.baidu.com"  

> Server Software:bfe/1.0.8.2  

> Server Port:443 

> Request Headers: 

>  Host:baike.baidu.com 

> 
> Document Path:/

> Document Length:443 

> Concurrency:2 

> Time Duration:36ms 

> Avg Time Taken:3ms 

> Complete Requests:20 

> Failed Request:0 

> Request Per Second:3 

> Connections Per Second:0 

> Non2XXCode:0 


e.g. HTTPS CPS

> go run iBenchmark -c 2 -t 5000 -u https://www.baidu.com/ -H "Host:baike.baidu.com"  

> Server Software: 

> Server Hostname:www.baidu.com 

> Server Port:443   

> Request Headers: 
>   Host:baike.baidu.com  

> Document Path:/  

> Document Length:0

> Concurrency:2  

> Time Duration:3826ms

> Avg Time Taken:356ms

> Complete Requests:1917 

> Failed Request:0  

> Request Per Second:0  

> Connections Per Second:24  

> Non2XXCode:0



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
