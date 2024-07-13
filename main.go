package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"io"
	"log"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	filename      = ""
	StopListening = false
	downloadsPath = ""
	warn          = lipgloss.NewStyle().Foreground(lipgloss.Color("#fbff00"))
	magenta       = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	placeholder   = lipgloss.NewStyle().Foreground(lipgloss.Color("#34eb98"))
	orange        = lipgloss.NewStyle().Foreground(lipgloss.Color("#fc9003"))
	red           = lipgloss.NewStyle().Foreground(lipgloss.Color("#d10000"))
	green         = lipgloss.NewStyle().Foreground(lipgloss.Color("#14db2e"))
)

type FileServer struct{}

func (fs *FileServer) start() {
	if !StopListening {
		ln, err := net.Listen("tcp", ":3000")
		if err != nil {
			log.Fatal(err)
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go fs.ReadLoop(conn)
			if StopListening {
				break
			}
		}
	}
	Loop()
}

func (fs *FileServer) ReadLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	var filenamesize int64
	binary.Read(conn, binary.LittleEndian, &filenamesize)
	_, err := io.CopyN(buf, conn, filenamesize)
	if err != nil {
		log.Fatal(err)
	}
	filename = string(buf.Bytes()[:len(buf.Bytes())])
	fmt.Println("we ge fil namme: " + filename)
	buf = new(bytes.Buffer)
	var size int64
	binary.Read(conn, binary.LittleEndian, &size)
	n, err := io.CopyN(buf, conn, size)
	if err != nil {
		log.Fatal(err)
	}
	file := string(buf.Bytes()[:len(buf.Bytes())])
	fmt.Println(file)
	fmt.Printf(green.Render("Incoming file (%s), %d bytes \n"), filename, n)
	fmt.Printf(placeholder.Render("Do you want to download "+filename+"?") + " (" + placeholder.Render("y") + "/" + red.Render("n") + ") ")
	var confirm string
	fmt.Scan(&confirm)
	if strings.ToLower(confirm) == "y" {
		os.WriteFile(downloadsPath+"\\"+filename, []byte(file), 0644)
		fmt.Println("saved file to: " + downloadsPath + "\\" + filename)
	} else {
		fmt.Println("File Ignored!")
	}
	StopListening = true
}

func sendFile(filename string) error {
	fileData, err := os.Stat(filename)
	if err != nil {
		return err
	}
	size := fileData.Size()
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	reader := strings.NewReader(string(fileContent[:len(fileContent)]))
	file := make([]byte, size)
	_, err = io.ReadFull(reader, file)
	if err != nil {
		return err
	}
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}
	filenamesize := len([]byte(filename))
	binary.Write(conn, binary.LittleEndian, int64(filenamesize))
	_, err = io.CopyN(conn, bytes.NewReader([]byte(filename)), int64(filenamesize))
	binary.Write(conn, binary.LittleEndian, int64(size))
	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
	if err != nil {
		return err
	}
	fmt.Printf("ooga shaka we send de byg uga (chunke numme %d) to de reseva\n", n)
	Loop()
	return nil
}

func main() {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return
	}

	// Construct the path to the Downloads folder
	downloadsPath = filepath.Join(currentUser.HomeDir, "Downloads")
	fmt.Println(magenta.Render("Welcome to Gofile!"))
	Loop()
}

func Loop() {
	fmt.Printf(placeholder.Render("type ") + orange.Render("r") + placeholder.Render(" to recieve, ") + orange.Render("s") + placeholder.Render(" to send and ") + orange.Render("q") + placeholder.Render(" to exit "))
	var send string
	fmt.Scan(&send)
	if strings.ToLower(send) == "s" {
		sendFile("main.go")
	} else if strings.ToLower(send) == "r" {
		StopListening = false
		server := &FileServer{}
		server.start()
	} else if strings.ToLower(send) == "q" {
		os.Exit(0)
	} else {
		fmt.Println(warn.Render("Please type s or r!"))
		Loop()
	}
}
