package main

import (
  "./proto"
  "strings"
  "strconv"
  "fmt"
  "regexp"
  "time"
  "encoding/json"
  "bytes"
)

type PingResponse struct {
  Latency float64
  Status struct {
    Players struct {
      Online int `json:"online"`
      Max int `json:"max"`
      Sample []struct {
        Name string `json:"name"`
        Id string `json:"id"`
      }
    }
    Description string `json:"-"`
    RawDescription interface{} `json:"description"`
    Version struct {
      Protocol int `json:"protocol"`
      Name string `json:"name"`
    }
    FaviconBase64 string `json:"favicon"`
    ModInfo struct {
      Type string
      ModList []string
    }
  }
}

type MCHost struct {
  Hostname string
  Resolved string
  Port uint16
}

func Ping(server string, protocol int) (resp PingResponse, resultingHost MCHost, err error) {
  serverParts := strings.Split(server, ":")
  if len(serverParts) > 2 {
    err = fmt.Errorf("The server provided is invalid! Wrong number of colons, detected %d", len(serverParts))
    return
  }

  serverHost := serverParts[0]
  var serverPort int
  if len(serverParts) > 1 {
    serverPort, err = strconv.Atoi(serverParts[1])
  } else {
    server = serverHost + ":25565"
    serverPort = 25565
  }

  resultingHost.Hostname = serverHost
  resultingHost.Port = uint16(serverPort)
  if err != nil || (serverPort < 0 || serverPort > 65535) {
    err = fmt.Errorf("Error parsing port from server host. Error: %e; Port: %d", err, serverPort)
    return
  }

  sent := time.Now().UnixNano() / 10E3
  conn, err := proto.Connect(server)
  if err != nil {
    return
  }
  defer conn.Close()
  resultingHost.Resolved = conn.RemoteAddr().String()
  go func() {
    for err := range conn.Errors() {
      fmt.Printf("err: %o\n", err)
    }
  }()
  conn.Serve()
  readPackets := conn.IncomingPackets()
  writePackets := conn.OutgoingPackets()
  writePackets <- &proto.HandshakePacket{proto.PVarInt(protocol), proto.PString(serverHost), proto.PUShort(serverPort), proto.PVarInt(1)}
  writePackets <- &proto.StatusRequest{}
  descData := string((<-readPackets).(*proto.StatusResponse).JSONResponse)
  // fmt.Printf("%s\n", descData)

  writePackets <- &proto.ClientPing{proto.PLong(sent)}
  if _, ok := (<-readPackets).(*proto.ServerPong); !ok {
    err = fmt.Errorf("Server did not respond to ping with pong correctly...")
    return
  }
  resp.Latency = (float64(time.Now().UnixNano()) / 10E6) - (float64(sent) / 10E2)

  json.Unmarshal([]byte(descData), &resp.Status)

  switch resp.Status.RawDescription.(type) {
  case map[string]interface{}:
    resp.Status.Description = changeJsonChatToClassicFormat((resp.Status.RawDescription).(map[string]interface{}))
  case string:
    resp.Status.Description = resp.Status.RawDescription.(string)
  }
  return
}

var regexes = map[*regexp.Regexp]func(string)string{
  regexp.MustCompile("§[a-f0-9A-Fl-okrKRL-O]") : getAnsiFor,
  regexp.MustCompile("\\s{2,}"): func(string) string {
    return " "
  },
  regexp.MustCompile("\n"): func(string) string {
    return " || "
  }}

func AsyncPing(host string, outChan chan<- hostResult) {
  resChan := make(chan hostResult)
  //the func that actually does the talking
  go func() {
    result, hostData, err := Ping(host, 315)
    if err != nil {
      resChan <- hostResult{Error: err, Host: hostData}
    } else {
      desc := "§f" + result.Status.Description + "§r"
      for key, f := range regexes {
        desc = string(key.ReplaceAllStringFunc(desc, f))
      }
      result.Status.Description = desc
      resChan <- hostResult{Ping: result, Host: hostData}
    }
  }()
  //the func that handles the timeout
  go func() {
    select {
    case line := <-resChan:
      outChan <- line
    case <-time.After(1 * time.Second):
      outChan <- hostResult{Error: fmt.Errorf("\t%s: Timeout\n", host)}
    }
  }()
}

var colorCodes = []string{"black","dark_blue","dark_green","dark_aqua","dark_red","dark_purple","gold","gray","dark_gray","blue","green","aqua","red","light_purple","yellow","white",}
const reset = "§r"

func changeJsonChatToClassicFormat(data map[string]interface{}) string {
	//string builder
	var buf bytes.Buffer
	//if this part of the message is colored, then we need to:
	if color, contains := data["color"]; contains {
		//go through all color codes (which are in order for the hex of the index to match the legacy code needed)
		for i := range colorCodes {
			//if we find one that matches
			if colorCodes[i] == color {
				//add the color code with the hex of the index which this color is in
				buf.WriteString("§" + strconv.FormatInt(int64(i), 16))
				//and break
				break
			}
		}
	}
	//now write the actual text
	buf.WriteString(data["text"].(string))
	//and if we have extra, we either append or recurse
	if extra, contains := data["extra"]; contains {
		//let's check to make sure it's an array
		switch extra.(type) {
		//if it is
		case []interface{}:
			//reset the color for safety
			buf.WriteString(reset)
			//type convert to an array
			extraArray := extra.([]interface{})
			for i := range extraArray {
				//then for each element check if it's a string or another object
				part := extraArray[i]
				switch part.(type) {
				case string: //when a string, append
					buf.WriteString(part.(string))
				case map[string]interface{}: //when an object, recurse
					buf.WriteString(changeJsonChatToClassicFormat(part.(map[string]interface{})))
				}
			}
		}
	}

	return buf.String()
}
