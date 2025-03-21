package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/ssh"
)

var validatorInstance *validator.Validate

type ConnWithTimeout struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (c *ConnWithTimeout) Read(b []byte) (int, error) {
	err := c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))

	if err != nil {
		return 0, err
	}

	return c.Conn.Read(b)
}

func (c *ConnWithTimeout) Write(b []byte) (int, error) {
	err := c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))

	if err != nil {
		return 0, err
	}

	return c.Conn.Write(b)
}

type SSHListener struct {
	SourceIP   string
	SourcePort uint16
	DestPort   uint16
	Protocol   string // Will be either 'tcp' or 'udp'
	Listeners  []net.Listener
}

type SSHBackendData struct {
	IP              string   `json:"ip" validate:"required"`
	Port            uint16   `json:"port" validate:"required"`
	Username        string   `json:"username" validate:"required"`
	PrivateKey      string   `json:"privateKey" validate:"required"`
	DisablePIDCheck bool     `json:"disablePIDCheck"`
	ListenOnIPs     []string `json:"listenOnIPs"`
}

type SSHBackend struct {
	config         *SSHBackendData
	conn           *ssh.Client
	clients        []*commonbackend.ProxyClientConnection
	proxies        []*SSHListener
	arrayPropMutex sync.Mutex
	pid            int
	isReady        bool
	inReinitLoop   bool
}

