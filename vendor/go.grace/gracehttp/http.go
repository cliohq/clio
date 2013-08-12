// Package gracehttp provides easy to use graceful restart
// functionality for HTTP server.
package gracehttp

import (
  "bytes"
  "errors"
  "flag"
  "fmt"
  "github.com/pallada/clio/vendor/go.grace"
  "github.com/pallada/clio/helpers"
  "log"
  "net"
  "net/http"
  "os"
)

type serverSlice []*http.Server

var (
  verbose           = flag.Bool("gracehttp.log", true, "Enable logging.")
  errListenersCount = errors.New("unexpected listeners count")
)

// Creates new listeners for all the given addresses.
func (servers serverSlice) newListeners() ([]grace.Listener, error) {
  listeners := make([]grace.Listener, len(servers))
  for index, pair := range servers {
    addr, err := net.ResolveTCPAddr("tcp", pair.Addr)
    if err != nil {
      return nil, fmt.Errorf(
        "Failed net.ResolveTCPAddr for %s: %s", pair.Addr, err)
    }
    l, err := net.ListenTCP("tcp", addr)
    if err != nil {
      return nil, fmt.Errorf("Failed net.ListenTCP for %s: %s", pair.Addr, err)
    }
    listeners[index] = grace.NewListener(l)
  }
  return listeners, nil
}

// Serve on the given listeners and wait for signals.
func (servers serverSlice) serveWait(listeners []grace.Listener) error {
  if len(servers) != len(listeners) {
    return errListenersCount
  }
  errch := make(chan error, len(listeners)+1) // listeners + grace.Wait
  for i, l := range listeners {
    go func(i int, l net.Listener) {
      err := servers[i].Serve(l)
      // The underlying Accept() will return grace.ErrAlreadyClosed
      // when a signal to do the same is returned, which we are okay with.
      if err != nil && err != grace.ErrAlreadyClosed {
        errch <- fmt.Errorf("Failed http.Serve: %s", err)
      }
    }(i, l)
  }
  go func() {
    err := grace.Wait(listeners)
    if err != nil {
      errch <- fmt.Errorf("Failed grace.Wait: %s", err)
    } else {
      errch <- nil
    }
  }()
  return <-errch
}

// Serve will serve the given pairs of addresses and listeners and
// will monitor for signals allowing for graceful termination (SIGTERM)
// or restart (SIGUSR2).
func Serve(servers ...*http.Server) error {
  sslice := serverSlice(servers)
  listeners, err := grace.Inherit()
  if err == nil {
    err = grace.CloseParent()
    if err != nil {
      return fmt.Errorf("Failed to close parent: %s", err)
    }
    if *verbose {
      ppid := os.Getppid()
      if ppid == 1 {
        log.Printf("Listening on init activated %s", pprintAddr(listeners))
      } else {
        log.Printf(
          "Graceful handoff of %s with new pid %d and old pid %d.",
          pprintAddr(listeners), os.Getpid(), ppid)

        // updating pid file
        helpers.UpdatePidFile(os.Getpid())
      }
    }
  } else if err == grace.ErrNotInheriting {
    listeners, err = sslice.newListeners()
    if err != nil {
      return err
    }

    // creating pid file
    helpers.UpdatePidFile(os.Getpid())

    if *verbose {
      log.Printf("Serving %s with pid %d.", pprintAddr(listeners), os.Getpid())
    }
  } else {
    return fmt.Errorf("Failed graceful handoff: %s", err)
  }
  err = sslice.serveWait(listeners)
  if err != nil {
    return err
  }
  if *verbose {
    log.Printf("Exiting pid %d.", os.Getpid())
  }
  return nil
}

// Used for pretty printing addresses.
func pprintAddr(listeners []grace.Listener) []byte {
  out := bytes.NewBuffer(nil)
  for i, l := range listeners {
    if i != 0 {
      fmt.Fprint(out, ", ")
    }
    fmt.Fprint(out, l.Addr())
  }
  return out.Bytes()
}
