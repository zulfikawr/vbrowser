package xorg

//go:generate ./keysymdef.sh

type KbdMod uint8

const (
	KbdModShift    KbdMod = 1 << 0
	KbdModCapsLock KbdMod = 1 << 1
	KbdModControl  KbdMod = 1 << 2
	KbdModAlt      KbdMod = 1 << 3
	KbdModNumLock  KbdMod = 1 << 4
	KbdModMeta     KbdMod = 1 << 5
	KbdModSuper    KbdMod = 1 << 6
	KbdModAltGr    KbdMod = 1 << 7
)

type ScreenConfiguration struct {
	Width  int
	Height int
	Rates  map[int]int16
}

var ScreenConfigurations = make(map[int]ScreenConfiguration)