func (backend *SSHBackend) StartBackend(bytes []byte) (bool, error) {
	log.Info("SSHBackend is initializing...")

	if validatorInstance == nil {
		validatorInstance = validator.New()
	}

	if backend.inReinitLoop {
		for !backend.isReady {
			time.Sleep(100 * time.Millisecond)
		}
	}

	var backendData SSHBackendData

	if err := json.Unmarshal(bytes, &backendData); err != nil {
		return false, err
	}

	if err := validatorInstance.Struct(&backendData); err != nil {
		return false, err
	}

	backend.config = &backendData

	if len(backend.config.ListenOnIPs) == 0 {
		backend.config.ListenOnIPs = []string{"0.0.0.0"}
	}

	signer, err := ssh.ParsePrivateKey([]byte(backendData.PrivateKey))

	if err != nil {
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

	addr := fmt.Sprintf("%s:%d", backendData.IP, backendData.Port)
	timeout := time.Duration(10 * time.Second)

	rawTCPConn, err := net.DialTimeout("tcp", addr, timeout)

	if err != nil {
		return false, err
	}

	connWithTimeout := &ConnWithTimeout{
		Conn:         rawTCPConn,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	c, chans, reqs, err := ssh.NewClientConn(connWithTimeout, addr, config)

	if err != nil {
		return false, err
	}

	client := ssh.NewClient(c, chans, reqs)
	backend.conn = client

	if !backendData.DisablePIDCheck {
		if backend.pid != 0 {
			session, err := client.NewSession()

			if err != nil {
				return false, err
			}

			err = session.Run(fmt.Sprintf("kill -9 %d", backend.pid))

			if err != nil {
				log.Warnf("Failed to kill process: %s", err.Error())
			}
		}

		session, err := client.NewSession()

		if err != nil {
			return false, err
		}

		// Get the parent PID of the shell so we can kill it if we disconnect
		output, err := session.Output("ps --no-headers -fp $$ | awk '{print $3}'")

		if err != nil {
			return false, err
		}

		// Strip the new line and convert to int
		backend.pid, err = strconv.Atoi(string(output)[:len(output)-1])

		if err != nil {
			return false, err
		}
	}

	go backend.backendDisconnectHandler()
	go backend.backendKeepaliveHandler()

	log.Info("SSHBackend has initialized successfully.")

	return true, nil
}

func (backend *SSHBackend) StopBackend() (bool, error) {
	err := backend.conn.Close()

	if err != nil {
		return false, err
	}

	return true, nil
}

func (backend *SSHBackend) GetBackendStatus() (bool, error) {
	return backend.conn != nil, nil
}

func (backend *SSHBackend) StartProxy(command *commonbackend.AddProxy) (bool, error) {
	listenerObject := &SSHListener{
		SourceIP:   command.SourceIP,
		SourcePort: command.SourcePort,
		DestPort:   command.DestPort,
		Protocol:   command.Protocol,
		Listeners:  []net.Listener{},
	}

	for _, ipListener := range backend.config.ListenOnIPs {
		ip := net.TCPAddr{
			IP:   net.ParseIP(ipListener),
			Port: int(command.DestPort),
		}

		listener, err := backend.conn.ListenTCP(&ip)

		if err != nil {
			// Incase we error out, we clean up all the other listeners
			for _, listener := range listenerObject.Listeners {
				err = listener.Close()

				if err != nil {
					log.Warnf("failed to close listener upon failure cleanup: %s", err.Error())
				}
			}

			return false, err
		}

		listenerObject.Listeners = append(listenerObject.Listeners, listener)

		go func() {
			for {
				forwardedConn, err := listener.Accept()

				if err != nil {
					log.Warnf("failed to accept listener connection: %s", err.Error())

					if err.Error() == "EOF" {
						return
					}

					continue
				}

				sourceConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", command.SourceIP, command.SourcePort))

				if err != nil {
					log.Warnf("failed to dial source connection: %s", err.Error())
					continue
				}

				clientIPAndPort := forwardedConn.RemoteAddr().String()
				clientIP := clientIPAndPort[:strings.LastIndex(clientIPAndPort, ":")]
				clientPort, err := strconv.Atoi(clientIPAndPort[strings.LastIndex(clientIPAndPort, ":")+1:])

				if err != nil {
					log.Warnf("failed to parse client port: %s", err.Error())
					continue
				}

				advertisedConn := &commonbackend.ProxyClientConnection{
					SourceIP:   command.SourceIP,
					SourcePort: command.SourcePort,
					DestPort:   command.DestPort,
					ClientIP:   clientIP,
					ClientPort: uint16(clientPort),

					// FIXME (imterah): shouldn't protocol be in here?
					// Protocol:   command.Protocol,
				}

				backend.arrayPropMutex.Lock()
				backend.clients = append(backend.clients, advertisedConn)
				backend.arrayPropMutex.Unlock()

				cleanupJob := func() {
					defer backend.arrayPropMutex.Unlock()
					err := sourceConn.Close()

					if err != nil {
						log.Warnf("failed to close source connection: %s", err.Error())
					}

					err = forwardedConn.Close()

					if err != nil {
						log.Warnf("failed to close forwarded/proxied connection: %s", err.Error())
					}

					backend.arrayPropMutex.Lock()

					for clientIndex, clientInstance := range backend.clients {
						// Check if memory addresses are equal for the pointer
						if clientInstance == advertisedConn {
							// Splice out the clientInstance by clientIndex

							// TODO: change approach. It works but it's a bit wonky imho
							backend.clients = slices.Delete(backend.clients, clientIndex, clientIndex+1)
							return
						}
					}

					log.Warn("failed to delete client from clients metadata: couldn't find client in the array")
				}

				sourceBuffer := make([]byte, 65535)
				forwardedBuffer := make([]byte, 65535)

				go func() {
					defer cleanupJob()

					for {
						len, err := forwardedConn.Read(forwardedBuffer)

						if err != nil {
							if err.Error() != "EOF" && !errors.Is(err, net.ErrClosed) {
								log.Errorf("failed to read from forwarded connection: %s", err.Error())
							}

							return
						}

						if _, err = sourceConn.Write(forwardedBuffer[:len]); err != nil {
							if err.Error() != "EOF" && !errors.Is(err, net.ErrClosed) {
								log.Errorf("failed to write to source connection: %s", err.Error())
							}

							return
						}
					}
				}()

				go func() {
					defer cleanupJob()

					for {
						len, err := sourceConn.Read(sourceBuffer)

						if err != nil {
							if err.Error() != "EOF" && !errors.Is(err, net.ErrClosed) {
								log.Errorf("failed to read from source connection: %s", err.Error())
							}

							return
						}

						if _, err = forwardedConn.Write(sourceBuffer[:len]); err != nil {
							if err.Error() != "EOF" && !errors.Is(err, net.ErrClosed) {
								log.Errorf("failed to write to forwarded connection: %s", err.Error())
							}

							return
						}
					}
				}()
			}
		}()
	}

	backend.arrayPropMutex.Lock()
	backend.proxies = append(backend.proxies, listenerObject)
	backend.arrayPropMutex.Unlock()

	return true, nil
}

func (backend *SSHBackend) StopProxy(command *commonbackend.RemoveProxy) (bool, error) {
	defer backend.arrayPropMutex.Unlock()
	backend.arrayPropMutex.Lock()

	for proxyIndex, proxy := range backend.proxies {
		if command.SourceIP == proxy.SourceIP && command.SourcePort == proxy.SourcePort && command.DestPort == proxy.DestPort && command.Protocol == proxy.Protocol {
			for _, listener := range proxy.Listeners {
				err := listener.Close()

				if err != nil {
					log.Warnf("failed to stop listener in StopProxy: %s", err.Error())
				}
			}

			// Splice out the proxy instance by proxyIndex
			// TODO: change approach. It works but it's a bit wonky imho
			backend.proxies = slices.Delete(backend.proxies, proxyIndex, proxyIndex+1)
			return true, nil
		}
	}

	return false, fmt.Errorf("could not find the proxy")
}

func (backend *SSHBackend) GetAllClientConnections() []*commonbackend.ProxyClientConnection {
	defer backend.arrayPropMutex.Unlock()
	backend.arrayPropMutex.Lock()

	return backend.clients
}

func (backend *SSHBackend) CheckParametersForConnections(clientParameters *commonbackend.CheckClientParameters) *commonbackend.CheckParametersResponse {
	if clientParameters.Protocol != "tcp" {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: "Only TCP is supported for SSH",
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHBackend) CheckParametersForBackend(arguments []byte) *commonbackend.CheckParametersResponse {
	var backendData SSHBackendData

	if validatorInstance == nil {
		validatorInstance = validator.New()
	}

	if err := json.Unmarshal(arguments, &backendData); err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("could not read json: %s", err.Error()),
		}
	}

	if err := validatorInstance.Struct(&backendData); err != nil {
		return &commonbackend.CheckParametersResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed validation of parameters: %s", err.Error()),
		}
	}

	return &commonbackend.CheckParametersResponse{
		IsValid: true,
	}
}

