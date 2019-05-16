package main

import (
  "fmt"
  "bufio"
  "os"
  "log"
  "strconv"
  "syscall"
  "strings"

  "github.com/pborman/getopt/v2"
  "github.com/fsnotify/fsnotify"
)

func readPidFile(pidFile string) int {
  f, err := os.Open(pidFile)
  defer f.Close()
  if err != nil {
    log.Println("WARNING: could not open pid file")
    return -1
  }
  r := bufio.NewReader(f)
  line, _ := r.ReadString('\n')
  s := strings.TrimSuffix(line, "\n")
  pid, err := strconv.Atoi(s)
  if err != nil {
    log.Println("WARNING: invalid pid file")
    return -1
  }
  return pid
}
func MapPidArgs(pidStrings []string) []int {
  var err error
    pids := make([]int, len(pidStrings))
    for i, pidString := range pidStrings {
        pids[i], err = strconv.Atoi(pidString)
        if err != nil {
          panic(err)
        }
    }
    return pids
}
func notifyProcess(pid int) {
  p, err := os.FindProcess(pid)
  if err != nil {
    panic(err)
  }
  p.Signal(syscall.SIGHUP)
}
func notifyPids(pids []int) {
  for _, pid := range pids {
    notifyProcess(pid)
  }
}
func notifyPidFiles(pidfiles []string) {
  for _, pidFile := range pidfiles {
    pid := readPidFile(pidFile)
    notifyProcess(pid)
  }
}

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

  pids := MapPidArgs(*optPid)

  watchNotify(*optWatch, pids, *optPidFile)
}

func watchNotify(watches []string, pids []int, pidfiles []string) {
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
        notifyPids(pids)
        notifyPidFiles(pidfiles)
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
