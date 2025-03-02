package ppu

import (
	"image/color"
)

const (
	spriteAttrPalette  = 0x03 // two bits
	spriteAttrPriority = 1 << 5
	spriteAttrFlipX    = 1 << 6
	spriteAttrFlipY    = 1 << 7
)

type Sprite struct {
	Index     int
	Pixels    [8][16]uint8
	PaletteID uint8
	X, Y      uint8
	FlipX     bool
	FlipY     bool
	Behind    bool
}

func flipPixels(pixels [8][16]uint8, flipX, flipY bool, height int) (flipped [8][16]uint8) {
	if !flipX && !flipY {
		return pixels
	}

	for y := 0; y < height; y++ {
		for x := 0; x < 8; x++ {
			fx, fy := x, y

			if flipX {
				fx = 7 - x
			}

			if flipY {
				fy = height - y - 1
			}

			flipped[fx][fy] = pixels[x][y]
		}
	}

	return flipped
}

func (p *PPU) spritePatternTableAddr() uint16 {
	if p.getCtrl(CtrlSpritePatternAddr) {
		return 0x1000
	}

	return 0
}

func (p *PPU) spriteHeight() int {
	if p.getCtrl(CtrlSpriteSize) {
		return 16
	}

	return 8
}

func (p *PPU) fetchSprite(idx int) Sprite {
	var (
		id      = p.oamData[idx*4+1]
		attr    = p.oamData[idx*4+2]
		spriteX = p.oamData[idx*4+3]
		spriteY = p.oamData[idx*4+0]
		height  = p.spriteHeight()
	)

	sprite := Sprite{
		Index:     idx,
		PaletteID: attr & spriteAttrPalette,
		Behind:    attr&spriteAttrPriority != 0,
		FlipX:     attr&spriteAttrFlipX != 0,
		FlipY:     attr&spriteAttrFlipY != 0,
		Y:         spriteY,
		X:         spriteX,
	}

	var addr uint16

	for y := 0; y < height; y++ {
		if height == 16 {
			table := id & 1
			tile := id & 0xFE
			if y >= 8 {
				tile++
			}
			addr = uint16(table)*0x1000 + uint16(tile)*16 + uint16(y&7)
		} else {
			addr = p.spritePatternTableAddr() + uint16(id)*16 + uint16(y)
		}

		p1 := p.readVRAM(addr + 0)
		p2 := p.readVRAM(addr + 8)

		for x := 0; x < 8; x++ {
			px := p1 & (0x80 >> x) >> (7 - x) << 0
			px |= (p2 & (0x80 >> x) >> (7 - x)) << 1
			sprite.Pixels[x][y] = px // two-bit pixel value (0-3)
		}
	}

	return sprite
}

func (p *PPU) evaluateSprites() {
	height := p.spriteHeight()
	scanline := p.scanline + 1
	p.spriteCount = 0

	for i := 0; i < 64; i++ {
		spriteY := int(p.oamData[i*4+0])

		if scanline < spriteY || scanline >= spriteY+height {
			continue
		}

		if p.spriteCount == 8 {
			p.setStatus(StatusSpriteOverflow, true)
			if !p.NoSpriteLimit {
				break
			}
		}

		p.spriteScanline[p.spriteCount] = p.fetchSprite(i)
		p.spriteCount++
	}
}

func (p *PPU) readSpriteColor(pixel, paletteID uint8) color.RGBA {
	colorAddr := 0x3F10 + uint16(paletteID)*4 + uint16(pixel)
	colorIdx := p.readVRAM(colorAddr)
	return Colors[colorIdx%64]
}

func (p *PPU) renderSpriteScanline() {
	frameY := p.scanline
	if frameY > 239 {
		return
	}

	var (
		height = p.spriteHeight()
	)

	for i := p.spriteCount - 1; i >= 0; i-- {
		sprite := p.spriteScanline[i]

		pixels := flipPixels(sprite.Pixels, sprite.FlipX, sprite.FlipY, height)
		pixelY := (p.scanline - int(sprite.Y)) % height

		for pixelX := 0; pixelX < 8; pixelX++ {
			frameX := int(sprite.X) + pixelX
			if frameX > 255 {
				continue
			}

			if !p.getMask(MaskShowLeftSprites) && frameX < 8 {
				continue
			}

			px := pixels[pixelX][pixelY]
			if px == 0 {
				continue
			}

			// Sprite zero hit detection.
			if sprite.Index == 0 && !p.transparent[frameX][frameY] {
				p.setStatus(StatusSpriteZeroHit, true)
			}

			// Sprite is behind the background, so don't render.
			if sprite.Behind && !p.transparent[frameX][frameY] {
				continue
			}

			p.Frame[frameX][frameY] = p.readSpriteColor(px, sprite.PaletteID)
		}
	}
}
