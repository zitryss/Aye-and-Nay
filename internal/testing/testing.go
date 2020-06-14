package testing

import (
	"math"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zitryss/aye-and-nay/domain/model"
)

const (
	tolerance = 0.000000000000001
)

func CheckStatusCode(t *testing.T, w *httptest.ResponseRecorder, code int) {
	t.Helper()
	got := w.Code
	want := code
	if got != want {
		t.Errorf("Status Code = %v; want %v", got, want)
	}
}

func CheckContentType(t *testing.T, w *httptest.ResponseRecorder, content string) {
	t.Helper()
	got := w.Result().Header.Get("Content-Type")
	want := content
	if got != want {
		t.Errorf("Content-Type = %v; want %v", got, want)
	}
}

func CheckBody(t *testing.T, w *httptest.ResponseRecorder, body string) {
	t.Helper()
	got := w.Body.String()
	want := body
	if got != want {
		t.Errorf("Body = %v; want %v", got, want)
	}
}

func IsIn(image model.Image, imgs []model.Image) bool {
	for _, img := range imgs {
		if reflect.DeepEqual(image, img) {
			return true
		}
	}
	return false
}

func AlbumEmptyFactory(id string) model.Album {
	img1 := model.Image{Id: "RcBj3m9vuYPbntAE", Src: "/aye-and-nay/albums/" + id + "/images/6sgsr8WwqudTDzhR"}
	img2 := model.Image{Id: "Q3NafBGuDH9PAtS4", Src: "/aye-and-nay/albums/" + id + "/images/2H7NpJkPwBWUk6gL"}
	img3 := model.Image{Id: "442BbctbQhcQHrgH", Src: "/aye-and-nay/albums/" + id + "/images/kUrtHH5hTLbcSJdu"}
	img4 := model.Image{Id: "VYFczQcF45x7gLYH", Src: "/aye-and-nay/albums/" + id + "/images/428PcLG7e7VZHyAJ"}
	img5 := model.Image{Id: "qBmu5KGTqCdvfgTU", Src: "/aye-and-nay/albums/" + id + "/images/gXR6VrL9h7E3pFVY"}
	imgs := []model.Image{img1, img2, img3, img4, img5}
	edgs := map[string]map[string]int{}
	edgs["RcBj3m9vuYPbntAE"] = map[string]int{}
	edgs["Q3NafBGuDH9PAtS4"] = map[string]int{}
	edgs["442BbctbQhcQHrgH"] = map[string]int{}
	edgs["VYFczQcF45x7gLYH"] = map[string]int{}
	edgs["qBmu5KGTqCdvfgTU"] = map[string]int{}
	alb := model.Album{id, imgs, edgs}
	return alb
}

func AlbumFullFactory(id string) model.Album {
	alb := AlbumEmptyFactory(id)
	alb.Images[0].Rating = 0.48954984
	alb.Images[1].Rating = 0.19186324
	alb.Images[2].Rating = 0.41218211
	alb.Images[3].Rating = 0.77920413
	alb.Images[4].Rating = 0.13278389
	alb.Edges["VYFczQcF45x7gLYH"]["442BbctbQhcQHrgH"]++
	alb.Edges["RcBj3m9vuYPbntAE"]["442BbctbQhcQHrgH"]++
	alb.Edges["RcBj3m9vuYPbntAE"]["VYFczQcF45x7gLYH"]++
	alb.Edges["Q3NafBGuDH9PAtS4"]["442BbctbQhcQHrgH"]++
	alb.Edges["Q3NafBGuDH9PAtS4"]["VYFczQcF45x7gLYH"]++
	alb.Edges["Q3NafBGuDH9PAtS4"]["RcBj3m9vuYPbntAE"]++
	alb.Edges["qBmu5KGTqCdvfgTU"]["442BbctbQhcQHrgH"]++
	alb.Edges["qBmu5KGTqCdvfgTU"]["VYFczQcF45x7gLYH"]++
	alb.Edges["qBmu5KGTqCdvfgTU"]["RcBj3m9vuYPbntAE"]++
	alb.Edges["qBmu5KGTqCdvfgTU"]["Q3NafBGuDH9PAtS4"]++
	return alb
}

func EqualMap(x, y map[string]float64) bool {
	if len(x) != len(y) {
		return false
	}
	for xk, xv := range x {
		yv, ok := y[xk]
		if !ok {
			return false
		}
		if !EqualFloat(xv, yv) {
			return false
		}
	}
	return true
}

func EqualFloat(x, y float64) bool {
	diff := math.Abs(x - y)
	if diff > tolerance {
		return false
	}
	return true
}

func Png() []byte {
	return []byte{137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82, 0, 0, 0, 1, 0, 0, 0, 1, 8, 6, 0, 0, 0, 31, 21, 196, 137, 0, 0, 0, 10, 73, 68, 65, 84, 120, 156, 99, 0, 1, 0, 0, 5, 0, 1, 13, 10, 45, 180, 0, 0, 0, 0, 73, 69, 78, 68, 174, 66, 96, 130}
}
