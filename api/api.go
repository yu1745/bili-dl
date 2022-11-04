package api

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/yu1745/bili/C"
	"io"
	"log"
	"net/http"
	url2 "net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var client = &http.Client{}

func videoInfo(bv string) ([]byte, error) {
	url := "http://api.bilibili.com/x/web-interface/view"
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

func videoFromUP(mid string, pn int) ([]byte, error) {
	url := "http://api.bilibili.com/x/space/arc/search?order=pubdate&ps=49"
	parse, _ := url2.Parse(url)
	query := parse.Query()
	query.Add("mid", mid)
	query.Add("pn", strconv.Itoa(pn))
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
	Title string
	BV    string
	Cid   string
}

func AllVideo(mid string) ([]Video, error) {
	bytes, err := videoFromUP(mid, 1)
	if err != nil {
		return nil, err
	}
	var videos []Video
	count := jsoniter.Get(bytes, "data", "page", "count").ToInt()
	var pn int
	if count%49 == 0 {
		pn = count / 49
	} else {
		pn = count/49 + 1
	}
	for i := 1; i <= pn; i++ {
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
				info, err := videoInfo(bvid.(string))
				if err != nil {
					return nil, err
				}
				cid := jsoniter.Get(info, "data", "cid").ToString()
				title := jsoniter.Get(info, "data", "title").ToString()
				video := Video{BV: bvid.(string), Cid: cid, Title: title}
				log.Printf("%+v\n", video)
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
	url := "http://api.bilibili.com/x/player/playurl?fnver=0&fnval=3216&fourk=1&qn=127"
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
	_ = file.Truncate(0)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	_ = file.Close()
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
	_ = file.Truncate(0)
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	_ = file.Close()
	return nil
}
