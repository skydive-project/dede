package main

import (
	"time"

	"github.com/tebeka/selenium"
)

func main() {
	caps := selenium.Capabilities{"browserName": "chrome"}
	webdriver, err := selenium.NewRemote(caps, "http://localhost:4444/wd/hub")
	if err != nil {
		panic(err)
	}
	webdriver.MaximizeWindow("")
	webdriver.SetAsyncScriptTimeout(5 * time.Second)
}
