package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nobonobo/t80nxbt/procon"
	"github.com/simulatedsimian/joystick"
)

var (
	deadzonePlus  = float32(0.0)
	deadzoneMinus = float32(0.0)
)

// normalize returns a value in the range [-1, 1].
// auto deadzone canceler
func normalize(v float32) float32 {
	if v == 0.0 {
		return 0.0
	}
	if v > 0.0 {
		if deadzonePlus == 0 {
			deadzonePlus = v
		} else {
			if v < deadzonePlus {
				deadzonePlus = v // minimum
			}
		}
		if v < deadzonePlus {
			return 0.0
		}
		return (v - deadzonePlus) / (1 - deadzonePlus)
	} else {
		if deadzoneMinus == 0 {
			deadzoneMinus = v
		} else {
			if v > deadzoneMinus {
				deadzoneMinus = v // maximum
			}
		}
		if v > deadzoneMinus {
			return 0.0
		}
		return (v - deadzoneMinus) / (1 + deadzoneMinus)
	}
}

func bind(input *procon.Input, state joystick.State) {
	input.LStick.XValue = int(normalize(float32(state.AxisData[0])/32767) * 100) // Wheel
	input.RStick.XValue = state.AxisData[3]*50/32767 + 50                        // Brake
	input.RStick.YValue = state.AxisData[4]*50/32767 + 50                        // Accel
	ps := state.Buttons&(1<<12) != 0
	input.Y = state.Buttons&(1<<0) != 0
	input.B = state.Buttons&(1<<1) != 0
	input.A = state.Buttons&(1<<2) != 0
	input.X = state.Buttons&(1<<3) != 0
	input.L = state.Buttons&(1<<4) != 0
	input.R = state.Buttons&(1<<5) != 0
	input.Minus = !ps && state.Buttons&(1<<8) != 0
	input.Plus = !ps && state.Buttons&(1<<9) != 0
	input.Zl = !ps && state.Buttons&(1<<10) != 0
	input.Zr = !ps && state.Buttons&(1<<11) != 0
	input.DpadLeft = float32(state.AxisData[6])/32767 < 0
	input.DpadRight = float32(state.AxisData[6])/32767 > 0
	input.DpadUp = float32(state.AxisData[7])/32767 < 0
	input.DpadDown = float32(state.AxisData[7])/32767 > 0
	input.Capture = ps && state.Buttons&(1<<8) != 0
	input.Home = ps && state.Buttons&(1<<9) != 0
	input.LStick.Pressed = ps && state.Buttons&(1<<10) != 0
	input.RStick.Pressed = ps && state.Buttons&(1<<11) != 0
}

var client = procon.New()

func connect(ctx context.Context, id int, script string) {
	deadzoneMinus = float32(0.0)
	deadzonePlus = float32(0.0)
	js, err := joystick.Open(id)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Printf("Joystick Name: %s", js.Name())
	fmt.Printf("   Axis Count: %d", js.AxisCount())
	fmt.Printf(" Button Count: %d", js.ButtonCount())
	fmt.Println()
	if err := client.Start(ctx, script); err != nil {
		log.Print(err)
		return
	}
	defer client.Stop()
	if err := client.Connect(); err != nil {
		log.Print(err)
		return
	}
	defer client.Disconnect()
	go func() {
		<-ctx.Done()
		js.Close()
	}()
	input := procon.Input{}
	defer log.Println("closing...")
	tick := time.NewTicker(16666 * time.Microsecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			state, err := js.Read()
			if err != nil {
				log.Print(err)
				return
			}
			bind(&input, state)
			fmt.Printf("Handle: %+4d Accel: %+4d Brake: %+4d\r",
				input.LStick.XValue, input.RStick.YValue, input.RStick.XValue)
			if err := client.Input(input); err != nil {
				log.Print(err)
				s, err := client.State()
				if err != nil {
					log.Print(err)
					return
				} else {
					log.Print(s)
				}
				if err := client.Connect(); err != nil {
					log.Print(err)
					return
				}
			}
		}
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	id := 0
	script := "./procon.py"
	flag.IntVar(&id, "id", 0, "Joystick ID")
	flag.StringVar(&script, "script", script, "stub script")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	timer := time.NewTimer(time.Second * 3)
	for {
		select {
		case <-ctx.Done():
			log.Println("terminated")
			return // exit
		case <-timer.C:
			timer.Stop()
			connect(ctx, id, script)
			timer.Reset(time.Second * 3)
		}
	}
}
