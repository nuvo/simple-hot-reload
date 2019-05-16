package main

import (
  "fmt"
  "os"
  "log"

  "github.com/pborman/getopt/v2"
  "github.com/fsnotify/fsnotify"
)

func main()  {
  optWatch := getopt.ListLong("watch", 'w', "", "Directory or file to watch")
  optPid := getopt.ListLong("pid", 'p', "", "A pid to send a SIGHUP when watched files change")
  optPidFile := getopt.ListLong("pidfile", 'f', "", "A pid file to take the pid from")
  optHelp := getopt.BoolLong("help", 0, "Help")
  getopt.Parse()

  if *optHelp {
      getopt.Usage()
      os.Exit(0)
  }

  if len(*optWatch) == 0 {
    fmt.Printf("Error: must watch at least one item\n")
    getopt.Usage()
    os.Exit(1)
  }

  if len(*optPid) == 0 && len(*optPidFile) == 0 {
    fmt.Printf("Error: must include at least one of --pid or --pidfile\n")
    getopt.Usage()
    os.Exit(1)
  }

  watchNotify(*optWatch, *optPid, *optPidFile)
}

func watchNotify(watches []string, pids []string, pidfiles []string) {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal(err)
  }
  defer watcher.Close()

  done := make(chan bool)
  go func() {
    for {
      select {
      case _, ok := <-watcher.Events:
        if !ok {
          return
        }
        for _, pid := range pids {
          fmt.Printf("sending SIGHUP to %s\n", pid)
        }
      case err, ok := <-watcher.Errors:
        if !ok {
          return
        }
        log.Println("error:", err)
      }
    }
  }()

  for i, watch := range watches {
    fmt.Printf("watchlist item [%d] %s\n", i, watch)
    watcher.Add(watch)
  }
  for i, pidFile := range pidfiles {
    fmt.Printf("pidFile item [%d] %s\n", i, pidFile)
  }

  if err != nil {
		log.Fatal(err)
	}
  <-done
}
