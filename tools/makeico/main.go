// makeico 将 PNG 图片转换为包含多个尺寸的 ICO 文件
// 用法: go run ./tools/makeico/
package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	data, err := os.ReadFile("launcher/icon.png")
	if err != nil {
		panic(err)
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	// ICO 包含多个尺寸，Windows 会自动选择合适的
	sizes := []int{16, 32, 48, 256}
	var images [][]byte

	for _, size := range sizes {
		pngData := resizeAndEncode(src, size)
		images = append(images, pngData)
	}

	ico := buildICO(images, sizes)
	if err := os.WriteFile("launcher/icon.ico", ico, 0644); err != nil {
		panic(err)
	}
}

func resizeAndEncode(src image.Image, size int) []byte {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			sx := x * sw / size
			sy := y * sh / size
			r, g, b, a := src.At(sb.Min.X+sx, sb.Min.Y+sy).RGBA()
			dst.Set(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, dst)
	return buf.Bytes()
}

func buildICO(images [][]byte, sizes []int) []byte {
	count := len(images)
	headerSize := 6 + 16*count // ICONDIR + N*ICONDIRENTRY

	// 计算每张图的偏移量
	offsets := make([]int, count)
	offsets[0] = headerSize
	for i := 1; i < count; i++ {
		offsets[i] = offsets[i-1] + len(images[i-1])
	}

	var buf bytes.Buffer

	// ICONDIR 头 (6 字节)
	binary.Write(&buf, binary.LittleEndian, uint16(0))     // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1))     // type = 1 (ICO)
	binary.Write(&buf, binary.LittleEndian, uint16(count)) // count

	// ICONDIRENTRY 列表 (每个 16 字节)
	for i, pngData := range images {
		size := sizes[i]
		w := byte(size)
		if size >= 256 {
			w = 0
		}
		h := byte(size)
		if size >= 256 {
			h = 0
		}
		buf.WriteByte(w)
		buf.WriteByte(h)
		buf.WriteByte(0)                                              // colorCount
		buf.WriteByte(0)                                              // reserved
		binary.Write(&buf, binary.LittleEndian, uint16(1))            // planes
		binary.Write(&buf, binary.LittleEndian, uint16(32))           // bitCount
		binary.Write(&buf, binary.LittleEndian, uint32(len(pngData))) // bytesInRes
		binary.Write(&buf, binary.LittleEndian, uint32(offsets[i]))   // imageOffset
	}

	// 所有图像数据
	for _, pngData := range images {
		buf.Write(pngData)
	}

	return buf.Bytes()
}
