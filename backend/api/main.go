package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"git.terah.dev/imterah/hermes/backend/api/backendruntime"
	"git.terah.dev/imterah/hermes/backend/api/controllers/v1/backends"
	"git.terah.dev/imterah/hermes/backend/api/controllers/v1/proxies"
	"git.terah.dev/imterah/hermes/backend/api/controllers/v1/users"
	"git.terah.dev/imterah/hermes/backend/api/db"
	"git.terah.dev/imterah/hermes/backend/api/jwt"
	"git.terah.dev/imterah/hermes/backend/api/state"
	"git.terah.dev/imterah/hermes/backend/commonbackend"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"
)

func apiEntrypoint(cCtx *cli.Context) error {
	developmentMode := false

	if os.Getenv("HERMES_DEVELOPMENT_MODE") != "" {
		log.Warn("You have development mode enabled. This may weaken security.")
		developmentMode = true
	}

	log.Info("Hermes is initializing...")
	log.Debug("Initializing database and opening it...")

	databaseBackendName := os.Getenv("HERMES_DATABASE_BACKEND")
	var databaseBackendParams string

	if databaseBackendName == "sqlite" {
		databaseBackendParams = os.Getenv("HERMES_SQLITE_FILEPATH")

		if databaseBackendParams == "" {
			log.Fatal("HERMES_SQLITE_FILEPATH is not set")
		}
	}

	if databaseBackendName == "postgres" {
		databaseBackendParams = os.Getenv("HERMES_POSTGRES_DSN")

		if databaseBackendParams == "" {
			log.Fatal("HERMES_POSTGRES_DSN is not set")
		}
	}

	dbInstance, err := db.New(databaseBackendName, databaseBackendParams)

	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err)
	}

	log.Debug("Running database migrations...")

	if err := dbInstance.DoMigrations(); err != nil {
		return fmt.Errorf("Failed to run database migrations: %s", err)
	}

	log.Debug("Initializing the JWT subsystem...")

	jwtDataString := os.Getenv("HERMES_JWT_SECRET")
	var jwtKey []byte
	var jwtValidityTimeDuration time.Duration

	if jwtDataString == "" {
		log.Fatalf("HERMES_JWT_SECRET is not set")
	}

	if os.Getenv("HERMES_JWT_BASE64_ENCODED") != "" {
		jwtKey, err = base64.StdEncoding.DecodeString(jwtDataString)

		if err != nil {
			log.Fatalf("Failed to decode base64 JWT: %s", err.Error())
		}
	} else {
		jwtKey = []byte(jwtDataString)
	}

	if developmentMode {
		jwtValidityTimeDuration = jwt.DevelopmentModeTimings
	} else {
		jwtValidityTimeDuration = jwt.NormalModeTimings
	}

	jwtInstance := jwt.New(jwtKey, dbInstance, jwtValidityTimeDuration)

	log.Debug("Initializing the backend subsystem...")

	backendMetadataPath := cCtx.String("backends-path")
	backendMetadata, err := os.ReadFile(backendMetadataPath)

	if err != nil {
		return fmt.Errorf("Failed to read backends: %s", err.Error())
	}

	availableBackends := []*backendruntime.Backend{}
	err = json.Unmarshal(backendMetadata, &availableBackends)

	if err != nil {
		return fmt.Errorf("Failed to parse backends: %s", err.Error())
	}

	for _, backend := range availableBackends {
		backend.Path = path.Join(filepath.Dir(backendMetadataPath), backend.Path)
	}

	backendruntime.Init(availableBackends)

	log.Debug("Enumerating backends...")

	backendList := []db.Backend{}

	if err := dbInstance.DB.Find(&backendList).Error; err != nil {
		return fmt.Errorf("Failed to enumerate backends: %s", err.Error())
	}

	for _, backend := range backendList {
		log.Infof("Starting up backend #%d: %s", backend.ID, backend.Name)

		var backendRuntimeFilePath string

		for _, runtime := range backendruntime.AvailableBackends {
			if runtime.Name == backend.Backend {
				backendRuntimeFilePath = runtime.Path
			}
		}

		if backendRuntimeFilePath == "" {
			log.Errorf("Unsupported backend recieved for ID %d: %s", backend.ID, backend.Backend)
			continue
		}

		backendInstance := backendruntime.NewBackend(backendRuntimeFilePath)

		backendInstance.OnCrashCallback = func(conn net.Conn) {
			backendParameters, err := base64.StdEncoding.DecodeString(backend.BackendParameters)

			if err != nil {
				log.Errorf("Failed to decode backend parameters for backend #%d: %s", backend.ID, err.Error())
				return
			}

			marshalledStartCommand, err := commonbackend.Marshal(&commonbackend.Start{
				Arguments: backendParameters,
			})

			if err != nil {
				log.Errorf("Failed to marshal start command for backend #%d: %s", backend.ID, err.Error())
				return
			}

			if _, err := conn.Write(marshalledStartCommand); err != nil {
				log.Errorf("Failed to send start command for backend #%d: %s", backend.ID, err.Error())
				return
			}

			backendResponse, err := commonbackend.Unmarshal(conn)

			if err != nil {
				log.Errorf("Failed to get start command response for backend #%d: %s", backend.ID, err.Error())
				return
			}

			switch responseMessage := backendResponse.(type) {
			case *commonbackend.BackendStatusResponse:
				if !responseMessage.IsRunning {
					log.Errorf("Failed to start backend #%d: %s", backend.ID, responseMessage.Message)
					return
				}

				log.Infof("Backend #%d has been reinitialized successfully", backend.ID)
			}

			log.Warnf("Backend #%d has reinitialized! Starting up auto-starting proxies...", backend.ID)

			autoStartProxies := []db.Proxy{}

			if err := dbInstance.DB.Where("backend_id = ? AND auto_start = true", backend.ID).Find(&autoStartProxies).Error; err != nil {
				log.Errorf("Failed to query proxies to autostart: %s", err.Error())
				return
			}

			for _, proxy := range autoStartProxies {
				log.Infof("Starting up route #%d for backend #%d: %s", proxy.ID, backend.ID, proxy.Name)

				marhalledCommand, err := commonbackend.Marshal(&commonbackend.AddProxy{
					SourceIP:   proxy.SourceIP,
					SourcePort: proxy.SourcePort,
					DestPort:   proxy.DestinationPort,
					Protocol:   proxy.Protocol,
				})

				if err != nil {
					log.Errorf("Failed to marshal proxy adding request for backend #%d and route #%d: %s", proxy.BackendID, proxy.ID, err.Error())
					continue
				}

				if _, err := conn.Write(marhalledCommand); err != nil {
					log.Errorf("Failed to send proxy adding request for backend #%d and route #%d: %s", proxy.BackendID, proxy.ID, err.Error())
					continue
				}

				backendResponse, err := commonbackend.Unmarshal(conn)

				if err != nil {
					log.Errorf("Failed to get response for backend #%d and route #%d: %s", proxy.BackendID, proxy.ID, err.Error())
					continue
				}

				switch responseMessage := backendResponse.(type) {
				case *commonbackend.ProxyStatusResponse:
					if !responseMessage.IsActive {
						log.Warnf("Failed to start proxy for backend #%d and route #%d", proxy.BackendID, proxy.ID)
					}
				default:
					log.Errorf("Got illegal response type for backend #%d and proxy #%d: %T", proxy.BackendID, proxy.ID, responseMessage)
					continue
				}
			}
		}

		err = backendInstance.Start()

		if err != nil {
			log.Errorf("Failed to start backend #%d: %s", backend.ID, err.Error())
			continue
		}

		backendParameters, err := base64.StdEncoding.DecodeString(backend.BackendParameters)

		if err != nil {
			log.Errorf("Failed to decode backend parameters for backend #%d: %s", backend.ID, err.Error())
			continue
		}

		backendStartResponse, err := backendInstance.ProcessCommand(&commonbackend.Start{
			Arguments: backendParameters,
		})

		if err != nil {
			log.Warnf("Failed to get response for backend #%d: %s", backend.ID, err.Error())

			err = backendInstance.Stop()

			if err != nil {
				log.Warnf("Failed to stop backend: %s", err.Error())
			}

			continue
		}

		switch responseMessage := backendStartResponse.(type) {
		case *commonbackend.BackendStatusResponse:
			if !responseMessage.IsRunning {
				err = backendInstance.Stop()

				if err != nil {
					log.Warnf("Failed to start backend: %s", err.Error())
				}

				if responseMessage.Message == "" {
					log.Errorf("Unkown error while trying to start the backend #%d", backend.ID)
				} else {
					log.Errorf("Failed to start backend: %s", responseMessage.Message)
				}

				continue
			}
		default:
			log.Errorf("Got illegal response type for backend #%d: %T", backend.ID, responseMessage)
			continue
		}

		backendruntime.RunningBackends[backend.ID] = backendInstance

		log.Infof("Successfully initialized backend #%d", backend.ID)

		autoStartProxies := []db.Proxy{}

		if err := dbInstance.DB.Where("backend_id = ? AND auto_start = true", backend.ID).Find(&autoStartProxies).Error; err != nil {
			log.Errorf("Failed to query proxies to autostart: %s", err.Error())
			continue
		}

		for _, proxy := range autoStartProxies {
			log.Infof("Starting up route #%d for backend #%d: %s", proxy.ID, backend.ID, proxy.Name)

			backendResponse, err := backendInstance.ProcessCommand(&commonbackend.AddProxy{
				SourceIP:   proxy.SourceIP,
				SourcePort: proxy.SourcePort,
				DestPort:   proxy.DestinationPort,
				Protocol:   proxy.Protocol,
			})

			if err != nil {
				log.Errorf("Failed to get response for backend #%d and route #%d: %s", proxy.BackendID, proxy.ID, err.Error())
				continue
			}

			switch responseMessage := backendResponse.(type) {
			case *commonbackend.ProxyStatusResponse:
				if !responseMessage.IsActive {
					log.Warnf("Failed to start proxy for backend #%d and route #%d", proxy.BackendID, proxy.ID)
				}
			default:
				log.Errorf("Got illegal response type for backend #%d and proxy #%d: %T", proxy.BackendID, proxy.ID, responseMessage)
				continue
			}
		}

		log.Infof("Successfully started backend #%d", backend.ID)
	}

	log.Debug("Initializing API...")

	if !developmentMode {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	listeningAddress := os.Getenv("HERMES_LISTENING_ADDRESS")

	if listeningAddress == "" {
		if developmentMode {
			listeningAddress = "localhost:8000"
		} else {
			listeningAddress = "0.0.0.0:8000"
		}
	}

	trustedProxiesString := os.Getenv("HERMES_TRUSTED_HTTP_PROXIES")

	if trustedProxiesString != "" {
		trustedProxies := strings.Split(trustedProxiesString, ",")

		engine.ForwardedByClientIP = true
		engine.SetTrustedProxies(trustedProxies)
	} else {
		engine.ForwardedByClientIP = false
		engine.SetTrustedProxies(nil)
	}

	state := state.New(dbInstance, jwtInstance, engine)

	// Initialize routes
	users.SetupCreateUser(state)
	users.SetupLoginUser(state)
	users.SetupRefreshUserToken(state)
	users.SetupRemoveUser(state)
	users.SetupLookupUser(state)

	backends.SetupCreateBackend(state)
	backends.SetupRemoveBackend(state)
	backends.SetupLookupBackend(state)

	proxies.SetupCreateProxy(state)
	proxies.SetupRemoveProxy(state)
	proxies.SetupLookupProxy(state)
	proxies.SetupStartProxy(state)
	proxies.SetupStopProxy(state)
	proxies.SetupGetConnections(state)

	log.Infof("Listening on '%s'", listeningAddress)
	err = engine.Run(listeningAddress)

	if err != nil {
		return fmt.Errorf("Error running web server: %s", err.Error())
	}

	return nil
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

	app := &cli.App{
		Name:  "hermes",
		Usage: "port forwarding across boundaries",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "backends-path",
				Aliases:  []string{"b"},
				Usage:    "path to the backend manifest file",
				Required: true,
			},
		},
		Action: apiEntrypoint,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
