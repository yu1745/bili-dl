package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/yu1745/bili-dl/C"
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

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	if C.Cookie != "" {
		req.AddCookie(&http.Cookie{Name: "SESSDATA", Value: C.Cookie})
	}

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
	Title string `json:"title,omitempty"`
	BV    string `json:"bv,omitempty"`
	Cid   string `json:"cid,omitempty"`
	// Author string `json:"author,omitempty"`
}

func AllVideo(mid string) ([]Video, error) {
	bytes, err := videoFromUP(mid, 1)
	if err != nil {
		return nil, err
	}
	var videos []Video
	count := jsoniter.Get(bytes, "data", "page", "count").ToInt()
	if count == 0 {
		// No videos or API error, try to parse what we got
		count = 1
	}
	// Calculate number of pages (ceiling division)
	pn := (count + 48) / 49
	if pn < 1 {
		pn = 1
	}

	for i := 1; i <= pn; i++ {
		time.Sleep(time.Second)
		bytes, err := videoFromUP(mid, i)
		if err != nil {
			log.Printf("Failed to fetch page %d: %v, continuing...", i, err)
			continue
		}
		vlist := jsoniter.Get(bytes, "data", "list", "vlist").ToString()
		if vlist == "" {
			continue
		}
		var m []map[string]any
		if err := jsoniter.Unmarshal([]byte(vlist), &m); err != nil {
			log.Printf("Failed to parse page %d: %v, continuing...", i, err)
			continue
		}
		for _, v := range m {
			if bvid := v["bvid"]; bvid != nil {
				if bvStr, ok := bvid.(string); ok {
					video := Video{BV: bvStr}
					videoJson, _ := jsoniter.MarshalToString(&video)
					println(videoJson)
					videos = append(videos, video)
				}
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
	}
	return 0
}

// Safe map accessors to avoid panics from type assertions
func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]any, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

type Stream struct {
	V string
	A string
	Video
}

func GetStream(v Video) (*Stream, error) {
	url := "https://api.bilibili.com/x/player/wbi/playurl?fnver=0&fnval=3216&fourk=1&qn=127"
	parse, _ := url2.Parse(url)
	query := parse.Query()
	query.Add("bvid", v.BV)
	query.Add("cid", v.Cid)
	parse.RawQuery = query.Encode()
	url = parse.String()
	var err error
	url, err = sign(url)
	if err != nil {
		return nil, err
	}
	method := "GET"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36 Edg/123.0.0.0")
	req.Header.Set("Referer", "https://www.bilibili.com/")

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
	if len(l) == 0 {
		return nil, fmt.Errorf("no video streams available")
	}
	// Find best video stream without full sort
	bestIdx := 0
	for i := 1; i < len(l); i++ {
		ci, cj := codec2i(getString(l[i], "codecs")), codec2i(getString(l[bestIdx], "codecs"))
		if ci > cj || (ci == cj && getFloat(l[i], "width") > getFloat(l[bestIdx], "width")) {
			bestIdx = i
		}
	}
	audios := jsoniter.Get(body, "data", "dash", "audio").ToString()
	var l2 []map[string]any
	err = jsoniter.Unmarshal([]byte(audios), &l2)
	if err != nil {
		return nil, err
	}
	if len(l2) == 0 {
		return nil, fmt.Errorf("no audio streams available")
	}
	// Find best audio stream without full sort
	bestAudioIdx := 0
	for i := 1; i < len(l2); i++ {
		if getFloat(l2[i], "bandwidth") > getFloat(l2[bestAudioIdx], "bandwidth") {
			bestAudioIdx = i
		}
	}
	stream := &Stream{V: getString(l[bestIdx], "base_url"), A: getString(l2[bestAudioIdx], "base_url"), Video: v}
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
	var file *os.File
	if C.AddBVSuffix {
		file, err = os.OpenFile(filepath.Join(C.O, stream.Title+"_"+stream.BV+".mp4"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(filepath.Join(C.O, stream.Title+".mp4"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
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
	var file *os.File
	if C.AddBVSuffix {
		file, err = os.OpenFile(filepath.Join(C.O, stream.Title+"_"+stream.BV+".mp3"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(filepath.Join(C.O, stream.Title+".mp3"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
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
	var video, audio, output string
	if C.AddBVSuffix {
		video = filepath.Join(C.O, stream.Title+"_"+stream.BV+".mp4")
		audio = filepath.Join(C.O, stream.Title+"_"+stream.BV+".mp3")
		output = filepath.Join(C.O, stream.Title+"_"+stream.BV+"-merged.mp4")
	} else {
		video = filepath.Join(C.O, stream.Title+".mp4")
		audio = filepath.Join(C.O, stream.Title+".mp3")
		output = filepath.Join(C.O, stream.Title+"-merged.mp4")
	}
	cmd := exec.Command("ffmpeg", "-y", "-i", video, "-i", audio, "-c", "copy", output)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg merge failed: %w", err)
	}
	if C.Delete {
		if err := os.Remove(video); err != nil {
			return err
		}
		if err := os.Remove(audio); err != nil {
			return err
		}
		if C.AddBVSuffix {
			if err := os.Rename(output, filepath.Join(C.O, stream.Title+"_"+stream.BV+".mp4")); err != nil {
				return err
			}
		} else {
			if err := os.Rename(output, filepath.Join(C.O, stream.Title+".mp4")); err != nil {
				return err
			}
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
