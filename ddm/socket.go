package ddm

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"

	// "github.com/labstack/echo"
	// socketio "github.com/googollee/go-socket.io"
	"github.com/rodaine/table"
)

func (d *DDMObserver) Socket() {
	listener, err := net.Listen("tcp", "0.0.0.0:9999")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		// If you want, you can increment a counter here and inject to handleClientRequest below as client identifier
		go handleClientRequest(con, d)
	}
}

func handleClientRequest(con net.Conn, d *DDMObserver) {
	defer con.Close()

	clientReader := bufio.NewReader(con)

	for {
		// Waiting for the client request
		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if clientRequest == ":QUIT" {
				log.Println("client requested server to close the connection so closing")
				return
			} else {
				log.Println(clientRequest)
				tbl := table.New("UUID", "Name", "Active")
				for _, disk := range d.DiskStat {
					tbl.AddRow(disk.UUID, disk.Name, disk.Active)
				}
				tbl.Print()
			}
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}

		// Responding to the client request
		if _, err = con.Write([]byte("GOT IT!\n")); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}
	}
}
