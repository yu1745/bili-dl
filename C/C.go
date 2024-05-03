package C

var (
	Cookie           string
	UP               string
	O                string
	FFMPEG           bool
	WD               string
	J                int
	BVs              string
	Merge            bool
	Delete           bool
	Debug            bool
	AddBVSuffix      bool
	DisableOverwrite bool
)

const (
	GetAllBV = `Array.from(document.querySelectorAll('#submit-video-list > ul.clearfix.cube-list > li')).map(e=>e.dataset['aid']).join(',')`
)
