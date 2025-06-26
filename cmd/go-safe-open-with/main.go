package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/DeRuina/timberjack"
	"github.com/stephanr/go-safe-open-with/config"
)

// constants for Logger
var (
	// Trace logs general information messages.
	Trace *log.Logger
	// Error logs error messages.
	Error *log.Logger
)

const (
	appVersion string = "1.0.1"
	subVersion string = "3"
)

type IncomingMessage struct {
	Command       string   `json:"cmd"`
	ExecCommand   any      `json:"command"`
	ExecKill      bool     `json:"kill"`
	ExecArguments []string `json:"arguments"`
}

// OutgoingMessage respresents a response to an incoming message query.
type OutgoingMessage struct {
	Error   string `json:"error,omitempty"`
	Command string `json:"cmd,omitempty"`
	Code    uint32 `json:"code,omitempty"`
	Version string `json:"version,omitempty"`
}

type OutErrorMessage struct {
	Error   string `json:"error"`
	Command string `json:"cmd,omitempty"`
	Code    uint32 `json:"code"`
}

type OutEnvMessage struct {
	Environment map[string]string `json:"env"`
}

type OutSpecMessage struct {
	Version     string            `json:"version"`
	Environment map[string]string `json:"env"`
	Separator   string            `json:"separator"`
	TmpDir      string            `json:"tmpdir"`
}

type OutExecMessage struct {
	Code   uint32 `json:"code"`
	StdOut string `json:"stdout"`
	StdErr string `json:"stderr"`
}

// bufferSize used to set size of IO buffer - adjust to accommodate message payloads
var bufferSize = 8192

// nativeEndian used to detect native byte order
var nativeEndian binary.ByteOrder

var configuration config.Configuration

