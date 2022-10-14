package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio"

	"github.com/uselagoon/deploy2lights/internal/lights"

	"github.com/shreddedbacon/machinery/api/lagoon"
	lclient "github.com/shreddedbacon/machinery/api/lagoon/client"
	"github.com/shreddedbacon/machinery/api/schema"
	"github.com/shreddedbacon/machinery/utils/lagoon/sshtoken"

	"github.com/mattes/go-asciibot"
)

func main() {
	var ledCount int
	var ledBrightness int
	var lagoonAPI, stripType, projectName, environmentName string
	var sshHost, sshPort, sshKey string
	var displayDuration int

	flag.StringVar(&lagoonAPI, "lagoon-api", "https://lagoon-api.apps.shreddedbacon.com/graphql",
		"The lagoon API url.")
	flag.StringVar(&stripType, "led-strip-type", "RGB",
		"The color order of the LED strip.")
	flag.StringVar(&projectName, "project-name", "ben",
		"The lagoon project name.")
	flag.StringVar(&environmentName, "environment-name", "master",
		"The lagoon environment name.")
	flag.IntVar(&displayDuration, "led-duration-interval", 1,
		"The duration between render intervals in ms.")

	flag.IntVar(&ledCount, "led-count", 6,
		"The total number of LEDs.")
	flag.IntVar(&ledBrightness, "led-brightness", 255,
		"The brightness max of the leds.")
	flag.Parse()

	stripType = getEnv("LED_STRIP_TYPE", stripType)

	ledCount = getEnvInt("LED_COUNT", ledCount)
	ledBrightness = getEnvInt("LED_BRIGHTNESS", ledBrightness)

	lagoonAPI = getEnv("LAGOON_API", lagoonAPI)
	sshHost = getEnv("LAGOON_SSHHOST", sshHost)
	sshPort = getEnv("LAGOON_SSHPORT", sshPort)
	projectName = getEnv("LAGOON_PROJECT", projectName)
	environmentName = getEnv("LAGOON_ENVIRONMENT", environmentName)
	sshKey = getEnv("SSH_KEY", "/home/pi/.ssh/id_rsa")

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

	builds := &[]string{}

	go func() {
		for {
			if len(*builds) > 0 {
				for _, c := range *builds {
					ls.Wipe(lights.HexToColor(c))
				}
			}
		}
	}()

	ls.Wipe(lights.HexToColor("06BA90")) //teal

	for {
		if pin.EdgeDetected() {
			fmt.Println("button pressed")
			builds = &[]string{"0000FF", "06BA90", "48D99F"}
			token := ""
			err = sshtoken.ValidateOrRefreshToken(sshKey, sshHost, sshPort, &token)
			if err != nil {
				fmt.Println("generate token error:", err)
				//red orange yellow
				builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
				time.Sleep(time.Second * 5)
				ls.Wipe(lights.HexToColor("06BA90")) //teal
				continue
			}
			ctx := context.Background()
			deploy := &schema.DeployEnvironmentLatestInput{
				Environment: schema.EnvironmentInput{
					Name: environmentName,
					Project: schema.ProjectInput{
						Name: projectName,
					},
				},
				BuildVariables: []schema.EnvKeyValueInput{
					{
						Name:  "LAGOON_BUILD_NAME",
						Value: base64.StdEncoding.EncodeToString([]byte(asciibot.Random())),
					},
				},
				ReturnData: true,
			}
			l := lclient.New(lagoonAPI, "deploy2lights", &token, false)
			project, err := lagoon.GetMinimalProjectByName(ctx, projectName, l)
			if err != nil {
				fmt.Println("project get error:", err)
				//red orange yellow
				builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
				time.Sleep(time.Second * 5)
				ls.Wipe(lights.HexToColor("06BA90")) //teal
				continue
			}
			deployment, err := lagoon.DeployLatest(ctx, deploy, l)
			if err != nil {
				fmt.Println("deploy error:", err)
				//red orange yellow
				builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
				time.Sleep(time.Second * 5)
				ls.Wipe(lights.HexToColor("06BA90")) //teal
				continue
			}
			fmt.Println("started", deployment.DeployEnvironmentLatest, project.Name, project.ID)
			timeout := 1
			for timeout <= 600 {
				err := sshtoken.ValidateOrRefreshToken(sshKey, sshHost, sshPort, &token)
				if err != nil {
					fmt.Println("token validation error:", err)
					//red orange yellow
					builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
					time.Sleep(time.Second * 5)
					ls.Wipe(lights.HexToColor("06BA90")) //teal
					break
				}
				environment, err := lagoon.GetDeploymentsByEnvironment(ctx, project.ID, environmentName, l)
				if err != nil {
					fmt.Println("list deploy error:", err)
					//red orange yellow
					builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
					time.Sleep(time.Second * 5)
					ls.Wipe(lights.HexToColor("06BA90")) //teal
					break
				}
				breakout := false
				for _, deploy := range environment.Deployments {
					if deploy.Name == deployment.DeployEnvironmentLatest {
						fmt.Println(deploy.Name, deploy.Status)
						// wipeCount := 4
						switch deploy.Status {
						case "new":
							//purple, lighter purple, lighter again purple
							// builds = &[]string{"6200FF", "A77BED", "8249AB"}
							//purple, darker purple, lighter purple
							builds = &[]string{"460ba3", "391f61", "925ee0"}
							time.Sleep(time.Second * 5)
						case "pending":
							//pink, darker pink, lighter pink
							// builds = &[]string{"F542B6", "87095B", "E681C2"}
							//pink, darker pink, lighter pink
							builds = &[]string{"9e0e6b", "6e2b56", "e36bb8"}
							time.Sleep(time.Second * 5)
						case "running":
							//light blue, cyan blue, teal blue
							// builds = &[]string{"00F7FF", "027399", "2D8385"}
							//teal, darker teal, lighter teal
							builds = &[]string{"06BA90", "2C7362", "67E0C3"}
							time.Sleep(time.Second * 5)
						case "complete":
							//green, lighter green, teal green
							builds = &[]string{"00FF00", "1a6b1a", "4ddb4d"}
							time.Sleep(time.Second * 10)
							breakout = true
						case "failed":
							//red, orange, yellow
							builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
							time.Sleep(time.Second * 10)
							breakout = true
						case "cancelled":
							//red, orange, yellow
							builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
							time.Sleep(time.Second * 10)
							breakout = true
						}
					}
				}
				if breakout {
					builds = &[]string{}
					ls.Wipe(lights.HexToColor("06BA90")) //teal
					break
				}
				timeout++
			}
			builds = &[]string{}
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
