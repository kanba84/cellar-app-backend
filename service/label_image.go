package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

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

	// EXIF情報をもとに回転補正
	img, err := autoRotateByExif(bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("failed to auto-rotate image: %w", err)
	}

	// 幅300pxにリサイズ（縦横比を維持）
	thumbnail := imaging.Resize(img, 300, 0, imaging.Lanczos)

	// 保存先パスを作成
	outPath := filepath.Join(IMAGE_OUTPUT_DIR, outputFileName)
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// JPEG形式で保存（品質85で十分）
	if err := jpeg.Encode(outFile, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode and save thumbnail: %w", err)
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
