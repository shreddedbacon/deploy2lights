package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/uselagoon/deploy2lights/internal/lights"
	"github.com/uselagoon/deploy2lights/internal/oled"

	"github.com/uselagoon/machinery/api/lagoon"
	lclient "github.com/uselagoon/machinery/api/lagoon/client"
	"github.com/uselagoon/machinery/api/schema"
	"github.com/uselagoon/machinery/utils/sshtoken"

	"github.com/mattes/go-asciibot"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
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
	// err = rpio.Open()
	// if err != nil {
	// 	log.Fatal(fmt.Sprint("unable to open gpio", err.Error()))
	// }

	// defer rpio.Close()

	// pin := rpio.Pin(17)
	// pin.Input()
	// pin.PullUp()
	// pin.Detect(rpio.FallEdge)

	// defer pin.Detect(rpio.NoEdge)

	pin := gpioreg.ByName("GPIO17")
	if pin == nil {
		log.Fatal("Failed to find GPIO17")
	}
	if err := pin.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		log.Fatal(err)
	}

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	spiPort, err := spireg.Open("") // spireg.Open(fmt.Sprintf("/dev/spidev0.%d", index))
	if err != nil {
		log.Fatal(err)
	}
	defer spiPort.Close()
	dc := gpioreg.ByName("GPIO23")
	if dc == nil {
		log.Fatal("Failed to find GPIO23")
	}
	rst := gpioreg.ByName("GPIO24")
	if rst == nil {
		log.Fatal("Failed to find GPIO24")
	}

	disp := oled.NewDisplay(128, 64, spiPort, dc, rst)
	disp.PrintLogo()

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

	for pin.WaitForEdge(-1) {
		// if pin.EdgeDetected() {
		fmt.Println("button pressed")
		disp.Clear()
		disp.DrawLine(oled.TextRow1)
		disp.DrawTextCenter("BUILD", oled.TextRow3)
		disp.DrawTextCenter("TRIGGERED", oled.TextRow4)
		disp.DrawLine(oled.TextRow5)
		builds = &[]string{"0000FF", "06BA90", "48D99F"}
		token := ""
		err = sshtoken.ValidateOrRefreshToken(sshKey, sshHost, sshPort, &token)
		if err != nil {
			disp.DrawLine(oled.TextRow1)
			disp.DrawTextCenter("FAILED TO AUTH", oled.TextRow3)
			disp.DrawLine(oled.TextRow5)
			fmt.Println("generate token error:", err)
			//red orange yellow
			builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
			time.Sleep(time.Second * 5)
			continue
		}
		ctx := context.Background()
		buildRobot := base64.StdEncoding.EncodeToString([]byte(asciibot.Random()))
		deploy := &schema.DeployEnvironmentLatestInput{
			Environment: schema.EnvironmentInput{
				Name: environmentName,
				Project: schema.ProjectInput{
					Name: projectName,
				},
			},
			BuildVariables: []schema.EnvKeyValueInput{
				{
					Name:  "BUILD_ROBOT",
					Value: buildRobot,
				},
			},
			ReturnData: true,
		}
		l := lclient.New(lagoonAPI, "deploy2lights", &token, false)
		project, err := lagoon.GetMinimalProjectByName(ctx, projectName, l)
		if err != nil {
			disp.DrawLine(oled.TextRow1)
			disp.DrawTextCenter("PROJECT GET FAILED", oled.TextRow3)
			disp.DrawLine(oled.TextRow5)
			fmt.Println("project get error:", err)
			//red orange yellow
			builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
			time.Sleep(time.Second * 5)
			continue
		}
		deployment, err := lagoon.DeployLatest(ctx, deploy, l)
		if err != nil {
			disp.DrawLine(oled.TextRow1)
			disp.DrawTextCenter("FAILED TO DEPLOY", oled.TextRow3)
			disp.DrawLine(oled.TextRow5)
			fmt.Println("deploy error:", err)
			//red orange yellow
			builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
			time.Sleep(time.Second * 5)
			continue
		}
		fmt.Println("started", deployment.DeployEnvironmentLatest, project.Name, project.ID)
		timeout := 1
		for timeout <= 600 {
			err := sshtoken.ValidateOrRefreshToken(sshKey, sshHost, sshPort, &token)
			if err != nil {
				disp.DrawLine(oled.TextRow1)
				disp.DrawTextCenter("FAILED TO DEPLOY", oled.TextRow3)
				disp.DrawLine(oled.TextRow5)
				fmt.Println("token validation error:", err)
				//red orange yellow
				builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
				time.Sleep(time.Second * 5)
				break
			}
			environment, err := lagoon.GetDeploymentsByEnvironment(ctx, project.ID, environmentName, l)
			if err != nil {
				disp.DrawLine(oled.TextRow1)
				disp.DrawTextCenter("FAILED TO DEPLOY", oled.TextRow3)
				disp.DrawLine(oled.TextRow5)
				fmt.Println("list deploy error:", err)
				//red orange yellow
				builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
				time.Sleep(time.Second * 5)
				break
			}
			breakout := false
			for _, deploy := range environment.Deployments {
				if deploy.Name == deployment.DeployEnvironmentLatest {
					fmt.Println(deploy.Name, deploy.Status)
					// wipeCount := 4
					switch deploy.Status {
					case "new":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: New", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//purple, darker purple, lighter purple
						builds = &[]string{"460ba3", "391f61", "925ee0"}
						time.Sleep(time.Second * 5)
					case "queued":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: Queued", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//pink, darker pink, lighter pink
						builds = &[]string{"9e0e6b", "6e2b56", "e36bb8"}
						time.Sleep(time.Second * 5)
					case "pending":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: Pending", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//pink, darker pink, lighter pink
						builds = &[]string{"9e0e6b", "6e2b56", "e36bb8"}
						time.Sleep(time.Second * 5)
					case "running":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: Running", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//teal, darker teal, lighter teal
						builds = &[]string{"06BA90", "2C7362", "67E0C3"}
						time.Sleep(time.Second * 5)
					case "complete":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: Complete", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//green, lighter green, teal green
						builds = &[]string{"00FF00", "1a6b1a", "4ddb4d"}
						time.Sleep(time.Second * 10)
						breakout = true
					case "failed":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: Failed", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//red, orange, yellow
						builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
						time.Sleep(time.Second * 10)
						breakout = true
					case "cancelled":
						disp.Clear()
						disp.DrawLine(oled.TextRow1)
						disp.DrawTextCenter(strings.Replace(deploy.Name, "lagoon-", "", -1), oled.TextRow2)
						disp.DrawLine(oled.TextRow3)
						disp.DrawTextCenter("STATUS: Cancelled", oled.TextRow4)
						disp.DrawLine(oled.TextRow5)
						//red, orange, yellow
						builds = &[]string{"FF0000", "EB8F34", "FFFF00"}
						time.Sleep(time.Second * 10)
						breakout = true
					}
				}
			}
			if breakout {
				break
			}
			timeout++
		}
		builds = &[]string{}
		time.Sleep(2 * time.Second)
		ls.Wipe(lights.HexToColor("06BA90")) //teal
		disp.Clear()
		disp.PrintLogo()
		// }
		// time.Sleep(time.Second)
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
