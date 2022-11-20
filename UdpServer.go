package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getHostFromIp(ip string) string {
	ip = strings.Split(ip, ":")[0]
	addr, err := net.LookupAddr(ip)
	if err != nil {
		return ip
	}
	return strings.TrimSuffix(addr[0], ".")
}

var knownMsgClasses = []string{ // List of known Classes, identified by regex
	`system.power.state`,
	`system.init`,
	`system.power.nightly`,
	`system.firealarm.state`,
	`system.touchpanel.page`,
	`system.connected.[a-z0-9-.]+`,
	`video.input.select.[a-z0-9-.]+`,
	`system.keepalive`,
}

func parseMessage(msg []byte) (string, int, error) {

	parts := strings.Split(strings.ToLower(string(bytes.Trim(msg, "\x00"))), "=") // Remove nil-Characters from []byte-Message and convert to string and lower string and split msgClass from msgValue
	msgClass := parts[0]
	if len(parts) < 2 {
		return "0", 0, errors.New("Message value not found")
	}
	msgValue, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return "0", 0, err
	}

	// Check that the message contains a valid metric

	for _, re := range knownMsgClasses {

		res, _ := regexp.MatchString(re, msgClass)

		if res {
			fmt.Println("Valid message detected", msgValue)

			// Is it a init command?
			res, _ := regexp.MatchString(`system.init`, msgClass)
			if res {
				msgValue = int(time.Now().Unix())
			}

			// Is it a nightly power shutdown command?
			res, _ = regexp.MatchString(`system.power.nightly`, msgClass)
			if res {
				msgValue = int(time.Now().Unix())
			}

			// Is it a keep alive command?
			res, _ = regexp.MatchString(`system.keepalive`, msgClass)
			if res {
				msgValue = int(time.Now().Unix())
			}

			return msgClass, msgValue, nil
		}

	}

	fmt.Println(msgClass, msgValue)

	return "0", 0, errors.New("Message Class not found")
}

func UdpServer(ctx context.Context, address string, redisClient *redis.Client) (err error) {

	pc, err := net.ListenPacket("udp", address)
	if err != nil {
		fmt.Println("Error opening udp connection", err)

		return
	}

	// At least, close connection
	defer pc.Close()

	doneChan := make(chan error, 1)
	buffer := make([]byte, 1024) // 1024 is the max buffer size

	fmt.Println("Start Server")

	go func() {
		for {

			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				fmt.Println("readfrom", err)

				doneChan <- err
				return
			}

			// fmt.Printf("packet-received: bytes=%d msg=%s\n", n, buffer)

			host := getHostFromIp(addr.String())

			//fmt.Println("Host:", host)

			msg, val, err := parseMessage(buffer[:n])

			/* Todo: Reset values on init
			if msg == "system.init"{  // reset values on init
				err := redisClient.HSet(host , msg, val).Err()
				if err != nil {
					panic(err)
				}
			}
			*/

			resp := "ACK"
			if err != nil {
				fmt.Println("Fehler", err)
				resp = "ERR"
			} else {

				err := redisClient.HSet(ctx, host, msg, val).Err()
				if err != nil {
					// Continue, if redis is not responding
					fmt.Println("Redis write timeout")
					continue
				}

				fmt.Println("Written", host, msg, val)

			}

			// Write the packet's contents back to the client.
			n, err = pc.WriteTo([]byte(resp+"\n"), addr)
			if err != nil {
				doneChan <- err

			}

		}
	}()
	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
		err = ctx.Err()
	case err = <-doneChan:
	}

	return

}