func (backend *SSHBackend) backendKeepaliveHandler() {
	for {
		if backend.conn != nil {
			_, _, err := backend.conn.SendRequest("keepalive@openssh.com", true, nil)

			if err != nil {
				log.Warn("Keepalive message failed!")
				return
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func (backend *SSHBackend) backendDisconnectHandler() {
	for {
		if backend.conn != nil {
			backend.conn.Wait()
			backend.conn.Close()

			backend.isReady = false
			backend.inReinitLoop = true

			log.Info("Disconnected from the remote SSH server. Attempting to reconnect in 5 seconds...")
		} else {
			log.Info("Retrying reconnection in 5 seconds...")
		}

		time.Sleep(5 * time.Second)

		// Make the connection nil to accurately report our status incase GetBackendStatus is called
		backend.conn = nil

		// Use the last half of the code from the main initialization
		signer, err := ssh.ParsePrivateKey([]byte(backend.config.PrivateKey))

		if err != nil {
			log.Errorf("Failed to parse private key: %s", err.Error())
			return
		}

		auth := ssh.PublicKeys(signer)

		config := &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			User:            backend.config.Username,
			Auth: []ssh.AuthMethod{
				auth,
			},
		}

		addr := fmt.Sprintf("%s:%d", backend.config.IP, backend.config.Port)
		timeout := time.Duration(10 * time.Second)

		rawTCPConn, err := net.DialTimeout("tcp", addr, timeout)

		if err != nil {
			log.Errorf("Failed to establish connection to the server: %s", err.Error())
			continue
		}

		connWithTimeout := &ConnWithTimeout{
			Conn:         rawTCPConn,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		}

		c, chans, reqs, err := ssh.NewClientConn(connWithTimeout, addr, config)

		if err != nil {
			log.Errorf("Failed to create SSH client connection: %s", err.Error())
			rawTCPConn.Close()
			continue
		}

		client := ssh.NewClient(c, chans, reqs)
		backend.conn = client

		if !backend.config.DisablePIDCheck {
			if backend.pid != 0 {
				session, err := client.NewSession()

				if err != nil {
					log.Warnf("Failed to create SSH command session: %s", err.Error())
					return
				}

				err = session.Run(fmt.Sprintf("kill -9 %d", backend.pid))

				if err != nil {
					log.Warnf("Failed to kill process: %s", err.Error())
				}
			}

			session, err := client.NewSession()

			if err != nil {
				log.Warnf("Failed to create SSH command session: %s", err.Error())
				return
			}

			// Get the parent PID of the shell so we can kill it if we disconnect
			output, err := session.Output("ps --no-headers -fp $$ | awk '{print $3}'")

			if err != nil {
				log.Warnf("Failed to execute command to fetch PID: %s", err.Error())
				return
			}

			// Strip the new line and convert to int
			backend.pid, err = strconv.Atoi(string(output)[:len(output)-1])

			if err != nil {
				log.Warnf("Failed to parse PID: %s", err.Error())
				return
			}
		}

		go backend.backendKeepaliveHandler()

		log.Info("SSHBackend has reconnected successfully. Attempting to set up proxies again...")

		for _, proxy := range backend.proxies {
			ok, err := backend.StartProxy(&commonbackend.AddProxy{
				SourceIP:   proxy.SourceIP,
				SourcePort: proxy.SourcePort,
				DestPort:   proxy.DestPort,
				Protocol:   proxy.Protocol,
			})

			if err != nil {
				log.Errorf("Failed to set up proxy: %s", err.Error())
				continue
			}

			if !ok {
				log.Errorf("Failed to set up proxy: OK status is false")
				continue
			}
		}

		log.Info("SSHBackend has reinitialized and restored state successfully.")
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

	backend := &SSHBackend{}

	application := backendutil.NewHelper(backend)
	err := application.Start()

	if err != nil {
		log.Fatalf("failed execution in application: %s", err.Error())
	}
}
