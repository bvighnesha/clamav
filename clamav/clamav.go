package clamav

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

const (
	RES_OK          = "OK"
	RES_FOUND       = "FOUND"
	RES_ERROR       = "ERROR"
	RES_PARSE_ERROR = "PARSE ERROR"
)

type ClamAV struct {
	address string
}

type Stats struct {
	Pools    string
	State    string
	Threads  string
	Memstats string
	Queue    string
}

type ScanResult struct {
	Raw         string
	Description string
	Path        string
	Hash        string
	Size        int
	Status      string
}

func (c *ClamAV) newConnection() (conn *Connection, err error) {

	var u *url.URL

	if u, err = url.Parse(c.address); err != nil {
		return
	}

	switch u.Scheme {
	case "tcp":
		conn, err = newCLAMDTcpConn(u.Host)
	case "unix":
		conn, err = newCLAMDUnixConn(u.Path)
	default:
		conn, err = newCLAMDUnixConn(c.address)
	}

	return
}

func (c *ClamAV) simpleCommand(command string) (chan *ScanResult, error) {
	conn, err := c.newConnection()
	if err != nil {
		return nil, err
	}

	err = conn.sendCommand(command)
	if err != nil {
		return nil, err
	}

	ch, wg, err := conn.readResponse()

	go func() {
		wg.Wait()
		conn.Close()
	}()

	return ch, err
}

/*
Check the daemon's state (should reply with PONG).
*/
func (c *ClamAV) Ping() error {
	ch, err := c.simpleCommand("PING")
	if err != nil {
		return err
	}

	select {
	case s := (<-ch):
		switch s.Raw {
		case "PONG":
			return nil
		default:
			return errors.New(fmt.Sprintf("Invalid response, got %s.", s))
		}
	}

	return nil
}

/*
Print program and database versions.
*/
func (c *ClamAV) Version() (chan *ScanResult, error) {
	dataArrays, err := c.simpleCommand("VERSION")
	return dataArrays, err
}

/*
On this command clamd provides statistics about the scan queue, contents of scan
queue, and memory usage. The exact reply format is subject to changes in future
releases.
*/
func (c *ClamAV) Stats() (*Stats, error) {
	ch, err := c.simpleCommand("STATS")
	if err != nil {
		return nil, err
	}

	stats := &Stats{}

	for s := range ch {
		if strings.HasPrefix(s.Raw, "POOLS") {
			stats.Pools = strings.Trim(s.Raw[6:], " ")
		} else if strings.HasPrefix(s.Raw, "STATE") {
			stats.State = s.Raw
		} else if strings.HasPrefix(s.Raw, "THREADS") {
			stats.Threads = s.Raw
		} else if strings.HasPrefix(s.Raw, "QUEUE") {
			stats.Queue = s.Raw
		} else if strings.HasPrefix(s.Raw, "MEMSTATS") {
			stats.Memstats = s.Raw
		} else if strings.HasPrefix(s.Raw, "END") {
		} else {
			//	return nil, errors.New(fmt.Sprintf("Unknown response, got %s.", s))
		}
	}

	return stats, nil
}

/*
Reload the databases.
*/
func (c *ClamAV) Reload() error {
	ch, err := c.simpleCommand("RELOAD")
	if err != nil {
		return err
	}

	select {
	case s := (<-ch):
		switch s.Raw {
		case "RELOADING":
			return nil
		default:
			return errors.New(fmt.Sprintf("Invalid response, got %s.", s))
		}
	}

	return nil
}

func (c *ClamAV) Shutdown() error {
	_, err := c.simpleCommand("SHUTDOWN")
	if err != nil {
		return err
	}

	return err
}

/*
Scan file or directory (recursively) with archive support enabled (a full path is
required).
*/
func (c *ClamAV) ScanFile(path string) (chan *ScanResult, error) {
	command := fmt.Sprintf("SCAN %s", path)
	ch, err := c.simpleCommand(command)
	return ch, err
}

/*
Scan file or directory (recursively) with archive and special file support disabled
(a full path is required).
*/
func (c *ClamAV) RawScanFile(path string) (chan *ScanResult, error) {
	command := fmt.Sprintf("RAWSCAN %s", path)
	ch, err := c.simpleCommand(command)
	return ch, err
}

/*
Scan file in a standard way or scan directory (recursively) using multiple threads
(to make the scanning faster on SMP machines).
*/
func (c *ClamAV) MultiScanFile(path string) (chan *ScanResult, error) {
	command := fmt.Sprintf("MULTISCAN %s", path)
	ch, err := c.simpleCommand(command)
	return ch, err
}

/*
Scan file or directory (recursively) with archive support enabled and don’t stop
the scanning when a virus is found.
*/
func (c *ClamAV) ContScanFile(path string) (chan *ScanResult, error) {
	command := fmt.Sprintf("CONTSCAN %s", path)
	ch, err := c.simpleCommand(command)
	return ch, err
}

/*
Scan file or directory (recursively) with archive support enabled and don’t stop
the scanning when a virus is found.
*/
func (c *ClamAV) AllMatchScanFile(path string) (chan *ScanResult, error) {
	command := fmt.Sprintf("ALLMATCHSCAN %s", path)
	ch, err := c.simpleCommand(command)
	return ch, err
}

/*
Scan a stream of data. The stream is sent to clamd in chunks, after INSTREAM,
on the same socket on which the command was sent. This avoids the overhead
of establishing new TCP connections and problems with NAT. The format of the
chunk is: <length><data> where <length> is the size of the following data in
bytes expressed as a 4 byte unsigned integer in network byte order and <data> is
the actual chunk. Streaming is terminated by sending a zero-length chunk. Note:
do not exceed StreamMaxLength as defined in clamd.conf, otherwise clamd will
reply with INSTREAM size limit exceeded and close the connection
*/
func (c *ClamAV) ScanStream(r io.Reader, abort chan bool) (chan *ScanResult, error) {
	conn, err := c.newConnection()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			_, allowRunning := <-abort
			if !allowRunning {
				break
			}
		}
		conn.Close()
	}()

	conn.sendCommand("INSTREAM")

	for {
		buf := make([]byte, CHUNK_SIZE)

		nr, err := r.Read(buf)
		if nr > 0 {
			conn.sendChunk(buf[0:nr])
		}

		if err != nil {
			break
		}

	}

	err = conn.sendEOF()
	if err != nil {
		return nil, err
	}

	ch, wg, err := conn.readResponse()

	go func() {
		wg.Wait()
		conn.Close()
	}()

	return ch, nil
}

func NewClamd(address string) *ClamAV {
	clamd := &ClamAV{address: address}
	return clamd
}