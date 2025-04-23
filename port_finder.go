package main

import (
	"log"
	"os"
	"os/exec"
	gc "github.com/gbin/goncurses"
	"strings"
	"slices"
)

const (
	// magic numbers from original
	LINE_LENGTH = 4 + 6 + 12
	PATTERN =  "::1:5"
	PORT_MAP = "/tmp/port-map.txt"
	//NETSTAT_EX = "./ex_netstat.txt"
)

var today string

func checkQuit(screen gc.Window) {
	if screen.GetChar() == 'q' {
		gc.End()
		os.Exit(0)
	}
}

func getSpacing(maxValue int, n int) string{
	spaces := string(" ")
	if maxValue >= 10 {
		if n < 10 {
			spaces += " "
			return spaces
		}
	}
	return spaces 
}

func sameDate(a string, b string) bool {
	if extractDate(a) == extractDate(b) {
		return true
	}
	return false
}

// only get month day year (mdy)
// Mon Apr 21 02:25:08 EDT 2025
// [0] [1] [2]   [3]   [4]  [5]
func extractDate(date string) string {
	parts := strings.Fields(date)
	mdy := []string{parts[1],parts[2],parts[5]}
	return strings.Join(mdy," ")
} 

func getToday() string {
	// date formatting will be 24-hour
	os.Setenv("LC_TIME","C.UTF-8")
	cmd := exec.Command("date")
	out_bytes, err := cmd.Output()
	if err != nil {
		log.Fatal("command: ", err)
	}

	return string(out_bytes)
}

func getPortMap() map[string]string {
	portmap := map[string]string{}

	contents, err := os.ReadFile(PORT_MAP)
	if err != nil {
		log.Fatal("reading file: ", err)
	}

	// port-map.txt format:
	// {hwaddr},{port},${date}
	// date command format eg
	// Mon Apr 21 02:25:08 PM EDT 2025
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		fields := strings.Split(line, ",")
		date := string(fields[2])
		mac := string(fields[0])
		port := string(fields[1])

		// compare if days are the same

		// if same, assign MAC to port
		if sameDate(date,today) {
			portmap[port] = mac 
		} else  {
			// empty, override pervious text if there was
			portmap[port] = "            "
		}
	}

	return portmap
}

func main() {
	screen, err := gc.Init()

	if err != nil {
		log.Fatal("init: ", err)
	}
	defer gc.End()
	go checkQuit(*screen)
	log.Print("Screen initialization successful.")

	if err := gc.StartColor(); err != nil {
		log.Fatal(err)
	}
	log.Print("Color set up successful.")

	gc.InitPair(1, gc.C_WHITE, gc.C_BLUE)

	maxY, maxX := screen.MaxYX()

	today = getToday()

	screen.Erase()
	gc.Echo(false)
	gc.Cursor(0)
	gc.Raw(true)
	screen.Box(gc.ACS_VLINE, gc.ACS_HLINE)
	
	var start_ports []string

	screen.SetBackground(gc.ColorPair(1))
	screen.Refresh()
	screen.Clear()
	for {
		screen.Erase()
		cmd := exec.Command("netstat", "-tunpl")
		out_bytes, err := cmd.Output()
		if err != nil {
			gc.End()
			log.Fatal("command: ",err)
		}
		output := string(out_bytes)

		//contents, err := os.ReadFile(NETSTAT_EX)
		//if err != nil {
		//	log.Fatal("error reading: ", err)
		//}
		//output := string(contents)

		screen.MovePrintln(1,1,"(press 'q' to exit)")
		screen.Box(gc.ACS_VLINE, gc.ACS_HLINE)
		lines := strings.Split(output, "\n")

		var ports []string

		var new_ports []string
		// get ports with ::1:5
		for _, line := range lines {

			if strings.Contains(line, PATTERN) {
				port := strings.Fields(line)[3][4:]
				ports = append(ports, port)
			}
		}
		// map ports to hwaddr IF the date last used by the port is today
		// otherwise, ommit providing a Port as it may be incorrect

		portmap := getPortMap()

		startY := ((maxY / 2) - len(ports)/2)
		startX := ((maxX/ 2) - (LINE_LENGTH / 2))
		// theyre all new if its first, dont highlight
		first_run := false
		if len(start_ports) == 0 {
			first_run = true
		}


		slices.Sort(ports)
		slices.Compact(ports)
		ports_n := len(ports)
		for dy, port := range ports {
			if first_run {
				start_ports = append(start_ports, port)
			} else if ! slices.Contains(start_ports,  port){
				err := screen.Standout()
				if err != nil {
					log.Fatal("Standout: ")
				}
				new_ports = append(new_ports, port)
			}

			spacing := getSpacing(ports_n, dy)
			screen.MovePrintf(startY+dy, startX, "[%d]%s%s %s",
				dy, spacing, port, portmap[port])
				err := screen.Standend()
				if err != nil {
					log.Fatal("Standend: ")
				}
		}
		// keep from original
		screen.Touch()
		screen.Refresh()
		gc.Update()
		// sleep for 500ms
		gc.Nap(500)
	}
}

