package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type Config struct {
	TerminalKey     byte `json:"terminal_key"`
	RingLeftKey     byte `json:"ring_left_key"`
	RingRightKey    byte `json:"ring_right_key"`
	RingModeKey     byte `json:"ring_mode_key"`
	CloseWindowKey  byte `json:"close_window_key"`
	DmenuKey        byte `json:"dmenu_key"`
	ExitWmKey       byte `json:"exit_wm_key"`
}

type WindowManager struct {
	X                *xgb.Conn
	Root             xproto.Window
	ScreenWidth      uint16
	ScreenHeight     uint16
	Workspaces       [][]xproto.Window
	CurrentWorkspace int
	RingMode         bool
	CurrentIndex     int
	ZoomMode         bool
	Cfg              Config
	Gaps             int
	WmProtocols      xproto.Atom
	WmDeleteWindow   xproto.Atom
}

func loadConfig() Config {
	defaultCfg := Config{
		TerminalKey:     36,  // Enter
		RingLeftKey:     113, // Left
		RingRightKey:    114, // Right
		RingModeKey:     27,  // R
		CloseWindowKey:  38,  // Q
		DmenuKey:        40,  // D
		ExitWmKey:       58,  // M
	}
	file, err := os.Open("config.json")
	if err != nil {
		return defaultCfg
	}
	defer file.Close()
	var cfg Config
	_ = json.NewDecoder(file).Decode(&cfg)
	return cfg
}

func main() {
	X, err := xgb.NewConn()
	if err != nil {
		fmt.Printf("ÐžÑˆÐ¸Ð±ÐºÐ° Ð¿Ð¾Ð´ÐºÐ»ÑŽÑ‡ÐµÐ½Ð¸Ñ Ðº X: %v\n", err)
		return
	}
	defer X.Close()

	setup := xproto.Setup(X)
	screen := setup.DefaultScreen(X)

	protocolsReply, _ := xproto.InternAtom(X, false, uint16(len("WM_PROTOCOLS")), "WM_PROTOCOLS").Reply()
	deleteReply, _ := xproto.InternAtom(X, false, uint16(len("WM_DELETE_WINDOW")), "WM_DELETE_WINDOW").Reply()

	wm := &WindowManager{
		X:                X,
		Root:             screen.Root,
		ScreenWidth:      screen.WidthInPixels,
		ScreenHeight:     screen.HeightInPixels,
		Workspaces:       make([][]xproto.Window, 9),
		CurrentWorkspace: 0,
		RingMode:         false,
		CurrentIndex:     0,
		ZoomMode:         false,
		Cfg:              loadConfig(),
		Gaps:             5,
		WmProtocols:      protocolsReply.Atom,
		WmDeleteWindow:   deleteReply.Atom,
	}

	for i := 0; i < 9; i++ {
		wm.Workspaces[i] = make([]xproto.Window, 0)
	}

	xproto.ChangeWindowAttributes(X, wm.Root, xproto.CwEventMask, []uint32{
		xproto.EventMaskSubstructureRedirect | xproto.EventMaskSubstructureNotify,
	})

	wm.grabKey(wm.Cfg.TerminalKey, xproto.ModMask4)
	wm.grabKey(wm.Cfg.RingLeftKey, xproto.ModMask4)
	wm.grabKey(wm.Cfg.RingRightKey, xproto.ModMask4)
	wm.grabKey(wm.Cfg.RingModeKey, xproto.ModMask4)
	wm.grabKey(wm.Cfg.CloseWindowKey, xproto.ModMask4)
	wm.grabKey(wm.Cfg.DmenuKey, xproto.ModMask4)
	wm.grabKey(wm.Cfg.ExitWmKey, xproto.ModMask4|xproto.ModMaskShift)

	wm.grabKey(

111, xproto.ModMask4) // Up
	wm.grabKey(116, xproto.ModMask4) // Down

	for i := byte(10); i <= 18; i++ {
		wm.grabKey(i, xproto.ModMask4)
	}

	fmt.Println("Ð‘Ð¾ÐµÐ²Ð¾Ð¹ ringwm Ð³Ð¾Ñ‚Ð¾Ð² Ðº ÑÐ±Ð¾Ñ€ÐºÐµ!")

	for {
		ev, err := X.WaitForEvent()
		if err != nil {
			continue
		}
		if ev == nil {
			break
		}

		switch e := ev.(type) {
		case xproto.MapRequestEvent:
			wm.manageWindow(e.Window)
		case xproto.UnmapNotifyEvent:
			wm.unmanageWindow(e.Window)
		case xproto.DestroyNotifyEvent:
			wm.unmanageWindow(e.Window)
		case xproto.EnterNotifyEvent:
			if !wm.RingMode {
				xproto.SetInputFocus(wm.X, xproto.InputFocusPointerRoot, e.Event, xproto.TimeCurrentTime)
			}
		case xproto.KeyPressEvent:
			if shouldQuit := wm.handleKeyPress(e.Detail, e.State); shouldQuit {
				fmt.Println("Ð’Ñ‹Ñ…Ð¾Ð´...")
				return
			}
		}
	}
}

