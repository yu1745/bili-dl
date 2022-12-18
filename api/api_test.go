package api

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"testing"
)

var testVideo = Video{BV: "BV1uP4y1S7ps", Cid: "873198432", Title: "会有人不喜欢玛奇玛？硬撑罢了！"}

func TestVideoInfo(t *testing.T) {
	bytes, err := videoInfo("BV1uP4y1S7ps")
	if err != nil {
		t.Error(err)
	}
	t.Log(jsoniter.Get(bytes, "data", "cid").ToString())
}

func TestAllVideo(t *testing.T) {
	/*videos*/ _, err := AllVideo("2223018")
	if err != nil {
		t.Log(err)
	}
	//fmt.Printf("%+v\n", videos)
}

func TestStream(t *testing.T) {
	stream, err := GetStream(testVideo)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", *stream)
}

func TestDl(t *testing.T) {
	stream, err := GetStream(testVideo)
	if err != nil {
		t.Error(err)
	}
	err = Dl(stream)
	if err != nil {
		t.Error(err)
	}
}

func Test_fileNameFix(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test", args{"a|a"}, "a a"},
		{"test", args{"a:a"}, "a a"},
		{"test", args{"a*a"}, "a a"},
		{"test", args{"a?a"}, "a a"},
		{"test", args{"a<a"}, "a a"},
		{"test", args{"a>a"}, "a a"},
		{"test", args{"a\"a"}, "a a"},
		{"test", args{"a\\a"}, "a a"},
		{"test", args{"a/a"}, "a a"},
		{"test", args{"a|a:a*a?a<a>a\"a\\a/a"}, "a a a a a a a a a a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileNameFix(tt.args.name); got != tt.want {
				t.Errorf("fileNameFix() = %v, want %v", got, tt.want)
			}
		})
	}
}
