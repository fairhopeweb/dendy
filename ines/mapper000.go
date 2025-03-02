package ines

import (
	"encoding/gob"
	"log"
)

// Mapper0 is the simplest mapper. It has no registers, and it only supports
// 16KB or 32KB PRG-ROM banks and 8KB CHR-ROM banks.
//
//	PRG-ROM is mapped to 0x8000-0xFFFF.
//	CHR-ROM is mapped to 0x0000-0x1FFF.
type Mapper0 struct {
	rom *ROM
}

func NewMapper0(cart *ROM) *Mapper0 {
	return &Mapper0{
		rom: cart,
	}
}

func (m *Mapper0) Reset() {
}

func (m *Mapper0) ScanlineTick() {
}

func (m *Mapper0) PendingIRQ() bool {
	return false
}

func (m *Mapper0) MirrorMode() MirrorMode {
	return m.rom.MirrorMode
}

func (m *Mapper0) ReadPRG(addr uint16) byte {
	switch {
	case addr >= 0x8000 && addr <= 0xFFFF:
		idx := addr % uint16(len(m.rom.PRG))
		return m.rom.PRG[idx]
	default:
		log.Printf("[WARN] mapper0: unhandled prg read at %04X", addr)
		return 0
	}
}

func (m *Mapper0) WritePRG(addr uint16, data byte) {
}

func (m *Mapper0) ReadCHR(addr uint16) byte {
	switch {
	case addr >= 0x0000 && addr <= 0x1FFF:
		return m.rom.CHR[addr]
	default:
		log.Printf("[WARN] mapper0: unhandled chr read at %04X", addr)
		return 0
	}
}

func (m *Mapper0) WriteCHR(addr uint16, data byte) {
}

func (m *Mapper0) Save(enc *gob.Encoder) error {
	return m.rom.Save(enc)
}

func (m *Mapper0) Load(dec *gob.Decoder) error {
	return m.rom.Load(dec)
}
