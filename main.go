package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nobonobo/t80nxbt/procon"
	"github.com/simulatedsimian/joystick"
)

var (
	deadzonePlus  = 0.0
	deadzoneMinus = 0.0
)

func gamma(v float64) float64 {
	const g = 1.5
	if v > 0 {
		return math.Pow(v, g)
	}
	return -1 * math.Pow(-v, g)
}

// normalize returns a value in the range [-1, 1].
// auto deadzone canceler
func normalize(v float64) float64 {
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

var VRallyMode = false

func bind(input *procon.Input, state joystick.State) {
	ps := state.Buttons&(1<<12) != 0
	input.LStick.XValue = int(gamma(normalize(float64(state.AxisData[0])/32767)) * 100) // Wheel
	input.LStick.YValue = state.AxisData[3]*50/32767 + 50                               // Brake
	if !ps {
		if !VRallyMode {
			input.RStick.YValue = state.AxisData[4]*50/32767 + 50 // Accel
		}
		// HAT switch to DPad
		input.DpadLeft = float32(state.AxisData[6])/32767 < 0
		input.DpadRight = float32(state.AxisData[6])/32767 > 0
		input.DpadUp = float32(state.AxisData[7])/32767 < 0
		input.DpadDown = float32(state.AxisData[7])/32767 > 0
	} else {
		// HAT switch to RStick
		input.RStick.XValue = int(float32(state.AxisData[6]) / 32767 * +100)
		input.RStick.YValue = int(float32(state.AxisData[7]) / 32767 * -100)
	}
	input.Y = state.Buttons&(1<<0) != 0
	input.B = state.Buttons&(1<<1) != 0
	input.A = state.Buttons&(1<<2) != 0
	input.X = state.Buttons&(1<<3) != 0
	input.L = !ps && state.Buttons&(1<<4) != 0
	input.R = !ps && state.Buttons&(1<<5) != 0
	input.Minus = !ps && state.Buttons&(1<<8) != 0
	input.Plus = !ps && state.Buttons&(1<<9) != 0
	input.Capture = ps && state.Buttons&(1<<8) != 0
	input.Home = ps && state.Buttons&(1<<9) != 0
	input.LStick.Pressed = state.Buttons&(1<<10) != 0
	input.RStick.Pressed = state.Buttons&(1<<11) != 0
	if !VRallyMode {
		input.Zl = !ps && state.Buttons&(1<<6) != 0
		input.Zr = !ps && state.Buttons&(1<<7) != 0
	}
	switch {
	case ps && state.Buttons&(1<<4) != 0:
		VRallyMode = false
	case ps && state.Buttons&(1<<5) != 0:
		VRallyMode = true
	}
}

var client = procon.New()

func connect(ctx context.Context, id int, script string) {
	deadzoneMinus = 0.0
	deadzonePlus = 0.0
	js, err := joystick.Open(id)
	if err != nil {
		log.Print(err)
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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
		defer js.Close()
		tick := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				_, err := os.Stat("/dev/input/js0")
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()
	var mutex sync.RWMutex
	input := procon.Input{}
	lastState := joystick.State{}
	go func() {
		for {
			state, err := js.Read()
			if err != nil {
				log.Print(err)
				return
			}
			mutex.Lock()
			lastState = state
			bind(&input, lastState)
			mutex.Unlock()
		}
	}()
	defer log.Println("closing...")
	const dt = time.Second / 60 / 20 // PWM 5 levels
	tick := time.NewTicker(time.Second / 60)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			pulse := int((time.Duration(time.Now().UnixNano()) % (100 * dt)) / dt)
			mutex.RLock()
			copyInput := input
			ps := lastState.Buttons&(1<<12) != 0
			brake := lastState.AxisData[3]*50/32767 + 50
			accel := lastState.AxisData[4]*50/32767 + 50
			mutex.RUnlock()
			if !ps {
				copyInput.Zl = copyInput.Zl || pulse < brake
				copyInput.Zr = copyInput.Zr || pulse < accel
			}
			fmt.Printf("Handle: %+4d Accel: %+4d Brake: %+4d Zl: %5v Zr: %5v\r",
				copyInput.LStick.XValue,
				copyInput.RStick.YValue,
				copyInput.LStick.YValue,
				copyInput.Zl, copyInput.Zr,
			)
			if err := client.Input(copyInput); err != nil {
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
