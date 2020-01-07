package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

/*
 from https://github.com/syncthing/syncthing/blob/master/lib/upnp/upnp.go
*/

func main() {
	/*
		ii, _ := net.Interfaces()
		for _, i := range ii {
			fmt.Println(i)
		}
	*/

	ssdp := &net.UDPAddr{IP: []byte{239, 255, 255, 250}, Port: 1900}

	var timeout time.Duration = time.Second * 15

	tpl := `M-SEARCH * HTTP/1.1
HOST: 239.255.255.250:1900
MAN: "ssdp:discover"
ST: ssdp:all
MX: %d
USER-AGENT: syncthing/1.0
`
	search := []byte(strings.Replace(fmt.Sprintf(tpl, timeout/time.Second), "\n", "\r\n", -1) + "\r\n")

	inf, err := net.InterfaceByIndex(7)
	if err != nil {
		fmt.Printf("UPnP discovery: get interface: %v", err)
		return
	}

	socket, err := net.ListenMulticastUDP("udp4", inf, &net.UDPAddr{IP: ssdp.IP})
	if err != nil {
		fmt.Printf("UPnP discovery: listening to udp multicast: %v", err)
		return
	}
	defer socket.Close()

	err = socket.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		fmt.Printf("UPnP discovery: setting socket deadline: %v", err)
		return
	}

	_, err = socket.WriteTo(search, ssdp)
	if err != nil {
		fmt.Printf("UPnP discovery: sending search request: %v", err)
		return
	}

	fmt.Println("Listening for UPnP response on " + inf.Name)

	// Listen for responses until a timeout is reached
	for {
		resp := make([]byte, 65536)
		n, _, err := socket.ReadFrom(resp)
		if err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				fmt.Printf("UPnP read: %v\n", err) //legitimate error, not a timeout.
			}
			break
		}
		fmt.Println(string(resp[:n]))
	}
	fmt.Println("Discovery on " + inf.Name + " finished.")
}
