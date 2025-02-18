package backendutil

import (
	"net"
	"os"

	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
)

type BackendApplicationHelper struct {
	Backend    BackendInterface
	SocketPath string

	socket net.Conn
}

func (helper *BackendApplicationHelper) Start() error {
	log.Debug("BackendApplicationHelper is starting")
	err := ConfigureProfiling()

	if err != nil {
		return err
	}

	log.Debug("Currently waiting for Unix socket connection...")

	helper.socket, err = net.Dial("unix", helper.SocketPath)

	if err != nil {
		return err
	}

	log.Debug("Sucessfully connected")

	for {
		commandRaw, err := commonbackend.Unmarshal(helper.socket)

		if err != nil {
			return err
		}

		switch command := commandRaw.(type) {
		case *commonbackend.Start:
			ok, err := helper.Backend.StartBackend(command.Arguments)

			var (
				message    string
				statusCode int
			)

			if err != nil {
				message = err.Error()
				statusCode = commonbackend.StatusFailure
			} else {
				statusCode = commonbackend.StatusSuccess
			}

			response := &commonbackend.BackendStatusResponse{
				IsRunning:  ok,
				StatusCode: statusCode,
				Message:    message,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.BackendStatusRequest:
			ok, err := helper.Backend.GetBackendStatus()

			var (
				message    string
				statusCode int
			)

			if err != nil {
				message = err.Error()
				statusCode = commonbackend.StatusFailure
			} else {
				statusCode = commonbackend.StatusSuccess
			}

			response := &commonbackend.BackendStatusResponse{
				IsRunning:  ok,
				StatusCode: statusCode,
				Message:    message,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.Stop:
			ok, err := helper.Backend.StopBackend()

			var (
				message    string
				statusCode int
			)

			if err != nil {
				message = err.Error()
				statusCode = commonbackend.StatusFailure
			} else {
				statusCode = commonbackend.StatusSuccess
			}

			response := &commonbackend.BackendStatusResponse{
				IsRunning:  !ok,
				StatusCode: statusCode,
				Message:    message,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.AddProxy:
			ok, err := helper.Backend.StartProxy(command)
			var hasAnyFailed bool

			if err != nil {
				log.Warnf("failed to add proxy (%s:%d -> remote:%d): %s", command.SourceIP, command.SourcePort, command.DestPort, err.Error())
				hasAnyFailed = true
			} else if !ok {
				log.Warnf("failed to add proxy (%s:%d -> remote:%d): StartProxy returned into failure state", command.SourceIP, command.SourcePort, command.DestPort)
				hasAnyFailed = true
			}

			response := &commonbackend.ProxyStatusResponse{
				SourceIP:   command.SourceIP,
				SourcePort: command.SourcePort,
				DestPort:   command.DestPort,
				Protocol:   command.Protocol,
				IsActive:   !hasAnyFailed,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.RemoveProxy:
			ok, err := helper.Backend.StopProxy(command)
			var hasAnyFailed bool

			if err != nil {
				log.Warnf("failed to remove proxy (%s:%d -> remote:%d): %s", command.SourceIP, command.SourcePort, command.DestPort, err.Error())
				hasAnyFailed = true
			} else if !ok {
				log.Warnf("failed to remove proxy (%s:%d -> remote:%d): RemoveProxy returned into failure state", command.SourceIP, command.SourcePort, command.DestPort)
				hasAnyFailed = true
			}

			response := &commonbackend.ProxyStatusResponse{
				SourceIP:   command.SourceIP,
				SourcePort: command.SourcePort,
				DestPort:   command.DestPort,
				Protocol:   command.Protocol,
				IsActive:   hasAnyFailed,
			}

			responseMarshalled, err := commonbackend.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		case *commonbackend.ProxyConnectionsRequest:
			connections := helper.Backend.GetAllClientConnections()

			serverParams := &commonbackend.ProxyConnectionsResponse{
				Connections: connections,
			}

			byteData, err := commonbackend.Marshal(serverParams)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case *commonbackend.CheckClientParameters:
			resp := helper.Backend.CheckParametersForConnections(command)
			resp.InResponseTo = "checkClientParameters"

			byteData, err := commonbackend.Marshal(resp)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case *commonbackend.CheckServerParameters:
			resp := helper.Backend.CheckParametersForBackend(command.Arguments)
			resp.InResponseTo = "checkServerParameters"

			byteData, err := commonbackend.Marshal(resp)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		default:
			log.Warnf("Unsupported command recieved: %T", command)
		}
	}
}

func NewHelper(backend BackendInterface) *BackendApplicationHelper {
	socketPath, ok := os.LookupEnv("HERMES_API_SOCK")

	if !ok {
		log.Warn("HERMES_API_SOCK is not defined! This will cause an issue unless the backend manually overwrites it")
	}

	helper := &BackendApplicationHelper{
		Backend:    backend,
		SocketPath: socketPath,
	}

	return helper
}
