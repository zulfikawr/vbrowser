//go:build !linux || !cgo

package xorg

import (
	"image"
	"time"

	"github.com/zulfikawr/vbrowser/pkg/types"
)

func GetScreenConfigurations() {}

func DisplayOpen(display string) bool {
	return false
}

func DisplayClose() {}

func Move(x, y int) {}

func GetCursorPosition() (int, int) {
	return 0, 0
}

func Scroll(deltaX, deltaY float64, controlKey bool) {}

func ButtonDown(code uint32) error {
	return nil
}

func KeyDown(code uint32) error {
	return nil
}

func ButtonUp(code uint32) error {
	return nil
}

func KeyUp(code uint32) error {
	return nil
}

func ResetKeys() {}

func CheckKeys(duration time.Duration) {}

func ChangeScreenSize(s types.ScreenSize) (types.ScreenSize, error) {
	return s, nil
}

func GetScreenSize() types.ScreenSize {
	return types.ScreenSize{
		Width:  1280,
		Height: 720,
		Rate:   60,
	}
}

func SetKeyboardModifier(mod KbdMod, active bool) {}

func GetKeyboardModifiers() KbdMod {
	return 0
}

func GetCursorImage() *types.CursorImage {
	return nil
}

func GetScreenshotImage() *image.RGBA {
	return nil
}
