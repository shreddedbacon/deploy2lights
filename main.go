package main

import (
	"context"
	"fmt"
	"log"
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
	lagoonAPI := "https://lagoon-api.apps.shreddedbacon.com/graphql"
	brightness := 255
	ledCount := 6
	stripType := "GBR"
	projectName := "ben"
	environmentName := "master"

	ls, err := lights.Setup(brightness, ledCount, stripType)
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
			ls.Wipe(lights.HexToColor("FF0000")) //red
			ls.Wipe(lights.HexToColor("00FF00")) //green
			ls.Wipe(lights.HexToColor("0000FF")) //blue
			ls.Wipe(lights.HexToColor("FF00FF")) //purple
			ls.Wipe(lights.HexToColor("FFFF00")) //yellow
			ls.Wipe(lights.HexToColor("00FFFF")) //cyan
			ls.Wipe(lights.HexToColor("EB8F34")) //orange
			ls.Wipe(lights.HexToColor("06BA90")) //teal
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
			fmt.Println(deployment, id.String())
			ls.Wipe(lights.HexToColor("00FF00")) //green
			ls.Wipe(lights.HexToColor("7BA832")) //lighter green
			ls.Wipe(lights.HexToColor("48D99F")) //teal green
			ls.Wipe(lights.HexToColor("00FF00")) //green
			ls.Wipe(lights.HexToColor("7BA832")) //lighter green
			ls.Wipe(lights.HexToColor("48D99F")) //teal green
			ls.Wipe(lights.HexToColor("06BA90")) //teal
		}
		time.Sleep(time.Second)
	}
}
