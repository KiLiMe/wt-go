package ui

import (
	"testing"

	myui "github.com/egoist/mygo/ui"
)

// TestChoice covers a control over values the settings hold: it opens on the
// label of the value in force, writes the value a chosen label stands for, and
// shows a value no option names as it is.
func TestChoice(t *testing.T) {
	shape := "block"
	family := "Cascadia Mono,monospace"
	tt := myui.NewTester(func(c *myui.Context) {
		myui.Form(c, func() {
			field(c, "Cursor", "", func() {
				choice(c, &shape, []option{
					{"bar", "Bar"},
					{"block", "Block"},
					{"underline", "Underline"},
				})
			})
			field(c, "Font", "", func() {
				choice(c, &family, []option{{"monospace", "monospace"}})
			})
		})
	}, 400, 200)

	if !tt.HasText("Block") {
		t.Fatalf("the control does not open on the label of its value: %q", tt.Texts())
	}
	if !tt.HasText("Cascadia Mono,monospace") {
		t.Fatalf("a value no option names is not shown as it is: %q", tt.Texts())
	}
	if err := tt.Click("Block"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Underline"); err != nil {
		t.Fatal(err)
	}
	if shape != "underline" {
		t.Fatalf("chose Underline, value is %q", shape)
	}
}

// TestCount covers a control over a whole number: it steps by one and stops at
// the bounds.
func TestCount(t *testing.T) {
	n := 1
	tt := myui.NewTester(func(c *myui.Context) {
		myui.Form(c, func() {
			field(c, "Panes", "", func() { count(c, &n, 0, 2) })
		})
	}, 400, 120)

	for i := 0; i < 4; i++ {
		if err := tt.Click("Increase"); err != nil {
			t.Fatal(err)
		}
	}
	if n != 2 {
		t.Fatalf("stepped past the bound: %d", n)
	}
	if err := tt.Click("Decrease"); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("did not step down: %d", n)
	}
}

// TestMeasureStep covers a measurement: a float32 step steps by what it was
// meant to be, not by the noise of its type, which would also ask the control
// for seventeen decimals.
func TestMeasureStep(t *testing.T) {
	line := float32(1.2)
	tt := myui.NewTester(func(c *myui.Context) {
		myui.Form(c, func() {
			field(c, "Line height", "", func() { measure(c, &line, 0.8, 2.4, 0.1) })
		})
	}, 400, 120)

	if err := tt.Click("Increase"); err != nil {
		t.Fatal(err)
	}
	if line != float32(1.3) {
		t.Fatalf("stepped to %v, want 1.3", line)
	}
}

// TestFieldDescription covers a field: what it does shows under its control,
// where the room of the form does not cut it.
func TestFieldDescription(t *testing.T) {
	name := ""
	tt := myui.NewTester(func(c *myui.Context) {
		myui.Form(c, func() {
			field(c, "Name", "The program the profile runs; empty runs the shell the system starts.", func() {
				myui.TextInput(c, &name)
			})
		})
	}, 400, 200)

	if !tt.HasText("The program the profile runs; empty runs the shell the system starts.") {
		t.Fatalf("the description is not under the control: %q", tt.Texts())
	}
}
