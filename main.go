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
	"sync"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate)
	flag.StringVar(&C.Cookie, "c", "", "cookie")
	flag.StringVar(&C.UP, "up", "", "up")
	flag.StringVar(&C.O, "o", ".", "output dir")
	flag.IntVar(&C.J, "j", 1, "concurrent threads")
	flag.Parse()
	C.WD, _ = os.Getwd()
	if !strings.HasPrefix(C.O, "/") {
		C.O = filepath.Join(C.WD, C.O)
	}
	println(C.O)
	cmd := exec.Command("ffmpeg", "-version")
	//cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	} else {
		C.FFMPEG = true
	}
}

func main() {
	videos, err := api.AllVideo(C.UP)
	if err != nil {
		log.Fatalln(err)
	}
	limit := util.NewGoLimit(C.J)
	wg := &sync.WaitGroup{}
	for _, v := range videos {
		limit.Add()
		wg.Add(1)
		go func(v api.Video) {
			defer limit.Done()
			defer wg.Done()
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
					go func() {
						wg.Add(1)
						defer wg.Done()
						cmd := exec.Command("ffmpeg", "-y", "-i", filepath.Join(C.O, stream.Title+".mp4"), "-i", filepath.Join(C.O, stream.Title+".mp3"), "-c", "copy", filepath.Join(C.O, stream.Title+"-merged.mp4"))
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err := cmd.Run()
						if err != nil {
							log.Fatalln(err)
						}
					}()
				}
				break
			}
		}(v)
	}
	wg.Wait()
}
