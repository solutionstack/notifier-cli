package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/solutionstack/notifier-cli/notifier"
	"io"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
)

type CliOptions struct {
	url      string
	interval string
}

var cliOptions CliOptions
var fileData []string

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)


	defer func() {
		if r := recover(); r != nil {
			fmt.Println("The program encountered and error: ", r)
		}
	}()

	cliOptions = CliOptions{}
	var helpFlag,helpFlagShort *bool

	flag.StringVar(&cliOptions.url, "url", "", "")
	flag.StringVar(&cliOptions.interval, "interval", "0s", "")
	flag.StringVar(&cliOptions.interval, "i", "0s", "")
	helpFlag = flag.Bool("help", false, "print help message")
	helpFlagShort = flag.Bool("h", false, "print help message")

	flag.Parse()

	// if user does not supply flags, print usage
	if flag.NFlag() == 0 {
		printHelp()
		os.Exit(0)
	}

	if *helpFlag == true || *helpFlagShort == true{
		printHelp()
		return
	}

	_, err := url.ParseRequestURI(cliOptions.url)
	if err != nil {
		panic(err)
	}

	readFile()
	processData()
}



func readFile() {

	if !stdInAvailable() {
		panic("no input stream detected")
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("reading file...")

	for scanner.Scan() {
		text := scanner.Text()
		fileData = append(fileData, text)
		//fmt.Println(fileData)
	}

	err := scanner.Err()
	if err != nil && err != io.EOF {
	}
	fmt.Println("reading file complete")

}

func stdInAvailable() bool {
	file := os.Stdin
	fi, err := file.Stat()
	if err != nil {
		panic(fmt.Sprint("file.Stat(): ", err))
	}
	size := fi.Size()
	if size > 0 {
		return true
	} else {
		return false
	}
}
func printHelp() {
	fmt.Println("usage: notify --url=URL [<flags>]")
	fmt.Println("Flags:\n\t--i --interval=5s Notification interval\n\t--h --help Print this help message")
}


func processNotifyEvents(event notifier.MessageEvent, messageId int, errBody string) {
	switch event {
	case notifier.CompletedEvent:
		fmt.Println("\ninput processing is complete, exit.")
		os.Exit(0)

	default:
		printEvent(event , messageId, errBody)
	}
}
func printEvent(event notifier.MessageEvent, messageId int, errBody string) {

	if event == notifier.SuccessEvent{
		fmt.Printf("message_id[%d] text:%s | processed succesfully\n", messageId, fileData[messageId][0:10]+"..." )
	}else{
		fmt.Printf("message_id[%d] text:%s | failed with :%s | error: %s \n", messageId, fileData[messageId][0:10]+"...", event, errBody)

	}
}



func processData() {

	fmt.Println("\nstarting processing")

	re := regexp.MustCompile("[0-9]+") //re to get the actual number from the interval
	interval, _ := strconv.ParseInt(re.FindAllString(cliOptions.interval,1)[0], 10, 32)

	n := notifier.NewNotifier(cliOptions.url, fileData, int(interval), processNotifyEvents)
	n.ProcessMessages()
}