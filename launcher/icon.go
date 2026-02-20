package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
)

//go:embed icon.png
var iconData []byte

// makeIcon 将嵌入的 PNG 缩放到 32x32 并包装为 ICO 格式
// Windows 系统托盘要求 ICO 格式，直接传 PNG 会导致图标不显示
func makeIcon() []byte {
	src, err := png.Decode(bytes.NewReader(iconData))
	if err != nil {
		return iconData
	}

	// 将图片缩放到 32x32（像素艺术用最近邻插值，保留像素风格）
	const size = 32
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

	// 将 32x32 图像重新编码为 PNG
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, dst); err != nil {
		return iconData
	}

	// 将 PNG 包装成 ICO 文件格式（现代 Windows Vista+ 支持 ICO 内嵌 PNG）
	return wrapPNGinICO(pngBuf.Bytes(), size, size)
}

// wrapPNGinICO 将 PNG 数据包装为标准 ICO 文件格式
func wrapPNGinICO(pngData []byte, width, height int) []byte {
	var buf bytes.Buffer

	// ICONDIR 头 (6 字节)
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // type = 1 (ICO)
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // count = 1 张图

	// ICONDIRENTRY (16 字节)
	w := byte(width)
	if width >= 256 {
		w = 0 // 0 表示 256
	}
	h := byte(height)
	if height >= 256 {
		h = 0
	}
	buf.WriteByte(w)                                              // width
	buf.WriteByte(h)                                              // height
	buf.WriteByte(0)                                              // colorCount (0 = 无调色板)
	buf.WriteByte(0)                                              // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1))            // planes
	binary.Write(&buf, binary.LittleEndian, uint16(32))           // bitCount
	binary.Write(&buf, binary.LittleEndian, uint32(len(pngData))) // bytesInRes
	binary.Write(&buf, binary.LittleEndian, uint32(6+16))         // imageOffset = ICONDIR(6) + ICONDIRENTRY(16)

	// PNG 图像数据
	buf.Write(pngData)
	return buf.Bytes()
}
