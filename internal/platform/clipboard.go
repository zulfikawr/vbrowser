package platform

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	lastContent   string
	lastContentMu sync.Mutex
)

// ReadClipboard reads the current clipboard content from the specified X11 display.
func ReadClipboard(display int) (string, error) {
	cmd := exec.Command("xclip", "-selection", "clipboard", "-o")
	cmd.Env = append(cmd.Env, fmt.Sprintf("DISPLAY=:%d", display))

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		// xclip returns error if clipboard is empty
		return "", nil
	}

	return out.String(), nil
}

// WriteClipboard writes the specified content to the X11 clipboard.
func WriteClipboard(display int, content string) error {
	log.Debug().Int("display", display).Int("len", len(content)).Msg("writing to X11 clipboard")
	cmd := exec.Command("xclip", "-selection", "clipboard", "-i")
	cmd.Env = append(cmd.Env, fmt.Sprintf("DISPLAY=:%d", display))
	cmd.Stdin = bytes.NewBufferString(content)

	return cmd.Run()
}

// SetLastContent updates the internal state of the clipboard watcher.
func SetLastContent(content string) {
	lastContentMu.Lock()
	defer lastContentMu.Unlock()
	lastContent = content
}

// WatchClipboard polls the clipboard for changes and calls the callback when it does.
func WatchClipboard(ctx context.Context, display int, callback func(string)) {
	log.Info().Int("display", display).Msg("starting clipboard watcher")

	initial, _ := ReadClipboard(display)
	SetLastContent(initial)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Int("display", display).Msg("stopping clipboard watcher")
			return
		case <-ticker.C:
			content, err := ReadClipboard(display)
			if err == nil {
				lastContentMu.Lock()
				isNew := content != lastContent && content != ""
				if isNew {
					lastContent = content
				}
				lastContentMu.Unlock()

				if isNew {
					log.Debug().Int("display", display).Int("len", len(content)).Msg("clipboard change detected")
					callback(content)
				}
			}
		}
	}
}
