package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

const TEMP_FILE string = "/tmp/screen_temp_file"
const DETACHED string = "Detached"

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

type Screen struct {
	PID    string
	Name   string
	Time   string
	Host   string
	Status bool
}

func InitLogs(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

//func usage() {
//	w := os.Stdout
//
//	getopt.PrintUsage(w)
//}
//
func main() {

	InitLogs(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	_, err := user.Current()
	if err != nil {
		Error.Println("Error getting current user info: ")
		Error.Println(err)
		os.Exit(1)
	}

	var opts struct {
		Verbose []bool `short:"v" long:"verbose" description:"Show verbose debug information"`
		Auto bool `short:"a" long:"auto" description:"Auto connect if there are detached screens"`
	}

	var parser = flags.NewParser(&opts, flags.Default)

	_, err = parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	verbose := 0
	for _, s := range opts.Verbose {
		if s {
			verbose = verbose + 1
		}
	}

	vw := ioutil.Discard
	if verbose > 0 {
		vw = os.Stdout
	}

	vi := ioutil.Discard
	if verbose > 1 {
		vi = os.Stdout
	}

	vt := ioutil.Discard
	if verbose > 2 {
		vt = os.Stdout
	}

	InitLogs(vt, vi, vw, os.Stderr)

	Info.Println("Verbose set to ", verbose)

	for mainMenu(opts.Auto) != 0 {

	}

}

func mainMenu(auto bool) int {
	screens := List()
	if len(screens) <= 1 {
		os.Exit(1)
	}

	screens = screens[1 : len(screens)-1]

	if auto {
		for i, s := range screens {
			if strings.Contains(s, DETACHED) {
				var list []string = strings.Split(strings.Trim(screens[i], "\t"), "\t")
				fmt.Println("screen -x "+ list[i])
				os.Exit(0)
			}
		}
	}

	fmt.Println("Available Screens:\n")
	for i, s := range screens {
		fmt.Println(i, s)
	}
	fmt.Println("X\tExit")
	fmt.Print("Choose a screen to open\n> ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	Info.Println("Pressed:", text)
	text = strings.Replace(strings.ToLower(text), "\n", "", -1)
	text = strings.Replace(strings.ToLower(text), "\r", "", -1)

	if text == "x" || text == "X" {
		Info.Println("Selected Exit option")
		os.Exit(1)
	}

	i, err := strconv.Atoi(text)

	if err != nil {
		Error.Println("Error in Atoi. is '" + text + "' not a number?")
		os.Exit(1)
	}
	var list []string = strings.Split(strings.Trim(screens[i], "\t"), "\t")

	Info.Println("Screen selected: " + list[0])

	// screen  -r <name of sesion>
	// screen  -S 7854.pts-1.aurum -p 0 -X hardcopy kk.txt

	Hardcopy(list[0])

	var cmd *exec.Cmd
	cmd = runPager(TEMP_FILE)
	cmd.Wait()

	fmt.Println("\nDo you want to open this screen? (y/n)")
	text, _ = reader.ReadString('\n')
	Info.Println("Pressed:", text)
	text = strings.Replace(strings.ToLower(text), "\n", "", -1)
	text = strings.Replace(strings.ToLower(text), "\r", "", -1)

	if text == "y" || text == "Y" {
		fmt.Println("screen -x "+ list[0])
		os.Exit(0)
	}

	return 1
}

func List() []string {
	var cmd *exec.Cmd = exec.Command("screen", "-ls")
	var output, err = cmd.Output()
	if err != nil {
		Error.Println("List exec Command Failed")
		os.Exit(1)
	}
	var list []string = strings.Split(strings.Trim(string(output), "\n"), "\n")
	return list
}

func Hardcopy(session string) {
	var cmd *exec.Cmd = exec.Command("screen", "-S", session, "-p", "0", "-X", "hardcopy", TEMP_FILE)
	var _, err = cmd.Output()
	if err != nil {
		Error.Println("Hardcopy exec Command Failed")
		os.Exit(1)
	}
}

// https://stackoverflow.com/a/54198703/945568

func runPager(file string) (*exec.Cmd) {
	var cmd *exec.Cmd
	pager := os.Getenv("PAGER")
	if pager == "" {
		cmd = exec.Command("less", "-X", "-N", "-R", "-S", "+G", file)
	} else {
		cmd = exec.Command(pager, file)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		Error.Println(err)
		os.Exit(1)
	}
	return cmd
}

func displayHardcopy(buf io.Writer) {
	dat, err := os.ReadFile(TEMP_FILE)
	if err != nil {
		Error.Println(err)
		os.Exit(1)
	}

	fmt.Fprintf(buf, string(dat))

	err = os.Remove(TEMP_FILE)
	if err != nil {
		Error.Println(err)
	}
}
