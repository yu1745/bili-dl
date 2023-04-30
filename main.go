package main

import (
	"flag"
	"github.com/yu1745/bili-dl/C"
	"github.com/yu1745/bili-dl/api"
	"github.com/yu1745/bili-dl/util"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&C.Cookie, "c", "", "cookie,cookie的key是SESSDATA,不设置只能下载480P")
	flag.StringVar(&C.UP, "up", "", "up主id,设置后会下载该up主的所有视频")
	flag.StringVar(&C.O, "o", ".", "下载路径,可填相对或绝对路径,建议在windows下使用相对路径避免无聊的正反斜杠问题")
	flag.IntVar(&C.J, "j", 1, "同时下载的任务数,默认为1")
	flag.StringVar(&C.BVs, "bv", "", "1-n个bv号,用逗号分隔,如:BVxxxxxx,BVyyyyyyy")
	flag.BoolVar(&C.Merge, "m", true, "是否合并视频,默认为true")
	flag.BoolVar(&C.Delete, "d", true, "合并后是否删除单视频和单音频,默认为true")
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
					//wg.Add(1)
					//go func() {
					//	defer wg.Done()
					err := api.Merge(stream)
					if err != nil {
						log.Println(err)
					}
					//}()
				}
			}()
		}
		//wg.Wait()
		limit.Wait()
	}
	if C.UP != "" {
		videos, err := api.AllVideo(C.UP)
		if err != nil {
			log.Fatalln(err)
		}
		limit := util.NewGoLimit(C.J)
		//wg := &sync.WaitGroup{}
		for _, v := range videos {
			limit.Add()
			//wg.Add(1)
			go func(v api.Video) {
				defer limit.Done()
				//defer wg.Done()
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
						//wg.Add(1)
						//go func() {
						//defer wg.Done()
						if C.Merge {
							err := api.Merge(stream)
							if err != nil {
								log.Println(err)
							}
						}
						//}()
					}
					break
				}
			}(v)
		}
		//wg.Wait()
		limit.Wait()
	}
	log.Println("done")
}
