package main

import (
  "fmt"
  "runtime"
  "./proto"
  "time"
)

type hostResult struct {
  Ping PingResponse
  Error error
  Host MCHost
}

var servers = []string{ "us.mineplex.com", "emenbee.net",
  "play.luckyprison.com", "server.cityprison.net",
  "mc.fearpvp.com", "play.furiouspvp.net",
  "play.primemc.org", "mineheroes.net",
  "play.extremecraft.net", "play.cubecraft.net",
  "play.lemoncloud.org", "play.gotpvp.com",
  "mc.desiredcraft.net", "mc-central.net",
  "mineverse.com", "mc.snapcraft.net",
  "omegarealm.com", "pixel.rc-gamers.com",
  "FadeCloud.com", "pvp.Desteria.com",
  "play.hypixel.net", "tmg.pw:42069"}
func main() {
  runtime.GOMAXPROCS(runtime.NumCPU())
  for {
    proto.SetDebug(false)
    res := make(chan hostResult)
    for i := range servers {
      AsyncPing(servers[i], res)
    }
    var count int
    fmt.Printf("\n----------------------------------------------------------------\n\n")
    start := time.Now()
    for server := range res {
      err := server.Error
      if err != nil {
        fmt.Printf("\x1B\x5B91m\x1B\x5B4m\x1B\x5B1m#%d [%s:%d ERR]:\x1B\x5B0m\n     %s\n\n", count + 1, server.Host.Hostname, server.Host.Port, err.Error())
      } else {
        fmt.Printf("#%d [%.02fms ---> %s AKA %s]:\n     (%d/%d) %s\n\n", count + 1, server.Ping.Latency, server.Host.Resolved, server.Host.Hostname, server.Ping.Status.Players.Online, server.Ping.Status.Players.Max, server.Ping.Status.Description)
      }
      count++
      if (count == len(servers)) {
        close(res)
      }
    }
    stop := time.Now()
    fmt.Printf("--------------------------[ took: %.02fms ]---------------------\n", float64(stop.UnixNano() - start.UnixNano()) / 10E6)
    <-time.After(2 * time.Second)
  }
}
