package main

import (
	"flag"
	"time"
	"fmt"
	"image"
	"github.com/disintegration/imaging"
)

func main() {
	// ----- flag parsing -----

	url := flag.String("url", "", "input URL")
	output := flag.String("output", "", "output image path")
	format := flag.Bool("strftime", false, "enable strftime formatting in URL")
	verbose := flag.Bool("verbose", false, "enable debug output")
	timezone := flag.String("timezone", "", "override timezone (tzinfo format)")
	xpath := flag.String("xpath", "", "xpath to <img> tag in url")
	test := flag.Bool("test", false, "disable wait-online and cooldown")
	mode := flag.String("mode", "fill", "image scaling mode (fill, center)")
	scale := flag.Float64("scale", 1, "scale image prior to centering")
	// top := flag.Int("top", 0, "crop from top")
	// left := flag.Int("left", 0, "crop from left")
	// right := flag.Int("right", 0, "crop from right")
	// bottom := flag.Int("bottom", 0, "crop from bottom")
	cooldown := flag.Int("cooldown", 3600, "minimum seconds to wait before attempting download again")
	flag.Parse()

	if *verbose {
		LOG_LEVEL = "debug"
	}

	// download/rescale image, then quit
	if *test {
		img, err := download(*url, *format, *timezone, *xpath)
		if err != nil {
			panic(err)
		}
		// img = adjust(img, *top, *left, *right, *bottom)
		img = adjust(img, *mode, *scale)
		imaging.Save(img, *output)
		debug("Image saved to ", *output)
	} else {
		time_last_success := time.Time{}

		online := make(chan int)
		go wait_online(online)

		// loop forever and wait for network online events
		for {
			// wait for network online message from wpa supplicant
			<- online
			debug("Network online")

			// FIXME - need to wait a few seconds for DNS?
			time.Sleep(5 * time.Second)

			var img image.Image
			// make sure we don't hammer server every time wifi is turned on
			if time.Now().Sub(time_last_success).Seconds() > float64(*cooldown) {
				var err error
				img, err = download(*url, *format, *timezone, *xpath)
				if err == nil {
					time_last_success = time.Now()
				} else {
					fmt.Println(err)
					continue
				}
			} else {
				debug("Hit cooldown limit")
				continue
			}

			// img = adjust(img, *top, *left, *right, *bottom)
			img = adjust(img, *mode, *scale)
			imaging.Save(img, *output)
			debug("Image saved to ", *output)
		}
	}
}
