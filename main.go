package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/yu1745/bili-dl/C"
	"github.com/yu1745/bili-dl/api"
	"github.com/yu1745/bili-dl/util"
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&C.Cookie, "c", "", "cookie,cookie的key是SESSDATA,不设置只能下载清晰度小于等于480P的视频")
	// flag.StringVar(&C.UP, "up", "", "up主id,设置后会下载该up主的所有视频")
	flag.StringVar(&C.O, "o", ".", "下载路径,可填相对或绝对路径,建议在windows下使用相对路径避免正反斜杠问题")
	flag.IntVar(&C.J, "j", 1, "同时下载的任务数\n机械硬盘不应超过5")
	flag.StringVar(&C.BVs, "bv", "", fmt.Sprintf("单或多个bv号, 多个时用逗号分隔, 如: \"BVxxxxxx,BVyyyyyyy\"\n可以通过在浏览器控制台输入以下代码来获取整页的BV\n%s", C.GetAllBV))
	flag.BoolVar(&C.Merge, "m", true, "是否合并视频流和音频流, 不合并将得到单独的视频(不含音频)和单独的音频(不含视频)文件, 不利于正常播放")
	flag.BoolVar(&C.Delete, "d", true, "合并后是否删除单视频和单音频")
	// flag.BoolVar(&C.Debug, "debug", false, "是否打印调试信息")
	flag.BoolVar(&C.AddBVSuffix, "suffix", true, "在下载的视频文件名后添加bv号\n用来解决视频重名问题\n关闭后跳过已下载功能将失效")
	flag.BoolVar(&C.DisableOverwrite, "no-overwrite", true, "跳过下载过的视频\n注意: 需要先前下载时没有指定suffix为false")
	flag.Parse()
	C.WD, _ = os.Getwd()
	if //goland:noinspection GoBoolExpressions
	runtime.GOOS == "windows" || runtime.GOOS == "nt" {
		pattern := `^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$`
		if matched, _ := regexp.MatchString(pattern, C.O); !matched {
			C.O = filepath.Join(C.WD, C.O)
			err := os.MkdirAll(C.O, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
	} else {
		if !strings.HasPrefix(C.O, "/") {
			C.O = filepath.Join(C.WD, C.O)
			err := os.MkdirAll(C.O, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
	}
	log.Println("下载路径: ", C.O)
	cmd := exec.Command("ffmpeg", "-version")
	//cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		log.Println("ffmpeg未找到，将不会合并音频和视频")
	} else {
		C.FFMPEG = true
	}
	reg := regexp.MustCompile(`.*_BV[a-zA-Z0-9]+\.mp4`)
	bvReg := regexp.MustCompile(`BV[a-zA-Z0-9]+`)
	bvs := make(map[string]struct{})
	for _, v := range strings.Split(C.BVs, ",") {
		if v != "" {
			bvs[v] = struct{}{}
		}
	}
	exists := make(map[string]struct{})
	if C.DisableOverwrite {
		err = filepath.WalkDir(C.O, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() && reg.MatchString(d.Name()) {
				exists[bvReg.FindString(d.Name())] = struct{}{}
			}
			return nil
		})
		if err != nil {
			log.Fatalln(err)
		}
		for k := range bvs {
			if _, ok := exists[k]; ok {
				log.Printf("%s已存在, 将不会下载\n", k)
				delete(bvs, k)
			}
		}
		keys := make([]string, 0, len(bvs))
		for k := range bvs {
			keys = append(keys, k)
		}
		C.BVs = strings.Join(keys, ",")
	}
}

func main() {
	if C.BVs != "" {
		split := strings.Split(C.BVs, ",")
		limit := util.NewGoLimit(C.J)
		//wg := &sync.WaitGroup{}
		for _, v := range split {
			limit.Add()
			//wg.Add(1)
			v := v
			go func() {
				defer limit.Done()
				//defer wg.Done()
				video, err := api.VideoFromBV(v)
				if err != nil {
					log.Println(err)
					return
				}
				_, err = api.ResolveVideo(video)
				if err != nil {
					log.Println(err)
					return
				}
				stream, err := api.GetStream(*video)
				if err != nil {
					log.Println(err)
					return
				}
				err = api.Dl(stream)
				if err != nil {
					log.Println(err)
					return
				}
				if C.FFMPEG {
					err := api.Merge(stream)
					if err != nil {
						log.Println(err)
					}
				}
			}()
		}
		limit.Wait()
	}
	if C.UP != "" {
		videos, err := api.AllVideo(C.UP)
		if err != nil {
			log.Fatalln(err)
		}
		limit := util.NewGoLimit(C.J)
		for _, v := range videos {
			limit.Add()
			go func(v api.Video) {
				defer limit.Done()
				_, err := api.ResolveVideo(&v)
				if err != nil {
					log.Println(err)
					return
				}
				for i := 0; i < 3; i++ {
					stream, err := api.GetStream(v)
					if err != nil {
						log.Println(err)
						continue
					}
					err = api.Dl(stream)
					if err != nil {
						log.Println(err)
						continue
					}
					if C.FFMPEG {
						if C.Merge {
							err := api.Merge(stream)
							if err != nil {
								log.Println(err)
							}
						}
					}
					break
				}
			}(v)
		}
		limit.Wait()
	}
	log.Println("下载完成")
}
