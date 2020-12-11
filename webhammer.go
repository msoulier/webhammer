package main

import (
    "github.com/op/go-logging"
    "flag"
    "os"
    "net/http"
)

var (
    debug = false
    log *logging.Logger
    waittime = 0
    surl = ""
    goroutines = 1
)

func init() {
    flag.BoolVar(&debug, "d", false, "Debug logging")
    flag.IntVar(&waittime, "w", 0, "Wait time")
    flag.StringVar(&surl, "u", "", "URL to hammer")
    flag.IntVar(&goroutines, "g", 1, "Number of goroutines to maintain")
    flag.Parse()

    format := logging.MustStringFormatter(
        `%{time:2006-01-02 15:04:05.000-0700} %{level} [%{shortfile}] %{message}`,
        )
    stderrBackend := logging.NewLogBackend(os.Stderr, "", 0)
    stderrFormatter := logging.NewBackendFormatter(stderrBackend, format)
    stderrBackendLevelled := logging.AddModuleLevel(stderrFormatter)
    logging.SetBackend(stderrBackendLevelled)
    if debug {
        stderrBackendLevelled.SetLevel(logging.DEBUG, "webhammer")
    } else {
        stderrBackendLevelled.SetLevel(logging.INFO, "webhammer")
    }
    log = logging.MustGetLogger("webhammer")
}

func main() {
    log.Infof("hammer of url %s requested, wait seconds %d, goroutines %d",
        surl, waittime, goroutines)

    var ch chan int
    ch = make(chan int)

    go func(ch chan int) {
        client := &http.Client{}
        if resp, err := client.Get(surl); err != nil {
            log.Errorf("%s", err)
            ch <- -1
        } else {
            if resp.StatusCode == 200 {
                ch <- 1
            } else {
                ch <- 0
            }
        }
    }(ch)

    rv := <- ch
    log.Infof("return %d", rv)
}
