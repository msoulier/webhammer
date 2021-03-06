package main

import (
    "github.com/op/go-logging"
    "flag"
    "os"
    "net/http"
    "time"
    "crypto/tls"
)

var (
    debug = false
    log *logging.Logger
    waittime = 0
    surl = ""
    goroutines = 1
    pause = 0
    disable_tls_validation = false
)

func init() {
    flag.BoolVar(&disable_tls_validation, "k", false, "Disable TLS validation")
    flag.BoolVar(&debug, "d", false, "Debug logging")
    flag.IntVar(&waittime, "w", 0, "Wait time")
    flag.StringVar(&surl, "u", "", "URL to hammer")
    flag.IntVar(&goroutines, "g", 1, "Number of goroutines to maintain")
    flag.IntVar(&pause, "p", 100, "Number of milliseconds to pause between starting goroutines")
    flag.Parse()

    if surl == "" {
        flag.PrintDefaults()
        os.Exit(1)
    }

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

    for i := 0; i < goroutines; i++ {
        log.Infof("starting goroutine %d", i+1)
        go func(ch chan int) {
            client := &http.Client{}
            if disable_tls_validation {
                tr := &http.Transport{
                        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
                    }
                client = &http.Client{
                    Transport: tr,
                    Timeout: time.Second*time.Duration(waittime),
                }
            } else {
                client = &http.Client{
                    Timeout: time.Second*time.Duration(waittime),
                }
            }
            if resp, err := client.Get(surl); err != nil {
                log.Errorf("%s", err)
                ch <- -1
            } else {
                defer resp.Body.Close()
                if resp.StatusCode == 200 {
                    ch <- 1
                } else {
                    ch <- 0
                }
            }
        }(ch)
        if pause > 0.0 {
            time.Sleep(time.Millisecond*time.Duration(pause))
        }
    }

    for i := 0; i < goroutines; i++ {
        rv := <- ch
        log.Infof("return %d - %d goroutines to go",
            rv, goroutines - (i+1))
    }
}
