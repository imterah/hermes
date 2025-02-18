package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/gaslighter"
	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type TCPProxy struct {
	proxyInformation *commonbackend.AddProxy
	connections      map[uint16]net.Conn
}

type UDPProxy struct {
	proxyInformation *commonbackend.AddProxy
}

type SSHAppBackendData struct {
	IP          string   `json:"ip" validate:"required"`
	Port        uint16   `json:"port" validate:"required"`
	Username    string   `json:"username" validate:"required"`
	PrivateKey  string   `json:"privateKey" validate:"required"`
	ListenOnIPs []string `json:"listenOnIPs"`
}

type SSHAppBackend struct {
	config      *SSHAppBackendData
	conn        *ssh.Client
	listener    net.Listener
	currentSock net.Conn

	tcpProxies map[uint16]*TCPProxy
	udpProxies map[uint16]*UDPProxy

	// globalNonCriticalMessageLock: Locks all messages that don't need low-latency transmissions & high
	// speed behind a lock. This ensures safety when it comes to handling messages correctly.
	globalNonCriticalMessageLock sync.Mutex
	// globalNonCriticalMessageChan: Channel for handling messages that need a reply / aren't critical.
	globalNonCriticalMessageChan chan interface{}
}

func (backend *SSHAppBackend) StartBackend(configBytes []byte) (bool, error) {
	log.Info("SSHAppBackend is initializing...")
	backend.globalNonCriticalMessageChan = make(chan interface{})
	backend.tcpProxies = map[uint16]*TCPProxy{}
	backend.udpProxies = map[uint16]*UDPProxy{}

	var backendData SSHAppBackendData

	if err := json.Unmarshal(configBytes, &backendData); err != nil {
		return false, err
	}

	if err := validator.New().Struct(&backendData); err != nil {
		return false, err
	}

	backend.config = &backendData

	if len(backend.config.ListenOnIPs) == 0 {
		backend.config.ListenOnIPs = []string{"0.0.0.0"}
	}

	signer, err := ssh.ParsePrivateKey([]byte(backendData.PrivateKey))

	if err != nil {
		log.Warnf("Failed to initialize: %s", err.Error())
		return false, err
	}

	auth := ssh.PublicKeys(signer)

	config := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            backendData.Username,
		Auth: []ssh.AuthMethod{
			auth,
		},
	}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", backendData.IP, backendData.Port), config)

	if err != nil {
		log.Warnf("Failed to initialize: %s", err.Error())
		return false, err
	}

	backend.conn = conn

	log.Debug("SSHAppBackend has connected successfully.")
	log.Debug("Getting CPU architecture...")

	session, err := backend.conn.NewSession()

	if err != nil {
		log.Warnf("Failed to create session: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf

	err = session.Run("uname -m")

	if err != nil {
		log.Warnf("Failed to run uname command: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	cpuArchBytes := make([]byte, stdoutBuf.Len())
	stdoutBuf.Read(cpuArchBytes)

	cpuArch := string(cpuArchBytes)
	cpuArch = cpuArch[:len(cpuArch)-1]

	var backendBinary string

	// Ordered in (subjective) popularity
	if cpuArch == "x86_64" {
		backendBinary = "remote-bin/rt-amd64"
	} else if cpuArch == "aarch64" {
		backendBinary = "remote-bin/rt-arm64"
	} else if cpuArch == "arm" {
		backendBinary = "remote-bin/rt-arm"
	} else if len(cpuArch) == 4 && string(cpuArch[0]) == "i" && strings.HasSuffix(cpuArch, "86") {
		backendBinary = "remote-bin/rt-386"
	} else {
		log.Warn("Failed to determine executable to use: CPU architecture not compiled/supported currently")
		conn.Close()
		backend.conn = nil
		return false, fmt.Errorf("CPU architecture not compiled/supported currently")
	}

	log.Debug("Checking if we need to copy the application...")

	var binary []byte
	needsToCopyBinary := true

	session, err = backend.conn.NewSession()

	if err != nil {
		log.Warnf("Failed to create session: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	session.Stdout = &stdoutBuf

	err = session.Start("[ -f /tmp/sshappbackend.runtime ] && md5sum /tmp/sshappbackend.runtime | cut -d \" \" -f 1")

	if err != nil {
		log.Warnf("Failed to calculate hash of possibly existing backend: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	fileExists := stdoutBuf.Len() != 0

	if fileExists {
		remoteMD5HashStringBuf := make([]byte, stdoutBuf.Len())
		stdoutBuf.Read(remoteMD5HashStringBuf)

		remoteMD5HashString := string(remoteMD5HashStringBuf)
		remoteMD5HashString = remoteMD5HashString[:len(remoteMD5HashString)-1]

		remoteMD5Hash, err := hex.DecodeString(remoteMD5HashString)

		if err != nil {
			log.Warnf("Failed to decode hex: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		binary, err = binFiles.ReadFile(backendBinary)

		if err != nil {
			log.Warnf("Failed to read file in the embedded FS: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, fmt.Errorf("(embedded FS): %s", err.Error())
		}

		localMD5Hash := md5.Sum(binary)

		log.Infof("remote: %s, local: %s", remoteMD5HashString, hex.EncodeToString(localMD5Hash[:]))

		if bytes.Compare(localMD5Hash[:], remoteMD5Hash) == 0 {
			needsToCopyBinary = false
		}
	}

	if needsToCopyBinary {
		log.Debug("Copying binary...")

		sftpInstance, err := sftp.NewClient(conn)

		if err != nil {
			log.Warnf("Failed to initialize SFTP: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		defer sftpInstance.Close()

		if len(binary) == 0 {
			binary, err = binFiles.ReadFile(backendBinary)

			if err != nil {
				log.Warnf("Failed to read file in the embedded FS: %s", err.Error())
				conn.Close()
				backend.conn = nil
				return false, fmt.Errorf("(embedded FS): %s", err.Error())
			}
		}

		var file *sftp.File

		if fileExists {
			file, err = sftpInstance.OpenFile("/tmp/sshappbackend.runtime", os.O_WRONLY)
		} else {
			file, err = sftpInstance.Create("/tmp/sshappbackend.runtime")
		}

		if err != nil {
			log.Warnf("Failed to create (or open) file: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		_, err = file.Write(binary)

		if err != nil {
			log.Warnf("Failed to write file: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		err = file.Chmod(0755)

		if err != nil {
			log.Warnf("Failed to change permissions on file: %s", err.Error())
			conn.Close()
			backend.conn = nil
			return false, err
		}

		log.Debug("Done copying file.")
		sftpInstance.Close()
	} else {
		log.Debug("Skipping copying as there's a copy on disk already.")
	}

	log.Debug("Initializing Unix socket...")

	socketPath := fmt.Sprintf("/tmp/sock-%d.sock", rand.Uint())
	listener, err := conn.ListenUnix(socketPath)

	if err != nil {
		log.Warnf("Failed to listen on socket: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	log.Debug("Starting process...")

	session, err = backend.conn.NewSession()

	if err != nil {
		log.Warnf("Failed to create session: %s", err.Error())
		conn.Close()
		backend.conn = nil
		return false, err
	}

	backend.listener = listener

	session.Stdout = WriteLogger{}
	session.Stderr = WriteLogger{}

	go func() {
		for {
			err := session.Run(fmt.Sprintf("HERMES_LOG_LEVEL=\"%s\" HERMES_API_SOCK=\"%s\" /tmp/sshappbackend.runtime", os.Getenv("HERMES_LOG_LEVEL"), socketPath))

			if err != nil && !errors.Is(err, &ssh.ExitError{}) && !errors.Is(err, &ssh.ExitMissingError{}) {
				log.Errorf("Critically failed during execution of remote code: %s", err.Error())
				return
			} else {
				log.Warn("Remote code failed for an unknown reason. Restarting...")
			}
		}
	}()

	go backend.sockServerHandler()

	log.Debug("Started process. Waiting for Unix socket connection...")

	for backend.currentSock == nil {
		time.Sleep(10 * time.Millisecond)
	}

	log.Debug("Detected connection. Sending initialization command...")

	proxyStatusRaw, err := backend.SendNonCriticalMessage(&commonbackend.Start{
		Arguments: []byte{},
	})

	if err != nil {
		return false, err
	}

	proxyStatus, ok := proxyStatusRaw.(*commonbackend.BackendStatusResponse)

	if !ok {
		return false, fmt.Errorf("recieved invalid response type: %T", proxyStatusRaw)
	}

	if proxyStatus.StatusCode == commonbackend.StatusFailure {
		if proxyStatus.Message == "" {
			return false, fmt.Errorf("failed to initialize backend in remote code")
		} else {
			return false, fmt.Errorf("failed to initialize backend in remote code: %s", proxyStatus.Message)
		}
	}

	log.Info("SSHAppBackend has initialized successfully.")

	return true, nil
}

func (backend *SSHAppBackend) StopBackend() (bool, error) {
	err := backend.conn.Close()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (backend *SSHAppBackend) GetBackendStatus() (bool, error) {
	return backend.conn != nil, nil
}

func (backend *SSHAppBackend) StartProxy(command *commonbackend.AddProxy) (bool, error) {
	proxyStatusRaw, err := backend.SendNonCriticalMessage(command)

	if err != nil {
		return false, err
	}

	proxyStatus, ok := proxyStatusRaw.(*datacommands.ProxyStatusResponse)

	if !ok {
		return false, fmt.Errorf("recieved invalid response type: %T", proxyStatusRaw)
	}

	if !proxyStatus.IsActive {
		return false, fmt.Errorf("failed to initialize proxy in remote code")
	}

	if command.Protocol == "tcp" {
		backend.tcpProxies[proxyStatus.ProxyID] = &TCPProxy{
			proxyInformation: command,
		}

		backend.tcpProxies[proxyStatus.ProxyID].connections = map[uint16]net.Conn{}
	} else if command.Protocol == "udp" {
		backend.udpProxies[proxyStatus.ProxyID] = &UDPProxy{
			proxyInformation: command,
		}
	}

	return true, nil
}

func (backend *SSHAppBackend) StopProxy(command *commonbackend.RemoveProxy) (bool, error) {
	if command.Protocol == "tcp" {
		for proxyIndex, proxy := range backend.tcpProxies {
			if proxy.proxyInformation.DestPort != command.DestPort {
				continue
			}

			onDisconnect := &datacommands.TCPConnectionClosed{
				ProxyID: proxyIndex,
			}

			for connectionIndex, connection := range proxy.connections {
				connection.Close()
				delete(proxy.connections, connectionIndex)

				onDisconnect.ConnectionID = connectionIndex
				disconnectionCommandMarshalled, err := datacommands.Marshal(onDisconnect)

				if err != nil {
					log.Errorf("failed to marshal disconnection message: %s", err.Error())
				}

				backend.currentSock.Write(disconnectionCommandMarshalled)
			}

			proxyStatusRaw, err := backend.SendNonCriticalMessage(&datacommands.RemoveProxy{
				ProxyID: proxyIndex,
			})

			if err != nil {
				return false, err
			}

			proxyStatus, ok := proxyStatusRaw.(*datacommands.ProxyStatusResponse)

			if !ok {
				log.Warn("Failed to stop proxy: typecast failed")
				return true, fmt.Errorf("failed to stop proxy: typecast failed")
			}

			if proxyStatus.IsActive {
				log.Warn("Failed to stop proxy: still running")
				return true, fmt.Errorf("failed to stop proxy: still running")
			}
		}
	} else if command.Protocol == "udp" {
		for proxyIndex, proxy := range backend.udpProxies {
			if proxy.proxyInformation.DestPort != command.DestPort {
				continue
			}

			proxyStatusRaw, err := backend.SendNonCriticalMessage(&datacommands.RemoveProxy{
				ProxyID: proxyIndex,
			})

			if err != nil {
				return false, err
			}

			proxyStatus, ok := proxyStatusRaw.(*datacommands.ProxyStatusResponse)

			if !ok {
				log.Warn("Failed to stop proxy: typecast failed")
				return true, fmt.Errorf("failed to stop proxy: typecast failed")
			}

			if proxyStatus.IsActive {
				log.Warn("Failed to stop proxy: still running")
				return true, fmt.Errorf("failed to stop proxy: still running")
			}

			// TODO: finish code for UDP
		}
	}

	return false, fmt.Errorf("could not find the proxy")
}

func (backend *SSHAppBackend) GetAllClientConnections() []*commonbackend.ProxyClientConnection {
	return []*commonbackend.ProxyClientConnection{}
}

func (backend *SSHAppBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHAppBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	var backendData SSHAppBackendData

	if err := json.Unmarshal(arguments, &backendData); err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("could not read json: %s", err.Error()),
		}
	}

	if err := validator.New().Struct(&backendData); err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed validation of parameters: %s", err.Error()),
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHAppBackend) OnTCPConnectionOpened(proxyID, connectionID uint16) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", backend.tcpProxies[proxyID].proxyInformation.SourceIP, backend.tcpProxies[proxyID].proxyInformation.SourcePort))

	if err != nil {
		log.Warnf("failed to dial sock: %s", err.Error())
	}

	go func() {
		dataBuf := make([]byte, 65535)

		tcpData := &datacommands.TCPProxyData{
			ProxyID:      proxyID,
			ConnectionID: connectionID,
		}

		for {
			len, err := conn.Read(dataBuf)

			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				} else if err.Error() != "EOF" {
					log.Warnf("failed to read from sock: %s", err.Error())
				}

				conn.Close()
				break
			}

			tcpData.DataLength = uint16(len)
			marshalledMessageCommand, err := datacommands.Marshal(tcpData)

			if err != nil {
				log.Warnf("failed to marshal message data: %s", err.Error())

				conn.Close()
				break
			}

			if _, err := backend.currentSock.Write(marshalledMessageCommand); err != nil {
				log.Warnf("failed to send marshalled message data: %s", err.Error())

				conn.Close()
				break
			}

			if _, err := backend.currentSock.Write(dataBuf[:len]); err != nil {
				log.Warnf("failed to send raw message data: %s", err.Error())

				conn.Close()
				break
			}
		}

		onDisconnect := &datacommands.TCPConnectionClosed{
			ProxyID:      proxyID,
			ConnectionID: connectionID,
		}

		disconnectionCommandMarshalled, err := datacommands.Marshal(onDisconnect)

		if err != nil {
			log.Errorf("failed to marshal disconnection message: %s", err.Error())
		}

		backend.currentSock.Write(disconnectionCommandMarshalled)
	}()

	backend.tcpProxies[proxyID].connections[connectionID] = conn
}

func (backend *SSHAppBackend) OnTCPConnectionClosed(proxyID, connectionID uint16) {
	proxy, ok := backend.tcpProxies[proxyID]

	if !ok {
		log.Warn("Could not find TCP proxy")
	}

	connection, ok := proxy.connections[connectionID]

	if !ok {
		log.Warn("Could not find connection in TCP proxy")
	}

	connection.Close()
	delete(proxy.connections, connectionID)
}

func (backend *SSHAppBackend) HandleTCPMessage(message *datacommands.TCPProxyData, data []byte) {
	proxy, ok := backend.tcpProxies[message.ProxyID]

	if !ok {
		log.Warn("Could not find TCP proxy")
	}

	connection, ok := proxy.connections[message.ConnectionID]

	if !ok {
		log.Warn("Could not find connection in TCP proxy")
	}

	connection.Write(data)
}

func (backend *SSHAppBackend) HandleUDPMessage(message *datacommands.UDPProxyData, data []byte) {}

func (backend *SSHAppBackend) SendNonCriticalMessage(iface interface{}) (interface{}, error) {
	if backend.currentSock == nil {
		return nil, fmt.Errorf("socket connection not initialized yet")
	}

	bytes, err := datacommands.Marshal(iface)

	if err != nil && err.Error() == "unsupported command type" {
		bytes, err = commonbackend.Marshal(iface)

		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	backend.globalNonCriticalMessageLock.Lock()

	if _, err := backend.currentSock.Write(bytes); err != nil {
		backend.globalNonCriticalMessageLock.Unlock()
		return nil, fmt.Errorf("failed to write message: %s", err.Error())
	}

	reply, ok := <-backend.globalNonCriticalMessageChan

	if !ok {
		backend.globalNonCriticalMessageLock.Unlock()
		return nil, fmt.Errorf("failed to get reply back: chan not OK")
	}

	backend.globalNonCriticalMessageLock.Unlock()
	return reply, nil
}

func (backend *SSHAppBackend) sockServerHandler() {
	for {
		conn, err := backend.listener.Accept()

		if err != nil {
			log.Warnf("Failed to accept remote connection: %s", err.Error())
		}

		log.Debug("Successfully connected.")

		backend.currentSock = conn

		commandID := make([]byte, 1)

		gaslighter := &gaslighter.Gaslighter{}
		gaslighter.ProxiedReader = conn

		dataBuffer := make([]byte, 65535)

		var commandRaw interface{}

		for {
			if _, err := conn.Read(commandID); err != nil {
				log.Warnf("Failed to read command ID: %s", err.Error())
				return
			}

			gaslighter.Byte = commandID[0]
			gaslighter.HasGaslit = false

			if gaslighter.Byte > 100 {
				commandRaw, err = datacommands.Unmarshal(gaslighter)
			} else {
				commandRaw, err = commonbackend.Unmarshal(gaslighter)
			}

			if err != nil {
				log.Warnf("Failed to parse command: %s", err.Error())
			}

			switch command := commandRaw.(type) {
			case *datacommands.TCPConnectionOpened:
				backend.OnTCPConnectionOpened(command.ProxyID, command.ConnectionID)
			case *datacommands.TCPConnectionClosed:
				backend.OnTCPConnectionClosed(command.ProxyID, command.ConnectionID)
			case *datacommands.TCPProxyData:
				if _, err := io.ReadFull(conn, dataBuffer[:command.DataLength]); err != nil {
					log.Warnf("Failed to read entire data buffer: %s", err.Error())
					break
				}

				backend.HandleTCPMessage(command, dataBuffer[:command.DataLength])
			case *datacommands.UDPProxyData:
				if _, err := io.ReadFull(conn, dataBuffer[:command.DataLength]); err != nil {
					log.Warnf("Failed to read entire data buffer: %s", err.Error())
					break
				}

				backend.HandleUDPMessage(command, dataBuffer[:command.DataLength])
			default:
				select {
				case backend.globalNonCriticalMessageChan <- command:
				default:
				}
			}
		}
	}
}

func main() {
	logLevel := os.Getenv("HERMES_LOG_LEVEL")

	if logLevel != "" {
		switch logLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)

		case "info":
			log.SetLevel(log.InfoLevel)

		case "warn":
			log.SetLevel(log.WarnLevel)

		case "error":
			log.SetLevel(log.ErrorLevel)

		case "fatal":
			log.SetLevel(log.FatalLevel)
		}
	}

	backend := &SSHAppBackend{}

	application := backendutil.NewHelper(backend)
	err := application.Start()

	if err != nil {
		log.Fatalf("failed execution in application: %s", err.Error())
	}
}