func (wm *WindowManager) grabKey(keycode byte, modifiers uint16) {
	xproto.GrabKey(wm.X, true, wm.Root, modifiers, xproto.Keycode(keycode),
		xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (wm *WindowManager) getActiveWindows() []xproto.Window {
	return wm.Workspaces[wm.CurrentWorkspace]
}

func (wm *WindowManager) setActiveWindows(wins []xproto.Window) {
	wm.Workspaces[wm.CurrentWorkspace] = wins
}

func (wm *WindowManager) handleKeyPress(keycode xproto.Keycode, state uint16) bool {
	code := byte(keycode)

	if code == wm.Cfg.ExitWmKey && (state&xproto.ModMask4) != 0 && (state&xproto.ModMaskShift) != 0 {
		return true
	}

	if code >= 10 && code <= 18 && (state&xproto.ModMaskShift) == 0 {
		targetWorkspace := int(code - 10)
		if targetWorkspace != wm.CurrentWorkspace {
			for _, win := range wm.getActiveWindows() {
				xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX, []uint32{uint32(wm.ScreenWidth + 2000)})
			}
			wm.CurrentWorkspace = targetWorkspace
			wm.RingMode = false
			wm.ZoomMode = false
			wm.arrangeWindows()
		}
		return false
	}

	switch code {
	case wm.Cfg.TerminalKey:
		cmd := exec.Command("alacritty")
		cmd.Env = []string{
			"DISPLAY=" + os.Getenv("DISPLAY"),
			"HOME=" + os.Getenv("HOME"),
			"PATH=" + os.Getenv("PATH"),
			"DBUS_SESSION_BUS_ADDRESS=",
			"XDG_RUNTIME_DIR=",
		}
		_ = cmd.Start()

	case wm.Cfg.DmenuKey:
		cmd := exec.Command("dmenu_run")
		cmd.Env = []string{"DISPLAY=" + os.Getenv("DISPLAY"), "PATH=" + os.Getenv("PATH")}
		_ = cmd.Start()

	case wm.Cfg.CloseWindowKey:
		wm.killActiveWindow()

	case wm.Cfg.RingModeKey:
		wins := wm.getActiveWindows()
		if len(wins) > 0 {
			wm.RingMode = !wm.RingMode
			wm.ZoomMode = false
			if wm.RingMode {
				wm.CurrentIndex = len(wins) - 1
			}
			wm.arrangeWindows()
		}

	case wm.Cfg.RingRightKey:
		wins := wm.getActiveWindows()
		if wm.RingMode && len(wins) > 0 {
			wm.CurrentIndex = (wm.CurrentIndex + 1) % len(wins)
			wm.ZoomMode = false
			wm.arrangeWindows()
		}

	case wm.Cfg.RingLeftKey:
		wins := wm.getActiveWindows()
		if wm.RingMode && len(wins) > 0 {
			wm.CurrentIndex = (wm.CurrentIndex - 1 + len(wins)) % len(wins)
			wm.ZoomMode = false
			wm.arrangeWindows()
		}

	case 111: // Win + Up
		if wm.RingMode {
			wm.ZoomMode = true
			wm.arrangeWindows()
		}

	case 116: // Win + Down
		if wm.RingMode {
			wm.ZoomMode = false
			wm.arrangeWindows()
		}
	}
	return false
}

func (wm *WindowManager) killActiveWindow() {
	wins := wm.getActiveWindows()
	if len(wins) == 0 {
		return
	}
	var target xproto.Window
	if wm.RingMode {
		target = wins[wm.CurrentIndex]
	} else {
		target = wins[len(wins)-1]
	}

	var data [5]uint32
	data[0] = uint32(wm.WmDeleteWindow)
	data[1] = uint32(xproto.TimeCurrentTime)

	buf := make([]byte, 32)
	buf[0] = 33
	buf[1] = 32

	buf[4] = byte(target & 0xFF)
	buf[5] = byte((target >> 8) & 0xFF)
	buf[6] = byte((target >> 16) & 0xFF)
	buf[7] = byte((target >> 24) & 0xFF)

	buf[8] = byte(wm.WmProtocols & 0xFF)
	buf[9] = byte((wm.WmProtocols >> 8) & 0xFF)
	buf[10] = byte((wm.WmProtocols >> 16) & 0xFF)
	buf[11] = byte((wm.WmProtocols >> 24) & 0xFF)

	for i, val := range data {
		base := 12 + (i * 4)
		buf[base] = byte(val & 0xFF)
		buf[base+1] = byte((val >> 8) & 0xFF)
		buf[base+2] = byte((val >> 16) & 0xFF)
		buf[base+3] = byte((val >> 24) & 0xFF)
	}

	_ = xproto.SendEventChecked(wm.X, false, target, 0, string(buf))
}

func (wm *WindowManager) manageWindow(win xproto.Window) {
	wins := wm.getActiveWindows()
	for _, w := range wins {
		if w == win {
			return
		}
	}
	attrs, _ := xproto.GetWindowAttributes(wm.X, win).Reply()
	if attrs != nil && attrs.OverrideRedirect {
		xproto.MapWindow(wm.X, win)
		return
	}

	xproto.ChangeWindowAttributes(wm.X, win, xproto.CwEventMask, []uint32{xproto.EventMaskEnterWindow})

	wins = append(wins, win)
	wm.setActiveWindows(wins)
	xproto.MapWindow(wm.X, win)
	wm.CurrentIndex = len(wins) - 1
	wm.arrangeWindows()
}

func (wm *WindowManager) unmanageWindow(win xproto.Window) {
	for idx := 0; idx < 9; idx++ {
		wins := wm.Workspaces[idx]
		for i, w := range wins {
			if w == win {
				wins = append(wins[:i], wins[i+1:]...)
				wm.Workspaces[idx] = wins
				if idx == wm.CurrentWorkspace {
					if len(wins) > 0 {
						wm.CurrentIndex = wm.CurrentIndex % len(wins)
					} else {
						wm.CurrentIndex = 0
						wm.RingMode = false
						wm.ZoomMode = false
					}
					wm.arrangeWindows()
				}
				return
			}
		}
	}
}

type RenderWindow struct {
	ID      xproto.Window
	X, Y    int
	W, H    int
	ZDepth  float64
	IsFocus bool
}

func (wm *WindowManager) arrangeWindows() {
	wins := wm.getActiveWindows()
	n := len(wins)
	if n == 0 {
		xproto.ClearArea(wm.X, false, wm.Root, 0, 0, wm.ScreenWidth, wm.ScreenHeight)
		return
	}

	cx := int(wm.ScreenWidth) / 2
	cy := int(wm.ScreenHeight) / 2
	g := uint32(wm.Gaps)

	if wm.RingMode {
		if wm.ZoomMode {
			targetWin := wins[wm.CurrentIndex]
			zw := int(float64(wm.ScreenWidth) * 0.85)
			zh := int(float64(wm.ScreenHeight) * 0.85)
			zx := cx - (zw / 2)
			zy := cy - (zh / 2)
			
			xproto.ConfigureWindow(wm.X, targetWin, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight|xproto.ConfigWindowStackMode,
				[]uint32{uint32(zx), uint32(zy), uint32(zw), uint32(zh), xproto.StackModeAbove})
			xproto.SetInputFocus(wm.X, xproto.InputFocusPointerRoot, targetWin, xproto.TimeCurrentTime)
			return
		}

		radiusX := int(float64(wm.ScreenWidth) * 0.35)
		radiusY := int(float64(wm.ScreenHeight) * 0.22)

		renderQueue := make([]RenderWindow, n)

		for i, win := range wins {
			offsetIndex := i - wm.CurrentIndex
			angle := (2.0 * math.Pi * float64(offsetIndex)) / float64(n) + (math.Pi / 2.0)

			zDepth := math.Sin(angle)
			scale := 0.45 + (zDepth+1.0)*0.25

			winW := int(float64(wm.ScreenWidth) * 0.45 * scale)
			winH := int(float64(wm.ScreenHeight) * 0.45 * scale)

			x := int(float64(cx) + float64(radiusX)*math.Cos(angle) - float64(winW/2))
			y := int(float64(cy) + float64(radiusY)*math.Sin(angle) - float64(winH/2))

			renderQueue[i] = RenderWindow{
				ID:      win,
				X:       x,
				Y:       y,
				W:       winW,
				H:       winH,
				ZDepth:  zDepth,
				IsFocus: i == wm.CurrentIndex,
			}
		}

		sort.Slice(renderQueue, func(i, j int) bool {
			return renderQueue[i].ZDepth < renderQueue[j].ZDepth
		})

		for _, rw := range renderQueue {
xproto.ConfigureWindow(wm.X, rw.ID, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight|xproto.ConfigWindowStackMode,
[]uint32{uint32(rw.X), uint32(rw.Y), uint32(rw.W), uint32(rw.H), xproto.StackModeAbove})
if rw.IsFocus {
xproto.SetInputFocus(wm.X, xproto.InputFocusPointerRoot, rw.ID, xproto.TimeCurrentTime)
}
}
} else {
sw := int(wm.ScreenWidth)
sh := int(wm.ScreenHeight)
var visible []xproto.Window
if n > 4 {
visible = wins[n-4:]
for i := 0; i < n-4; i++ {
xproto.ConfigureWindow(wm.X, wins[i], xproto.ConfigWindowX, []uint32{uint32(sw + 2000)})
}
} else {
visible = wins
}
count := len(visible)
halfW := uint32(sw / 2)
halfH := uint32(sh / 2)
usw := uint32(sw)
ush := uint32(sh)
for i, win := range visible {
xproto.MapWindow(wm.X, win)
switch count {
case 1:
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{g, g, usw - (2 * g), ush - (2 * g)})
case 2:
if i == 0 {
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{g, g, halfW - g - (g / 2), ush - (2 * g)})
} else {
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{halfW + (g / 2), g, halfW - g - (g / 2), ush - (2 * g)})
}
case 3:
if i == 0 {
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{g, g, halfW - g - (g / 2), ush - (2 * g)})
} else if i == 1 {
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{halfW + (g / 2), g, halfW - g - (g / 2), halfH - g - (g / 2)})
} else {
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{halfW + (g / 2), halfH + (g / 2), halfW - g - (g / 2), halfH - g - (g / 2)})
}
case 4:
col := uint32(i % 2)
row := uint32(i / 2)
xPos := g
if col == 1 { xPos = halfW + (g / 2) }
yPos := g
if row == 1 { yPos = halfH + (g / 2) }
xproto.ConfigureWindow(wm.X, win, xproto.ConfigWindowX|xproto.ConfigWindowY|xproto.ConfigWindowWidth|xproto.ConfigWindowHeight,
[]uint32{xPos, yPos, halfW - g - (g / 2), halfH - g - (g / 2)})
}
}
if count > 0 {
xproto.SetInputFocus(wm.X, xproto.InputFocusPointerRoot, visible[count-1], xproto.TimeCurrentTime)
}
}
}
