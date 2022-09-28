package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stianeikeland/go-rpio"
	"github.com/uselagoon/deploy2lights/internal/lagoon"
	lclient "github.com/uselagoon/deploy2lights/internal/lagoon/client"
	"github.com/uselagoon/deploy2lights/internal/lights"
	"github.com/uselagoon/deploy2lights/internal/schema"
	"github.com/uselagoon/deploy2lights/internal/sshtoken"
)

func main() {
	lagoonAPI := "https://lagoon-api.apps.shreddedbacon.com/graphql"
	ls, err := lights.Setup(255, 6)
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
			ls.Wipe(0x00FF00)
			ls.Wipe(0x00FFFF)
			ls.Wipe(0x00FF00)
			ls.Wipe(0x00FFFF)
			ls.Wipe(0x0690BA)
			token, err := sshtoken.GetToken("/home/pi", "lagoon-ssh.apps.shreddedbacon.com", "32222")
			if err != nil {
				ls.Wipe(0xFF0000)
				ls.Wipe(0xEB8F34)
				ls.Wipe(0xFF0000)
				ls.Wipe(0xEB8F34)
				ls.Wipe(0x0690BA)
				fmt.Println(err)
				time.Sleep(time.Second)
				continue
			}
			ctx := context.Background()
			deploy := &schema.DeployEnvironmentLatestInput{
				Environment: schema.EnvironmentInput{
					Name: "master",
					Project: schema.ProjectInput{
						Name: "ben",
					},
				},
			}
			l := lclient.New(lagoonAPI, token, "deploy2lights", false)
			deployment, err := lagoon.DeployLatest(ctx, deploy, l)
			if err != nil {
				ls.Wipe(0xFF0000)
				ls.Wipe(0xEB8F34)
				ls.Wipe(0xFF0000)
				ls.Wipe(0xEB8F34)
				ls.Wipe(0x0690BA)
				fmt.Println(err)
				time.Sleep(time.Second)
				continue
			}
			fmt.Println(deployment)
			ls.Wipe(0x00FF00)
			ls.Wipe(0x00FFFF)
			ls.Wipe(0xEB8F34)
			ls.Wipe(0xFF0000)
			ls.Wipe(0x0690BA)
		}
		time.Sleep(time.Second)
	}
}
