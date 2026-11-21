package service

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

var IMAGE_OUTPUT_DIR = func() string {
	if v := os.Getenv("IMAGE_OUTPUT_DIR"); v != "" {
		return v
	}
	return "./"
}()

func (s *Service) UploadLabelImage(file *multipart.FileHeader, outputFileName string) error {
	// ファイルを開く
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// アップロードファイルをバッファに読み込む
	buf, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read uploaded file: %w", err)
	}

	return s.SaveLabelImageBytes(buf, outputFileName, "")
}

const MaxLabelImageBytes = 5 * 1024 * 1024

func (s *Service) SaveLabelImageBytes(data []byte, outputFileName, declaredMIME string) error {
	if len(data) == 0 || (declaredMIME != "" && len(data) > MaxLabelImageBytes) {
		return errors.New("label image must be between 1 byte and 5 MiB")
	}
	if outputFileName == "" || filepath.Base(outputFileName) != outputFileName || strings.ContainsAny(outputFileName, `/\`) {
		return errors.New("invalid generated label image filename")
	}
	detectedMIME := http.DetectContentType(data)
	if declaredMIME != "" && detectedMIME != "image/jpeg" && detectedMIME != "image/png" {
		return errors.New("label image must be a JPEG or PNG image")
	}
	if declaredMIME != "" && declaredMIME != detectedMIME {
		return errors.New("declared image MIME type does not match image content")
	}

	img, err := autoRotateByExif(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to auto-rotate image: %w", err)
	}

	// 幅300pxにリサイズ（縦横比を維持）
	thumbnail := imaging.Resize(img, 300, 0, imaging.Lanczos)

	// 保存先パスを作成
	outPath := filepath.Join(IMAGE_OUTPUT_DIR, outputFileName)
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	if err := jpeg.Encode(outFile, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		outFile.Close()
		os.Remove(outPath)
		return fmt.Errorf("failed to encode and save thumbnail: %w", err)
	}
	if err := outFile.Close(); err != nil {
		os.Remove(outPath)
		return fmt.Errorf("failed to close label image: %w", err)
	}

	return nil
}

func autoRotateByExif(r io.Reader) (image.Image, error) {
	// ファイルをバッファに保持（Exif解析とDecodeに使用）
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// 画像をデコード
	img, _, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	// EXIFを解析
	x, err := exif.Decode(bytes.NewReader(buf))
	if err != nil {
		// EXIFがなければ回転せず返す
		return img, nil
	}

	orientTag, err := x.Get(exif.Orientation)
	if err != nil || orientTag == nil {
		return img, nil
	}

	orientVal, err := orientTag.Int(0)
	if err != nil {
		return img, nil
	}

	// Orientationに応じて回転
	switch orientVal {
	case 3:
		img = imaging.Rotate180(img)
	case 6:
		img = imaging.Rotate270(img)
	case 8:
		img = imaging.Rotate90(img)
	}

	return img, nil
}
