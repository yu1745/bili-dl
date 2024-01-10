package api

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/yu1745/bili-dl/C"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		//禁止复用连接，防止同一个连接长时间大流量被限速
		DisableKeepAlives: true,
	},
}

func videoInfo(bv string) ([]byte, error) {
	url := "https://api.bilibili.com/x/web-interface/view"
	parse, _ := url2.Parse(url)
	query := parse.Query()
	query.Add("bvid", bv)
	parse.RawQuery = query.Encode()
	url = parse.String()
	method := "GET"
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://www.apifox.cn)")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//log.Println(string(body))
	return body, nil
}

func ResolveVideo(v *Video) (*Video, error) {
	info, err := videoInfo(v.BV)
	if err != nil {
		return nil, err
	}
	cid := jsoniter.Get(info, "data", "cid").ToString()
	title := jsoniter.Get(info, "data", "title").ToString()
	//v.BV = bv
	v.Cid = cid
	v.Title = title
	return v, nil
}

func videoFromUP(mid string, pn int) (rt []byte, err error) {
	url := "https://api.bilibili.com/x/space/wbi/arc/search?order=pubdate&ps=49"
	parse, _ := url2.Parse(url)
	query := parse.Query()
	query.Add("mid", mid)
	query.Add("pn", strconv.Itoa(pn))
	parse.RawQuery = query.Encode()
	url = parse.String()
	url, err = sign(url)
	if err != nil {
		return nil, err
	}
	method := "GET"

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://www.apifox.cn)")

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//log.Println(string(body))
	return body, err
}

type Video struct {
	Title  string `json:"title,omitempty"`
	BV     string `json:"bv,omitempty"`
	Cid    string `json:"cid,omitempty"`
	Author string `json:"author,omitempty"`
}

func AllVideo(mid string) ([]Video, error) {
	bytes, err := videoFromUP(mid, 1)
	if err != nil {
		return nil, err
	}
	var videos []Video
	if C.Debug {
		log.Println(string(bytes))
	}
	count := jsoniter.Get(bytes, "data", "page", "count").ToInt()
	var pn int
	n := 49
	if count%n == 0 {
		pn = count / n
	} else {
		pn = count/n + 1
	}
	for i := 1; i <= pn; i++ {
		time.Sleep(time.Second)
		bytes, err := videoFromUP(mid, i)
		if err != nil {
			return nil, err
		}
		vlist := jsoniter.Get(bytes, "data", "list", "vlist").ToString()
		var m []map[string]any
		err = jsoniter.Unmarshal([]byte(vlist), &m)
		if err != nil {
			return nil, err
		}
		for _, v := range m {
			if bvid := v["bvid"]; bvid != nil {
				//log.Println(bvid)
				/*info, err := videoInfo(bvid.(string))
				if err != nil {
					return nil, err
				}
				cid := jsoniter.Get(info, "data", "cid").ToString()
				title := jsoniter.Get(info, "data", "title").ToString()*/
				video := Video{BV: bvid.(string), Author: mid /*, Cid: cid, Title: title*/}
				//log.Printf("%+v\n", video)
				videoJson, err := jsoniter.MarshalToString(&video)
				if err != nil {
					return nil, err
				}
				println(videoJson)
				videos = append(videos, video)
			}
		}
	}
	return videos, nil
}

func codec2i(codec string) int {
	if strings.HasPrefix(codec, "avc") {
		return 1
	} else if strings.HasPrefix(codec, "hev") {
		return 2
	} else if strings.HasPrefix(codec, "av01") {
		return 3
	} else {
		return 0
	}
}

type Stream struct {
	V string
	A string
	Video
}

func GetStream(v Video) (*Stream, error) {
	url := "https://api.bilibili.com/x/player/playurl?fnver=0&fnval=3216&fourk=1&qn=127"
	parse, _ := url2.Parse(url)
	query := parse.Query()
	query.Add("bvid", v.BV)
	query.Add("cid", v.Cid)
	parse.RawQuery = query.Encode()
	url = parse.String()
	method := "GET"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{Name: "SESSDATA", Value: C.Cookie})

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	videos := jsoniter.Get(body, "data", "dash", "video").ToString()
	var l []map[string]any
	err = jsoniter.Unmarshal([]byte(videos), &l)
	if err != nil {
		return nil, err
	}
	sort.Slice(l, func(i, j int) bool {
		if codec2i(l[i]["codecs"].(string)) == codec2i(l[j]["codecs"].(string)) {
			return l[i]["width"].(float64) > l[j]["width"].(float64)
		} else {
			return codec2i(l[i]["codecs"].(string)) > codec2i(l[j]["codecs"].(string))
		}
	})
	audios := jsoniter.Get(body, "data", "dash", "audio").ToString()
	var l2 []map[string]any
	err = jsoniter.Unmarshal([]byte(audios), &l2)
	if err != nil {
		return nil, err
	}
	sort.Slice(l2, func(i, j int) bool {
		return l2[i]["bandwidth"].(float64) > l2[j]["bandwidth"].(float64)
	})
	stream := &Stream{V: l[0]["base_url"].(string), A: l2[0]["base_url"].(string), Video: v}
	return stream, nil
}

func Dl(stream *Stream) error {
	stream.Title = fileNameFix(stream.Title)
	err := DV(stream)
	if err != nil {
		return err
	}
	err = DA(stream)
	if err != nil {
		return err
	}
	log.Println(stream.Title, "下载完成")
	return nil
}

func DV(stream *Stream) error {
	req, err := http.NewRequest("GET", stream.V, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://www.bilibili.com")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(C.O, stream.Title+".mp4"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	_ = file.Truncate(0)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func DA(stream *Stream) error {
	req, err := http.NewRequest("GET", stream.A, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://www.bilibili.com")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(C.O, stream.Title+".mp3"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	_ = file.Truncate(0)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func VideoFromBV(bv string) (*Video, error) {
	info, err := videoInfo(bv)
	if err != nil {
		return nil, err
	}
	cid := jsoniter.Get(info, "data", "cid").ToString()
	title := jsoniter.Get(info, "data", "title").ToString()
	video := Video{BV: bv, Cid: cid, Title: title}
	log.Printf("%+v\n", video)
	return &video, nil
}

func Merge(stream *Stream) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", filepath.Join(C.O, stream.Title+".mp4"), "-i", filepath.Join(C.O, stream.Title+".mp3"), "-c", "copy", filepath.Join(C.O, stream.Title+"-merged.mp4"))
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	if C.Delete {
		err := os.Remove(filepath.Join(C.O, stream.Title+".mp4"))
		if err != nil {
			return err
		}
		err = os.Remove(filepath.Join(C.O, stream.Title+".mp3"))
		if err != nil {
			return err
		}
		err = os.Rename(filepath.Join(C.O, stream.Title+"-merged.mp4"), filepath.Join(C.O, stream.Title+".mp4"))
		if err != nil {
			return err
		}
	}
	log.Println(stream.Title, "合并完成")
	return nil
}

var reg = regexp.MustCompile(`[/\\:*?"<>|]`)

// 去掉文件名中的非法字符
func fileNameFix(name string) string {
	return reg.ReplaceAllString(name, " ")
}
