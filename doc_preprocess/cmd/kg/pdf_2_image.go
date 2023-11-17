package main

// import (
// 	"fmt"
// 	"image/jpeg"
// 	"os"
// 	"path/filepath"

// 	"github.com/karmdip-mi/go-fitz"
// )

// func pdf2Image(pdfPath string) {
// 	doc, err := fitz.New(pdfPath)
// 	if err != nil {
// 		panic(err)
// 	}

// 	defer doc.Close()

// 	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Extract pages as images
// 	for n := 0; n < doc.NumPage(); n++ {
// 		img, err := doc.Image(n)
// 		if err != nil {
// 			panic(err)
// 		}

// 		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.jpg", n)))
// 		if err != nil {
// 			panic(err)
// 		}

// 		err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
// 		if err != nil {
// 			panic(err)
// 		}

// 		f.Close()
// 	}
// }
