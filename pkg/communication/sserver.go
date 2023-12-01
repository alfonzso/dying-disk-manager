package communication

import (
	"bufio"
	"bytes"
	b64 "encoding/base64"
	"io"
	"net"
	"strings"

	"github.com/alfonzso/dying-disk-manager/ddm"
	"github.com/rodaine/table"
	log "github.com/sirupsen/logrus"
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

		go handleClientRequest(con, ddmd)
	}
}

func printDiskStat(ddmd *ddm.DDMData) *bytes.Buffer {
	table := table.New("UUID", "Name", "Active")
	buff := new(bytes.Buffer)
	table.WithWriter(buff)
	for _, disk := range ddmd.DiskStat {
		table.AddRow(disk.UUID, disk.Name, disk.Active)
	}
	table.Print()
	return buff
}

func printActionts(ddmd *ddm.DDMData) *bytes.Buffer {
	table := table.New("DiskName", "Action", "Status", "ThreadIsRunning", "DisabledByAction")
	buff := new(bytes.Buffer)
	table.WithWriter(buff)
	for _, disk := range ddmd.DiskStat {
		table.AddRow(disk.Name, "Mount", disk.Mount.Status, disk.Mount.ThreadIsRunning, disk.Mount.DisabledByAction)
		table.AddRow(disk.Name, "Test", disk.Test.Status, disk.Test.ThreadIsRunning, disk.Test.DisabledByAction)
		table.AddRow(disk.Name, "Repair", disk.Repair.Status, disk.Repair.ThreadIsRunning, disk.Repair.DisabledByAction)
	}
	table.Print()
	return buff
}

func printTasks(ddmd *ddm.DDMData) *bytes.Buffer {
	table := table.New("Name", "NextRun", "Tags")
	buff := new(bytes.Buffer)
	table.WithWriter(buff)
	for _, job := range ddmd.Scheduler.Jobs() {
		nextRun, _ := job.NextRun()
		table.AddRow(job.Name(), nextRun, job.Tags())
	}
	table.Print()
	return buff
}

func handleClientRequest(con net.Conn, ddmd *ddm.DDMData) {
	defer con.Close()

	clientReader := bufio.NewReader(con)

	for {
		buff := new(bytes.Buffer)
		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if clientRequest == ":q" {
				log.Debug("client requested server to close the connection so closing")
				return
			}

			if clientRequest == ":status" {
				buff = printDiskStat(ddmd)
			}
			if clientRequest == ":mount" {
				buff = printActionts(ddmd)
			}
			if clientRequest == ":tasks" {
				buff = printTasks(ddmd)
			}
		case io.EOF:
			return
		default:
			log.Errorf("error: %v\n", err)
			return
		}

		sEnc := b64.StdEncoding.EncodeToString([]byte("\n"+buff.String())) + "\000"
		if _, err = con.Write([]byte(sEnc[:])); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}

	}
}
