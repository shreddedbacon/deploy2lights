package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stianeikeland/go-rpio"
	"github.com/uselagoon/deploy2lights/internal/lagoon"
	lclient "github.com/uselagoon/deploy2lights/internal/lagoon/client"
	"github.com/uselagoon/deploy2lights/internal/lights"
	"github.com/uselagoon/deploy2lights/internal/schema"
	"github.com/uselagoon/deploy2lights/internal/sshtoken"
)

func main() {
	var ledCount int
	var ledBrightness int
	var lagoonAPI, stripType, projectName, environmentName string

	flag.StringVar(&lagoonAPI, "lagoon-api", "https://lagoon-api.apps.shreddedbacon.com/graphql",
		"The lagoon API url.")
	flag.StringVar(&stripType, "led-strip-type", "RGB",
		"The color order of the LED strip.")
	flag.StringVar(&projectName, "project-name", "ben",
		"The lagoon project name.")
	flag.StringVar(&environmentName, "environment-name", "master",
		"The lagoon environment name.")

	flag.IntVar(&ledCount, "led-count", 6,
		"The total number of LEDs.")
	flag.IntVar(&ledBrightness, "led-brightness", 255,
		"The brightness max of the leds.")
	flag.Parse()

	stripType = getEnv("LED_STRIP_TYPE", stripType)

	ledCount = getEnvInt("LED_COUNT", ledCount)
	ledBrightness = getEnvInt("LED_BRIGHTNESS", ledBrightness)

	ls, err := lights.Setup(ledBrightness, ledCount, stripType)
	if err != nil {
		log.Fatal(err)
	}
	// startup animation, once this is complete, builds can start
	ls.Startup()

	fmt.Println("opening gpio")
	err = rpio.Open()
	if err != nil {
		panic(fmt.Sprint("unable to open gpio", err.Error()))
	}

	defer rpio.Close()

	pin := rpio.Pin(17)
	pin.Input()
	pin.PullUp()
	pin.Detect(rpio.FallEdge)

	defer pin.Detect(rpio.NoEdge)

	for {
		if pin.EdgeDetected() {
			fmt.Println("button pressed")
			// ls.Wipe(lights.HexToColor("FF0000")) //red
			// ls.Wipe(lights.HexToColor("00FF00")) //green
			// ls.Wipe(lights.HexToColor("0000FF")) //blue
			// ls.Wipe(lights.HexToColor("FF00FF")) //purple
			// ls.Wipe(lights.HexToColor("FFFF00")) //yellow
			// ls.Wipe(lights.HexToColor("00FFFF")) //cyan
			// ls.Wipe(lights.HexToColor("EB8F34")) //orange
			// ls.Wipe(lights.HexToColor("06BA90")) //teal
			ls.Wipe(lights.HexToColor("0000FF")) //blue
			ls.Wipe(lights.HexToColor("06BA90")) //teal
			ls.Wipe(lights.HexToColor("48D99F")) //teal green
			token, err := sshtoken.GetToken("/home/pi", "lagoon-ssh.apps.shreddedbacon.com", "32222")
			if err != nil {
				ls.Wipe(lights.HexToColor("FF0000")) //red
				ls.Wipe(lights.HexToColor("EB8F34")) //orange
				ls.Wipe(lights.HexToColor("FFFF00")) //yellow
				ls.Wipe(lights.HexToColor("FF0000")) //red
				ls.Wipe(lights.HexToColor("EB8F34")) //orange
				ls.Wipe(lights.HexToColor("FFFF00")) //yellow
				fmt.Println(err)
				time.Sleep(time.Second)
				ls.Wipe(lights.HexToColor("06BA90")) //teal
				continue
			}
			id := uuid.New()
			ctx := context.Background()
			deploy := &schema.DeployEnvironmentLatestInput{
				Environment: schema.EnvironmentInput{
					Name: environmentName,
					Project: schema.ProjectInput{
						Name: projectName,
					},
				},
				BulkID: id.String(),
			}
			l := lclient.New(lagoonAPI, token, "deploy2lights", false)
			deployment, err := lagoon.DeployLatest(ctx, deploy, l)
			if err != nil {
				ls.Wipe(lights.HexToColor("FF0000")) //red
				ls.Wipe(lights.HexToColor("EB8F34")) //orange
				ls.Wipe(lights.HexToColor("FFFF00")) //yellow
				ls.Wipe(lights.HexToColor("FF0000")) //red
				ls.Wipe(lights.HexToColor("EB8F34")) //orange
				ls.Wipe(lights.HexToColor("FFFF00")) //yellow
				fmt.Println(err)
				time.Sleep(time.Second)
				ls.Wipe(lights.HexToColor("06BA90")) //teal
				continue
			}
			ls.Wipe(lights.HexToColor("0000FF")) //blue
			ls.Wipe(lights.HexToColor("06BA90")) //teal
			ls.Wipe(lights.HexToColor("48D99F")) //teal green
			fmt.Println(deployment.DeployEnvironmentLatest, id.String())
			timeout := 1
			for timeout <= 50 {
				deployments, err := lagoon.GetDeploymentsByBulkID(ctx, id.String(), l)
				if err != nil {
					ls.Wipe(lights.HexToColor("FF0000")) //red
					ls.Wipe(lights.HexToColor("EB8F34")) //orange
					ls.Wipe(lights.HexToColor("FFFF00")) //yellow
					ls.Wipe(lights.HexToColor("FF0000")) //red
					ls.Wipe(lights.HexToColor("EB8F34")) //orange
					ls.Wipe(lights.HexToColor("FFFF00")) //yellow
					fmt.Println(err)
					time.Sleep(time.Second)
					ls.Wipe(lights.HexToColor("06BA90")) //teal
					break
				}
				breakout := false
				for _, deploy := range *deployments {
					if deploy.Name == deployment.DeployEnvironmentLatest {
						fmt.Println(deploy.Name, deploy.Status)
						switch deploy.Status {
						case "new":
							for j := 1; j <= 8; j++ {
								ls.Wipe(lights.HexToColor("6200ff")) //purple
								ls.Wipe(lights.HexToColor("a77bed")) //lighter purple
								ls.Wipe(lights.HexToColor("8249ab")) //lighter again purple
							}
						case "pending":
							for j := 1; j <= 8; j++ {
								ls.Wipe(lights.HexToColor("f542b6")) //pink
								ls.Wipe(lights.HexToColor("87095b")) //darker pink
								ls.Wipe(lights.HexToColor("e681c2")) //lighter pink
							}
						case "running":
							for j := 1; j <= 8; j++ {
								ls.Wipe(lights.HexToColor("00f7ff")) //light blue
								ls.Wipe(lights.HexToColor("027399")) //cyan blue
								ls.Wipe(lights.HexToColor("2d8385")) //teal blue
							}
						case "complete":
							for j := 1; j <= 8; j++ {
								ls.Wipe(lights.HexToColor("00FF00")) //green
								ls.Wipe(lights.HexToColor("7BA832")) //lighter green
								ls.Wipe(lights.HexToColor("BBFF00")) //teal green
							}
							breakout = true
						case "failed":
							for j := 1; j <= 8; j++ {
								ls.Wipe(lights.HexToColor("FF0000")) //red
								ls.Wipe(lights.HexToColor("EB8F34")) //orange
								ls.Wipe(lights.HexToColor("FFFF00")) //yellow
							}
							breakout = true
						case "cancelled":
							for j := 1; j <= 8; j++ {
								ls.Wipe(lights.HexToColor("FF0000")) //red
								ls.Wipe(lights.HexToColor("EB8F34")) //orange
								ls.Wipe(lights.HexToColor("FFFF00")) //yellow
							}
							breakout = true
						}
					}
				}
				if breakout {
					break
				}
				// time.Sleep(5 * time.Second) // sleep for 5 seconds, wait total 250 seconds (token will expire)
				timeout++
			}
			ls.Wipe(lights.HexToColor("06BA90")) //teal
		}
		time.Sleep(time.Second)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		valueInt, e := strconv.Atoi(value)
		if e == nil {
			return valueInt
		}
	}
	return fallback
}