func Init(traceHandle io.Writer, errorHandle io.Writer) {
	Trace = log.New(traceHandle, "TRACE: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// determine native byte order so that we can read message size correctly
	nativeEndian = binary.NativeEndian

	configuration = config.ReadConfiguration()
	Trace.Printf("Loaded Configuration with %v values", len(configuration.Allowed))
}

func main() {
	logger := &timberjack.Logger{
		Filename:         "chrome-native-host-log.log", // Choose an appropriate path
		MaxSize:          15,                           // megabytes
		MaxBackups:       3,                            // backups
		MaxAge:           28,                           // days
		Compress:         false,                        // default: false
		LocalTime:        true,                         // default: false (use UTC)
		RotationInterval: time.Hour * 24,               // Rotate daily if no other rotation met
		RotateAtMinutes:  []int{0, 15, 30, 45},         // Also rotate at HH:00, HH:15, HH:30, HH:45
	}

	Init(logger, logger)
	// ensure we close the log file when we're done
	defer logger.Close()

	Trace.Printf("Chrome native messaging host started. Native byte order: %v. Build %v", nativeEndian, subVersion)
	read()
	Trace.Print("Chrome native messaging host exited.")
}

// read Creates a new buffered I/O reader and reads messages from Stdin.
func read() {
	v := bufio.NewReader(os.Stdin)
	// adjust buffer size to accommodate your json payload size limits; default is 4096
	s := bufio.NewReaderSize(v, bufferSize)
	Trace.Printf("IO buffer reader created with buffer size of %v.", s.Size())

	lengthBytes := make([]byte, 4)
	lengthNum := int(0)

	// we're going to indefinitely read the first 4 bytes in buffer, which gives us the message length.
	// if stdIn is closed we'll exit the loop and shut down host
	cnt := 0
	for b, err := s.Read(lengthBytes); b > 0 && err == nil; b, err = s.Read(lengthBytes) {
		Trace.Printf("Read Loop %v", cnt)
		cnt++

		// convert message length bytes to integer value
		lengthNum = readMessageLength(lengthBytes)
		Trace.Printf("Message size in bytes: %v", lengthNum)

		// If message length exceeds size of buffer, the message will be truncated.
		// This will likely cause an error when we attempt to unmarshal message to JSON.
		if lengthNum > bufferSize {
			Error.Printf("Message size of %d exceeds buffer size of %d. Message will be truncated and is unlikely to unmarshal to JSON.", lengthNum, bufferSize)
		}

		// read the content of the message from buffer
		content := make([]byte, lengthNum)
		_, err := s.Read(content)
		if err != nil && err != io.EOF {
			Error.Fatal(err)
		}

		// message has been read, now parse and process
		parseMessage(content)
	}

	Trace.Print("Stdin closed.")
}

// readMessageLength reads and returns the message length value in native byte order.
func readMessageLength(msg []byte) int {
	var length uint32
	buf := bytes.NewBuffer(msg)
	err := binary.Read(buf, nativeEndian, &length)
	if err != nil {
		Error.Printf("Unable to read bytes representing message length: %v", err)
	}
	return int(length)
}

func appendSplit(arg string, splitOnSpace bool) []string {
	if splitOnSpace {
		return strings.Split(arg, " ")
	} else {
		return []string{arg}
	}
}

func validateIncoming(msg IncomingMessage, cmd string) []string {

ALLOWED:
	for allowedIdx, allowed := range configuration.Allowed {
		res := []string{}
		if cmd != allowed.Command {
			continue
		}
		res = append(res, cmd)
		if len(msg.ExecArguments) != len(allowed.Arguments) {
			continue
		}
		for argIndex, argument := range allowed.Arguments {
			value := msg.ExecArguments[argIndex]
			for _, cut := range argument.TrimLeft {
				value = strings.TrimPrefix(value, cut)
			}
			for _, cut := range argument.TrimRight {
				value = strings.TrimSuffix(value, cut)
			}

			switch argument.Type {
			case "string":
				if len(argument.Value) == 0 || value != argument.Value {
					Trace.Printf("%v-%v Value '%v' not allowed", allowedIdx, argIndex, value)
					continue ALLOWED
				}
			case "list":
				if len(argument.Values) == 0 || !slices.Contains(argument.Values, value) {
					Trace.Printf("%v-%v Value '%v' not in list", allowedIdx, argIndex, value)
					continue ALLOWED
				}
			case "url":
				url, err := url.Parse(value)
				if err != nil {
					Trace.Printf("%v-%v URL '%s' could not be parsed: %v", allowedIdx, argIndex, value, err)
					continue ALLOWED
				}
				if url.Scheme != "http" && url.Scheme != "https" {
					Trace.Printf("%v-%v URL '%s' is not of type http/https", allowedIdx, argIndex, value)
					continue ALLOWED
				}
			default:
				continue
			}
			res = append(res, argument.InsertBefore...)
			res = append(res, appendSplit(value, argument.SplitSpace)...)
			res = append(res, argument.InsertAfter...)
		}
		Trace.Printf("%v Call matched", allowedIdx)
		return res
	}
	return []string{}
}

// parseMessage parses incoming message
func parseMessage(msg []byte) {
	iMsg := decodeMessage(msg)
	Trace.Printf("Message received: %s", msg)

	switch iMsg.Command {
	case "version":
		send(OutgoingMessage{
			Version: appVersion,
		})
	case "exec":
		_, cmdIsString := iMsg.ExecCommand.(string)
		cmdAndArgs := validateIncoming(iMsg, iMsg.ExecCommand.(string))
		if cmdIsString && len(cmdAndArgs) > 0 {

			cmd := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)
			cmd.Env = os.Environ()

			Trace.Printf("Executing command: %v", cmd.String())

			stdout, err := cmd.Output()
			if err != nil {
				errorMsg := "Command failed without output"
				if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
					errorMsg = string(ee.Stderr)
					Trace.Printf("Exit Code: %v", ee.ExitCode())
				}

				send(OutExecMessage{
					Code:   5,
					StdOut: string(stdout),
					StdErr: errorMsg,
				})
			} else {
				send(OutExecMessage{
					Code:   0,
					StdOut: string(stdout),
					StdErr: "",
				})
			}
		} else {
			send(OutErrorMessage{
				Error: "Usafe exec detected, ignored",
				Code:  1000,
			})
		}
	case "env":
		envMsg := OutEnvMessage{}
		envMsg.Environment = make(map[string]string)
		envMsg.Environment["TMP"] = os.Getenv("TMP")
		envMsg.Environment["TEMP"] = os.Getenv("TEMP")
		send(envMsg)
	case "spec":
		specMsg := OutSpecMessage{}
		specMsg.Version = appVersion
		specMsg.Environment = make(map[string]string)
		specMsg.Environment["TMP"] = os.Getenv("TMP")
		specMsg.Environment["TEMP"] = os.Getenv("TEMP")
		specMsg.TmpDir = os.Getenv("TMP")
		specMsg.Separator = string(os.PathSeparator)
		send(specMsg)
	case "echo":
	case "spawn":
	case "clean-tmp":
	case "ifup":
	case "dir":
	case "save-data":
	case "net":
	case "copy":
	case "remove":
	case "move":
	default:
		send(OutErrorMessage{
			Error:   "cmd is unknown",
			Command: iMsg.Command,
			Code:    1000,
		})
	}

}

// decodeMessage unmarshals incoming json request and returns query value.
func decodeMessage(msg []byte) IncomingMessage {
	var iMsg IncomingMessage
	err := json.Unmarshal(msg, &iMsg)
	if err != nil {
		Error.Printf("Unable to unmarshal json to struct: %v", err)
	}
	return iMsg
}

// send sends an OutgoingMessage to os.Stdout.
func send(msg any) {
	byteMsg := dataToBytes(msg)

	Trace.Printf("Response to send: %s", byteMsg)

	writeMessageLength(byteMsg)

	var msgBuf bytes.Buffer
	_, err := msgBuf.Write(byteMsg)
	if err != nil {
		Error.Printf("Unable to write message length to message buffer: %v", err)
	}

	_, err = msgBuf.WriteTo(os.Stdout)
	if err != nil {
		Error.Printf("Unable to write message buffer to Stdout: %v", err)
	}
}

// dataToBytes marshals OutgoingMessage struct to slice of bytes
func dataToBytes(msg any) []byte {
	byteMsg, err := json.Marshal(msg)
	if err != nil {
		Error.Printf("Unable to marshal OutgoingMessage struct to slice of bytes: %v", err)
	}
	return byteMsg
}

// writeMessageLength determines length of message and writes it to os.Stdout.
func writeMessageLength(msg []byte) {
	err := binary.Write(os.Stdout, nativeEndian, uint32(len(msg)))
	if err != nil {
		Error.Printf("Unable to write message length to Stdout: %v", err)
	}
}
