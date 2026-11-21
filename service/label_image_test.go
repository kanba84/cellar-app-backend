package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLabelImageBytesConvertsPNGAndRejectsUnsafeInput(t *testing.T) {
	outputDir := t.TempDir()
	originalDir := IMAGE_OUTPUT_DIR
	IMAGE_OUTPUT_DIR = outputDir
	t.Cleanup(func() { IMAGE_OUTPUT_DIR = originalDir })

	var source bytes.Buffer
	imageData := image.NewRGBA(image.Rect(0, 0, 1, 1))
	imageData.Set(0, 0, color.White)
	if err := png.Encode(&source, imageData); err != nil {
		t.Fatal(err)
	}
	service := &Service{}
	if err := service.SaveLabelImageBytes(source.Bytes(), "thumb_test.jpg", "image/png"); err != nil {
		t.Fatalf("valid PNG conversion failed: %v", err)
	}
	thumbnail, err := os.Open(filepath.Join(outputDir, "thumb_test.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := jpeg.Decode(thumbnail); err != nil {
		t.Fatalf("saved thumbnail is not JPEG: %v", err)
	}
	thumbnail.Close()

	if err := service.SaveLabelImageBytes(source.Bytes(), "thumb_legacy.png", ""); err != nil {
		t.Fatalf("legacy HTTP filename format should remain accepted: %v", err)
	}
	if err := service.SaveLabelImageBytes([]byte("not an image"), "thumb_bad.jpg", "image/png"); err == nil {
		t.Fatal("invalid image data must fail")
	}
	if err := service.SaveLabelImageBytes(source.Bytes(), "../outside.jpg", "image/png"); err == nil {
		t.Fatal("unsafe output filename must fail")
	}
}
