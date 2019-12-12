package net

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	n "net"
	"peterdekok.nl/logger"
	"peterdekok.nl/net/message"
	"peterdekok.nl/trap"
	"sync"
	"sync/atomic"
	"time"
)

type Connection struct {
	messageCallback  MessageCallback
	messageSeparator string

	readBuffer  []byte
	readTimeout time.Duration

	addr *n.TCPAddr
	tcp  *n.TCPConn

	writeMux     sync.RWMutex
	writeTimeout time.Duration

	closing   int32
	closed    int32 // TODO Reconnect ??
	waitClose sync.WaitGroup
}

type MessageCallback func(msg *message.Message)

var (
	log logger.Logger
)

func init() {
	log = logger.New("net")
}

func Connect(host string, port int, messageCallback MessageCallback) (conn *Connection, err error) {
	l := log.WithField("host", host).WithField("port", port)

	l.Info("Starting new connection")

	conn = &Connection{
		messageCallback:  messageCallback,
		messageSeparator: "\r\n",

		readBuffer:  make([]byte, 0),
		readTimeout: 2 * time.Second, // TODO config/param?

		writeTimeout: 5 * time.Second, // TODO config/param?

		closing: 0,
		closed:  1,
	}

	// Ensure the connection is closed (or not open) on kill or exit
	trap.OnKill(conn.Close)

	if ip, err := n.ResolveIPAddr("", host); err != nil {
		l.WithError(err).Error("Unable to resolve host")

		return conn, err
	} else {
		conn.addr = &n.TCPAddr{
			IP:   ip.IP,
			Port: port,
		}
	}

	conn.tcp, err = n.DialTCP("tcp", nil, conn.addr)

	if err != nil {
		l.WithError(err).Error("Failed to create connection")

		return conn, err
	}

	// Right now there is no way to have a race condition here,
	// but this is the last time this is possible, as after the below goroutine is started,
	// this is no longer true
	conn.closed = 0

	conn.waitClose.Add(1)

	go conn.read(l)

	l.Info("Connection started")

	return conn, nil
}

func (conn *Connection) Close() {
	log.Info("Signaling shutdown to readers and writers")

	// Increment the closing integer, if anything is reading they will close
	atomic.AddInt32(&conn.closing, 1)

	log.Debug("Waiting on close of readers and writers")

	conn.waitClose.Wait()

	log.Debug("Closing connection complete")
}

func (conn *Connection) SetMessageSeparator(sep string) {
	conn.messageSeparator = sep
}

func (conn *Connection) Write(msg *message.Message) (err error) {
	return conn.writeBytes(msg.Marshal())
}

func (conn *Connection) writeBytes(cmd []byte) (err error) {
	conn.writeMux.Lock()
	defer conn.writeMux.Unlock()

	conn.waitClose.Add(1)
	defer conn.waitClose.Done()

	if atomic.LoadInt32(&conn.closing) > 0 || atomic.LoadInt32(&conn.closed) > 0 {
		log.Error("Writing on closed connection")

		return errors.New("connection closed")
	}

	if err := conn.tcp.SetWriteDeadline(time.Now().Add(conn.writeTimeout)); err != nil {
		log.WithError(err).Error("Failed to set write deadline")

		return err
	}

	// TODO REMOVE DEBUG About to write command, or maybe log separately??
	fmt.Printf("====================================\n%s====================================\n\n", cmd)

	if _, err := conn.tcp.Write(cmd); err != nil {
		log.WithField("cmd", cmd).WithError(err).Error("Failed to send command")

		return err
	}

	return nil
}

func (conn *Connection) read(l *logrus.Entry) {
	// Close connection when this function ends
	defer func() {
		l.Info("TCP reader stopped")
		l.Info("Closing connection")

		atomic.AddInt32(&conn.closing, 1)

		l.Debug("Waiting on writers to close")

		conn.writeMux.Lock()
		defer conn.writeMux.Unlock()

		if err := conn.tcp.Close(); err != nil {
			l.WithError(err).Error("Failed to close connection")
		}

		atomic.AddInt32(&conn.closed, 1)

		conn.waitClose.Done()

		l.Info("Connection closed")
	}()

	l.Info("Start TCP reader")

	bufReader := bufio.NewReader(conn.tcp)

	for {
		if atomic.LoadInt32(&conn.closing) > 0 {
			return
		}

		// Set a deadline for reading. Read operation will fail if no data is received after deadline.
		if err := conn.tcp.SetReadDeadline(time.Now().Add(conn.readTimeout)); err != nil {
			log.WithError(err).Error("Failed to set read deadline")

			return
		}

		// Read tokens delimited by newline and ignore any errors
		b, err := bufReader.ReadBytes('\n')
		
		if err == io.EOF {
			log.WithError(err).Warn("Connection reached EOF, probably closed by server")

			return
		} else if err, ok := err.(n.Error); ok && err.Timeout() {
			continue
		} else if err != nil {
			log.WithError(err).Warn("Failed to read tokens from connection")

			continue
		}

		// TODO Debug statement
		str, _ := json.Marshal(string(b))
		fmt.Printf(">>>%s<<<\n", str)

		conn.readBuffer = append(conn.readBuffer, b...)

		// Look for the end of the message
		if bytes.HasSuffix(conn.readBuffer, []byte(conn.messageSeparator)) {
			msg := message.Unmarshal(conn.readBuffer)

			conn.readBuffer = []byte(nil)

			conn.messageCallback(msg)
		}
	}
}
