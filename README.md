# iBenchmark

Bechmark for HTTP,HTTPS,SPDY Server.

iBenchmark is a benchmark send queries to a web application which similar to wrk and Apache Bench(ab).But iBenchmark is more powerful then them.It not only can send queries on alive connections but also can established short connections which send queries on depand on your willings.So you can use it to test the web application's both QPS (queries per second) and CPS (connections per second) performance.

####Framework:

iBenchmark can take full advantage of your servers' resources.The M parameter set the max cpu cores to use,so you can make enough pressure to the web service which depends on you servers' resources.It init M goroutines to run M workers.And each worker has two goroutines:one to handle the http/https/spdy request,when the request roundTrip is finished,it notify the second.The second to generate some info and finially nofity the worker how many requests has finished.Then the worker exits depending on the numbers of finished requests.
When all of the workers exits,the main exit and prints the report.
<img src="http://www.shaheng.me/images_pri/ibench/frame.png" width = "800" height = "500" alt="ibench_frame" align=center />
####Result
<img src="http://www.shaheng.me/images_pri/ibench/result.png" width = "800" height = "300" alt="ibench_result" align=center />

> iBenchmark使用Go语言自研发，为测试HTTPS Server的QPS、CPS性能指标而设计。囊括了ab、wrk的特性，意在为HTTPS的性能测试量身打造，亦支持HTTP的性能测试。

> 特性：

> 支持参数调整iBenchmark的CPU使用量

> 支持参数调整iBenchmark的并发量、请求量，以及运行时间。

> 支持参数调整iBenchmark的已建立的链接特性，可设置是否为长连接。故可参数调整性能测试目的—测试QPS(Queries Per Second)、CPS(Connection Per Second)性能。以及在短连接上可参数调配每个短连接可发送的请求数。

> 支持参数调整HTTPS的加密套件的选择，为HTTPS不同加密套件下的CPS、QPS指标量身打造。

> 支持参数调整HTTP请求的方法(GET/HEAD/PUT/POST…)，以及HEADER和BODY体。

> 支持自适应URL解析：schema(http or https),host,port,path.

> 简洁的性能测试结果总结:服务器的Server信息，Query信息，响应文档信息，QPS/CPS性能总结，持续时间，响应延时，以及成功率

> 简洁易懂的参数输入以及帮助

> 使用GO语言编写，简洁易维护，并且高性能，可以充分使用服务器的多核CPU优势，能够发送足量压力。


> 与开源HTTP性能测试工具的对比，例如ab/wrk：

> ab最大使用单核CPU，性能不够。并且只支持HTTP的性能测试，不支持HTTPS。

> wrk只支持长连接，不支持短连接的CPS性能测试。虽支持HTTPS，但不能参数调配加密套件。



#Install
Simple as it takes to type the following command(online):

> ./install.sh

Or if you wanna build it on the local if you have already download the src,you can type following command(offline):

> ./install_local.sh <br>
> In this way,the lib won't install on go lib dir,this will install them on the {$pwd}/.ibenchmark.



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
