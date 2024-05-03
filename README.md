# bili-dl
## 安装
``` shell
go install github.com/yu1745/bili-dl@latest
```

## 注意

* 要想下载画质高于480P的视频请指定cookie, cookie获取方式为用浏览器登录b站后，按F12打开控制台，点击右上角加号，选择"应用"或"Application",选择存储,选择Cookie,选择www.bilibili.com然后找到名称是SESSDATA那一行，将值复制出来
* 需要环境变量中有ffmpeg，软件使用dash的方式取流，取得的音视频流是分开的，需要调用ffmpeg合并

## 功能

下载b站视频，支持批量下载，支持指定cookie实现高画质视频下载

``` shell
bili-dl -bv "BVfcasdsd,BVdsa1232das" ...其他参数
```
-j 指定下载并发数
-c 指定cookie下载高画质视频
-o 指定下载路径

更多功能请使用
``` shell
bili-dl -h 获取
```
