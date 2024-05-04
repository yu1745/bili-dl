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
-bv string
    单或多个bv号, 多个时用逗号分隔, 如: "BVxxxxxx,BVyyyyyyy"
    可以通过在浏览器控制台输入以下代码来获取整页的BV
    Array.from(document.querySelectorAll('#submit-video-list > ul.clearfix.cube-list > li')).map(e=>e.dataset['aid']).join(',')
-c string
    cookie,cookie的key是SESSDATA,不设置只能下载清晰度小于等于480P的视频
-d    合并后是否删除单视频和单音频 (default true)
-j int
    同时下载的任务数
    机械硬盘不应超过5 (default 1)
-m    是否合并视频流和音频流, 不合并将得到单独的视频(不含音频)和单独的音频(不含视频)文件, 不利于正常播放 (default true)
-no-overwrite
    跳过下载过的视频
    注意: 需要先前下载时没有指定suffix为false (default true)
-o string
    下载路径,可填相对或绝对路径,建议在windows下使用相对路径避免正反斜杠问题 (default ".")
-suffix
    在下载的视频文件名后添加bv号
    用来解决视频重名问题
    关闭后跳过已下载功能将失效 (default true)
```
