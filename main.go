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
	"strings"
)

func init() {
	log.SetFlags(log.Lshortfile)
	flag.StringVar(&C.Cookie, "c", "", "cookie, if not set, only low resolution video available")
	flag.StringVar(&C.UP, "up", "", "download all video from this up")
	flag.StringVar(&C.O, "o", ".", "output dir")
	flag.IntVar(&C.J, "j", 1, "concurrent threads")
	flag.StringVar(&C.BVs, "bv", "", "bvids, split by comma")
	flag.BoolVar(&C.Merge, "m", true, "merge audio and video")
	flag.BoolVar(&C.Delete, "d", true, "delete pure audio and pure video after merge, only remain merged video, only work when -m is true")
	flag.Parse()
	C.WD, _ = os.Getwd()
	if !strings.HasPrefix(C.O, "/") {
		C.O = filepath.Join(C.WD, C.O)
		err := os.MkdirAll(C.O, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	log.Println("Download Path:", C.O)
	cmd := exec.Command("ffmpeg", "-version")
	//cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		log.Println("ffmpeg not found, will not merge audio and video")
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
