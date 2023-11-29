package communication

import (
	"bufio"
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/alfonzso/dying-disk-manager/ddm"
	"github.com/rodaine/table"
)

func Socket(ddmd *ddm.DDMData) {
	listener, err := net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		log.Fatalln(err)
	}

	go handleConnection(listener, ddmd)
}

func handleConnection(listener net.Listener, ddmd *ddm.DDMData) {
	defer listener.Close()
	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}

		// If you want, you can increment a counter here and inject to handleClientRequest below as client identifier
		// tbl := table.New("UUID", "Name", "Active", "Mount")
		// tbl := table.New("UUID", "Name", "Active")
		// buf := new(bytes.Buffer)
		// tbl.WithWriter(buf)
		// go handleClientRequest(con, ddmd, tbl, buf)
		go handleClientRequest(con, ddmd)
	}
}

func printDiskStat(ddmd *ddm.DDMData) *bytes.Buffer {
	table := table.New("UUID", "Name", "Active")
	buff := new(bytes.Buffer)
	// // f(buf)
	table.WithWriter(buff)
	for _, disk := range ddmd.DiskStat {
		table.AddRow(disk.UUID, disk.Name, disk.Active) // , disk.Mount.Print(), disk.Test, disk.Repair)
	}
	table.Print()
	return buff
	// return buf.String()
}

func printActionts(ddmd *ddm.DDMData) *bytes.Buffer {
	table := table.New("Action", "Status", "ThreadIsRunning", "DisabledByAction")
	buff := new(bytes.Buffer)
	table.WithWriter(buff)
	for _, disk := range ddmd.DiskStat {
		table.AddRow("Mount", disk.Mount.Status, disk.Mount.ThreadIsRunning, disk.Mount.DisabledByAction)
		table.AddRow("Test", disk.Test.Status, disk.Test.ThreadIsRunning, disk.Test.DisabledByAction)
		table.AddRow("Repair", disk.Repair.Status, disk.Repair.ThreadIsRunning, disk.Repair.DisabledByAction)
	}
	table.Print()
	return buff
}

func handleClientRequest(con net.Conn, ddmd *ddm.DDMData) {
	defer con.Close()

	clientReader := bufio.NewReader(con)

	for {
		buff := new(bytes.Buffer)
		// Waiting for the client request
		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if clientRequest == ":QUIT" {
				log.Println("client requested server to close the connection so closing")
				return
			}

			log.Println(clientRequest)

			if clientRequest == ":status" {
				buff = printDiskStat(ddmd)
			}
			if clientRequest == ":mount" {
				buff = printActionts(ddmd)
			}
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}

		// msg := []byte("12345678901234567890123456789012345678901234567890123456789012345678901234567890\nasdfasdfadfa")
		// msg := '\n' + buf.Bytes()
		// msg := append([]byte("\n"), buf.Bytes()...)
		// // Responding to the client request
		// // __n, err := con.Write(len(msg))
		// _, err = con.Write([]byte(strconv.Itoa(len(msg)) + "\n"))
		// if err != nil {
		// 	log.Printf("failed to respond to client: %v\n", err)
		// }

		// sEnc := b64.StdEncoding.
		// sEnc := b64.StdEncoding.EncodeToString([]byte("\n" + buf.Bytes() + "\n"))
		// msg := append([]byte("\n"), buf.Bytes()...)
		// msg = append(msg, '\n')

		// mm := []byte("\n"+ buf.Bytes()... + "\000")
		msg := []byte("\n" + buff.String() + "\000")
		sEnc := b64.StdEncoding.EncodeToString(msg)
		fafa := []byte(sEnc[:] + "\000")
		__n, err := con.Write(fafa)
		if err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}

		fmt.Println("----------> ", __n)
		// fmt.Println("----------> ", string(msg[:]))
		// Responding to the client request
		// if _, err = con.Write(buf.Bytes()); err != nil {
		// 	log.Printf("failed to send BUF respond to client: %v\n", err)
		// }
	}
}
