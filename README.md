# bili-dl
安装：
``` shell
go install github.com/yu1745/bili-dl@latest
```
功能：

* 下载指定视频，若要下载多个视频请用逗号分隔，例子

``` shell
bili-dl -bv "BVfcasdsd,BVdsa1232das" ...其他参数
```

* -j 指定下载并发数
* -c 指定cookie下载高画质视频，cookie的key为SESSDATA，可从浏览器登录b站然后f12获得，例子
``` shell
bili-dl -c 129a3230%2E1681271434%2D39bcc%2Ab2
```
上面的例子仅作格式参考，非真实cookie
* -o 指定下载路径

更多功能请使用
``` shell
bili-dl -h 获取
```
