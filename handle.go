package function

//go:generate go install github.com/rakyll/statik@latest
//go:generate statik -src=data -f -include "*"

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/mattn/knative-func-anonymousface/statik"

	pigo "github.com/esimov/pigo/core"
	"github.com/nfnt/resize"
	"github.com/rakyll/statik/fs"
	"golang.org/x/image/draw"
)

var (
	maskImg    image.Image
	classifier *pigo.Pigo
)

func init() {
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	f, err := statikFS.Open("/mask.png")
	if err != nil {
		log.Fatal("cannot open mask.png:", err)
	}
	defer f.Close()

	maskImg, _, err = image.Decode(f)
	if err != nil {
		log.Fatal("cannot decode mask.png:", err)
	}

	f, err = statikFS.Open("/facefinder")
	if err != nil {
		log.Fatal("cannot open facefinder:", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("cannot read facefinder:", err)
	}

	pigo := pigo.NewPigo()
	classifier, err = pigo.Unpack(b)
	if err != nil {
		log.Fatal("cannot unpack facefinder:", err)
	}
}

func Handle(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	if req.Method == "GET" {
		fmt.Fprintln(w, "curl -i -X POST -F image=@input.jpg http://anonymousface.default.127.0.0.1.nip.io:8080 > out.jpg")
		return
	}

	f, _, err := req.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bounds := img.Bounds().Max
	param := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     2000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: pigo.RgbToGrayscale(pigo.ImgToNRGBA(img)),
			Rows:   bounds.Y,
			Cols:   bounds.X,
			Dim:    bounds.X,
		},
	}
	faces := classifier.RunCascade(param, 0)
	faces = classifier.ClusterDetections(faces, 0.18)

	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, img.Bounds(), img, image.Point{0, 0}, draw.Over)
	for _, face := range faces {
		pt := image.Point{face.Col - face.Scale/2, face.Row - face.Scale/2}
		fimg := resize.Resize(uint(face.Scale), uint(face.Scale), maskImg, resize.NearestNeighbor)
		draw.Copy(canvas, pt, fimg, fimg.Bounds(), draw.Over, nil)
	}

	w.Header().Add("Content-Type", "image/jpeg")
	err = jpeg.Encode(w, canvas, &jpeg.Options{Quality: 100})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
