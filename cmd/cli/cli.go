package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	// "os"

	// "strconv"
	b64 "encoding/base64"
	// "strings"
)

func main() {
	// Simple client to talk to default-http example
	con, err := net.Dial("tcp", "0.0.0.0:9999")
	if err != nil {
		log.Fatalln(err)
	}
	defer con.Close()

	// clientReader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(con)

	// buf := new(bytes.Buffer)
	// bufAll := []byte("")
	// bufAll = append(bufAll, 'a')
	// buf := make([]byte, 10)

	for _, command := range []string{":status", ":mount", ":tasks"} {
		// Waiting for the client request
		// clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			// clientRequest := strings.TrimSpace(clientRequest)
			clientRequest := command
			if _, err = con.Write([]byte(clientRequest + "\n")); err != nil {
				log.Printf("failed to send the client request: %v\n", err)
			}
		case io.EOF:
			log.Println("client closed the connection")
			return
		default:
			log.Printf("client error: %v\n", err)
			return
		}

		// Waiting for the server response
		serverResponse, err := serverReader.ReadString('\000')
		// serverResponse, err := serverReader.ReadLine()
		// line, _, err := serverReader.ReadLine()
		// e := serverReader.UnreadByte()
		// fmt.Println("e ", e)
		// serverResponse, err := serverReader.ReadString(';')
		// allMsg, _ := strconv.Atoi(strings.TrimSpace(serverResponse))
		// cntAll := 0
		// for {
		// 	cnt, err := serverReader.Read(buf)
		// 	cntAll += cnt
		// 	bufAll = append(bufAll, buf[:]...)
		// 	if err != nil {
		// 		break
		// 	}
		// 	// log.Println(cnt, err, string(bufAll[:]))
		// 	if cntAll == allMsg {
		// 		break
		// 	}
		// 	// log.Println(cnt, err, string(bufAll[:]))
		// }

		switch err {
		case nil:
			// log.Println(string(line[:]))
			// line, _, err = serverReader.ReadLine()

			// log.Println(string(line[:]), err)
			// log.Println(string(bufAll[:]))
			sDec, _ := b64.StdEncoding.DecodeString(serverResponse)
			// log.Println(string(sDec[:]))
			fmt.Print(string(sDec[:]))
			// log.Println(string(sDec[:]))
			// log.Println(strings.TrimSpace(serverResponse))
		case io.EOF:
			log.Println("server closed the connection")
			return
		default:
			log.Printf("server error: %v\n", err)
			return
		}
	}
}
