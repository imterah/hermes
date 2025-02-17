package backendutil_custom

import (
	"net"
	"os"

	"git.terah.dev/imterah/hermes/backend/backendutil"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"git.terah.dev/imterah/hermes/backend/sshappbackend/datacommands"
	"github.com/charmbracelet/log"
)

type BackendApplicationHelper struct {
	Backend    BackendInterface
	SocketPath string

	socket net.Conn
}

func (helper *BackendApplicationHelper) Start() error {
	log.Debug("BackendApplicationHelper is starting")
	err := backendutil.ConfigureProfiling()

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
		commandRaw, err := datacommands.Unmarshal(helper.socket)

		if err != nil && err.Error() != "couldn't match command ID" {
			return err
		}

		switch command := commandRaw.(type) {
		case *datacommands.ProxyConnectionsRequest:
			connections := helper.Backend.GetAllClientConnections(command.ProxyID)

			serverParams := &datacommands.ProxyConnectionsResponse{
				Connections: connections,
			}

			byteData, err := datacommands.Marshal(serverParams)

			if err != nil {
				return err
			}

			if _, err = helper.socket.Write(byteData); err != nil {
				return err
			}
		case *datacommands.RemoveProxy:
			ok, err := helper.Backend.StopProxy(command)
			var hasAnyFailed bool

			if !ok {
				log.Warnf("failed to remove proxy (ID %d): RemoveProxy returned into failure state", command.ProxyID)
				hasAnyFailed = true
			} else if err != nil {
				log.Warnf("failed to remove proxy (ID %d): %s", command.ProxyID, err.Error())
				hasAnyFailed = true
			}

			response := &datacommands.ProxyStatusResponse{
				ProxyID:  command.ProxyID,
				IsActive: hasAnyFailed,
			}

			responseMarshalled, err := datacommands.Marshal(response)

			if err != nil {
				log.Error("failed to marshal response: %s", err.Error())
				continue
			}

			helper.socket.Write(responseMarshalled)
		default:
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
			case *commonbackend.AddProxy:
				id, ok, err := helper.Backend.StartProxy(command)
				var hasAnyFailed bool

				if !ok {
					log.Warnf("failed to add proxy (%s:%d -> remote:%d): StartProxy returned into failure state", command.SourceIP, command.SourcePort, command.DestPort)
					hasAnyFailed = true
				} else if err != nil {
					log.Warnf("failed to add proxy (%s:%d -> remote:%d): %s", command.SourceIP, command.SourcePort, command.DestPort, err.Error())
					hasAnyFailed = true
				}

				response := &datacommands.ProxyStatusResponse{
					ProxyID:  id,
					IsActive: !hasAnyFailed,
				}

				responseMarshalled, err := datacommands.Marshal(response)

				if err != nil {
					log.Error("failed to marshal response: %s", err.Error())
					continue
				}

				helper.socket.Write(responseMarshalled)
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
