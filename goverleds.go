package main

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"reflect"

	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/api"
	"github.com/hybridgroup/gobot/platforms/firmata"
	"github.com/hybridgroup/gobot/platforms/gpio"
)

func main() {

	gbot := gobot.NewGobot()

	a := api.NewAPI(gbot)
	a.AddHandler(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q \n", html.EscapeString(r.URL.Path))
	})
	a.Debug()
	a.Start()

	firmataAdaptor := firmata.NewFirmataAdaptor("arduino", "/dev/tty.usbmodemfa131")
	button := gpio.NewButtonDriver(firmataAdaptor, "buttonOff", "10")

	robot := gobot.NewRobot("bot",
		[]gobot.Connection{firmataAdaptor},
		[]gobot.Device{button},
	)

	work := func() {
		gobot.On(button.Event("push"), func(data interface{}) {
			actionOnLeds(robot, "off")
		})
	}

	robot.Work = work

	robot.AddCommand("initLight", func(params map[string]interface{}) interface{} {
		var buffer bytes.Buffer
		for k, v := range params {
			led := gpio.NewLedDriver(firmataAdaptor, k, v.(string))
			robot.AddDevice(led)
			buffer.WriteString("Init Light ")
			buffer.WriteString(k)
			buffer.WriteString(v.(string))
		}
		return fmt.Sprintf("initLight: %s", buffer.String())
	})

	robot.AddCommand("led", func(params map[string]interface{}) interface{} {
		var buffer bytes.Buffer
		for ledName, action := range params {
			buffer.WriteString("led %s %s")
			buffer.WriteString(ledName)
			buffer.WriteString(action.(string))
			actionOnLed(robot.Device(ledName).(*gpio.LedDriver), action.(string))
		}
		return fmt.Sprintf("Led(s) %s aa", buffer.String())
	})

	robot.AddCommand("leds", func(params map[string]interface{}) interface{} {
		actionOnLeds(robot, params["action"].(string))
		return fmt.Sprintf("all leds action")
	})

	gbot.AddRobot(robot)
	gbot.Start()
}

func actionOnLeds(robot *gobot.Robot, action string) {
	robot.Devices().Each(func(device gobot.Device) {
		if reflect.TypeOf(device).String() == "*gpio.LedDriver" {
			actionOnLed(device.(*gpio.LedDriver), action)
		}
	})
}

func actionOnLed(led *gpio.LedDriver, action string) {
	if action == "on" {
		led.On()
	} else if action == "off" {
		led.Off()
	}
}
