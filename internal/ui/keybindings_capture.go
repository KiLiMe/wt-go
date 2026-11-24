package ui

import (
	myui "github.com/egoist/mygo/ui"
)

// Boundary: recording one action's keys -- what the page listens for while a
// row is armed, and how a key heard becomes a combination. The row that arms it
// is keybindings.go.

// Capture holds the keys the actions page listens for while an action is being
// recorded. The window hears keys as shortcuts only, so recording one means
// listening for all 1,152 of them, sixteen modifier combinations of
// seventy-two keys, and a window keeps the shortcuts of the frame that declared
// them: the keys have to be declared again on every frame, but making them,
// which measured 1.39 ms and 59,248 allocations a frame, is done once.
type Capture struct {
	// act and action are the page's, set on every frame: the action a key
	// recorded now belongs to, and "" while none is being recorded.
	act    Actions
	action string
	// keys are the shortcuts to declare, and fns what each one records.
	keys []captureKey
	fns  []func()
}

// captureKey is one key a recording listens for: how the window hears it, and
// the name a keystroke on it gives the action.
type captureKey struct {
	mods myui.Modifiers
	key  myui.Key
	name string
}

// capture listens for the key the action being recorded takes. The window
// hears keys as shortcuts only, so it listens for every key there is, with
// every combination of the modifiers it knows.
func capture(c *myui.Context, keys *Capture, act Actions, action string) {
	if keys == nil {
		// A page drawn without a Capture keeps nothing past its frame, as a
		// page drawn without KeysCache keeps no answer to the file.
		keys = &Capture{}
	}
	keys.listen(c, act, action)
}

// listen declares the keys of a recording, making them on the first frame.
func (k *Capture) listen(c *myui.Context, act Actions, action string) {
	k.ready(act, action)
	for i := range k.keys {
		c.OnWindowShortcut(k.keys[i].mods, k.keys[i].key, k.fns[i])
	}
}

// ready says which action a key recorded now belongs to, making the keys when
// the frames before it did not.
func (k *Capture) ready(act Actions, action string) {
	k.act, k.action = act, action
	if k.fns != nil {
		return
	}
	seen := make(map[captureKey]bool)
	for _, m := range captureMods() {
		for _, n := range KeyNames() {
			name, ok := FormatKey(m, n.Key)
			if !ok {
				continue
			}
			// KeyNames lists a key more than once where the keyboard has two
			// of them under one name, as the plus does: the window hears them
			// as one shortcut.
			one := captureKey{mods: m, key: n.Key, name: name}
			if seen[one] {
				continue
			}
			seen[one] = true
			k.keys = append(k.keys, one)
			k.fns = append(k.fns, func() {
				// The action is the one being recorded when the key is
				// pressed, not the one the keys were made for.
				if k.action == "" {
					return
				}
				k.act.SetKeybinding(k.action, name)
			})
		}
	}
}

// captureMods are the modifiers capture listens for: every combination of the
// modifiers it records, with and without Super.
func captureMods() []myui.Modifiers {
	base := []myui.Modifiers{
		0,
		myui.Ctrl,
		myui.Shift,
		myui.Alt,
		myui.Ctrl | myui.Shift,
		myui.Ctrl | myui.Alt,
		myui.Shift | myui.Alt,
		myui.Ctrl | myui.Shift | myui.Alt,
	}
	mods := make([]myui.Modifiers, 0, len(base)*2)
	mods = append(mods, base...)
	for _, m := range base {
		mods = append(mods, m|myui.Super)
	}
	return mods
}

const (
	keyRestoreDefaults = "restoreDefaults"
	keySearchActions   = "keySearch"
	keyOnlyChanged     = "onlyChanged"
)
